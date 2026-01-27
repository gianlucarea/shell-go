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

var builtinsRegistry = map[string]BuiltinHandler{}

func initMap() {
	builtinsRegistry["exit"] = exitCmd
	builtinsRegistry["echo"] = echoCmd
	builtinsRegistry["type"] = typeCmd
	builtinsRegistry["pwd"] = pwdCmd
	builtinsRegistry["cd"] = cdCmd
}

func main() {
	initMap()
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("$ ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			continue
		}
		handleInput(strings.TrimSpace(input))
	}
}

func handleInput(input string) {
	if input == "" {
		return
	}

	parts := parseInput(input)

	var args []string
	var outputFile string
	var redirectError bool = false
	redirectIndex := -1

	for i, part := range parts {
		if (part == ">" || part == "1>" || part == "2>") && i+1 < len(parts) {
			redirectIndex = i
			outputFile = parts[i+1]
			if part == "2>" {
				redirectError = true
			}
			break
		}
	}

	if redirectIndex != -1 {
		args = parts[:redirectIndex]
	} else {
		args = parts
	}

	cmdName := args[0]
	cmdArgs := args[1:]

	stdout := os.Stdout
	stderr := os.Stderr

	if outputFile != "" {
		f, err := os.OpenFile(outputFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			fmt.Fprintln(stderr, err)
		}
		defer f.Close()
		if redirectError {
			stderr = f
		} else {
			stdout = f
		}
	}

	if execute, exists := builtinsRegistry[cmdName]; exists {
		if err := execute(cmdArgs, stdout, stderr); err != nil {
			fmt.Fprintln(stderr, err)
		}
		return
	}

	if path, found := findInPath(cmdName); found {
		runExtenalCommand(path, cmdName, cmdArgs, stdout, stderr)
		return
	}

	fmt.Fprintf(stderr, "%s: command not found\n", cmdName)
}

func exitCmd(args []string, stdout, stderr *os.File) error {
	os.Exit(0)
	return nil
}

func echoCmd(args []string, stdout, stderr *os.File) error {
	fmt.Fprintln(stdout, strings.Join(args, " "))
	return nil
}

func pwdCmd(args []string, stdout, stderr *os.File) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	fmt.Fprintln(stdout, dir)
	return nil
}

func cdCmd(args []string, stdout, stderr *os.File) error {
	var newDirectory string
	if args[0] == "~" {
		newDirectory = os.Getenv("HOME")
	} else {
		newDirectory = args[0]
	}
	err := os.Chdir(newDirectory)
	if err != nil {
		return fmt.Errorf("cd: %s: No such file or directory", newDirectory)
	}
	return nil

}

func typeCmd(args []string, stdout, stderr *os.File) error {
	if len(args) == 0 {
		return nil
	}
	command := args[0]

	if _, isBuiltins := builtinsRegistry[command]; isBuiltins {
		fmt.Fprintf(stdout, "%s is a shell builtin\n", command)
		return nil
	}

	if path, found := findInPath(command); found {
		fmt.Fprintf(stdout, "%s is %s\n", command, path)
		return nil
	}

	return fmt.Errorf("%s: not found", command)
}

func findInPath(command string) (string, bool) {
	pathEnv := os.Getenv("PATH")
	paths := filepath.SplitList(pathEnv)

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

func runExtenalCommand(path, cmdName string, args []string, stdout, stderr *os.File) {
	cmd := exec.Command(path, args...)
	cmd.Args[0] = cmdName
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			fmt.Fprintln(os.Stderr, "Execution error:", err)
		}
	}
}

func parseInput(input string) []string {
	var args []string
	var current strings.Builder
	isSingleQuotes := false
	isDoubleQuotes := false

	for i := 0; i < len(input); i++ {
		char := input[i]

		if isSingleQuotes {
			if char == '\'' {
				isSingleQuotes = false
			} else {
				current.WriteByte(char)
			}
		} else if isDoubleQuotes {
			if char == '"' {
				isDoubleQuotes = false
			} else if char == '\\' && i+1 < len(input) {
				nextChar := input[i+1]
				if nextChar == '$' || nextChar == '`' || nextChar == '"' || nextChar == '\\' || nextChar == '\n' {
					current.WriteByte(nextChar)
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
				isSingleQuotes = true
			} else if char == '"' {
				isDoubleQuotes = true
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
		current.Reset()
	}
	return args
}
