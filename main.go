package main

import (
	"fmt"
	"lamina/cmd"
	"os"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Println("Lamina Error:", err)
		os.Exit(1)
	}
}
