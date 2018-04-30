package parser

type Token struct {
	tokenType string
	value     string
	file      string
	lineno    int
}
