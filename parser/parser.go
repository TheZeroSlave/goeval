package parser

import (
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"unicode"
)

type token struct {
	data      string
	operation bool
	cmp       bool
	dot       bool
	comma     bool
	number    bool
	name      bool
	grouping  bool
}

func isLetter(c byte) bool {
	return unicode.IsLetter(rune(c))
}

func ParseExpression(e string) []token {
	var i int
	res := make([]token, 0)
	for i < len(e) {
		switch e[i] {
		case '+', '*', '/', '-':
			t := token{data: string(e[i]), operation: true}
			res = append(res, t)
			i++
		case '>', '<', '=':
			t := token{data: string(e[i]), cmp: true}
			if i+1 < len(e) && e[i+1] == '=' {
				t.data = e[i : i+1]
				i++
			}
			res = append(res, t)
			i++
		case '(', ')':
			t := token{data: string(e[i]), grouping: true}
			res = append(res, t)
			i++
		case '.':
			t := token{data: string(e[i]), dot: true}
			res = append(res, t)
			i++
		case ',':
			t := token{data: string(e[i]), comma: true}
			res = append(res, t)
			i++
		case ' ':
			for i < len(e) && e[i] == ' ' && i < len(e) {
				i++
			}
		}
		if i < len(e) && e[i] >= '0' && e[i] <= '9' {
			t := token{}
			j := i
			for i < len(e) && e[i] >= '0' && e[i] <= '9' {
				i++
			}
			t.data = e[j:i]
			t.number = true
			res = append(res, t)
		}
		if i < len(e) && isLetter(e[i]) {
			t := token{}
			j := i
			for i < len(e) && isLetter(e[i]) {
				i++
			}
			t.data = e[j:i]
			t.name = true
			res = append(res, t)
		}
	}
	return res
}

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
	t          *token
	ts         []token // in case of function
	isFunction bool
	nodes      []*node
	ctx        string
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
	fmt.Println("val:", *n)
	if n.t.number {
		fmt.Println("number")
		v, err := strconv.ParseFloat(n.t.data, 64)
		if err != nil {
			panic(err)
		}
		return v
	} else if n.t.name && !n.isFunction {
		fmt.Println("func")
		val := ctx[n.ts[0].data]
		fmt.Println("looking for name")
		rval := reflect.Indirect(reflect.ValueOf(val))
		for i := 1; i < len(n.ts); i++ {
			if !n.ts[i].dot {
				continue
			}
			fmt.Println("lokking for ", n.ts[i].data)
			rval = reflect.Indirect(rval.FieldByName(n.ts[i].data))
		}
		fmt.Println("found: ", rval.Interface())
		switch rval.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return float64(rval.Int())
		case reflect.Float32, reflect.Float64:
			return rval.Float()
		}
		panic("not valuable expression!")
	} else if n.t.name && n.isFunction {
		fmt.Println("found func", n.ts)
		val := ctx[n.ts[0].data]
		fmt.Println("looking for func name")
		rval := reflect.Indirect(reflect.ValueOf(val))
		for i := 1; i < len(n.ts)-1; i++ {
			if n.ts[i].dot {
				continue
			}
			fmt.Println("lokking for ", n.ts[i].data)
			rval = reflect.Indirect(rval.FieldByName(n.ts[i].data))
		}
		fmt.Println("len of ts=", n.ts[len(n.ts)-1].data)
		methodType := rval.MethodByName(n.ts[len(n.ts)-1].data)
		if !methodType.IsValid() {
			methodType = rval.Addr().MethodByName(n.ts[len(n.ts)-1].data)
			if !methodType.IsValid() {
				panic("not found: " + n.ts[len(n.ts)-1].data)
			}
		}
		vals := []reflect.Value{}
		for _, child := range n.nodes {
			vals = append(vals, reflect.ValueOf(child.Execute(ctx)))
		}
		output := methodType.Call(vals)[0]
		switch output.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return float64(output.Int())
		case reflect.Float32, reflect.Float64:
			return output.Float()
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

/*
pr: (g || pr)
e : p + e (>e)
p: f * p
f: [0-9]
f: [a-z]
f: [a-z].[a-z](e, e, ...)
f: (e)
*/

func group(tokens []token, c int) (*node, int) {
	traceId := rand.Int()
	fmt.Println("enter to group", c)
	if c >= len(tokens) {
		return nil, c
	}
	fmt.Println("call expr1", c, traceId)
	e1, c := expr(tokens, c)
	fmt.Println("get answer from expr ", traceId)
	if c >= len(tokens) || !tokens[c].cmp {
		return e1, c
	}
	root := &node{
		ctx:       "group",
		operation: getOp(tokens[c].data),
		t:         &tokens[c],
	}
	c++
	fmt.Println("call another group ", traceId)
	e2, c := expr(tokens, c)
	if e2 == nil {
		panic("nothing to compare with")
	}
	root.nodes = append(root.nodes, e1, e2)
	fmt.Println("return from group ", c, traceId)
	return root, c
}

func factor(tokens []token, c int) (*node, int) {
	if c >= len(tokens) {
		return nil, c
	}
	fmt.Println("enter to factor", c, tokens[c])
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
		fmt.Println("choose naming", tokens[c], c)
		for c < len(tokens) && (tokens[c].name || tokens[c].dot) {
			c++
		}
		fmt.Println("choose naming end", tokens[c], c)
		endNameC := startNameC
		if c < len(tokens) && tokens[c].data == "(" {
			endNameC = c
			fmt.Println("finding args func", c, tokens[c].data)
			n.isFunction = true
			for {
				e, c2 := expr(tokens, c+1)
				c = c2
				if e == nil {
					break
				}
				n.nodes = append(n.nodes, e)
				if !tokens[c].comma {
					break
				}
			}
			fmt.Println("chosed name", c, tokens[c].data)
			if tokens[c].data != ")" {
				panic("not closed bracket in func")
			}
			fmt.Println("here ", c, tokens[startNameC: endNameC])
			n.ts = tokens[startNameC : endNameC]
			return n, c + 1
		} else {
			n.ts = tokens[startNameC:c]
			fmt.Println("return naming", tokens[c].data, c)
			return n, c
		}
	} else if tokens[c].grouping && tokens[c].data == "(" {
		fmt.Println("call group from factor", tokens[c], c)
		n, c := group(tokens, c+1)
		if tokens[c].data != ")" {
			panic("not closed bracket...")
		}
		return n, c + 1
	} else if tokens[c].grouping && tokens[c].data == ")" {
		return nil, c
	}
	fmt.Println("not expected to be here...", tokens[c])
	panic("not expected")
	return nil, 0
}

