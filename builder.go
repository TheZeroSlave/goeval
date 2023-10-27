package goeval

import (
	"fmt"
	"reflect"
	"strconv"
)

const (
	none  = 0
	mul   = 1
	div   = 2
	plus  = 3
	minus = 4

	cmp = 5
)

type node struct {
	operation  int
	t          *Token
	ts         []Token // in case of function
	isFunction bool
	nodes      []*node
	ctx        string
}

/*
group : expr(>expr)*
expr: product (+ expr)
product: factor (* product)
f: [0-9]
f: [a-z]
f: [a-z].[a-z](e, e, ...)
f: (expr)
*/

func BuildTree(tokens []Token) (n *node, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered from buildtree %v", r)
		}
	}()

	n, _ = group(tokens, 0)

	return
}

func (n *node) PrintTree(lvl int) {
	fmt.Printf("level:%d token:%v op:%v ctx:%v \n", lvl, n.t, n.operation, n.ctx)
	for _, c := range n.nodes {
		c.PrintTree(lvl + 1)
	}
}

func boolToFloat64(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
}

func (n *node) Execute(ctx map[string]interface{}) float64 {
	if n.t.number {
		v, err := strconv.ParseFloat(n.t.data, 64)
		if err != nil {
			panic(err)
		}
		return v
	} else if n.t.name && !n.isFunction {
		val := ctx[n.ts[0].data]
		rval := reflect.Indirect(reflect.ValueOf(val))

		// find the last node in the chain
		for i := 1; i < len(n.ts); i++ {
			if !n.ts[i].IsDot() {
				continue
			}
			rval = reflect.Indirect(rval.FieldByName(n.ts[i].data))
		}

		switch rval.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return float64(rval.Int())
		case reflect.Float32, reflect.Float64:
			return rval.Float()
			// TODO: add check for boolean and other types
		}

		panic("not valuable expression!")
	} else if n.t.name && n.isFunction {
		val := ctx[n.ts[0].data]
		rval := reflect.Indirect(reflect.ValueOf(val))

		// find the last node in the chain
		for i := 1; i < len(n.ts)-1; i++ {
			if n.ts[i].IsDot() {
				continue
			}
			rval = reflect.Indirect(rval.FieldByName(n.ts[i].data))
		}

		methodName := n.ts[len(n.ts)-1].data
		methodType := rval.MethodByName(methodName)

		if !methodType.IsValid() {
			methodType = rval.Addr().MethodByName(methodName)
			if !methodType.IsValid() {
				panic("not found: " + methodName)
			}
		}

		vals := []reflect.Value{}
		for _, child := range n.nodes {
			vals = append(vals, reflect.ValueOf(child.Execute(ctx)))
		}

		if len(vals) != methodType.Type().NumIn() {
			panic(
				fmt.Errorf(
					"invalid count of input args for method \"%v\". Real=%v, Passed=%v, Args=%v",
					methodName,
					methodType.Type().NumIn(),
					len(vals),
					vals,
				),
			)
		}

		output := methodType.Call(vals)[0]
		switch output.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return float64(output.Int())
		case reflect.Float32, reflect.Float64:
			return output.Float()
			// TODO: add checking for boolean and other types
		}
	}

	arg1 := n.nodes[0].Execute(ctx)
	if len(n.nodes) == 1 {
		return arg1
	}
	arg2 := n.nodes[1].Execute(ctx)
	switch n.operation {
	case plus:
		return arg1 + arg2
	case minus:
		return arg1 - arg2
	case div:
		return arg1 / arg2
	case mul:
		return arg1 * arg2
	case cmp:
		switch n.t.data {
		case ">":
			//fmt.Println("arg1", arg1, "arg2", arg2)
			return boolToFloat64(arg1 > arg2)
		case ">=":
			return boolToFloat64(arg1 >= arg2)
		case "<":
			return boolToFloat64(arg1 < arg2)
		case "<=":
			return boolToFloat64(arg1 <= arg2)
		case "==":
			return boolToFloat64(arg1 == arg2)
		}
	}

	return 0.0
}

