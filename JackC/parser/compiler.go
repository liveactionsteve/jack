package parser

import (
	"fmt"
	"jack/JackC/symbols"
	"os"
	"path/filepath"
	"runtime"
)

var inputC chan Token
var token Token
var classSymbolTable symbols.SymbolTable
var subroutineSymbolTable symbols.SymbolTable

type ClassVarDec struct {
	staticOrField Token
	varType       Token
	varNames      []Token
}

type Parameter struct {
	parmType Token
	name     Token
}

type VarDec struct {
	varType Token
	names   []Token
}

type SingleTokenTerm struct {
	value Token
}

type ArrayAccessTerm struct {
	varName Token
	index   Expression
}

type UnaryOpTerm struct {
	unaryOp Token
	term    Term
}

type OpTerm struct {
	op  Token
	rhs Term
}

type Term struct {
	term interface{}
}

type Expression struct {
	term    Term
	opTerms []OpTerm
}

type LetStatement struct {
	varName         Token
	isArray         bool
	indexExpression Expression
	rhs             Expression
}

type IfStatement struct {
	condition  Expression
	thenClause []Statement
	isElse     bool
	elseClause []Statement
}

type WhileStatement struct {
	condition Expression
	stmts     []Statement
}

type MethodCall struct {
	classOrVarName Token
	methodName     Token
	arguments      []Expression
}

type FunctionCall struct {
	functionName Token
	arguments    []Expression
}

type DoStatement struct {
	subroutineCall interface{} // either MethodCall or FunctionCall
}

type ReturnStatement struct {
	isEmpty          bool
	returnExpression Expression
}

type Statement struct {
	stmt interface{}
}

type SubroutineBody struct {
	varDecs    []VarDec
	statements []Statement
}

type SubroutineDec struct {
	ctrOrFuncOrMethod Token
	returnType        Token
	name              Token
	parameters        []Parameter
	body              SubroutineBody
	symbolTable       symbols.SymbolTable
}

type ClassTree struct {
	className      Token
	classVarDecs   []ClassVarDec
	subroutineDecs []SubroutineDec
	symbolTable    symbols.SymbolTable
}

func compileFile(ch chan Token, filename string) {
	inputC = ch
	outDir := filepath.Dir(filename)
	outputFile := filepath.Base(filename)
	outputFile = outputFile[0:len(outputFile)-5] + "Steve.xml"
	outputFile = filepath.Join(outDir, outputFile)
	writer, err := os.Create(outputFile)
	if err != nil {
		panic(err)
	}
	classTree := compileClass()
	outputClass(classTree, writer)
}

func errorOut(msg string) {
	fmt.Printf("Error line %d file %s: %s\n", token.lineno, token.file, msg)
	os.Exit(1)
}

func getToken() {
	token = <-inputC
	// pc, fn, line, _ := runtime.Caller(1)
	// fmt.Printf("  in %s[%s:%d]\n", runtime.FuncForPC(pc).Name(), fn, line)
	// fmt.Printf("file: %s(%d): token type %s, value %s\n", token.file, token.lineno, token.tokenType, token.value)
}

func unless(condition bool, msg string) bool {
	pc, fn, line, _ := runtime.Caller(1)
	errmsg := fmt.Sprintf("[error] in %s[%s:%d] %s", runtime.FuncForPC(pc).Name(), fn, line, msg)
	if !condition {
		errorOut(errmsg)
	}
	return true
}

func unlessIdentifier(msg string) bool {
	unless(token.tokenType == "identifier", msg)
	return true
}

func compileClass() ClassTree {
	result := ClassTree{}
	getToken()
	unless(token.value == "class", "File must declare a class")
	getToken()
	unlessIdentifier("Expected classname")
	result.className = token
	result.symbolTable = symbols.NewSymbolTable(token.value)
	classSymbolTable = result.symbolTable
	getToken()
	unless(token.value == "{", "expected {")
	getToken()
	result.classVarDecs = compileClassVarDecs()
	result.subroutineDecs = compileSubroutineDecs()
	unless(token.value == "}", "expected }")
	return result
}

func compileSubroutineDecs() []SubroutineDec {
	var result []SubroutineDec
	for token.value == "constructor" || token.value == "function" || token.value == "method" {
		subroutineDec := SubroutineDec{ctrOrFuncOrMethod: token}
		getToken()
		unless(token.value == "void" || token.tokenType == "identifier", "Expected 'void' or identifier")
		subroutineDec.returnType = token
		getToken()
		unlessIdentifier("Expected subroutine name")
		subroutineDec.name = token
		symbolTable := symbols.NewSymbolTable(token.value)
		subroutineDec.symbolTable = symbolTable
		subroutineSymbolTable = symbolTable
		getToken()
		unless(token.value == "(", "Expected (")
		getToken()
		subroutineDec.parameters = compileParmList()
		unless(token.value == ")", "Expected )")
		getToken()
		subroutineDec.body = compileSubroutineBody()
		result = append(result, subroutineDec)
	}
	return result
}