func product(tokens []token, c int) (*node, int) {
	if c >= len(tokens) {
		return nil, c
	}
	fmt.Println("enter to product", c, tokens[c])
	f1, c := factor(tokens, c)
	root := &node{ctx: "product"}
	if f1 != nil {
		root.nodes = append(root.nodes, f1)
	} else {
		return nil, c
	}
	if c < len(tokens) {
		if tokens[c].operation && (tokens[c].data == "*" || tokens[c].data == "/") {
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

func BuildTree(tokens []token) *node {
	root, _ := group(tokens, 0)
	return root
}

func expr(tokens []token, c int) (*node, int) {
	if c >= len(tokens) {
		return nil, c
	}
	fmt.Println("enter to expr", c, tokens[c])
	root := &node{ctx: "expr"}
	n1, c := product(tokens, c)
	if n1 != nil {
		root.nodes = append(root.nodes, n1)
	} else {
		fmt.Println("return ", c)
		return nil, c
	}

	needToFindAnother := false
	if c < len(tokens) {
		if tokens[c].operation && (tokens[c].data == "+" || tokens[c].data == "-") {
			root.operation = getOp(tokens[c].data)
			root.t = &tokens[c]
			c++
			needToFindAnother = true
		}
	}
	//	fmt.Println("needToFindAnother", needToFindAnother, tokens[c].data)
	if needToFindAnother {
		fmt.Println("call from expr to expr")
		n2, c2 := expr(tokens, c)
		if n2 != nil {
			root.nodes = append(root.nodes, n2)
		}
		c = c2
	}

	if len(root.nodes) == 1 {
		return root.nodes[0], c
	}
	return root, c
}
