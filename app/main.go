package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("$ ")
	command, _ := reader.ReadString('\n')
	command = strings.TrimSpace(command)
	fmt.Printf("%s: command not found \n", command)
}
