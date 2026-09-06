FROM node:22-bookworm-slim

RUN npm install -g @earendil-works/pi-coding-agent@latest

WORKDIR /workspace

ENTRYPOINT ["pi", "--provider", "openrouter", "--model", "deepseek/deepseek-v4-flash-0731", "--thinking", "off"]