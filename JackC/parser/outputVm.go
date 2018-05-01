package parser

import (
	"fmt"
	"io"
	"jack/JackC/symbols"
	"strconv"
)

type VmWriter struct {
	writer             io.WriteCloser
	classTree          ClassTree
	currentSymbolTable *symbols.SymbolTable
	currentSubroutine  *SubroutineDec
	labelGenerator     LabelGenerator
}

func (vw VmWriter) outputStaticVariables() {
	prt("// static variables")
	for _, dec := range vw.classTree.classVarDecs {
		if dec.staticOrField.value != "static" {
			continue
		}
		for _, varName := range dec.varNames {
			name := varName.value
			symbol := vw.Lookup(name)
			prt("//  name:%s, access:%s", name, symbol.Access())
		}
	}
}

func (vw VmWriter) outputFieldVariables() {
	prt("// fields")
	for _, dec := range vw.classTree.classVarDecs {
		if dec.staticOrField.value != "field" {
			continue
		}
		for _, varName := range dec.varNames {
			name := varName.value
			symbol := vw.Lookup(name)
			prt("//  name:%s, access:%s", name, symbol.Access())
		}
	}
}

func (vw VmWriter) printSymbolTable(symbolTable symbols.SymbolTable) {
	if len(symbolTable.Map) == 0 {
		return
	}
	prt("//symbols")
	for name, symbol := range symbolTable.Map {
		prt("//  name:%s, type:%v, kind:%v, index:%d", name, symbol.TypeOf(), symbol.KindOf(), symbol.IndexOf())
	}
}

func (vw VmWriter) evaluateTerm(t Term) {
	switch term := t.term.(type) {
	case UnaryOpTerm:
		vw.evaluateUnaryOpTerm(term)
	case SingleTokenTerm:
		vw.evaluateSingleTokenTerm(term)
	case Expression:
		vw.evaluateExpression(term)
	case ArrayAccessTerm:
		vw.evaluateArrayAccessTerm(term)
	case FunctionCall:
		vw.outputFunctionCall(term)
	case MethodCall:
		vw.outputMethodCall(term)
	default:
		prt("panic: unknown type of Term: %v", t)
	}
}

func (vw VmWriter) evaluateUnaryOpTerm(term UnaryOpTerm) {
	vw.evaluateTerm(term.term)
	vw.evaluateUnaryOp(term.unaryOp)
}

func (vw VmWriter) evaluateUnaryOp(unaryOp Token) {
	switch unaryOp.value {
	case "+":
		// do nothing
	case "-":
		prt("neg")
	case "~":
		prt("not")
	}
}

func (vw VmWriter) evaluateSingleTokenTerm(term SingleTokenTerm) {
	token := term.value
	if token.tokenType == "keyword" {
		switch token.value {
		case "true":
			prt("push constant 1")
			prt("neg")
		case "false", "null":
			prt("push constant 0")
		case "this":
			prt("push pointer 0")
		}
		return
	}
	if token.tokenType == "stringConstant" {
		bytes := []byte(token.value)
		prt("push constant %d", len(bytes))
		prt("call String.new 1")
		for _, b := range bytes {
			prt("push constant %d", b)
			prt("call String.appendChar 2")
		}
		return
	}
	if token.tokenType == "integerConstant" {
		num, _ := strconv.Atoi(token.value)
		prt("push constant %d", num)
		return
	}
	if token.tokenType == "identifier" {
		symbol := vw.Lookup(token.value)
		prt("push %s", symbol.Access())
		return
	}
	prt("panic: evaluateSingleTokenTerm unhandled token type:%s token value:%s", token.tokenType, token.value)
}

func (vw VmWriter) evaluateArrayAccessTerm(term ArrayAccessTerm) {
	symbol := vw.Lookup(term.varName.value)
	prt("push %s", symbol.Access())
	vw.evaluateExpression(term.index)
	prt("add")
	prt("pop pointer 1")
	prt("push that 0")
}

func (vw VmWriter) evaluateExpression(expr Expression) {
	vw.evaluateTerm(expr.term)
	for _, opterm := range expr.opTerms {
		vw.evaluateTerm(opterm.rhs)
		vw.evaluateOp(opterm.op.value)
	}
}

func (vw VmWriter) evaluateOp(op string) {
	switch op {
	case "+":
		prt("add")
	case "-":
		prt("sub")
	case "*":
		prt("call Math.multiply 2")
	case "/":
		prt("call Math.divide 2")
	case "&":
		prt("and")
	case "|":
		prt("or")
	case "<":
		prt("lt")
	case ">":
		prt("gt")
	case "=":
		prt("eq")
	}
}

func (vw VmWriter) outputLetStatement(statement LetStatement) {
	name := statement.varName.value
	var symbol symbols.Symbol
	symbol = vw.Lookup(name)
	if statement.isArray {
		vw.evaluateExpression(statement.rhs)
		prt("push %s", symbol.Access()) // push address of array
		vw.evaluateExpression(statement.indexExpression)
		prt("add")
		prt("pop pointer 1")
		prt("pop that 0")
	} else {
		vw.evaluateExpression(statement.rhs)
		prt("pop %s", symbol.Access())
	}

}

func (vw VmWriter) outputIfStatement(statement IfStatement) {
	l1 := vw.labelGenerator.generateLabel("IF")
	vw.evaluateExpression(statement.condition)
	prt("not")
	prt("if-goto %s", l1) // skip to 'else' clause, if exists, otherwise skip to end of statement
	vw.outputStatements(statement.thenClause)
	if statement.isElse {
		l2 := vw.labelGenerator.generateLabel("IF")
		prt("goto %s", l2) // at end of 'then' clause, skip over the 'else' clause
		prt("label %s", l1)
		vw.outputStatements(statement.elseClause)
		prt("label %s", l2)
	} else {
		prt("label %s", l1)
	}
}

