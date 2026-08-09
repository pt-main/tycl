package lcproc

import (
	"github.com/dlclark/regexp2"
	"github.com/pt-main/lc/parsing/stringParsing"
	"github.com/pt-main/lc/parsing/stringParsing/parser3"
)

func NewLexer() *stringParsing.Lexer {
	rules := []stringParsing.LexerRule{
		{Type: "STRING", Pattern: regexp2.MustCompile(`"(?:\\.|[^"\\])*"`, 0)},
		{Type: "STRING", Pattern: regexp2.MustCompile(`'(?:\\.|[^'\\])*'`, 0)},
		// {Type: "LITERAL", Pattern: regexp2.MustCompile("(?s)`"+`(?:\\.|[^`+"`"+`\\])*`+"`", 0)},
		{Type: "COMMENT", Pattern: regexp2.MustCompile(`(?s)/\*(?<value>.*?)\*/`, 0)},
		{Type: "FLOAT", Pattern: regexp2.MustCompile(`-?\d+\.\d+`, 0)},
		{Type: "INT", Pattern: regexp2.MustCompile(`-?\d+`, 0)},
		{Type: "BOOL", Pattern: regexp2.MustCompile(`true|false`, 0)},
		{Type: "NULL", Pattern: regexp2.MustCompile(`null`, 0)},
		{Type: "IDENT", Pattern: regexp2.MustCompile(`[a-zA-Z_][a-zA-Z0-9_\-]*`, 0)},
		{Type: "ASSIGN", Pattern: regexp2.MustCompile(`=`, 0)},
		{Type: "COLON", Pattern: regexp2.MustCompile(`:`, 0)},
		{Type: "LBRACE", Pattern: regexp2.MustCompile(`\{`, 0)},
		{Type: "RBRACE", Pattern: regexp2.MustCompile(`\}`, 0)},
		{Type: "LBRACK", Pattern: regexp2.MustCompile(`\[`, 0)},
		{Type: "RBRACK", Pattern: regexp2.MustCompile(`\]`, 0)},
		{Type: "LPAREN", Pattern: regexp2.MustCompile(`\(`, 0)},
		{Type: "RPAREN", Pattern: regexp2.MustCompile(`\)`, 0)},
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
						parser3.OptionalExpr{
							Expr: parser3.TokenExpr{TokenType: "COMMENT"},
						},
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
						parser3.OptionalExpr{
							Expr: parser3.TokenExpr{TokenType: "COMMENT"},
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
						parser3.OptionalExpr{
							Expr: parser3.TokenExpr{TokenType: "COMMENT"},
						},
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
						parser3.OptionalExpr{
							Expr: parser3.TokenExpr{TokenType: "COMMENT"},
						},
						parser3.TokenExpr{TokenType: "RPAREN"},
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
