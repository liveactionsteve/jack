package main

import (
	"flag"
	"fmt"
	"jack"
	"jack/VMtranslator/parser"
	"os"
)

func printErrorAndExit(err interface{}) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

//func BoolVar(p *bool, name string, value bool, usage string)

func main() {
	var noBoot bool
	flag.BoolVar(&noBoot, "noboot", false, "do not add boot code")
	flag.Parse()
	arg := flag.Arg(0)
	if len(flag.Args()) != 1 {
		printErrorAndExit("Usage: VMtranslator <vm file or directory>")
	}
	basename, filenames, err := jack.Resolve(arg, ".vm")
	if err != nil {
		printErrorAndExit(err)
	}
	asmFilename := basename + ".asm"
	asmFile, err := os.Create(asmFilename)
	if err != nil {
		printErrorAndExit(err)
	}
	defer asmFile.Close()
	parser.SetWriter(asmFile)
	parser.ParseFiles(filenames, !noBoot)
}
