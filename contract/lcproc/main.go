package lcproc

import (
	"github.com/pt-main/lc/v2/engine/core"
	"github.com/pt-main/lc/v2/parsing/stringParsing"
	"github.com/pt-main/lc/v2/parsing/stringParsing/parser3"
)

func NewLexer() (*stringParsing.Lexer, core.ErrorInterface) {
	rules := []stringParsing.LexerRule{
		{Type: "COMMENT", Pattern: `(?s)/\*(?<value>.*?)\*/`},
		{Type: "COMMENT_LINE", Pattern: `//[^\n]*`},
		{Type: "CONTRACT", Pattern: `strict|flexible|dynamic`},
		{Type: "IDENT", Pattern: `[a-zA-Z_][a-zA-Z0-9_\-]*`},
		{Type: "COLON", Pattern: `:`},
		{Type: "ASSERT", Pattern: `=`},
		{Type: "LBRACE", Pattern: `\{`},
		{Type: "RBRACE", Pattern: `\}`},
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
						parser3.TokenExpr{TokenType: "CONTRACT"},
						parser3.TokenExpr{TokenType: "LBRACE"},
						parser3.OptionalExpr{
							Expr: parser3.TokenExpr{TokenType: "COMMENT"},
						},
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
						parser3.OptionalExpr{
							Expr: parser3.TokenExpr{TokenType: "COMMENT"},
						},
						parser3.TokenExpr{TokenType: "RBRACE"},
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
						parser3.SequenceExpr{
							Exprs: []parser3.Expr{
								parser3.TokenExpr{TokenType: "COLON"},
								parser3.TokenExpr{TokenType: "IDENT"},
							},
						},
						parser3.OptionalExpr{
							Expr: parser3.SequenceExpr{
								Exprs: []parser3.Expr{
									parser3.TokenExpr{TokenType: "ASSERT"},
									parser3.NamedExpr{RuleName: "object"},
								},
							},
						},
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
