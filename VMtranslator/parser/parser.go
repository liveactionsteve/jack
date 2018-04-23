package parser

import (
	"fmt"
	"io"
	"jack"
	"regexp"
	"strings"
)

const (
	C_UNRECOGNIZED_COMMAND = iota
	C_ARITHMETIC           = iota
	C_PUSH                 = iota
	C_POP                  = iota
	C_LABEL                = iota
	C_GOTO                 = iota
	C_IF                   = iota
	C_FUNCTION             = iota
	C_RETURN               = iota
	C_CALL                 = iota
)

type command struct {
	ctype    int
	function string
	command  string
	arg1     string
	arg2     string
}

var currentFunction string
var asmFile io.Writer

func ParseFiles(filenames []string, output io.Writer) {
	asmFile = output
	writeBoot()
	for _, filename := range filenames {
		parseFile(filename)
	}
}

func parseFile(filename string) {
	// fmt.Printf("parseFile called with filename:%s\n", filename)
	// fmt.Printf("asmfile:%v\n", asmFile)
	jack.ForLinesInFile(filename, processLine)
}

func parseCommand(line string) (cmd command, err error) {
	cmd = command{}
	cmd.function = currentFunction
	var command, rest string
	command, rest = nextWord(line)
	command = strings.ToLower(command)
	cmd.command = command
	switch command {
	case "add", "sub", "neg", "eq", "gt", "lt", "and", "or", "not":
		cmd.ctype = C_ARITHMETIC
		return
	case "push", "pop":
		if command == "pop" {
			cmd.ctype = C_POP
		} else {
			cmd.ctype = C_PUSH
		}
	case "label":
		cmd.ctype = C_LABEL
	case "goto":
		cmd.ctype = C_GOTO
	case "if-goto":
		cmd.ctype = C_IF
	case "function":
		cmd.ctype = C_FUNCTION
	case "call":
		cmd.ctype = C_CALL
	case "return":
		cmd.ctype = C_RETURN
	default:
		err = fmt.Errorf("unrecognized command:%s", command)
		return
	}
	cmd.arg1, rest = nextWord(rest)
	cmd.arg2, rest = nextWord(rest)
	// no command takes more than 2 arguments
	if len(rest) > 0 {
		err = fmt.Errorf("too many arguments")
	}
	return
}

func vetCommand(cmd command) error {
	switch cmd.ctype {
	case C_ARITHMETIC, C_RETURN:
		if len(cmd.arg1) > 0 {
			return fmt.Errorf("too many arguments: %s takes zero arguments", cmd.command)
		}
	case C_LABEL, C_GOTO, C_IF:
		if len(cmd.arg1) == 0 {
			return fmt.Errorf("no arguments: %s takes one argument", cmd.command)
		} else if len(cmd.arg2) > 0 {
			return fmt.Errorf("too many arguments: %s takes one argument", cmd.command)
		}
		if matches := labelRegexp.MatchString(cmd.arg1); !matches {
			return fmt.Errorf("invalid form of label:%s", cmd.arg1)
		}
	case C_CALL, C_FUNCTION, C_POP, C_PUSH:
		if len(cmd.arg1) == 0 || len(cmd.arg2) == 0 {
			return fmt.Errorf("too few arguments: %s takes two arguments", cmd.command)
		}
		if matches := allDigitsRegexp.MatchString(cmd.arg2); !matches {
			return fmt.Errorf("second argument of %s must be nonnegative decimal number", cmd.command)
		}
		if cmd.ctype == C_CALL || cmd.ctype == C_FUNCTION {
			if matches := labelRegexp.MatchString(cmd.arg1); !matches {
				return fmt.Errorf("invalid form of function name:%s", cmd.arg1)
			}
		} else {
			// C_POP or C_PUSH
			switch cmd.arg1 {
			case "argument", "local", "static", "constant", "this", "that", "pointer", "temp":
				// segment is good
			default:
				return fmt.Errorf("first argument of command %s must name a valid segment", cmd.command)
			}
			if cmd.ctype == C_POP && cmd.arg1 == "constant" {
				return fmt.Errorf("cannot pop into read-only 'constant' segment")
			}
		}
	}
	return nil
}

func nextWord(restOfLine string) (word, rest string) {
	restOfLine = strings.TrimSpace(restOfLine)
	if len(restOfLine) == 0 {
		return
	}
	wsIdx := strings.IndexAny(restOfLine, " \t")
	if wsIdx < 0 {
		word = restOfLine
		return
	}
	word = restOfLine[0:wsIdx]
	rest = strings.TrimSpace(restOfLine[wsIdx+1:])
	return
}

var allDigitsRegexp *regexp.Regexp
var labelRegexp *regexp.Regexp

func init() {
	var err error
	allDigitsRegexp, err = regexp.Compile("^[0-9]+$")
	if err != nil {
		panic(err)
	}
	labelRegexp, err = regexp.Compile("^[._:a-zA-Z][._:a-zA-Z0-9]*$")
	if err != nil {
		panic(err)
	}
}

func processLine(line string, lineno int, origLine string) error {
	if line == "" {
		return nil
	}
	cmd, err := parseCommand(line)
	if err != nil {
		return fmt.Errorf("ERROR:%v\n", err)
	}
	err = vetCommand(cmd)
	if err != nil {
		return err
	}
	write("// %s", origLine)
	writeCode(cmd)
	return nil
}
