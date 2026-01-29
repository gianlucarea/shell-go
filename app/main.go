package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/chzyer/readline"
	"github.com/codecrafters-io/shell-starter-go/app/shell"
)

type bellCompleter struct {
	base interface{ Do([]rune, int) ([][]rune, int) }
	terminal *readline.Terminal
}

func (b *bellCompleter) Do(line []rune, pos int) ([][]rune, int) {
	if b == nil || b.base == nil {
		return nil, 0
	}
	res, off := b.base.Do(line, pos)
	if len(res) == 0 && b.terminal != nil {
		b.terminal.Bell()
	}
	return res, off
}

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
	// wrap the existing completer so we can call the terminal bell
	rl.Config.AutoComplete = &bellCompleter{base: completer, terminal: rl.Terminal}
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
