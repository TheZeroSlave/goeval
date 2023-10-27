package goeval_test

import (
	"reflect"
	"testing"

	"github.com/TheZeroSlave/goeval"
	"github.com/stretchr/testify/assert"
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
	tokens, err := goeval.ParseExpression("1 + (a.BB.Test(15) * 2 + 1)")
	assert.NoError(t, err)

	root, err := goeval.BuildTree(tokens)
	assert.NoError(t, err)

	answer := root.Execute(map[string]interface{}{
		"a": a,
	})

	if answer != 62.0 {
		t.Fatal("the answer should be 62")
		t.Fail()
	}
}

func TestSeveralAdd(t *testing.T) {
	a := &A{Data: "aaa", BB: B{C: 10}}
	tokens, err := goeval.ParseExpression("1 + 4 + 5 + 6")
	assert.NoError(t, err)

	root, err := goeval.BuildTree(tokens)
	assert.NoError(t, err)

	answer := root.Execute(map[string]interface{}{
		"a": a,
	})

	assert.Equal(t, 16.0, answer)
}

func TestSeveralMul(t *testing.T) {
	a := &A{Data: "aaa", BB: B{C: 10}}
	tokens, err := goeval.ParseExpression("1 * 2 * 3 * 4")
	assert.NoError(t, err)

	root, err := goeval.BuildTree(tokens)
	assert.NoError(t, err)

	answer := root.Execute(map[string]interface{}{
		"a": a,
	})

	if answer != 24.0 {
		t.Fatal("the answer should be 24")
	}
}

func TestComplexGroup(t *testing.T) {
	a := &A{Data: "aaa", BB: B{C: 10}}
	tokens, err := goeval.ParseExpression("1 + 2 * (3 * (1 + 1))")
	assert.NoError(t, err)

	root, err := goeval.BuildTree(tokens)

	assert.NoError(t, err)

	answer := root.Execute(map[string]interface{}{
		"a": a,
	})

	if answer != 13.0 {
		t.Fatal("the answer should be 13")
	}
}

func TestCmp(t *testing.T) {
	a := &A{Data: "aaa", BB: B{C: 10}}
	tokens, err := goeval.ParseExpression("0 + 1 > 3")
	assert.NoError(t, err)

	root, err := goeval.BuildTree(tokens)
	assert.NoError(t, err)

	answer := root.Execute(map[string]interface{}{
		"a": a,
	})

	if answer != 0.0 {
		t.Fatal("the answer should be 0")
	}
}

func TestMultipleGroups(t *testing.T) {
	a := &A{Data: "aaa", BB: B{C: 10}}
	tokens, err := goeval.ParseExpression("a.Test(8) > 7")
	assert.NoError(t, err)

	root, err := goeval.BuildTree(tokens)

	assert.NoError(t, err)

	answer := root.Execute(map[string]interface{}{
		"a": a,
	})

	assert.Equal(t, 1.0, answer)
}

func TestCmpObj(t *testing.T) {
	tokens, err := goeval.ParseExpression("((a + 2) > 3)")
	assert.NoError(t, err)

	root, err := goeval.BuildTree(tokens)

	assert.NoError(t, err)

	answer := root.Execute(map[string]interface{}{
		"a": 2,
	})

	if answer != 1.0 {
		t.Fatal("the answer should be 1")
	}
}

func TestCmpFunc(t *testing.T) {
	tokens, err := goeval.ParseExpression("((a.Test(2) + 2) > 3)")
	assert.NoError(t, err)
	assert.Len(t, tokens, 14)

	root, err := goeval.BuildTree(tokens)

	assert.NoError(t, err)

	answer := root.Execute(map[string]interface{}{
		"a": &A{},
	})

	if answer != 1.0 {
		t.Fatal("the answer should be 1")
	}
}

func TestNotCompleteExpr(t *testing.T) {
	tokens, err := goeval.ParseExpression("5+")
	assert.NoError(t, err)
	assert.Len(t, tokens, 2)

	root, err := goeval.BuildTree(tokens)
	assert.Nil(t, root)
}

func TestParseLen(t *testing.T) {
	tokens, err := goeval.ParseExpression("((a.Test(2) + 2) > 3)")
	assert.NoError(t, err)
	assert.Len(t, tokens, 14)

	tests := []struct {
		grouping  bool
		number    bool
		name      bool
		dot       bool
		comma     bool
		operation bool
		cmp       bool
	}{
		{grouping: true},
		{grouping: true},
		{name: true},
		{dot: true},
		{name: true},
		{grouping: true},
		{number: true},
		{grouping: true},
		{operation: true},
		{number: true},
		{grouping: true},
		{cmp: true},
		{number: true},
		{grouping: true},
	}

	for ix, token := range tokens {
		assert.Equal(t, token.IsCMP(), tests[ix].cmp)
		assert.Equal(t, token.IsOP(), tests[ix].operation)
		assert.Equal(t, token.IsComma(), tests[ix].comma)
		assert.Equal(t, token.IsDot(), tests[ix].dot)
		assert.Equal(t, token.IsNum(), tests[ix].number)
		assert.Equal(t, token.IsName(), tests[ix].name)
		assert.Equal(t, token.IsGrouping(), tests[ix].grouping)
	}
}

func TestParseExpression_Cases(t *testing.T) {
	type args struct {
		e string
	}
	tests := []struct {
		name    string
		args    args
		want    goeval.TokenList
		wantErr bool
	}{
		{
			name:    "invalid tokens, single equal",
			args:    args{e: "1+5=7"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid tokens, #",
			args:    args{e: "1+5#7"},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := goeval.ParseExpression(tt.args.e)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseExpression() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseExpression() = %v, want %v", got, tt.want)
			}
		})
	}
}
