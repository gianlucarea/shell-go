package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type BuiltinHandler func(args []string) error

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

	parts := strings.Fields(input)
	cmdName := parts[0]
	args := parts[1:]

	if execute, exists := builtinsRegistry[cmdName]; exists {
		if err := execute(args); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		return
	}

	if path, found := findInPath(cmdName); found {
		runExtenalCommand(path, cmdName, args)
		return
	}

	fmt.Fprintf(os.Stderr, "%s: command not found\n", cmdName)
}

func exitCmd(args []string) error {
	os.Exit(0)
	return nil
}

func echoCmd(args []string) error {
	fmt.Println(strings.Join(args, " "))
	return nil
}

func pwdCmd(args []string) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	fmt.Println(dir)
	return nil
}

func cdCmd(args []string) error {
	var newDirectory string
	if args[0] == "~" {
		newDirectory = os.Getenv("HOME")
	} else {
		newDirectory = args[0]
	}
	err := os.Chdir(newDirectory)
	if err != nil {
		return errors.New("cd: " + args[0] + ": No such file or directory")
	}
	return nil

}

func typeCmd(args []string) error {
	if len(args) == 0 {
		return nil
	}
	command := args[0]

	if _, isBuiltins := builtinsRegistry[command]; isBuiltins {
		fmt.Printf("%s is a shell builtin\n", command)
		return nil
	}

	if path, found := findInPath(command); found {
		fmt.Printf("%s is %s\n", command, path)
		return nil
	}

	fmt.Printf("%s: not found\n", command)
	return nil
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

func runExtenalCommand(path, cmdName string, args []string) {
	cmd := exec.Command(path, args...)
	cmd.Args[0] = cmdName
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			fmt.Fprintln(os.Stderr, "Execution error:", err)
		}
	}
}
