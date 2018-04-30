package symbols

import "fmt"

type SymbolKind int

const (
	STATIC SymbolKind = iota
	FIELD
	ARGUMENT
	VAR
)

var symbolKindStrings []string = []string{
	"STATIC",
	"FIELD",
	"ARGUMENT",
	"VAR",
}

func (sk SymbolKind) String() string {
	return symbolKindStrings[sk]
}

type SymbolType struct {
	simpleType SimpleType
	className  string
}

type SimpleType int

const (
	INT SimpleType = iota
	CHAR
	BOOLEAN
	CLASS
)

var simpleTypeStrings []string = []string{
	"Int",
	"Char",
	"Boolean",
}

func (st SymbolType) String() string {
	if st.simpleType == CLASS {
		return "class " + st.className
	}
	return simpleTypeStrings[st.simpleType]
}

type Symbol struct {
	kindOf SymbolKind
	typeOf SymbolType
	index  int
}

type SymbolTable struct {
	Name string
	Map  map[string]Symbol
}

var classSymbolTable SymbolTable
var subroutingSymbolTable SymbolTable

func NewSymbolTable(name string) SymbolTable {
	return SymbolTable{Name: name, Map: make(map[string]Symbol)}
}

func TypeFromString(typeName string) SymbolType {
	for i, simpleTypeString := range simpleTypeStrings {
		if typeName == simpleTypeString {
			return SymbolType{simpleType: SimpleType(i)}
		}
	}
	return SymbolType{simpleType: CLASS, className: typeName}
}

func KindFromString(kindName string) SymbolKind {
	for i, symbolKindString := range symbolKindStrings {
		if kindName == symbolKindString {
			return SymbolKind(i)
		}
	}
	panic("in KindFromString, should never happen, unrecognized symbol kind")
}

func (st SymbolTable) VarCount(symbolKind SymbolKind) int {
	count := 0
	for _, symbol := range st.Map {
		if symbol.kindOf == symbolKind {
			count++
		}
	}
	return count
}

func (st SymbolTable) define(name string, symbolType SymbolType, symbolKind SymbolKind) {
	symbol := Symbol{kindOf: symbolKind, typeOf: symbolType, index: st.VarCount(symbolKind)}
	// fmt.Printf("in table %s, defining %s, type:%v, kind:%v, index:%d\n", st.Name, name, symbol.typeOf, symbol.kindOf, symbol.index)
	st.Map[name] = symbol
}

func (st SymbolTable) Define(name string, typeName string, kind SymbolKind) {
	symbolType := TypeFromString(typeName)
	st.define(name, symbolType, kind)
}

func (st SymbolTable) Lookup(name string) Symbol {
	symbol, ok := st.Map[name]
	if ok {
		return symbol
	} else {
		panic(fmt.Sprintf("name %s not found in symbol table %s", name, st.Name))
	}
}

func (s Symbol) String() string {
	return fmt.Sprintf("symbol type:%v, kind:%v, index:%d", s.typeOf, s.kindOf, s.index)
}

func (s Symbol) KindOf() SymbolKind {
	return s.kindOf
}

func (s Symbol) TypeOf() SymbolType {
	return s.typeOf
}

func (s Symbol) IndexOf() int {
	return s.index
}
