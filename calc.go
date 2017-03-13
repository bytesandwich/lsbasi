package main

import "fmt"
import "unicode/utf8"
import "unicode"
import "strconv"

// LEXER

const PLUS = "PLUS"
const MINUS = "MINUS"
const MULTIPLY = "MULTIPLY"
const DIVIDE = "DIVIDE"
const LEFTPAREN = "LEFTPAREN"
const RIGHTPAREN = "RIGHTPAREN"
const INTEGER = "INTEGER"
const EOF = "EOF"

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
		if l.currentRune == '*' {
			l.advance()
			return Token{MULTIPLY, l.currentRune}
		}
		if l.currentRune == '/' {
			l.advance()
			return Token{DIVIDE, l.currentRune}
		}
		if l.currentRune == '+' {
			l.advance()
			return Token{PLUS, l.currentRune}
		}
		if l.currentRune == '-' {
			l.advance()
			return Token{MINUS, l.currentRune}
		}
		if l.currentRune == '(' {
			l.advance()
			return Token{LEFTPAREN, l.currentRune}
		}
		if l.currentRune == ')' {
			l.advance()
			return Token{RIGHTPAREN, l.currentRune}
		}
		if unicode.IsDigit(l.currentRune) {
			return Token{INTEGER, l.integer()}
		}
	}
}

// PARSER
type Parser struct {
	l            Lexer
	currentToken Token
}

func NewParser(l Lexer) Parser {
	return Parser{l, l.getNextToken()}
}

func (p *Parser) eat(tokenType string) error {
	if p.currentToken.tokenType == tokenType {
		p.currentToken = p.l.getNextToken()
	} else {
		return fmt.Errorf("Expected token type %s but got %s", tokenType, p.currentToken.tokenType)
	}
	return nil
}

// PARENS : TERM | LEFT_PAREN (PARENS|LOWER) RIGHT_PAREN
func (p *Parser) parens() (int, error) {
	if p.currentToken.tokenType == INTEGER {
		d := p.currentToken.value.(int)
		p.eat(INTEGER)
		return d, nil
	}
	if p.currentToken.tokenType == LEFTPAREN {
		p.eat(LEFTPAREN)
		d, err := p.outer()
		if err != nil {
			return 0, err
		}
		p.eat(RIGHTPAREN)
		return d, nil
	}
	return 0, fmt.Errorf("Expected PARENS or TERM but got %s", p.currentToken.tokenType)
}

// inner : term ((MUL|DIV)term)*
func (p *Parser) inner() (int, error) {
	result, err := p.parens()
	if err != nil {
		return 0, err
	}

	for p.currentToken.tokenType == MULTIPLY || p.currentToken.tokenType == DIVIDE {
		if p.currentToken.tokenType == MULTIPLY {
			p.eat(MULTIPLY)
			d, err := p.parens()
			if err != nil {
				return 0, err
			}
			result = result * d
		} else if p.currentToken.tokenType == DIVIDE {
			p.eat(DIVIDE)
			d, err := p.parens()
			if err != nil {
				return 0, err
			}
			result = result / d
		}
	}

	return result, nil
}

// inner : outer ((PLUS|MINUS)outer)*
func (p *Parser) outer() (int, error) {
	result, err := p.inner()
	if err != nil {
		return 0, err
	}
	for p.currentToken.tokenType == PLUS || p.currentToken.tokenType == MINUS {
		if p.currentToken.tokenType == PLUS {
			p.eat(PLUS)
			d, err := p.outer()
			if err != nil {
				return 0, err
			}
			result = result + d
		} else if p.currentToken.tokenType == MINUS {
			p.eat(MINUS)
			d, err := p.outer()
			if err != nil {
				return 0, err
			}
			result = result - d
		}
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
	p := NewParser(l)
	fmt.Println(p.outer())
}
