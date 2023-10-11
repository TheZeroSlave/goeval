package goeval

type TokenList []Token

type Token struct {
	data      string
	operation bool
	cmp       bool
	number    bool
	name      bool
}

func (t Token) Data() string {
	return t.data
}

func (t Token) IsNum() bool {
	return t.number
}

func (t Token) IsGrouping() bool {
	return t.data == "(" || t.data == ")"
}

func (t Token) IsDot() bool {
	return t.data == "."
}

func (t Token) IsComma() bool {
	return t.data == ","
}
