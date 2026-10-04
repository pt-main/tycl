package lcproc

import (
	"github.com/pt-main/lc/v2/engine/core"
	"github.com/pt-main/lc/v2/parsing/stringParsing"
	"github.com/pt-main/lc/v2/parsing/stringParsing/parser3"
)

func NewLexer() (*stringParsing.Lexer, core.ErrorInterface) {
	rules := []stringParsing.LexerRule{
		{Type: "STRING", Pattern: `"(?:\\.|[^"\\])*"`},
		{Type: "STRING", Pattern: `'(?:\\.|[^'\\])*'`},
		{Type: "COMMENT", Pattern: `(?s)/\*(?<value>.*?)\*/`},
		{Type: "COMMENT_LINE", Pattern: `//[^\n]*`},
		{Type: "FLOAT", Pattern: `-?\d+\.\d+`},
		{Type: "INT", Pattern: `-?\d+`},
		{Type: "BOOL", Pattern: `true|false`},
		{Type: "NULL", Pattern: `null`},
		{Type: "IDENT", Pattern: `[a-zA-Z_][a-zA-Z0-9_\-]*`},
		{Type: "ASSIGN", Pattern: `=`},
		{Type: "COLON", Pattern: `:`},
		{Type: "LBRACE", Pattern: `\{`},
		{Type: "RBRACE", Pattern: `\}`},
		{Type: "LBRACK", Pattern: `\[`},
		{Type: "RBRACK", Pattern: `\]`},
		{Type: "LPAREN", Pattern: `\(`},
		{Type: "RPAREN", Pattern: `\)`},
		{Type: "SEPARATOR", Pattern: `,`},

		{Type: "WHITESPACE", Pattern: `\s+`},
	}
	config := &stringParsing.LexerConfig{UseBracketBalance: false}
	return stringParsing.NewLexer(rules, config)
}

func createGrammar() parser3.Grammar {
	return parser3.Grammar{
		"config": {
			Name: "config",
			Expr: parser3.NodeExpr{
				NodeType: "config",
				Expr:     parser3.NamedExpr{RuleName: "object"},
			},
		},

		"object": {
			Name: "object",
			Expr: parser3.NodeExpr{
				NodeType: "object",
				Expr: parser3.SequenceExpr{
					Exprs: []parser3.Expr{
						parser3.TokenExpr{TokenType: "LBRACE"},
						parser3.OptionalExpr{
							Expr: parser3.TokenExpr{TokenType: "COMMENT"},
						},
						parser3.SequenceExpr{
							Exprs: []parser3.Expr{
								parser3.RepeatExpr{
									Expr: parser3.SequenceExpr{
										Exprs: []parser3.Expr{
											parser3.NamedExpr{RuleName: "pair"},
											parser3.TokenExpr{TokenType: "SEPARATOR"},
										},
									},
									Min: 0,
								},
								parser3.OptionalExpr{
									Expr: parser3.NamedExpr{RuleName: "pair"},
								},
							},
						},
						parser3.OptionalExpr{
							Expr: parser3.TokenExpr{TokenType: "COMMENT"},
						},
						parser3.TokenExpr{TokenType: "RBRACE"},
					},
				},
			},
		},

		"array": {
			Name: "array",
			Expr: parser3.NodeExpr{
				NodeType: "array",
				Expr: parser3.SequenceExpr{
					Exprs: []parser3.Expr{
						parser3.TokenExpr{TokenType: "LBRACK"},
						parser3.RepeatExpr{
							Expr: parser3.SequenceExpr{
								Exprs: []parser3.Expr{
									parser3.NamedExpr{RuleName: "value"},
									parser3.TokenExpr{TokenType: "SEPARATOR"},
								},
							},
							Min: 0,
						},
						parser3.OptionalExpr{
							Expr: parser3.NamedExpr{RuleName: "value"},
						},
						parser3.TokenExpr{TokenType: "RBRACK"},
					},
				},
			},
		},

		"pair": {
			Name: "pair",
			Expr: parser3.NodeExpr{
				NodeType: "pair",
				Expr: parser3.SequenceExpr{
					Exprs: []parser3.Expr{
						parser3.TokenExpr{TokenType: "IDENT"},
						parser3.OptionalExpr{
							Expr: parser3.SequenceExpr{
								Exprs: []parser3.Expr{
									parser3.TokenExpr{TokenType: "COLON"},
									parser3.TokenExpr{TokenType: "IDENT"},
								},
							},
						},
						parser3.TokenExpr{TokenType: "ASSIGN"},
						parser3.NamedExpr{RuleName: "value"},
					},
				},
			},
		},

		"value": {
			Name: "value",
			Expr: parser3.ChoiceExpr{
				Alternatives: []parser3.Expr{
					parser3.NamedExpr{RuleName: "action"},
					parser3.NamedExpr{RuleName: "object"},
					parser3.NamedExpr{RuleName: "array"},
					parser3.TokenExpr{TokenType: "STRING"},
					parser3.TokenExpr{TokenType: "NULL"},
					parser3.TokenExpr{TokenType: "INT"},
					parser3.TokenExpr{TokenType: "FLOAT"},
					parser3.TokenExpr{TokenType: "BOOL"},
				},
			},
		},

		"action": {
			Name: "action",
			Expr: parser3.NodeExpr{
				NodeType: "action",
				Expr: parser3.SequenceExpr{
					Exprs: []parser3.Expr{
						parser3.TokenExpr{TokenType: "IDENT"},
						parser3.TokenExpr{TokenType: "LPAREN"},
						parser3.RepeatExpr{
							Expr: parser3.SequenceExpr{
								Exprs: []parser3.Expr{
									parser3.NamedExpr{RuleName: "value"},
									parser3.TokenExpr{TokenType: "SEPARATOR"},
								},
							},
							Min: 0,
						},
						parser3.OptionalExpr{
							Expr: parser3.NamedExpr{RuleName: "value"},
						},
						parser3.TokenExpr{TokenType: "RPAREN"},
					},
				},
			},
		},
	}
}

func NewParser() (*parser3.Parser, core.ErrorInterface) {
	lexer, err := NewLexer()
	if err != nil {
		return nil, err
	}
	return parser3.NewParser(lexer, createGrammar(), "config", []string{
		"WHITESPACE",
		"COMMENT_LINE",
	}), nil
}
