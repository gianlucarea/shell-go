package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Command struct {
	name    string
	execute func() error
}

var commandList = map[string]Command{
	"exit": {name: "exit", execute: func() error { return nil }},
	"echo": {name: "echo", execute: func() error { return nil }},
	"type": {name: "type", execute: func() error { return nil }},
}

func printPrompt() {
	fmt.Print("$ ")
}

func main() {
	printPrompt()
	handleInput()
	main()
}

func handleInput() {
	input, err := bufio.NewReader(os.Stdin).ReadString('\n')
	path := os.Getenv("PATH")

	if err != nil {
		fmt.Fprintln(os.Stderr, "Error reading input:", err)
		os.Exit(1)
	}

	input = strings.TrimSpace(input)
	parts := strings.Fields(input)

	if len(parts) <= 0 {
		return
	}

	commandAsString := parts[0]
	args := parts[1:]

	if command, exists := commandList[commandAsString]; exists {
		switch command.name {
		case "exit":
			{
				os.Exit(0)
			}
		case "echo":
			{
				fmt.Fprintln(os.Stdout, strings.Join(args, " "))
			}
		case "type":
			{
				typeCommand(args, path)
			}
		}

		if err := command.execute(); err != nil {
			fmt.Fprintln(os.Stderr, "Error executing command:", err)
		}
	} else {
		executeExternalProgram(parts, path)
	}
}

func typeCommand(args []string, path string) {
	if _, isCommand := commandList[args[0]]; isCommand {
		fmt.Fprintf(os.Stdout, "%s is a shell builtin\n", args[0])
	} else if path != "" {
		_ = searchExternalCommand(path, args)
	} else {
		fmt.Fprintf(os.Stdout, "%s: not found\n", args[0])
	}
}

func searchExternalCommand(path string, args []string) string {
	paths := strings.Split(path, ":")
	found := false
	for _, p := range paths {
		fullPath := fmt.Sprintf("%s/%s", p, args[0])
		if _, err := os.Stat(fullPath); err == nil {
			if isExecutable(fullPath) {
				fmt.Fprintf(os.Stdout, "%s is %s\n", args[0], fullPath)
				found = true
				return fullPath
			}
		}
	}
	if !found {
		fmt.Fprintf(os.Stdout, "%s: not found\n", args[0])
	}
	return ""
}

func isExecutable(filePath string) bool {
	info, err := os.Stat(filePath)
	if err != nil {
		return false
	}
	return info.Mode().IsRegular() && info.Mode()&0111 != 0
}

func findExecutablePath(path string, command string) string {
	paths := strings.Split(path, ":")
	for _, p := range paths {
		fullPath := fmt.Sprintf("%s/%s", p, command)
		if _, err := os.Stat(fullPath); err == nil {
			if isExecutable(fullPath) {
				return fullPath
			}
		}
	}
	return ""
}

func executeExternalProgram(args []string, path string) {
	programPath := findExecutablePath(path, args[0])
	if programPath == "" {
		fmt.Fprintf(os.Stdout, "%s: not found\n", args[0])
		return
	}
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Path = programPath
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error:", err)
	}
	fmt.Print(string(out))
}
