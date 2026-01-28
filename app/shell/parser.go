package shell

import (
	"strings"
)

var operatorsSet = map[string]bool{
	">": true, "1>": true, "2>": true,
	">>": true, "1>>": true, "2>>": true,
}

func ParseInput(input string) []string {
	var args []string
	var current strings.Builder
	isSingleQuote, isDoubleQuote := false, false

	for i := 0; i < len(input); i++ {
		char := input[i]
		if isSingleQuote {
			if char == '\'' {
				isSingleQuote = false
			} else {
				current.WriteByte(char)
			}
		} else if isDoubleQuote {
			if char == '"' {
				isDoubleQuote = false
			} else if char == '\\' && i+1 < len(input) {
				if strings.ContainsRune("$`\"\\\n", rune(input[i+1])) {
					current.WriteByte(input[i+1])
					i++
				} else {
					current.WriteByte(char)
				}
			} else {
				current.WriteByte(char)
			}
		} else {
			if char == '\\' && i+1 < len(input) {
				current.WriteByte(input[i+1])
				i++
			} else if char == '\'' {
				isSingleQuote = true
			} else if char == '"' {
				isDoubleQuote = true
			} else if char == ' ' || char == '\t' {
				if current.Len() > 0 {
					args = append(args, current.String())
					current.Reset()
				}
			} else {
				current.WriteByte(char)
			}
		}
	}
	if current.Len() > 0 {
		args = append(args, current.String())
	}
	return args
}

func ExtractRedirection(parts []string) ([]string, *RedirectConfig) {
	for i, part := range parts {
		if operatorsSet[part] && i+1 < len(parts) {
			return parts[:i], &RedirectConfig{
				File:     parts[i+1],
				IsError:  strings.HasPrefix(part, "2"),
				IsAppend: strings.Contains(part, ">>"),
			}
		}
	}
	return parts, nil
}
