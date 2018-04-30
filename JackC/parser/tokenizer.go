package parser

import (
	"fmt"
	"jack"
	"strings"
	"unicode"
	"unicode/utf8"
)

var OutputC chan Token
var filename string

func isSymbol(ch rune) bool {
	symbols := []rune{'{', '}', '(', ')', '[', ']', '.', ',', ';', '+', '-', '*', '/', '&', '|', '<', '>', '=', '~'}
	for _, symbol := range symbols {
		if ch == symbol {
			return true
		}
	}
	return false
}

func isKeyword(identifier string) bool {
	keywords := []string{"class", "constructor", "function", "method", "field", "static", "var", "int", "char",
		"boolean", "void", "true", "false", "null", "this", "let", "do", "if", "else", "while", "return"}
	for _, keyword := range keywords {
		if identifier == keyword {
			return true
		}
	}
	return false
}

var inComment bool = false

func tokenizeLine(line string, lineno int, origLine string) error {
	var endIdx int

	for len(line) > 0 {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "/*") {
			inComment = true
			line = line[2:]
		}
		if inComment {
			endIdx = strings.Index(line, "*/")
			if endIdx >= 0 {
				inComment = false
				line = line[endIdx+2:]
				continue
			} else {
				// comment extends past the rest of line, so return
				return nil
			}
		}

		firstRune, width := utf8.DecodeRuneInString(line)
		var arune rune
		var token Token
		if isSymbol(firstRune) {
			token = Token{tokenType: "symbol", value: line[0:1]}
			line = line[1:]
		} else if firstRune == '"' {
			line = line[width:]
			endIdx = strings.IndexRune(line, '"')
			if endIdx < 0 {
				return fmt.Errorf("no final quote found on line %d", lineno)
			}
			token = Token{tokenType: "stringConstant", value: line[0:endIdx]}
			line = line[endIdx+1:]
		} else if unicode.IsDigit(firstRune) {
			for endIdx, arune = range line {
				if unicode.IsDigit(arune) {
					continue
				} else {
					break
				}
			}
			token = Token{tokenType: "integerConstant", value: line[0:endIdx]}
			line = line[endIdx:]
		} else if unicode.IsLetter(firstRune) || firstRune == '_' {
			for endIdx, arune = range line {
				if unicode.IsLetter(arune) || unicode.IsDigit(arune) || arune == '_' {
					continue
				} else {
					break
				}
			}
			value := line[0:endIdx]
			tokenType := "identifier"
			if isKeyword(value) {
				tokenType = "keyword"
			}
			token = Token{tokenType: tokenType, value: value}
			line = line[endIdx:]
		}
		token.file = filename
		token.lineno = lineno
		OutputC <- token
	}
	return nil
}

func TokenizeFiles(filenames []string) {
	for _, filename = range filenames {
		OutputC = make(chan Token)
		go jack.ForLinesInFile(filename, tokenizeLine)
		compileFile(OutputC, filename)
	}
}
