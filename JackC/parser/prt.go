package parser

import (
	"fmt"
	"io"
)

var prtWriter io.Writer

func setPrtOutput(wrt io.Writer) {
	prtWriter = wrt
}

func prt(format string, a ...interface{}) {
	outString := fmt.Sprintf(format, a...) + "\r\n"
	io.WriteString(prtWriter, outString)
}
