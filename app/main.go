package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/chzyer/readline"
	"github.com/codecrafters-io/shell-starter-go/app/shell"
)

func main() {
	completer := readline.NewPrefixCompleter(
		readline.PcItem("echo"),
		readline.PcItem("exit"),
	)

	rl, err := readline.NewEx(&readline.Config{
		Prompt:       "$ ",
		AutoComplete: completer,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "readline setup:", err)
		os.Exit(1)
	}
	defer rl.Close()

	for {
		line, err := rl.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			}
			continue
		} else if err == io.EOF {
			break
		} else if err != nil {
			fmt.Fprintln(os.Stderr, "read error:", err)
			break
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		shell.HandleInput(line)
	}
}
