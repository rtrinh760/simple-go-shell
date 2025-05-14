package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	executor "github.com/rtrinh760/simple-go-shell/executor"
)

func main() {
    fmt.Println("Welcome to Simple Go Shell. Type 'exit' to quit.")

    scanner := bufio.NewScanner(os.Stdin)

    for {
        fmt.Print("> ")
        // if there is no scanner
        if !scanner.Scan() {
            break
        }
        line := scanner.Text()
        if strings.TrimSpace(line) == "exit" {
            break
        }
        err := executor.RunCommand(line)
        if err != nil {
            fmt.Println("Error: ", err)
        }
    }

    fmt.Println("Exiting shell...")
}