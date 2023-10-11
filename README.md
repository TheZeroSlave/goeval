# goeval
Golang library for evaluating simple predicates dynamically.
Maybe usefull in some data driven cases.

```golang
type A struct {
	Data string
	BB   B
}

func (a A) Test(v float64) int {
	return int(v)
}

func something() {
	tokens, err := goeval.ParseExpression("((a.Test(2) + 2) > 3)")
	if err != nil {
		panic(err)
	}

	root := goeval.BuildTree(tokens)

	answer := root.Execute(map[string]interface{}{
		"a": &A{},
	})

	if answer != 1.0 {
		panic("the answer should be 1")
	}
}

```

# run tests
`go test . -failfast -v`

