package parser

import (
	"bytes"
	"fmt"
	"html"
	"io"
	"jack/JackC/symbols"
)

var writer io.WriteCloser
var spacesToIndent int = 0

func indent() {
	var buf bytes.Buffer
	for i := 0; i < spacesToIndent; i++ {
		buf.WriteString(" ")
	}
	writer.Write(buf.Bytes())
}

func write(format string, a ...interface{}) {
	outString := fmt.Sprintf(format, a...) + "\r\n"
	io.WriteString(writer, outString)
}

func writeOpenTag(tag string) {
	indent()
	write("<%s>", tag)
	spacesToIndent += 2
}

func writeCloseTag(tag string) {
	spacesToIndent -= 2
	indent()
	write("</%s>", tag)
}

func writeToken(token Token) {
	indent()
	val := html.EscapeString(token.value)
	write("<%s> %s </%s>", token.tokenType, val, token.tokenType)
}

func writeTagValue(tag, value string) {
	indent()
	write("<%s> %s </%s>", tag, value, tag)
}

func writeKeyword(value string) {
	indent()
	write("<keyword> %s </keyword>", value)
}

func writeSymbol(value string) {
	indent()
	write("<symbol> %s </symbol>", html.EscapeString(value))
}

func outputClass(tree ClassTree, wr io.WriteCloser) {
	writer = wr
	writeOpenTag("class")
	writeKeyword("class")
	writeClassName(tree.className.value)
	writeSymbol("{")
	writeClassVarDecs(tree.classVarDecs)
	writeSubroutineDecs(tree.subroutineDecs)
	writeSymbol("}")
	writeCloseTag("class")
}

func writeSubroutineDecs(decs []SubroutineDec) {
	for _, dec := range decs {
		writeOpenTag("subroutineDec")
		writeToken(dec.ctrOrFuncOrMethod)
		writeType(dec.returnType)
		writeElement("subroutine-name", dec.name.value)
		writeSymbol("(")
		writeOpenTag("parameterList")
		writeParameters(dec)
		writeCloseTag("parameterList")
		writeSymbol(")")
		writeSubroutineBody(dec.body)
		writeCloseTag("subroutineDec")
	}
}

func writeSubroutineBody(body SubroutineBody) {
	writeOpenTag("subroutineBody")
	writeSymbol("{")
	for _, varDec := range body.varDecs {
		writeOpenTag("varDec")
		writeKeyword("var")
		writeToken(varDec.varType)
		for i, name := range varDec.names {
			writeToken(name)
			if i < len(varDec.names)-1 {
				writeSymbol(",")
			}
		}
		writeSymbol(";")
		writeCloseTag("varDec")
	}
	writeStatements(body.statements)
	writeSymbol("}")
	writeCloseTag("subroutineBody")
}

func writeStatements(stmts []Statement) {
	writeOpenTag("statements")
	for _, stmt := range stmts {
		switch statement := stmt.stmt.(type) {
		case LetStatement:
			writeLetStatement(statement)
		case IfStatement:
			writeIfStatement(statement)
		case WhileStatement:
			writeWhileStatement(statement)
		case DoStatement:
			writeDoStatement(statement)
		case ReturnStatement:
			writeReturnStatement(statement)
		}
	}
	writeCloseTag("statements")
}

func writeReturnStatement(statement ReturnStatement) {
	writeOpenTag("returnStatement")
	writeKeyword("return")
	if statement.returnExpression.term.term != nil {
		writeExpression(statement.returnExpression)
	}
	writeSymbol(";")
	writeCloseTag("returnStatement")
}

func writeDoStatement(statement DoStatement) {
	writeOpenTag("doStatement")
	writeKeyword("do")
	writeSubroutineCall(statement.subroutineCall)
	writeCloseTag("doStatement")
}

func writeSubroutineCall(subroutineCall interface{}) {
	switch call := subroutineCall.(type) {
	case FunctionCall:
		writeFunctionCall(call)
	case MethodCall:
		writeMethodCall(call)
	}
}

func writeExpressionList(expressions []Expression) {
	writeOpenTag("expressionList")
	for i, expr := range expressions {
		writeExpression(expr)
		if i < len(expressions)-1 {
			writeSymbol(",")
		}
	}
	writeCloseTag("expressionList")
}

func writeMethodCall(call MethodCall) {
	writeToken(call.classOrVarName)
	writeSymbol(".")
	writeToken(call.methodName)
	writeSymbol("(")
	writeExpressionList(call.arguments)
	writeSymbol(")")
	writeSymbol(";")
}

func writeFunctionCall(call FunctionCall) {
	writeToken(call.functionName)
	writeSymbol("(")
	writeExpressionList(call.arguments)
	writeSymbol(")")
	writeSymbol(";")
}

