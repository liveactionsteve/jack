package main

import (
	"fmt"
	"jack"
	"os"

	"jack/JackC/parser"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: <jack file or directory>\n")
		os.Exit(1)
	}
	basename, filenames, err := jack.Resolve(os.Args[1], ".jack")
	if err != nil {
		panic(err)
	}
	outputFile := basename + "Steve.xml"
	writer, err := os.Create(outputFile)
	if err != nil {
		panic(err)
	}
	parser.SetWriter(writer)

	parser.TokenizeFiles(filenames)
}
