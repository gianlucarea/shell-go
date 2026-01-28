package shell

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func HandleInput(input string) {
	if input == "" {
		return
	}

	parts := ParseInput(input)
	args, redir := ExtractRedirection(parts)
	if len(args) == 0 {
		return
	}

	stdout, stderr, cleanup := SetupStreams(redir)
	defer cleanup()

	cmdName := args[0]
	cmdArgs := args[1:]

	if execute, exists := BuiltinsRegistry[cmdName]; exists {
		if err := execute(cmdArgs, stdout, stderr); err != nil {
			fmt.Fprintln(stderr, err)
		}
		return
	}

	if path, found := FindInPath(cmdName); found {
		runExternalCommand(path, cmdName, cmdArgs, stdout, stderr)
		return
	}

	fmt.Fprintf(stderr, "%s: command not found\n", cmdName)
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

func FindInPath(command string) (string, bool) {
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
