package parser

import (
	"fmt"
	"testing"
)

type B struct {
	C int
}

func (b *B) Test(a float64) int {
	return int(a * 2)
}

type A struct {
	Data string
	BB   B
}

func (a A) Test(v float64) int {
	return int(v)
}

func TestFunctionExpr(t *testing.T) {
	a := &A{Data: "aaa", BB: B{C: 10}}
	tokens := ParseExpression("1 + (a.BB.Test(15) * 2 + 1)")
	root := BuildTree(tokens)
	fmt.Println("tokens=", tokens, root)
	root.PrintTree(0)
	answer := root.Execute(map[string]interface{}{
		"a": a,
	})
	fmt.Println("answer", answer)
	if answer != 62.0 {
		t.Fatal("the answer should be 62")
		t.Fail()
	}
}

func TestSeveralAdd(t *testing.T) {
	a := &A{Data: "aaa", BB: B{C: 10}}
	tokens := ParseExpression("1 + 4 + 5 + 6")
	root := BuildTree(tokens)
	fmt.Println("tokens=", tokens, root)
	root.PrintTree(0)
	answer := root.Execute(map[string]interface{}{
		"a": a,
	})

	if answer != 16.0 {
		t.Fatal("the answer should be 16")
	}
}

func TestSeveralMul(t *testing.T) {
	a := &A{Data: "aaa", BB: B{C: 10}}
	tokens := ParseExpression("1 * 2 * 3 * 4")
	root := BuildTree(tokens)
	fmt.Println("tokens=", tokens, root)
	root.PrintTree(0)
	answer := root.Execute(map[string]interface{}{
		"a": a,
	})

	if answer != 24.0 {
		t.Fatal("the answer should be 24")
	}
}

func TestComplexGroup(t *testing.T) {
	a := &A{Data: "aaa", BB: B{C: 10}}
	tokens := ParseExpression("1 + 2 * (3 * (1 + 1))")
	root := BuildTree(tokens)
	fmt.Println("tokens=", tokens, root)
	root.PrintTree(0)
	answer := root.Execute(map[string]interface{}{
		"a": a,
	})

	if answer != 13.0 {
		t.Fatal("the answer should be 13")
	}
}


func TestCmp(t *testing.T) {
	a := &A{Data: "aaa", BB: B{C: 10}}
	tokens := ParseExpression("0 + 1 > 3")
	root := BuildTree(tokens)
	fmt.Println("tokens=", tokens, root)
	root.PrintTree(0)
	answer := root.Execute(map[string]interface{}{
		"a": a,
	})

	if answer != 0.0 {
		t.Fatal("the answer should be 0")
	}
}


func TestCmpObj(t *testing.T) {
	tokens := ParseExpression("((a + 2) > 3)")
	root := BuildTree(tokens)
	fmt.Println("tokens=", tokens, root)
	root.PrintTree(0)
	answer := root.Execute(map[string]interface{}{
		"a": 2,
	})

	if answer != 1.0 {
		t.Fatal("the answer should be 1")
	}
}


func TestCmpFunc(t *testing.T) {
	tokens := ParseExpression("((a.Test(2) + 2) > 3)")
	root := BuildTree(tokens)
	fmt.Println("tokens=", tokens, root)
	root.PrintTree(0)
	answer := root.Execute(map[string]interface{}{
		"a": &A{},
	})

	if answer != 1.0 {
		t.Fatal("the answer should be 1")
	}
}