func compileSubroutineBody() SubroutineBody {
	var result SubroutineBody
	unless(token.value == "{", "Expected beginning of subroutine body '{'")
	getToken()
	result = SubroutineBody{}
	result.varDecs = compileVarDecs()
	result.statements = compileStatements()
	unless(token.value == "}", "Expected { at end of subroutine body")
	getToken()
	return result
}

func isStatement() bool {
	return token.value == "let" || token.value == "if" || token.value == "while" || token.value == "do" || token.value == "return"
}

func compileStatements() []Statement {
	var result []Statement
	for isStatement() {
		statement := compileStatement()
		result = append(result, statement)
	}
	return result
}

func compileStatement() Statement {
	switch token.value {
	case "let":
		return compileLetStatement()
	case "if":
		return compileIfStatement()
	case "while":
		return compileWhileStatement()
	case "do":
		return compileDoStatement()
	case "return":
		return compileReturnStatement()
	}
	return Statement{}
}

func compileReturnStatement() Statement {
	result := ReturnStatement{}
	getToken()
	if token.value == ";" {
		getToken()
		result.isEmpty = true
		return Statement{stmt: result}
	}
	result.returnExpression = compileExpression()
	result.isEmpty = false
	unless(token.value == ";", "Expecting ;")
	getToken()
	return Statement{stmt: result}
}

func compileDoStatement() Statement {
	result := DoStatement{}
	getToken()
	unlessIdentifier("expecting name of subroutine or variable or class")
	initialToken := token
	getToken()
	result.subroutineCall = compileSubroutineCall(initialToken)
	unless(token.value == ";", "Expecting ;")
	getToken()
	return Statement{stmt: result}
}

func compileWhileStatement() Statement {
	result := WhileStatement{}
	getToken()
	result.condition = compileParenthesizedExpression()
	result.stmts = compileBlockOfStatements()
	return Statement{stmt: result}
}

func compileParenthesizedExpression() Expression {
	unless(token.value == "(", "Expecting (")
	getToken()
	result := compileExpression()
	unless(token.value == ")", "Expecting )")
	getToken()
	return result
}

func compileBlockOfStatements() []Statement {
	unless(token.value == "{", "Expecting {")
	getToken()
	result := compileStatements()
	unless(token.value == "}", "Expecting }")
	getToken()
	return result
}

func compileIfStatement() Statement {
	result := IfStatement{isElse: false}
	getToken()
	result.condition = compileParenthesizedExpression()
	result.thenClause = compileBlockOfStatements()
	if token.value == "else" {
		result.isElse = true
		getToken()
		result.elseClause = compileBlockOfStatements()
	}
	return Statement{stmt: result}
}

func compileLetStatement() Statement {
	result := LetStatement{isArray: false}
	getToken()
	unlessIdentifier("Expecting variable name")
	result.varName = token
	getToken()
	if token.value == "[" {
		result.isArray = true
		getToken()
		result.indexExpression = compileExpression()
		unless(token.value == "]", "Expected ]")
		getToken()
	}
	unless(token.value == "=", "Expected =")
	getToken()
	result.rhs = compileExpression()
	unless(token.value == ";", "Expecting ;")
	getToken()
	return Statement{stmt: result}
}

func compileVarDecs() []VarDec {
	var result []VarDec
	for token.value == "var" {
		varDec := compileVarDec()
		result = append(result, varDec)
	}
	return result
}

func compileVarDec() VarDec {
	result := VarDec{}
	getToken()
	unless(IsType(), "Expected type")
	result.varType = token
	getToken()
	for unlessIdentifier("Expected variable name") {
		subroutineSymbolTable.Define(token.value, result.varType.value, symbols.VAR)
		result.names = append(result.names, token)
		getToken()
		if token.value == "," {
			getToken()
			continue
		} else {
			break
		}
	}
	unless(token.value == ";", "Expected ;")
	getToken()
	return result
}

func IsType() bool {
	return token.value == "int" || token.value == "char" || token.value == "boolean" || token.tokenType == "identifier"
}