func writeWhileStatement(statement WhileStatement) {
	writeOpenTag("whileStatement")
	writeKeyword("while")
	writeSymbol("(")
	writeExpression(statement.condition)
	writeSymbol(")")
	writeSymbol("{")
	writeStatements(statement.stmts)
	writeSymbol("}")
	writeCloseTag("whileStatement")
}

func writeIfStatement(stmt IfStatement) {
	writeOpenTag("ifStatement")
	writeKeyword("if")
	writeSymbol("(")
	writeExpression(stmt.condition)
	writeSymbol(")")
	writeSymbol("{")
	writeStatements(stmt.thenClause)
	writeSymbol("}")
	if stmt.isElse {
		writeKeyword("else")
		writeSymbol("{")
		writeStatements(stmt.elseClause)
		writeSymbol("}")
	}
	writeCloseTag("ifStatement")
}

func writeLetStatement(stmt LetStatement) {
	writeOpenTag("letStatement")
	writeKeyword("let")
	writeToken(stmt.varName)
	if stmt.isArray {
		writeSymbol("[")
		writeExpression(stmt.indexExpression)
		writeSymbol("]")
	}
	writeSymbol("=")
	writeExpression(stmt.rhs)
	writeSymbol(";")
	writeCloseTag("letStatement")
}

func writeExpression(expr Expression) {
	writeOpenTag("expression")
	//io.WriteString(writer, "  some expression  \r\n")
	writeTerm(expr.term)
	for _, opterm := range expr.opTerms {
		writeToken(opterm.op)
		writeTerm(opterm.rhs)
	}
	writeCloseTag("expression")
}

func writeTerm(term Term) {
	writeOpenTag("term")
	switch actualTerm := term.term.(type) {
	case SingleTokenTerm:
		writeToken(actualTerm.value)
	case UnaryOpTerm:
		writeToken(actualTerm.unaryOp)
		writeTerm(actualTerm.term)
	case Expression:
		writeSymbol("(")
		writeExpression(actualTerm)
		writeSymbol(")")
	case ArrayAccessTerm:
		writeToken(actualTerm.varName)
		writeSymbol("[")
		writeExpression(actualTerm.index)
		writeSymbol("]")
	case FunctionCall:
		writeToken(actualTerm.functionName)
		writeSymbol("(")
		writeArguments(actualTerm.arguments)
		writeSymbol(")")
	case MethodCall:
		writeToken(actualTerm.classOrVarName)
		writeSymbol(".")
		writeToken(actualTerm.methodName)
		writeSymbol("(")
		writeArguments(actualTerm.arguments)
		writeSymbol(")")
	}
	writeCloseTag("term")
}

func writeArguments(args []Expression) {
	writeOpenTag("expressionList")
	for i, arg := range args {
		writeExpression(arg)
		if i < len(args)-1 {
			writeSymbol(",")
		}
	}
	writeCloseTag("expressionList")
}

func writeClassVarDecs(decs []ClassVarDec) {
	for _, dec := range decs {
		writeOpenTag("classVarDec")
		writeToken(dec.staticOrField)
		writeToken(dec.varType)
		writeClassVarDec(dec)
		writeSymbol(";")
		writeCloseTag("classVarDec")
	}
}

func writeElement(tag, contents string) {
	indent()
	val := html.EscapeString(contents)
	write("<%s> %s </%s>", tag, val, tag)
}

func writeClassVarDec(dec ClassVarDec) {
	writeOpenTag(dec.staticOrField.value)
	for _, name := range dec.varNames {
		writeDefinition(name.value, classSymbolTable)
	}
	writeCloseTag(dec.staticOrField.value)
}

func writeDefinition(name string, symbolTable symbols.SymbolTable) {
	symbol := symbolTable.Lookup(name)
	writeElement("name", name)
	writeElement("reference-type", "definition")
	writeElement("access", fmt.Sprintf("%v %d", symbol.KindOf(), symbol.IndexOf()))
}

func writeClassName(className string) {
	writeElement("class-name", className)
}

func writeType(token Token) {
	writeOpenTag("type")
	if token.tokenType == "identifier" {
		writeElement("class", token.value)
	} else {
		writeToken(token)
	}
	writeCloseTag("type")
}

func writeParameters(dec SubroutineDec) {
	for i, parm := range dec.parameters {
		writeType(parm.parmType)
		writeOpenTag("argument")
		writeDefinition(parm.name.value, dec.symbolTable)
		writeCloseTag("argument")
		if i < len(dec.parameters)-1 {
			writeSymbol(",")
		}
	}
}
