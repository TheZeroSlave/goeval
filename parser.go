package goeval

import (
	"errors"
	"fmt"
	"unicode"
)

func isLetter(c byte) bool {
	return unicode.IsLetter(rune(c))
}

var (
	errSingleEqual = errors.New("we found a '=' single")
)

func ParseExpression(e string) (TokenList, error) {
	//var i int
	res := make([]Token, 0)

	queue := stringQ{str: e}

	for queue.ok() {
		ch, ok := queue.next()

		if !ok {
			break
		}

		switch ch {
		case '+', '*', '/', '-', '(', ')', '.', ',':
			t := Token{data: string(ch)}
			res = append(res, t)
		case '>', '<', '=':
			t := Token{data: string(ch), cmp: true}
			if ch == '=' {
				if nextCh, ok := queue.next(); ok && nextCh == '=' {
					t.data = string([]byte{ch, nextCh})
				} else {
					return nil, errSingleEqual
				}
			}
			res = append(res, t)
		case ' ':
			for queue.hasNext() {
				c, _ := queue.next()
				if c != ' ' {
					queue.rollback()
					break
				}
			}
		default:
			if ch >= '0' && ch <= '9' {
				t := Token{}
				dt := []byte{ch}

				for queue.hasNext() {
					c, _ := queue.next()
					if c >= '0' && c <= '9' {
						dt = append(dt, c)
					} else {
						queue.rollback()
						break
					}
				}
				t.data = string(dt)
				t.number = true
				res = append(res, t)
				continue
			}

			if isLetter(ch) {
				t := Token{name: true}
				dt := []byte{ch}

				for queue.hasNext() {
					c, _ := queue.next()

					if !isLetter(c) {
						queue.rollback()
						break
					}

					dt = append(dt, c)
				}
				t.data = string(dt)

				res = append(res, t)

				continue
			}

			return nil, fmt.Errorf("not valid char %v", ch)
		}
	}
	return TokenList(res), nil
}