func getOp(s string) int {
	switch s {
	case "+":
		return plus
	case "-":
		return minus
	case "*":
		return mul
	case "/":
		return div
	case ">", "<", "==", ">=", "<=":
		return cmp
	}
	return none
}

func group(tokens []Token, c int) (*node, int) {
	if c >= len(tokens) {
		return nil, c
	}
	e1, c := expr(tokens, c)
	if c >= len(tokens) || !tokens[c].cmp {
		return e1, c
	}
	root := &node{
		ctx:       "group",
		operation: getOp(tokens[c].data),
		t:         &tokens[c],
	}
	c++
	e2, c := expr(tokens, c)
	if e2 == nil {
		panic("nothing to compare with")
	}
	root.nodes = append(root.nodes, e1, e2)
	return root, c
}

func factor(tokens []Token, c int) (*node, int) {
	if c >= len(tokens) {
		return nil, c
	}
	if tokens[c].number {
		return &node{
			operation: 0,
			t:         &tokens[c],
			ctx:       "factor-number",
		}, c + 1
	} else if tokens[c].name {
		startNameC := c
		n := &node{
			t:   &tokens[c],
			ctx: "name",
		}
		for c < len(tokens) && (tokens[c].name || tokens[c].IsDot()) {
			c++
		}
		endNameC := startNameC
		if c < len(tokens) && tokens[c].data == "(" {
			endNameC = c
			n.isFunction = true
			for {
				e, c2 := expr(tokens, c+1)
				c = c2
				if e == nil {
					break
				}
				n.nodes = append(n.nodes, e)
				if !tokens[c].IsComma() {
					break
				}
			}
			if tokens[c].data != ")" {
				panic("not closed bracket in func")
			}
			n.ts = tokens[startNameC:endNameC]
			return n, c + 1
		} else {
			n.ts = tokens[startNameC:c]
			return n, c
		}
	} else if tokens[c].IsGrouping() && tokens[c].data == "(" {
		n, c := group(tokens, c+1)
		if tokens[c].data != ")" {
			panic("not closed bracket...")
		}
		return n, c + 1
	} else if tokens[c].IsGrouping() && tokens[c].data == ")" {
		return nil, c
	}
	panic("not expected")
	return nil, 0
}

func product(tokens []Token, c int) (*node, int) {
	if c >= len(tokens) {
		return nil, c
	}

	f1, c := factor(tokens, c)

	root := &node{ctx: "product"}

	if f1 == nil {
		return nil, c
	}

	root.nodes = append(root.nodes, f1)

	if c < len(tokens) {
		if tokens[c].IsOP() && (tokens[c].data == "*" || tokens[c].data == "/") {
			root.operation = getOp(tokens[c].data)
			root.t = &tokens[c]
			c++
		} else {
			if len(root.nodes) == 1 {
				return root.nodes[0], c
			}
			return root, c
		}
	}
	f2, c := product(tokens, c)
	if f2 != nil {
		root.nodes = append(root.nodes, f2)
	}
	if len(root.nodes) == 1 {
		return root.nodes[0], c
	}
	return root, c
}

func expr(tokens []Token, c int) (*node, int) {
	if c >= len(tokens) {
		return nil, c
	}
	root := &node{ctx: "expr"}
	n1, c := product(tokens, c)

	if n1 == nil {
		return nil, c
	}

	root.nodes = append(root.nodes, n1)

	needToFindAnother := false
	if c < len(tokens) {
		if tokens[c].IsOP() && (tokens[c].data == "+" || tokens[c].data == "-") {
			root.operation = getOp(tokens[c].data)
			root.t = &tokens[c]
			c++
			needToFindAnother = true
		}
	}

	if needToFindAnother {
		n2, c2 := expr(tokens, c)
		if n2 != nil {
			root.nodes = append(root.nodes, n2)
		} else {
			panic(fmt.Errorf("no next valid node found after operator %v", tokens[c].data))
		}
		c = c2
	}

	if len(root.nodes) == 1 {
		return root.nodes[0], c
	}
	return root, c
}
