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

func (t Token) String() string {
	return fmt.Sprintf("Token %s of %v\n", t.tokenType, t.value)
}

type Lexer struct {
	text        string
	index       int
	currentRune rune
}

func NewLexer(text string) Lexer {
	r, _ := utf8.DecodeRuneInString(text)
	return Lexer{text, 0, r}
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
type Parser struct {
	l Lexer
}

func (p *Parser) eat(tokenType string) error {
	t := p.l.getNextToken()
	if t.tokenType != tokenType {
		return fmt.Errorf("Expected token type %s but got %s", tokenType, t.tokenType)
	}
	return nil
}

func (p *Parser) term() (int, error) {
	t := p.l.getNextToken()
	if t.tokenType != INTEGER {
		return 0, fmt.Errorf("Expected INTEGER but got %s", t.tokenType)
	}
	return t.value.(int), nil
}

func (p *Parser) exp() (int, error) {
	result, err := p.term()
	if err != nil {
		return 0, err
	}

	op := p.l.getNextToken()

	for op.tokenType != EOF {
		if op.tokenType == MINUS {
			d, err := p.term()
			if err != nil {
				return 0, err
			}
			result = result - d
		} else if op.tokenType == PLUS {
			d, err := p.term()
			if err != nil {
				return 0, err
			}
			result = result + d
		}
		op = p.l.getNextToken()
	}

	return result, nil
}

func main() {
	var b []byte
	fmt.Print("calc> ")
	fmt.Scanln(&b)
	input := string(b)
	l := NewLexer(input)
	fmt.Println(input)
	p := Parser{l}
	fmt.Println(p.exp())
}