func (vw VmWriter) outputWhileStatement(statement WhileStatement) {
	l1 := vw.labelGenerator.generateLabel("WHILE")
	l2 := vw.labelGenerator.generateLabel("WHILE")
	prt("label %s", l1)
	vw.evaluateExpression(statement.condition)
	prt("not")
	prt("if-goto %s", l2)
	vw.outputStatements(statement.stmts)
	prt("goto %s", l1)
	prt("label %s", l2)
}

func (vw VmWriter) outputDoStatement(statement DoStatement) {
	switch call := statement.subroutineCall.(type) {
	case FunctionCall:
		vw.outputFunctionCall(call)
	case MethodCall:
		vw.outputMethodCall(call)
	}
	prt("pop temp 0")
}

func (vw VmWriter) outputReturnStatement(statement ReturnStatement) {
	if vw.currentSubroutine.returnType.value == "void" {
		prt("push constant 0")
	} else if vw.currentSubroutine.ctrOrFuncOrMethod.value == "constructor" {
		prt("push pointer 0")
	} else if statement.isEmpty {
		prt("push constant 0")
	} else {
		vw.evaluateExpression(statement.returnExpression)
	}
	prt("return")
}

func (vw VmWriter) outputFunctionCall(call FunctionCall) {
	prt("push pointer 0")
	for i, argExp := range call.arguments {
		prt("// push value of arg %d", i)
		vw.evaluateExpression(argExp)
	}
	prt("call %s.%s %d", vw.classTree.className.value, call.functionName.value, len(call.arguments)+1)
}

func (vw VmWriter) outputMethodCall(call MethodCall) {
	classOrVarName := call.classOrVarName.value
	symbol := vw.Lookup(classOrVarName)
	if symbol.Exists() {
		// this is a genuine method call
		prt("push %s", symbol.Access())
	}
	for i, argExp := range call.arguments {
		prt("// push value of arg %d", i)
		vw.evaluateExpression(argExp)
	}
	if symbol.Exists() {
		prt("call %v.%s %d", symbol.TypeOf(), call.methodName.value, len(call.arguments)+1)
	} else {
		// calling function in some other class
		prt("call %s.%s %d", call.classOrVarName.value, call.methodName.value, len(call.arguments))
	}
}

func (vw VmWriter) outputStatements(statements []Statement) {
	for _, statement := range statements {
		vw.outputStatement(statement)
	}
}

func (vw VmWriter) outputStatement(stmt Statement) {
	switch statement := stmt.stmt.(type) {
	case LetStatement:
		vw.outputLetStatement(statement)
	case IfStatement:
		vw.outputIfStatement(statement)
	case WhileStatement:
		vw.outputWhileStatement(statement)
	case DoStatement:
		vw.outputDoStatement(statement)
	case ReturnStatement:
		vw.outputReturnStatement(statement)
	}
}

func (vw VmWriter) outputMethod(dec SubroutineDec) {
	prt("// set 'this' pointer")
	prt("push argument 0")
	prt("pop pointer 0")
	// adjust visible argument indexes, since argument 0 is 'this'
	for _, parm := range dec.parameters {
		symbol := vw.Lookup(parm.name.value)
		symbol.IncIndex()
	}
	vw.outputStatements(dec.body.statements)
}

func (vw VmWriter) outputFunction(dec SubroutineDec) {
	vw.outputStatements(dec.body.statements)
}

func (vw VmWriter) outputConstructor(dec SubroutineDec) {
	numFields := vw.classTree.symbolTable.VarCount(symbols.FIELD)
	prt("push constant %d", numFields)
	prt("call Memory.alloc 1")
	prt("pop pointer 0")
	vw.outputStatements(dec.body.statements)
}

func (vw VmWriter) outputSubroutine(dec SubroutineDec) {
	vw.currentSymbolTable = &dec.symbolTable
	vw.currentSubroutine = &dec
	name := fmt.Sprintf("%s.%s", vw.classTree.className.value, dec.name.value)
	numLocalVariables := dec.symbolTable.VarCount(symbols.VAR)
	prt("function %s %d", name, numLocalVariables)
	prt("//type of subroutine: %s", dec.ctrOrFuncOrMethod.value)
	prt("//returns %s", dec.returnType.value)
	vw.printSymbolTable(dec.symbolTable)
	switch dec.ctrOrFuncOrMethod.value {
	case "method":
		vw.outputMethod(dec)
	case "function":
		vw.outputFunction(dec)
	case "constructor":
		vw.outputConstructor(dec)
	}

}

func (vw VmWriter) outputSubroutines() {
	for _, dec := range vw.classTree.subroutineDecs {
		vw.outputSubroutine(dec)
	}
}

func outputJackVM(tree ClassTree, wrt io.WriteCloser) {
	vw := VmWriter{writer: wrt, classTree: tree}
	vw.labelGenerator = newLabelGenerator()
	setPrtOutput(wrt)
	vw.outputStaticVariables()
	vw.outputFieldVariables()
	vw.outputSubroutines()
}

func (vw VmWriter) Lookup(name string) symbols.Symbol {
	result := symbols.NoSymbol()
	if vw.currentSymbolTable != nil {
		result = vw.currentSymbolTable.Lookup(name)
	}
	if !result.Exists() {
		result = vw.classTree.symbolTable.Lookup(name)
	}
	return result
}
