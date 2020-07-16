package main

import (
	"fmt"
	"goeval/parser"
)

type A struct {
}

func (a *A) Test() int {
	return 10
}

func main() {
	tokens := parser.ParseExpression("a.Test()")
	fmt.Println("tokens", tokens)
	root := parser.BuildTree(tokens)
	fmt.Println("tokens=", tokens, root)
	root.PrintTree(0)
	answer := root.Execute(map[string]interface{}{
		"a": &A{},
	})

	fmt.Println("answer is", answer)
}