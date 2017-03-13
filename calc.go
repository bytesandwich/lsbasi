package main

import "fmt"
import "unicode/utf8"
import "unicode"
import "strconv"

// LEXER

const PLUS = "PLUS"
const MINUS = "MINUS"
const EOF = "EOF"
const INTEGER = "INTEGER"

type Token struct {
	tokenType string
	value     interface{}
}

type Lexer struct {
	text        string
	index       int
	currentRune rune
}

func (l *Lexer) advance() {
	_, size := utf8.DecodeRuneInString(l.text[l.index:])
	l.index += size
	l.currentRune, _ = utf8.DecodeRuneInString(l.text[l.index:])
}
func (l *Lexer) integer() int {
	var s string = string(l.currentRune)
	l.advance()
	for unicode.IsDigit(l.currentRune) {
		s += string(l.currentRune)
		l.advance()
	}
	d, _ := strconv.Atoi(s)
	return d
}

func (l *Lexer) getNextToken() Token {
	for {
		if l.index > len(l.text)-1 {
			return Token{EOF, ""}
		}
		if unicode.IsSpace(l.currentRune) {
			l.advance()
			continue
		}
		if l.currentRune == '+' {
			l.advance()
			return Token{PLUS, l.currentRune}
		}
		if l.currentRune == '-' {
			l.advance()
			return Token{MINUS, l.currentRune}
		}
		if unicode.IsDigit(l.currentRune) {
			return Token{INTEGER, l.integer()}
		}
	}
}

// PARSER

func main() {
	var b []byte
	fmt.Print("calc> ")
	fmt.Scanln(&b)
	input := string(b)
	l := Lexer{input, 0, ' '}
	fmt.Println(string(b))
	for t := l.getNextToken(); t.tokenType != EOF; t = l.getNextToken() {
		fmt.Println(t.tokenType)
	}
}
