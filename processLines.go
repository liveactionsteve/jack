package jack

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func trimLine(line string) string {
	commentIdx := strings.Index(line, "//")
	if commentIdx >= 0 {
		line = line[0:commentIdx]
	}
	line = strings.TrimSpace(line)
	return line
}

func ForLinesInFile(filename string, processLine func(line string, lineno int, origLine string) error) {
	lineno := 0
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lineno++
		origLine := scanner.Text()
		line := trimLine(origLine)
		err = processLine(line, lineno, origLine)
		if err != nil {
			fmt.Printf("line %d, %v\n", lineno, err)
			os.Exit(1)
		}
	}
}
