package shell

import (
	"bufio"
	"fmt"
	"strings"
)

func completeCommand(partial string) string {
	matches := []string{}

	allowedCompletions := []string{"echo", "exit"}

	for _, cmd := range allowedCompletions {
		if strings.HasPrefix(cmd, partial) {
			matches = append(matches, cmd)
		}
	}

	if len(matches) == 1 {
		return matches[0] + " "
	}

	return partial
}

func ReadLineWithCompletion(reader *bufio.Reader) string {
	var line strings.Builder
	for {
		b, err := reader.ReadByte()
		if err != nil {
			return line.String()
		}

		switch b {
		case '\t':
			partial := line.String()
			completed := completeCommand(partial)
			if completed != partial {
				for i := 0; i < len(partial); i++ {
					fmt.Print("\b \b")
				}
				fmt.Print(completed)
				line.Reset()
				line.WriteString(completed)
			}
		case '\n':
			fmt.Print("\n")
			return line.String()
		case '\b', 127:
			if line.Len() > 0 {
				s := line.String()
				line.Reset()
				line.WriteString(s[:len(s)-1])
				fmt.Print("\b \b")
			}
		default:
			line.WriteByte(b)
			fmt.Print(string(b))
		}
	}
}
