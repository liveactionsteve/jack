package main

import (
	"fmt"
	"jack"
	"jack/VMtranslator/parser"
	"os"
)

func printErrorAndExit(err interface{}) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func main() {
	if len(os.Args) != 2 {
		printErrorAndExit("Usage: VMtranslator <vm file or directory>")
	}
	basename, filenames, err := jack.Resolve(os.Args[1], ".vm")
	if err != nil {
		printErrorAndExit(err)
	}
	asmFilename := basename + ".asm"
	asmFile, err := os.Create(asmFilename)
	if err != nil {
		printErrorAndExit(err)
	}
	defer asmFile.Close()
	parser.ParseFiles(filenames, asmFile)
}
