package main

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

func main() {
	var s string
	fmt.Scan(&s)

	ans, err := f(s)
	if err != nil {
		fmt.Printf("error occured, err = %v", err)
		return
	}

	fmt.Println(ans)
}

//это будет работать только если у нас цифры от 0 до 9
func f(s string) (string, error) {
	var prevChar rune
	var flag bool
	sb := strings.Builder{}
	for _, ch := range s {
		if unicode.IsDigit(ch) {
			if !flag {
				return "", errors.New("bad string input")
			}
			digit := ch - '0'
			for range int(digit) {
				sb.Write([]byte(string(prevChar)))
			}
			flag = false
		} else {
			flag = true
			prevChar = ch
		}
	}
	return sb.String(), nil
}