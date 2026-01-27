package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type BuiltinHandler func(args []string, stdout, stderr *os.File) error

type RedirectConfig struct {
	file     string
	isError  bool
	isAppend bool
}

var (
	builtinsRegistry = map[string]BuiltinHandler{}
	operatorsSet     = map[string]bool{
		">": true, "1>": true, "2>": true,
		">>": true, "1>>": true, "2>>": true,
	}
)

func init() {
	builtinsRegistry["exit"] = exitCmd
	builtinsRegistry["echo"] = echoCmd
	builtinsRegistry["type"] = typeCmd
	builtinsRegistry["pwd"] = pwdCmd
	builtinsRegistry["cd"] = cdCmd
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("$ ")
		input, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		handleInput(strings.TrimSpace(input))
	}
}

func handleInput(input string) {
	if input == "" {
		return
	}

	parts := parseInput(input)
	args, redir := extractRedirection(parts)

	if len(args) == 0 {
		return
	}

	stdout, stderr, cleanup := setupStreams(redir)
	defer cleanup()

	cmdName := args[0]
	cmdArgs := args[1:]

	if execute, exists := builtinsRegistry[cmdName]; exists {
		if err := execute(cmdArgs, stdout, stderr); err != nil {
			fmt.Fprintln(stderr, err)
		}
		return
	}

	if path, found := findInPath(cmdName); found {
		runExternalCommand(path, cmdName, cmdArgs, stdout, stderr)
		return
	}

	fmt.Fprintf(stderr, "%s: command not found\n", cmdName)
}

func extractRedirection(parts []string) ([]string, *RedirectConfig) {
	for i, part := range parts {
		if operatorsSet[part] && i+1 < len(parts) {
			return parts[:i], &RedirectConfig{
				file:     parts[i+1],
				isError:  strings.HasPrefix(part, "2"),
				isAppend: strings.Contains(part, ">>"),
			}
		}
	}
	return parts, nil
}

func setupStreams(redir *RedirectConfig) (*os.File, *os.File, func()) {
	stdout, stderr := os.Stdout, os.Stderr
	var closer *os.File

	if redir != nil {
		flags := os.O_WRONLY | os.O_CREATE
		if redir.isAppend {
			flags |= os.O_APPEND
		} else {
			flags |= os.O_TRUNC
		}

		f, err := os.OpenFile(redir.file, flags, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "shell: %v\n", err)
		} else {
			closer = f
			if redir.isError {
				stderr = f
			} else {
				stdout = f
			}
		}
	}

	cleanup := func() {
		if closer != nil {
			closer.Close()
		}
	}
	return stdout, stderr, cleanup
}

func runExternalCommand(path, name string, args []string, stdout, stderr *os.File) {
	cmd := exec.Command(path, args...)
	cmd.Args[0] = name
	cmd.Stdout, cmd.Stderr, cmd.Stdin = stdout, stderr, os.Stdin

	if err := cmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			fmt.Fprintln(stderr, "Execution error:", err)
		}
	}
}

func exitCmd(_ []string, _, _ *os.File) error {
	os.Exit(0)
	return nil
}

func echoCmd(args []string, stdout, _ *os.File) error {
	fmt.Fprintln(stdout, strings.Join(args, " "))
	return nil
}

func pwdCmd(_ []string, stdout, _ *os.File) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	fmt.Fprintln(stdout, dir)
	return nil
}

func cdCmd(args []string, _, _ *os.File) error {
	if len(args) == 0 {
		return nil
	}
	path := args[0]
	if path == "~" {
		path = os.Getenv("HOME")
	}
	if err := os.Chdir(path); err != nil {
		return fmt.Errorf("cd: %s: No such file or directory", path)
	}
	return nil
}

func typeCmd(args []string, stdout, _ *os.File) error {
	if len(args) == 0 {
		return nil
	}
	cmd := args[0]

	if _, ok := builtinsRegistry[cmd]; ok {
		fmt.Fprintf(stdout, "%s is a shell builtin\n", cmd)
		return nil
	}
	if path, found := findInPath(cmd); found {
		fmt.Fprintf(stdout, "%s is %s\n", cmd, path)
		return nil
	}
	return fmt.Errorf("%s: not found", cmd)
}

func findInPath(command string) (string, bool) {
	paths := filepath.SplitList(os.Getenv("PATH"))
	for _, dir := range paths {
		fullPath := filepath.Join(dir, command)
		if info, err := os.Stat(fullPath); err == nil && !info.IsDir() {
			if info.Mode()&0111 != 0 {
				return fullPath, true
			}
		}
	}
	return "", false
}

func parseInput(input string) []string {
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
