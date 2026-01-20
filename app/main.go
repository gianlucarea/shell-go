package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Command struct {
	name    string
	execute func() error
}

var commandList = map[string]Command{
	"exit": {name: "exit", execute: func() error { return nil }},
	"echo": {name: "echo", execute: func() error { return nil }},
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
		if command.name == "exit" {
			os.Exit(0)
		}

		if command.name == "echo" {
			fmt.Fprintln(os.Stdout, strings.Join(args, " "))
		}

		if err := command.execute(); err != nil {
			fmt.Fprintln(os.Stderr, "Error executing command:", err)
		}
	} else {
		fmt.Fprintf(os.Stdout, "%s: command not found\n", commandAsString)
	}
}
