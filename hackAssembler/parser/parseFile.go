package parser

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var writer io.WriteCloser
var lineno int
var instrno int

type symbol struct {
	name    string
	address int
}

var symbols []symbol = []symbol{
	{name: "SCREEN", address: 16384},
	{name: "KBD", address: 24576},
	{name: "SP", address: 0},
	{name: "LCL", address: 1},
	{name: "ARG", address: 2},
	{name: "THIS", address: 3},
	{name: "THAT", address: 4},
	{name: "R0", address: 0},
	{name: "R1", address: 1},
	{name: "R2", address: 2},
	{name: "R3", address: 3},
	{name: "R4", address: 4},
	{name: "R5", address: 5},
	{name: "R6", address: 6},
	{name: "R7", address: 7},
	{name: "R8", address: 8},
	{name: "R9", address: 9},
	{name: "R10", address: 10},
	{name: "R11", address: 11},
	{name: "R12", address: 12},
	{name: "R13", address: 13},
	{name: "R14", address: 14},
	{name: "R15", address: 15},
}

func trimLine(line string) string {
	commentIdx := strings.Index(line, "//")
	if commentIdx >= 0 {
		line = line[0:commentIdx]
	}
	line = strings.TrimSpace(line)
	return line
}

func forLinesInFile(filename string, parseLine func(line string) error) {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lineno++
		line := scanner.Text()
		line = trimLine(line)
		err = parseLine(line)
		if err != nil {
			fmt.Printf("line %d, %v", lineno, err)
			panic(err)
		}
	}
}

func ParseFile(filename string) {
	dir := filepath.Dir(filename)
	inputBase := filepath.Base(filename)
	asmExt := filepath.Ext(filename)
	if asmExt != ".asm" {
		fmt.Println("file must have 'asm' extenstion")
		return
	}
	hackFile := filepath.Join(dir, inputBase[0:len(inputBase)-4]+"1.hack")
	var err error
	writer, err = os.Create(hackFile)
	if err != nil {
		panic(err)
	}
	defer writer.Close()
	forLinesInFile(filename, firstPass)
	resolveVariables()
	printSymbolTable(false)
	instrno = 0
	forLinesInFile(filename, secondPass)
}

func resolveVariables() {
	address := 16
	for i := range symbols {
		if symbols[i].address == -1 {
			symbols[i].address = address
			address++
		}
	}
}

func printSymbolTable(actuallyPrint bool) {
	if !actuallyPrint {
		return
	}
	fmt.Printf("symbols:\n")
	for _, symbol := range symbols {
		fmt.Printf("  %v\n", symbol)
	}
}

func instructionType(line string) string {
	firstChar := string(line[0])
	if firstChar == "(" {
		return "L"
	} else if firstChar == "@" {
		return "A"
	} else {
		return "C"
	}
}

type Cinstruction struct {
	dest  int
	value int
	jmp   int
}

func parseValue(value string) int {
	value = strings.TrimSpace(value)
	switch value {
	case "0":
		return 0x2a
	case "1":
		return 0x3f
	case "-1":
		return 0x3a
	case "D":
		return 0x0c
	case "A":
		return 0x30
	case "M":
		return 0x70
	case "!D":
		return 0x0d
	case "!A":
		return 0x31
	case "!M":
		return 0x71
	case "-D":
		return 0x0f
	case "-A":
		return 0x33
	case "-M":
		return 0x73
	case "D+1":
		return 0x1f
	case "A+1":
		return 0x37
	case "M+1":
		return 0x77
	case "D-1":
		return 0x0e
	case "A-1":
		return 0x32
	case "M-1":
		return 0x72
	case "D+A":
		return 0x02
	case "D+M":
		return 0x42
	case "D-A":
		return 0x13
	case "D-M":
		return 0x53
	case "A-D":
		return 0x07
	case "M-D":
		return 0x47
	case "D&A":
		return 0x00
	case "D&M":
		return 0x40
	case "D|A":
		return 0x15
	case "D|M":
		return 0x55
	}
	return -1
}

func parseCinstruction(a string) Cinstruction {
	// dest=value;jmp
	instr := Cinstruction{}
	equalsIdx := strings.Index(a, "=")
	if equalsIdx >= 0 {
		dest := a[0:equalsIdx]
		for _, rune := range dest {
			ch := string(rune)
			if ch == "A" {
				instr.dest |= 4
			} else if ch == "D" {
				instr.dest |= 2
			} else if ch == "M" {
				instr.dest |= 1
			} else {
				fmt.Printf("Syntax error on line %d\n", lineno)
			}
		}
	}
	value := a[equalsIdx+1:]
	jump := ""
	semiIndex := strings.Index(a, ";")
	if semiIndex >= 0 {
		jump = value[semiIndex+1:]
		value = value[0:semiIndex]
	}
	instr.value = parseValue(value)
	instr.jmp = parseJump(jump)
	return instr
}

func parseJump(jump string) int {
	switch jump {
	case "JMP":
		return 7
	case "JLT":
		return 4
	case "JLE":
		return 6
	case "JEQ":
		return 2
	case "JGE":
		return 3
	case "JGT":
		return 1
	case "JNE":
		return 5
	}
	return 0
}

func parseAinstruction(a string) int {
	address := a[1:]
	firstChar := address[0]
	if firstChar >= '0' && firstChar <= '9' {
		addr, _ := strconv.Atoi(address)
		return addr
	} else {
		return getAddress(address)
	}
}

func getAddress(name string) int {
	for _, symbol := range symbols {
		if symbol.name == name {
			return symbol.address
		}
	}
	panic(fmt.Errorf("symbol %s not in symbol table", name))
}

func setSymbol(name string, value int) {
	for i, symbol := range symbols {
		if symbol.name == name {
			if symbol.address == -1 {
				symbols[i].address = value
			}
			return
		}
	}
	newSymbol := symbol{name: name, address: value}
	symbols = append(symbols, newSymbol)
}

func firstPass(line string) error {
	if len(line) == 0 {
		return nil
	}
	iType := instructionType(line)
	if iType == "L" {
		if strings.Index(line, ")") != len(line)-1 {
			return fmt.Errorf("expected ')' at end of L instruction")
		}
		label := line[1 : len(line)-1]
		setSymbol(label, instrno)
	} else if iType == "A" {
		firstChar := line[1]
		if firstChar >= '0' && firstChar <= '9' {
		} else {
			name := line[1:]
			setSymbol(name, -1)
		}
		instrno++
	} else {
		instrno++
	}
	return nil
}

func secondPass(line string) error {
	if len(line) == 0 {
		return nil
	}
	iType := instructionType(line)
	var cinstr Cinstruction
	var ainstr int
	pc := instrno
	var outstr string
	if iType == "C" {
		cinstr = parseCinstruction(line)
		outstr = outputCinstruction(cinstr)
		instrno++
	} else if iType == "A" {
		ainstr = parseAinstruction(line)
		outstr = outputAinstruction(ainstr)
		instrno++
	}
	fmt.Printf("pc:%d:%s:%s\n", pc, line, outstr)
	if len(outstr) > 0 {
		fmt.Fprintf(writer, "%s\n", outstr)
	}
	return nil
}

func outputAinstruction(instr int) string {
	return fmt.Sprintf("%016b", instr)
}

func outputCinstruction(instr Cinstruction) string {
	// 111v vvvv vvdd djjj
	code := 0xe000 // turn on 3 hi bits
	code |= (instr.value << 6)
	code |= (instr.dest << 3)
	code |= instr.jmp
	return fmt.Sprintf("%016b", code)
}
