package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var commandList = map[string]int{"exit": 1}

func main() {
	fmt.Print("$ ")
	command, err := bufio.NewReader(os.Stdin).ReadString('\n')

	if err != nil {
		fmt.Fprintln(os.Stderr, "Error reading input:", err)
		os.Exit(1)
	}

	command = strings.TrimSpace(command)
	commandType, exists := commandList[command]
	if !exists {
		fmt.Println(command + ": command not found")
	}
	if commandType == 1 {
		os.Exit(0)
	}
	main()
}
