package main

import (
	"fmt"
	"jack/hackAssembler/parser"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: hackAssember <file>")
		os.Exit(1)
	}
	asmFile := os.Args[1]
	parser.ParseFile(asmFile)
}
