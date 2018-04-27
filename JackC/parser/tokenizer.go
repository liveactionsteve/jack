package parser

import (
	"fmt"
	"io"
	"jack"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"
)

var writer io.Writer

type Token struct {
	tokenType string
	value     string
}

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
	keywords := []string{"class", "constructor", "function", "method", "field", "static", "var", "int", "char", "boolean", "void", "true", "false", "this", "let", "do", "if", "else", "while", "return"}
	for _, keyword := range keywords {
		if identifier == keyword {
			return true
		}
	}
	return false
}

var inComment bool = false

func tokenizeLine(line string, lineno int, origLine string) error {
	var tokens []Token
	var token Token
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
		if isSymbol(firstRune) {
			token = Token{tokenType: "symbol", value: line[0:1]}
			line = line[1:]
		} else if firstRune == '"' {
			line = line[width:]
			endIdx = strings.IndexRune(line, '"')
			if endIdx < 0 {
				return fmt.Errorf("no final quote found on line %d", lineno)
			}
			token = Token{tokenType: "stringConstant", value: line[0 : endIdx-1]}
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
		tokens = append(tokens, token)
	}
	for _, token = range tokens {
		write("<%s> %s </%s>", token.tokenType, token.value, token.tokenType)
	}
	fmt.Printf("line %d, :%s:, tokens: %v\n", lineno, origLine, tokens)
	return nil
}

func write(format string, a ...interface{}) {
	outString := fmt.Sprintf(format, a...) + "\n"
	io.WriteString(writer, outString)
}

func tokenizeFile(filename string) {
	outputFile := filepath.Base(filename)
	outputFile = outputFile[0:len(outputFile)-5] + "Steve.xml"

}

func TokenizeFiles(filenames []string) {
	for _, filename := range filenames {
		jack.ForLinesInFile(filename, tokenizeLine)
	}
}
