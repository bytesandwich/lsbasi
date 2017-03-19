package main

import "runtime/debug"
import "fmt"
import "os"
import "reflect"
import "regexp"
import "strings"
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

const BEGIN = "BEGIN"
const END = "END"
const DOT = "DOT"
const ASSIGN = "ASSIGN"
const SEMI = "SEMI"
const ID = "ID"

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

func (l *Lexer) advanceN(n int) {
	for i := 0; i < n; i++ {
		l.advance()
	}
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

var re = regexp.MustCompile("^[[:alpha:]][[:alnum:]]*")

func (l *Lexer) getNextToken() Token {
	for {
		switch {
		case l.index > len(l.text)-1:
			return Token{EOF, ""}
		case unicode.IsSpace(l.currentRune):
			l.advance()
			continue
		case strings.HasPrefix(l.text[l.index:], ":="):
			l.advanceN(2)
			return Token{ASSIGN, ":="}
		case l.currentRune == ';':
			l.advance()
			return Token{SEMI, l.currentRune}
		case l.currentRune == '.':
			l.advance()
			return Token{DOT, l.currentRune}
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
		default:
			matched := re.FindString(l.text[l.index:])
			if len(matched) > 0 {
				l.advanceN(len(matched))
				if matched == "END" {
					return Token{END, "END"}
				}
				if matched == "BEGIN" {
					return Token{BEGIN, "BEGIN"}
				}
				return Token{ID, matched}
			}
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

// program : compound_statement DOT
func (p *Parser) program() (AST, error) {
	compoundStatement, err := p.compoundStatement()
	if err != nil {
		return nil, err
	}
	p.eat(DOT)
	return compoundStatement, nil
}

// compound_statement : BEGIN statement_list END
func (p *Parser) compoundStatement() (AST, error) {
	p.eat(BEGIN)
	list, err := p.statementList()
	if err != nil {
		return nil, err
	}
	p.eat(END)
	return Compound{list}, nil
}

// statement_list : statement
//                | statement SEMI statement_list
func (p *Parser) statementList() ([]AST, error) {
	statement, err := p.statement()
	if err != nil {
		return nil, err
	}
	list := []AST{statement}
	for p.currentToken.tokenType == SEMI {
		p.eat(SEMI)
		next, err := p.statement()
		if err != nil {
			return nil, err
		}
		list = append(list, next)
	}
	return list, nil
}

// statement : compound_statement
//           | assignment_statement
//           | empty
func (p *Parser) statement() (AST, error) {
	if p.currentToken.tokenType == BEGIN {
		return p.compoundStatement()
	} else if p.currentToken.tokenType == ID {
		return p.assignmentStatement()
	} else {
		fmt.Printf("got empty at %v", p.currentToken)
		return Empty{}, nil
	}

}

// assignment_statement : variable ASSIGN expr
func (p *Parser) assignmentStatement() (AST, error) {
	variable, err := p.variable()
	if err != nil {
		return nil, err
	}
	op := p.currentToken
	p.eat(ASSIGN)
	expr, err := p.expr()
	if err != nil {
		return nil, err
	}
	return Assign{variable, op, expr}, nil
}

func (p *Parser) variable() (Var, error) {
	fmt.Println(reflect.TypeOf(p.currentToken.value))
	if p.currentToken.tokenType != ID {
		debug.PrintStack()
		return Var{}, fmt.Errorf("variable expected token type ID but got %s", p.currentToken.tokenType)
	}
	name, ok := p.currentToken.value.(string)
	if !ok {
		return Var{}, fmt.Errorf("Expected token type ID with string but got %s", p.currentToken.value)
	}
	v := Var{p.currentToken, name}
	p.eat(ID)
	return v, nil
}

// expr: term ((PLUS | MINUS) term)*
func (p *Parser) expr() (AST, error) {
	ast, err := p.term()
	if err != nil {
		return nil, err
	}
	for p.currentToken.tokenType == PLUS || p.currentToken.tokenType == MINUS {
		if p.currentToken.tokenType == PLUS {
			op := p.currentToken
			p.eat(PLUS)
			d, err := p.term()
			if err != nil {
				return nil, err
			}
			ast = BinOp{op, ast, d}
		} else if p.currentToken.tokenType == MINUS {
			op := p.currentToken
			p.eat(MINUS)
			d, err := p.term()
			if err != nil {
				return nil, err
			}
			ast = BinOp{op, ast, d}
		}
	}

	return ast, nil

}

//
// term: factor ((MUL | DIV) factor)*
func (p *Parser) term() (AST, error) {
	ast, err := p.factor()
	if err != nil {
		return nil, err
	}

	for p.currentToken.tokenType == MULTIPLY || p.currentToken.tokenType == DIVIDE {
		if p.currentToken.tokenType == MULTIPLY {
			op := p.currentToken
			p.eat(MULTIPLY)
			d, err := p.factor()
			if err != nil {
				return nil, err
			}
			ast = BinOp{op, ast, d}
		} else if p.currentToken.tokenType == DIVIDE {
			op := p.currentToken
			p.eat(DIVIDE)
			d, err := p.factor()
			if err != nil {
				return nil, err
			}
			ast = BinOp{op, ast, d}
		}
	}

	return ast, nil
}

//
// factor : PLUS factor
//        | MINUS factor
//        | INTEGER
//        | LPAREN expr RPAREN
//        | variable
func (p *Parser) factor() (AST, error) {

	if p.currentToken.tokenType == MINUS || p.currentToken.tokenType == PLUS {
		t := p.currentToken
		p.eat(p.currentToken.tokenType)
		d, err := p.factor()
		if err != nil {
			return nil, err
		}
		return UnaryOp{t, d}, nil
	} else if p.currentToken.tokenType == INTEGER {
		fmt.Println("AT INT", p.currentToken, reflect.TypeOf(p.currentToken.value))
		t := p.currentToken
		d := p.currentToken.value.(int)
		p.eat(INTEGER)
		return Num{t, d}, nil
	} else if p.currentToken.tokenType == LEFTPAREN {
		p.eat(LEFTPAREN)
		d, err := p.expr()
		if err != nil {
			return nil, err
		}
		p.eat(RIGHTPAREN)
		return d, nil
	} else if p.currentToken.tokenType == ID {
		fmt.Println("AT ID", p.currentToken, reflect.TypeOf(p.currentToken.value))
		v, err := p.variable()
		if err != nil {
			return nil, err
		}
		return v, nil
	}
	return nil, fmt.Errorf("Expected PARENS or TERM but got %s", p.currentToken.tokenType)
}

// AST
type AST interface {
}

type Compound struct {
	children []AST
}

type Assign struct {
	left  Var
	op    Token
	right AST
}

type Var struct {
	token Token
	value string
}

type Empty struct{}

type UnaryOp struct {
	op  Token
	arg AST
}

func (u UnaryOp) String() string {
	return fmt.Sprintf("%s(%v)", u.op.tokenType, u.arg)
}

type BinOp struct {
	op    Token
	left  AST
	right AST
}

func (b BinOp) String() string {
	return fmt.Sprintf("%s of (%v and %v)\n", b.op.tokenType, b.left, b.right)
}

type Num struct {
	token Token
	value int
}

func (n Num) String() string {
	return fmt.Sprintf("%d", n.value)
}

type Interpreter struct {
	globals map[string]interface{}
}

func NewInterpereter() Interpreter {
	return Interpreter{make(map[string]interface{})}
}

func (i Interpreter) Eval(a AST) (interface{}, error) {
	var v interface{}
	var err error
	switch a := a.(type) {
	case Compound:
		v, err = i.eval_compound(a)
	case Assign:
		v, err = i.eval_assign(a)
	case UnaryOp:
		v, err = i.eval_unaryOp(a)
	case BinOp:
		v, err = i.eval_binOp(a)
	case Num:
		v = i.eval_num(a)
	case Var:
		v = i.eval_var(a)
	case Empty:
		return nil, nil
	default:
		return nil, fmt.Errorf("Unexpected AST type %v", a)
	}
	return v, err
}

func (i Interpreter) eval_compound(c Compound) (interface{}, error) {
	var value interface{}
	var err error
	for _, child := range c.children {
		value, err = i.Eval(child)
		if err != nil {
			return nil, err
		}
	}
	return value, nil
}

func (i Interpreter) eval_assign(a Assign) (interface{}, error) {
	v, err := i.Eval(a.right)
	if err != nil {
		return nil, err
	}
	i.globals[a.left.value] = v
	return v, nil
}

func (i Interpreter) eval_unaryOp(u UnaryOp) (int, error) {
	argVal, err := i.Eval(u.arg)
	if err != nil {
		return -1, err
	}
	num, ok := argVal.(int)
	if !ok {
		return -1, fmt.Errorf("Expected num but got %v", argVal)
	}
	switch op := u.op.tokenType; op {
	case MINUS:
		return -1 * num, nil
	case PLUS:
		return num, nil
	default:
		return -1, fmt.Errorf("Unexpected op in eval_unary, %v", op)
	}
}

func (i Interpreter) eval_binOp(b BinOp) (int, error) {
	leftVal, err := i.Eval(b.left)
	if err != nil {
		return -1, err
	}
	leftNum, ok := leftVal.(int)
	if !ok {
		return -1, fmt.Errorf("Expected num but got %v", leftVal)
	}
	rightVal, err := i.Eval(b.right)
	if err != nil {
		return -1, err
	}
	rightNum, ok := rightVal.(int)
	if !ok {
		return -1, fmt.Errorf("Expected num but got %v", rightVal)
	}
	switch op := b.op.tokenType; op {
	case MINUS:
		return leftNum - rightNum, nil
	case PLUS:
		return leftNum + rightNum, nil
	case MULTIPLY:
		return leftNum * rightNum, nil
	case DIVIDE:
		return leftNum / rightNum, nil
	default:
		return -1, fmt.Errorf("Unexpected BinOp op %v", b.op.tokenType)
	}
}

func (i Interpreter) eval_num(n Num) int {
	return n.value
}

func (i Interpreter) eval_var(v Var) interface{} {
	return i.globals[v.value]
}

func main() {
	fmt.Print("calc> ")
	// reader := bufio.NewReader(os.Stdin)
	// input, _ := reader.ReadString('\n')
	input := "BEGIN a := 1; b := a + 2 END"
	fmt.Println("[" + input + "]")
	// l := NewLexer(input)
	// fmt.Println(input)
	// for t := l.getNextToken(); t.tokenType != EOF; t = l.getNextToken() {
	// 	fmt.Println(t)
	// }
	l := NewLexer(input)
	p := NewParser(l)
	ast, err := p.program()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	i := NewInterpereter()
	outcome, err2 := i.Eval(ast)
	if err2 != nil {
		fmt.Println(err2)
		os.Exit(1)
	}
	fmt.Println(outcome)
}
