package main

import "bufio"
import "fmt"
import "os"
import "strconv"
import "unicode"
import "unicode/utf8"

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
		switch {
		case l.index > len(l.text)-1:
			return Token{EOF, ""}
		case unicode.IsSpace(l.currentRune):
			l.advance()
			continue
		case l.currentRune == '*':
			l.advance()
			return Token{MULTIPLY, l.currentRune}
		case l.currentRune == '/':
			l.advance()
			return Token{DIVIDE, l.currentRune}
		case l.currentRune == '+':
			l.advance()
			return Token{PLUS, l.currentRune}
		case l.currentRune == '-':
			l.advance()
			return Token{MINUS, l.currentRune}
		case l.currentRune == '(':
			l.advance()
			return Token{LEFTPAREN, l.currentRune}
		case l.currentRune == ')':
			l.advance()
			return Token{RIGHTPAREN, l.currentRune}
		case unicode.IsDigit(l.currentRune):
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
func (p *Parser) parens() (AST, error) {
	if p.currentToken.tokenType == MINUS || p.currentToken.tokenType == PLUS {
		t := p.currentToken
		p.eat(p.currentToken.tokenType)
		d, err := p.parens()
		if err != nil {
			return nil, err
		}
		return UnaryOp{t, d}, nil
	}
	if p.currentToken.tokenType == INTEGER {
		t := p.currentToken
		d := p.currentToken.value.(int)
		p.eat(INTEGER)
		return Num{t, d}, nil
	}
	if p.currentToken.tokenType == LEFTPAREN {
		p.eat(LEFTPAREN)
		d, err := p.outer()
		if err != nil {
			return nil, err
		}
		p.eat(RIGHTPAREN)
		return d, nil
	}
	return nil, fmt.Errorf("Expected PARENS or TERM but got %s", p.currentToken.tokenType)
}

// inner : term ((MUL|DIV)term)*
func (p *Parser) inner() (AST, error) {
	ast, err := p.parens()
	if err != nil {
		return nil, err
	}

	for p.currentToken.tokenType == MULTIPLY || p.currentToken.tokenType == DIVIDE {
		if p.currentToken.tokenType == MULTIPLY {
			op := p.currentToken
			p.eat(MULTIPLY)
			d, err := p.parens()
			if err != nil {
				return nil, err
			}
			ast = BinOp{op, ast, d}
		} else if p.currentToken.tokenType == DIVIDE {
			op := p.currentToken
			p.eat(DIVIDE)
			d, err := p.parens()
			if err != nil {
				return nil, err
			}
			ast = BinOp{op, ast, d}
		}
	}

	return ast, nil
}

// outer : inner ((PLUS|MINUS)inner)*
func (p *Parser) outer() (AST, error) {
	ast, err := p.inner()
	if err != nil {
		return nil, err
	}
	for p.currentToken.tokenType == PLUS || p.currentToken.tokenType == MINUS {
		if p.currentToken.tokenType == PLUS {
			op := p.currentToken
			p.eat(PLUS)
			d, err := p.inner()
			if err != nil {
				return nil, err
			}
			ast = BinOp{op, ast, d}
		} else if p.currentToken.tokenType == MINUS {
			op := p.currentToken
			p.eat(MINUS)
			d, err := p.inner()
			if err != nil {
				return nil, err
			}
			ast = BinOp{op, ast, d}
		}
	}

	return ast, nil

}

// AST
type AST interface {
	Eval() int
}

type UnaryOp struct {
	op  Token
	arg AST
}

func (u UnaryOp) String() string {
	return fmt.Sprintf("%s(%v)", u.op.tokenType, u.arg)
}

func (u UnaryOp) Eval() int {
	argVal := u.arg.Eval()
	switch op := u.op.tokenType; op {
	case MINUS:
		return -1 * argVal
	case PLUS:
		return argVal
	default:
		return -100000000
	}
}

type BinOp struct {
	op    Token
	left  AST
	right AST
}

func (b BinOp) String() string {
	return fmt.Sprintf("%s of (%v and %v)\n", b.op.tokenType, b.left, b.right)
}

func (b BinOp) Eval() int {
	leftVal := b.left.Eval()
	rightVal := b.right.Eval()
	switch op := b.op.tokenType; op {
	case MINUS:
		return leftVal - rightVal
	case PLUS:
		return leftVal + rightVal
	case MULTIPLY:
		return leftVal * rightVal
	case DIVIDE:
		return leftVal / rightVal
	default:
		return -1000000000
	}
}

type Num struct {
	token Token
	value int
}

func (n Num) String() string {
	return fmt.Sprintf("%d", n.value)
}

func (n Num) Eval() int {
	return n.value
}

func main() {
	fmt.Print("calc> ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	fmt.Println("[" + input + "]")
	l := NewLexer(input)
	fmt.Println(input)
	p := NewParser(l)
	ast, _ := p.outer()
	fmt.Println(ast.Eval())
}