func compileParmList() []Parameter {
	var result []Parameter
	for IsType() {
		parm := Parameter{parmType: token}
		getToken()
		unlessIdentifier("expected parameter name")
		parm.name = token
		getToken()
		subroutineSymbolTable.Define(parm.name.value, parm.parmType.value, symbols.ARGUMENT)
		result = append(result, parm)
		if token.value == "," {
			getToken()
			continue
		} else {
			break
		}
	}
	return result
}

func compileClassVarDec() ClassVarDec {
	result := ClassVarDec{staticOrField: token}
	var symbolKind symbols.SymbolKind
	if token.value == "static" {
		symbolKind = symbols.STATIC
	} else {
		symbolKind = symbols.FIELD
	}
	getToken()
	unless(token.value == "int" || token.value == "char" || token.value == "boolean" || token.tokenType == "identifier",
		"expected type or identifier")
	result.varType = token
	getToken()
	for unlessIdentifier("expected variable name") {
		classSymbolTable.Define(token.value, result.varType.value, symbolKind)
		result.varNames = append(result.varNames, token)
		getToken()
		if token.value == "," {
			getToken()
		} else {
			break
		}
	}
	unless(token.value == ";", "expecting ;")
	getToken()
	return result
}

func compileClassVarDecs() []ClassVarDec {
	var result []ClassVarDec
	for token.value == "static" || token.value == "field" {
		classVar := compileClassVarDec()
		result = append(result, classVar)
	}
	return result
}

func IsTerm() bool {
	returnValue := token.tokenType == "integerConstant" || token.tokenType == "stringConstant" ||
		token.value == "true" || token.value == "false" || token.value == "null" || token.value == "this" ||
		token.value == "(" || token.value == "-" || token.value == "~" || token.tokenType == "identifier"
	return returnValue
}

func compileTerm() Term {
	var term Term
	if token.value == "-" || token.value == "~" {
		unaryOpTerm := UnaryOpTerm{unaryOp: token}
		getToken()
		unaryOpTerm.term = compileTerm()
		term.term = unaryOpTerm
		return term
	}
	if token.tokenType == "integerConstant" || token.tokenType == "stringConstant" ||
		token.value == "true" || token.value == "false" || token.value == "null" || token.value == "this" {
		singleTokenTerm := SingleTokenTerm{value: token}
		term.term = singleTokenTerm
		getToken()
		return term
	}
	if token.value == "(" {
		getToken()
		innerExpression := compileExpression()
		term = Term{term: innerExpression}
		unless(token.value == ")", "Expecting )")
		getToken()
		return term
	}
	if unlessIdentifier("expecting term") {
		term = Term{term: compileTermWithIdentifier()}
		return term
	}
	return term
}

func compileTermWithIdentifier() interface{} {
	initialToken := token
	getToken()
	if token.value == "[" {
		return compileArrayAccessTerm(initialToken)
	} else if token.value == "(" || token.value == "." {
		return compileSubroutineCall(initialToken)
	} else {
		return SingleTokenTerm{value: initialToken}
	}
}

func compileArrayAccessTerm(initialToken Token) ArrayAccessTerm {
	result := ArrayAccessTerm{varName: initialToken}
	getToken()
	result.index = compileExpression()
	unless(token.value == "]", "Expecting ]")
	getToken()
	return result
}

func compileSubroutineCall(initialToken Token) interface{} {
	if token.value == "(" {
		functionCall := FunctionCall{functionName: initialToken}
		getToken()
		functionCall.arguments = compileArgumentList()
		unless(token.value == ")", "Expecting )")
		getToken()
		return functionCall
	}
	if token.value == "." {
		methodCall := MethodCall{classOrVarName: initialToken}
		getToken()
		unlessIdentifier("expecting identifier after .")
		methodCall.methodName = token
		getToken()
		unless(token.value == "(", "expecting (")
		getToken()
		methodCall.arguments = compileArgumentList()
		unless(token.value == ")", "Expecting )")
		getToken()
		return methodCall
	}
	return nil
}

func IsOp() bool {
	returnValue := token.value == "+" || token.value == "-" || token.value == "*" || token.value == "/" ||
		token.value == "&" || token.value == "|" || token.value == "<" || token.value == ">" || token.value == "="
	return returnValue
}

func compileExpression() Expression {
	result := Expression{term: compileTerm()}
	for IsOp() {
		opTerm := OpTerm{op: token}
		getToken()
		opTerm.rhs = compileTerm()
		result.opTerms = append(result.opTerms, opTerm)
	}
	return result
}

func compileArgumentList() []Expression {
	var result []Expression
	for IsTerm() {
		expression := compileExpression()
		result = append(result, expression)
		if token.value == "," {
			getToken()
			continue
		} else {
			break
		}
	}
	return result
}
