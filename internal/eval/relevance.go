package eval

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"agenttest/internal/tasks"
)

type RelevanceResult struct {
	RelevantFilesFound     []string `json:"relevant_files_found"`
	RelevantFilesMissing   []string `json:"relevant_files_missing"`
	AlsoOKFilesFound       []string `json:"also_ok_files_found"`
	RelevantSymbolsFound   []string `json:"relevant_symbols_found"`
	RelevantSymbolsMissing []string `json:"relevant_symbols_missing"`
}

func (r RelevanceResult) RelevantFilesFoundCount() int {
	return len(r.RelevantFilesFound)
}

func (r RelevanceResult) RelevantFilesTotal() int {
	return len(r.RelevantFilesFound) + len(r.RelevantFilesMissing)
}

func (r RelevanceResult) RelevantSymbolsFoundCount() int {
	return len(r.RelevantSymbolsFound)
}

func (r RelevanceResult) RelevantSymbolsTotal() int {
	return len(r.RelevantSymbolsFound) + len(r.RelevantSymbolsMissing)
}

func EvaluateFile(
	expected tasks.Expected,
	eventsFilename string,
	answer string,
) (RelevanceResult, error) {
	if eventsFilename == "" {
		return RelevanceResult{}, fmt.Errorf(
			"events filename must not be empty",
		)
	}

	data, err := os.ReadFile(eventsFilename)
	if err != nil {
		return RelevanceResult{}, fmt.Errorf(
			"read events file %s: %w",
			eventsFilename,
			err,
		)
	}

	return Evaluate(
		expected,
		data,
		answer,
	), nil
}

func Evaluate(
	expected tasks.Expected,
	events []byte,
	answer string,
) RelevanceResult {
	var result RelevanceResult

	content := bytes.ToLower(events)
	answerLower := strings.ToLower(answer)

	for _, expectedFile := range expected.Files {
		if containsFile(content, expectedFile) ||
			containsFile([]byte(answerLower), expectedFile) {
			result.RelevantFilesFound = append(
				result.RelevantFilesFound,
				expectedFile,
			)
		} else {
			result.RelevantFilesMissing = append(
				result.RelevantFilesMissing,
				expectedFile,
			)
		}
	}

	for _, alsoOKFile := range expected.AlsoOK {
		if containsFile(content, alsoOKFile) ||
			containsFile([]byte(answerLower), alsoOKFile) {
			result.AlsoOKFilesFound = append(
				result.AlsoOKFilesFound,
				alsoOKFile,
			)
		}
	}

	for _, expectedSymbol := range expected.Symbols {
		if containsSymbol(content, expectedSymbol) ||
			containsSymbol([]byte(answerLower), expectedSymbol) {
			result.RelevantSymbolsFound = append(
				result.RelevantSymbolsFound,
				expectedSymbol,
			)
		} else {
			result.RelevantSymbolsMissing = append(
				result.RelevantSymbolsMissing,
				expectedSymbol,
			)
		}
	}

	return result
}

func containsFile(
	content []byte,
	expected string,
) bool {
	normalizedExpected := normalizePath(
		expected,
	)

	if normalizedExpected == "" {
		return false
	}

	text := string(content)

	paths := []string{
		normalizedExpected,
		"./" + normalizedExpected,
		"/workspace/" + normalizedExpected,
	}

	for _, path := range paths {
		if strings.Contains(
			text,
			strings.ToLower(path),
		) {
			return true
		}
	}

	return false
}

func containsSymbol(
	content []byte,
	expected string,
) bool {
	normalized := strings.TrimSpace(
		expected,
	)

	if normalized == "" {
		return false
	}

	text := string(content)
	symbol := strings.ToLower(
		normalized,
	)

	if strings.Contains(text, symbol) {
		return true
	}

	parts := strings.Split(
		normalized,
		".",
	)

	if len(parts) < 2 {
		return false
	}

	method := strings.ToLower(
		parts[len(parts)-1],
	)

	receiver := strings.ToLower(
		parts[len(parts)-2],
	)

	patterns := []string{
		receiver + "." + method,
		"func (" + receiver,
		") " + method + "(",
		"." + method + "(",
	}

	for _, pattern := range patterns {
		if strings.Contains(
			text,
			pattern,
		) {
			return true
		}
	}

	return false
}

func normalizePath(
	value string,
) string {
	value = strings.TrimSpace(value)

	value = strings.Trim(
		value,
		"\"'`",
	)

	value = filepath.ToSlash(
		value,
	)

	value = strings.TrimPrefix(
		value,
		"/workspace/",
	)

	value = strings.TrimPrefix(
		value,
		"./",
	)

	return strings.ToLower(value)
}
