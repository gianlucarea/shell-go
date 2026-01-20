package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	for true {
		fmt.Print("$ ")
		command, err := bufio.NewReader(os.Stdin).ReadString('\n')

		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			os.Exit(1)
		}

		command = strings.TrimSpace(command)
		fmt.Println(command + ": command not found")
	}

}
