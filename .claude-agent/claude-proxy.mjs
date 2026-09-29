import http from "node:http";
import { Readable } from "node:stream";

const PORT = 8080;
const TARGET = "https://openrouter.ai";

const reasoningMode =
    process.env.CLAUDE_PROXY_REASONING || "off";

if (!["off", "on"].includes(reasoningMode)) {
    console.error(
        `invalid CLAUDE_PROXY_REASONING: ${reasoningMode}`,
    );
    process.exit(1);
}

function buildRequestBody(rawBody) {
    if (!rawBody || rawBody.length === 0) {
        return rawBody;
    }

    let payload;

    try {
        payload = JSON.parse(rawBody.toString("utf8"));
    } catch {
        return rawBody;
    }

    if (
        !payload ||
        typeof payload !== "object" ||
        Array.isArray(payload)
    ) {
        return rawBody;
    }

    if (reasoningMode === "off") {
        // Полностью убираем параметры, которыми Claude Code
        // может включить reasoning.
        delete payload.reasoning;
        delete payload.reasoning_effort;
        delete payload.thinking;

        // Claude Code добавляет:
        // "output_config": {"effort": "high"}
        //
        // Для benchmark режима off это надо убрать.
        if (
            payload.output_config &&
            typeof payload.output_config === "object"
        ) {
            delete payload.output_config.effort;

            if (
                Object.keys(payload.output_config).length === 0
            ) {
                delete payload.output_config;
            }
        }
    } else {
        // Для режима on явно разрешаем reasoning.
        payload.reasoning = {
            enabled: true,
        };
    }

    return Buffer.from(
        JSON.stringify(payload),
        "utf8",
    );
}

function copyRequestHeaders(request) {
    const headers = {};

    for (const [name, value] of Object.entries(
        request.headers,
    )) {
        const lower = name.toLowerCase();

        if (
            lower === "host" ||
            lower === "content-length" ||
            lower === "connection" ||
            lower === "transfer-encoding"
        ) {
            continue;
        }

        if (value !== undefined) {
            headers[name] = value;
        }
    }

    return headers;
}

async function handleRequest(request, response) {
    if (
        request.method === "GET" &&
        request.url === "/healthz"
    ) {
        response.writeHead(200, {
            "content-type": "text/plain; charset=utf-8",
        });

        response.end("ok\n");
        return;
    }

    const chunks = [];

    for await (const chunk of request) {
        chunks.push(chunk);
    }

    const originalBody = Buffer.concat(chunks);
    const body = buildRequestBody(originalBody);

    // ВАЖНО:
    // Claude Code ходит на:
    //   http://127.0.0.1:8080/api/v1/messages
    //
    // Поэтому request.url уже содержит /api/...
    // и его можно напрямую приклеивать к openrouter.ai.
    const targetURL = `${TARGET}${request.url}`;

    const headers = copyRequestHeaders(request);

    if (body.length > 0) {
        headers["content-length"] = String(body.length);
    }

    try {
        const upstream = await fetch(
            targetURL,
            {
                method: request.method,
                headers,
                body:
                    request.method === "GET" ||
                    request.method === "HEAD"
                        ? undefined
                        : body,
                redirect: "manual",
            },
        );

        const responseHeaders = {};

        for (const [name, value] of upstream.headers) {
            responseHeaders[name] = value;
        }

        response.writeHead(
            upstream.status,
            responseHeaders,
        );

        if (!upstream.body) {
            response.end();
            return;
        }

        Readable.fromWeb(
            upstream.body,
        ).pipe(response);
    } catch (error) {
        console.error(
            "proxy request failed:",
            error instanceof Error
                ? error.message
                : String(error),
        );

        if (!response.headersSent) {
            response.writeHead(502, {
                "content-type": "text/plain; charset=utf-8",
            });
        }

        response.end("Bad Gateway\n");
    }
}

const server = http.createServer(
    (request, response) => {
        handleRequest(
            request,
            response,
        ).catch((error) => {
            console.error(
                "unhandled proxy error:",
                error,
            );

            if (!response.headersSent) {
                response.writeHead(500);
            }

            response.end();
        });
    },
);

server.listen(
    PORT,
    "127.0.0.1",
    () => {
        console.log(
            `claude proxy listening on 127.0.0.1:${PORT}`,
        );

        console.log(
            `reasoning mode: ${reasoningMode}`,
        );
    },
);