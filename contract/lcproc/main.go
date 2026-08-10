package lcproc

import (
	"github.com/dlclark/regexp2"
	"github.com/pt-main/lc/parsing/stringParsing"
	"github.com/pt-main/lc/parsing/stringParsing/parser3"
)

func NewLexer() *stringParsing.Lexer {
	rules := []stringParsing.LexerRule{
		{Type: "COMMENT", Pattern: regexp2.MustCompile(`(?s)/\*(?<value>.*?)\*/`, 0)},
		{Type: "CONTRACT", Pattern: regexp2.MustCompile(`strict|flexible|dynamic`, 0)},
		{Type: "IDENT", Pattern: regexp2.MustCompile(`[a-zA-Z_][a-zA-Z0-9_\-]*`, 0)},
		{Type: "COLON", Pattern: regexp2.MustCompile(`:`, 0)},
		{Type: "ASSERT", Pattern: regexp2.MustCompile(`=`, 0)},
		{Type: "LBRACE", Pattern: regexp2.MustCompile(`\{`, 0)},
		{Type: "RBRACE", Pattern: regexp2.MustCompile(`\}`, 0)},
		{Type: "SEPARATOR", Pattern: regexp2.MustCompile(`,`, 0)},

		{Type: "WHITESPACE", Pattern: regexp2.MustCompile(`\s+`, 0)},
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

func NewParser() *parser3.Parser {
	return parser3.NewParser(NewLexer(), createGrammar(), "config", []string{
		"WHITESPACE",
	})
}
