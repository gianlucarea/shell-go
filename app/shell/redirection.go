package shell

import (
	"fmt"
	"os"
)

type RedirectConfig struct {
	File     string
	IsError  bool
	IsAppend bool
}

func SetupStreams(redir *RedirectConfig) (*os.File, *os.File, func()) {
	stdout, stderr := os.Stdout, os.Stderr
	var closer *os.File

	if redir != nil {
		flags := os.O_WRONLY | os.O_CREATE
		if redir.IsAppend {
			flags |= os.O_APPEND
		} else {
			flags |= os.O_TRUNC
		}

		f, err := os.OpenFile(redir.File, flags, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "shell: %v\n", err)
		} else {
			closer = f
			if redir.IsError {
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
