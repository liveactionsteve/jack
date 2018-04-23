package main

import (
	"fmt"
	"jack/VMtranslator/parser"
	"os"
	"path/filepath"
	"strings"
)

func printErrorAndExit(err interface{}) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func main() {
	if len(os.Args) != 2 {
		printErrorAndExit("Usage: VMtranslator <vm file or directory>")
	}
	path := os.Args[1]
	path, err := filepath.Abs(path)
	if err != nil {
		printErrorAndExit(err)
	}

	// check file or directory exists
	fileinfo, err := os.Stat(path)
	if err != nil {
		printErrorAndExit(err)
	}

	var dirpath string
	var filenames []string
	var asmFilename string

	if fileinfo.IsDir() {
		// parse all .vm files in directory
		dirpath = path
		asmFilename = filepath.Base(path) + ".asm"
		dir, err := os.Open(path)
		if err != nil {
			printErrorAndExit(err)
		}
		allnames, err := dir.Readdirnames(0)
		if err != nil {
			printErrorAndExit(err)
		}
		for _, fname := range allnames {
			if strings.HasSuffix(fname, ".vm") {
				filenames = append(filenames, filepath.Join(dirpath, fname))
			}
		}
	} else {
		if !strings.HasSuffix(path, ".vm") {
			printErrorAndExit("File type must be .vm ")
		}
		dirpath = filepath.Dir(path)
		base := filepath.Base(path)
		filenames = []string{filepath.Join(dirpath, base)}
		asmFilename = base[0:len(base)-3] + ".asm"
	}
	asmFilename = filepath.Join(dirpath, asmFilename)
	asmFile, err := os.Create(asmFilename)
	if err != nil {
		printErrorAndExit(err)
	}
	defer asmFile.Close()
	parser.ParseFiles(filenames, asmFile)
}
