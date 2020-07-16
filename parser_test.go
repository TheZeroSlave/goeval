package goeval

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

func (a A) Test() {

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
	if answer != 61.0 {
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


