package shell

import (
	"fmt"
	"os"
	"strings"
)

type BuiltinHandler func(args []string, stdout, stderr *os.File) error

var BuiltinsRegistry = map[string]BuiltinHandler{}

func init() {
	BuiltinsRegistry["exit"] = exitCmd
	BuiltinsRegistry["echo"] = echoCmd
	BuiltinsRegistry["type"] = typeCmd
	BuiltinsRegistry["pwd"] = pwdCmd
	BuiltinsRegistry["cd"] = cdCmd
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

	if _, ok := BuiltinsRegistry[cmd]; ok {
		fmt.Fprintf(stdout, "%s is a shell builtin\n", cmd)
		return nil
	}
	if path, found := FindInPath(cmd); found {
		fmt.Fprintf(stdout, "%s is %s\n", cmd, path)
		return nil
	}
	return fmt.Errorf("%s: not found", cmd)
}
