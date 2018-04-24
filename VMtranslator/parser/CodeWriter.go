package parser

import (
	"fmt"
	"io"
	"strconv"
)

var labelsn = 1000

func writeCode(cmd command) {
	switch cmd.ctype {
	case C_ARITHMETIC:
		writeArithmetic(cmd)
	case C_PUSH, C_POP:
		writePushOrPop(cmd)
	case C_LABEL:
		writeLabel(cmd)
	case C_GOTO:
		writeGoto(cmd)
	case C_IF:
		writeIf(cmd)
	case C_CALL:
		writeCall(cmd)
	case C_FUNCTION:
		writeFunction(cmd)
	case C_RETURN:
		writeReturn(cmd)
	}
}

func writeBoot() {
	write("@256")
	write("D=A")
	write("@SP")
	write("M=D")
	callSysInit := command{
		function: "Boot",
		arg1:     "Sys.init",
		arg2:     "0",
	}
	writeCall(callSysInit)
}

func pushAddressAt(address string) {
	write("@%s", address)
	write("D=M")
	pushDRegister()
}

func writeCall(cmd command) {
	returnAddress := fmt.Sprintf("Ret%s%s", cmd.function, newLabel())
	write("@%s", returnAddress)
	write("D=A")
	pushDRegister()
	pushAddressAt("LCL")
	pushAddressAt("ARG")
	pushAddressAt("THIS")
	pushAddressAt("THAT")
	// SP = SP - n - 5
	n, _ := strconv.Atoi(cmd.arg2)
	offset := n + 5
	write("@SP")
	write("D=M")
	write("@%d", offset)
	write("D=D-A")
	write("@ARG")
	write("M=D")
	write("@SP")
	write("D=M")
	write("@LCL")
	write("M=D")
	write("@%s", cmd.arg1)
	write("D;JMP")
	write("(%s)", returnAddress)
}

func writeFunction(cmd command) {
	write("(%s)", cmd.arg1)
	write("D=0")
	n, _ := strconv.Atoi(cmd.arg2)
	for i := 0; i < n; i++ {
		pushDRegister()
	}
}

func writeReturn(cmd command) {
	write("//FRAME = LCL")
	write("@LCL")
	write("D=M")
	write("@R13")
	write("M=D")
	write("// RET = *(FRAME - 5)")
	write("@5")
	write("A=D-A")
	write("D=M") // D holds MEM[FRAME-5]; return address
	write("@R14")
	write("M=D") // R14 holds return address
	write("// *ARG = pop()")
	popIntoDRegister() // returned value into D
	write("@ARG")
	write("A=M")
	write("M=D") // store returned value at *ARG
	write("// SP = ARG+1")
	write("@ARG")
	write("D=M+1")
	write("@SP")
	write("M=D")
	decrementR13AndRestoreAddressIn("THAT")
	decrementR13AndRestoreAddressIn("THIS")
	decrementR13AndRestoreAddressIn("ARG")
	decrementR13AndRestoreAddressIn("LCL")
	write("// goto RET")
	write("@R14")
	write("A=M")
	write("D;JMP")
}

func decrementR13AndRestoreAddressIn(regName string) {
	write("// decrement R13 and store %s at that location", regName)
	// decrement R13
	write("@R13")
	write("AM=M-1")
	// THAT = *R13
	write("D=M")
	write("@%s", regName)
	write("M=D")
}

func writeGoto(cmd command) {
	write("@%s", cmd.arg1)
	write("D;JMP")
}

func writeIf(cmd command) {
	popIntoDRegister()
	write("@%s", cmd.arg1)
	write("D;JNE")
}

func writeLabel(cmd command) {
	write("(%s)", cmd.arg1)
}

func writePushOrPop(cmd command) {
	switch cmd.arg1 {
	case "constant":
		write("@%s", cmd.arg2)
		write("D=A")
		pushDRegister()
		return
	case "local", "argument", "this", "that":
		write("@%s", cmd.arg2)
		write("D=A") // store offset in D
		if cmd.arg1 == "local" {
			write("@LCL")
		} else if cmd.arg1 == "argument" {
			write("@ARG")
		} else if cmd.arg1 == "this" {
			write("@THIS")
		} else if cmd.arg1 == "that" {
			write("@THAT")
		}
		if cmd.ctype == C_POP {
			write("D=D+M") // D holds address of cell we want to pop into
			write("@R13")
			write("M=D") // R13 holds address we want to pop into
			popIntoDRegister()
			write("@R13")
			write("A=M")
			write("M=D")
		} else {
			write("A=D+M") // A holds address of cell we want to push from
			write("D=M")   // D holds value we want to push
			pushDRegister()
		}
	case "temp", "pointer":
		offset, err := strconv.Atoi(cmd.arg2)
		if err != nil {
			panic(err)
		}
		if cmd.arg1 == "temp" {
			offset += 5
		} else if cmd.arg1 == "pointer" {
			offset += 3
		}
		// offset is the address of the relevant cell
		if cmd.ctype == C_POP {
			popIntoDRegister()
			write("@%d", offset)
			write("M=D")
		} else {
			write("@%d", offset)
			write("D=M")
			pushDRegister()
		}
	case "static":
		write("// steve was here")
		label := cmd.module + "." + cmd.arg2
		if cmd.ctype == C_POP {
			popIntoDRegister()
			write("@%s", label)
			write("M=D")
		} else {
			write("@%s", label)
			write("D=M")
			pushDRegister()
		}
	}
}

func writeArithmetic(cmd command) {
	// unary functions leave SP unchanged
	if cmd.command == "not" || cmd.command == "neg" {
		write("@SP")
		write("A=M-1") // A is address of argument
		write("D=M")   // D is argument
		if cmd.command == "not" {
			write("M=!M")
		} else {
			write("M=-M")
		}
		return
	}
	popIntoDRegister()
	// store in R13
	write("@R13")
	write("M=D")
	popIntoDRegister()

	write("@R13") // M is second argument
	switch cmd.command {
	case "add":
		write("D=D+M")
	case "sub":
		write("D=D-M")
	case "and":
		write("D=D&M")
	case "or":
		write("D=D|M")
	case "eq", "gt", "lt":
		falseLabel := newLabel()
		trueLabel := newLabel()
		write("D=D-M")
		write("@%s", trueLabel)
		switch cmd.command {
		case "eq":
			write("D;JEQ")
		case "gt":
			write("D;JGT")
		case "lt":
			write("D;JLT")
		}
		write("D=0")
		write("@%s", falseLabel)
		write("0;JMP")
		write("(%s)", trueLabel)
		write("D=-1")
		write("(%s)", falseLabel)
	}
	pushDRegister()
}

func newLabel() string {
	labelsn++
	return fmt.Sprintf("L%d", labelsn)
}

func write(format string, a ...interface{}) {
	outString := fmt.Sprintf(format, a...) + "\n"
	io.WriteString(asmFile, outString)
}

func popIntoDRegister() {
	write("@SP")    // A=0, so M will be RAM[0]
	write("AM=M-1") // decrement value in RAM[2] and store it in RAM[2] and A register
	// now A register is address of value to be popped
	write("D=M") // so now the value at that address is pulled into D
}

func pushDRegister() {
	write("@SP")   // A=2, so M will be RAM[2]
	write("A=M")   // A is RAM[2], which is address where value will be pushed
	write("M=D")   // put value of D into RAM[A]
	write("@SP")   // A=2 again, so M is RAM[2]
	write("M=M+1") // increment value in RAM[2] and overwrite it
}
