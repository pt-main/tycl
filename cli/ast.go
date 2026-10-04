package cli

import (
	"fmt"
	"strings"

	"github.com/pt-main/lc/v2/parsing/stringParsing"
	"github.com/pt-main/lc/v2/tooling/astools"
	lcprocC "github.com/pt-main/tycl/contract/lcproc"
	"github.com/pt-main/tycl/diag"
	"github.com/pt-main/tycl/generation"
	"github.com/pt-main/tycl/lang"
	lcprocL "github.com/pt-main/tycl/lang/lcproc"
	"github.com/pt-main/tycl/shared"
)

// ASTNode is a JSON-serialisable syntax tree node.
type ASTNode struct {
	Type     string     `json:"type"`
	Raw      string     `json:"raw,omitempty"`
	Span     *diag.Span `json:"span,omitempty"`
	Children []ASTNode  `json:"children,omitempty"`
	Comment  string     `json:"comment,omitempty"`
}

// astHandler implements `tycl ast`: dump the syntax tree as JSON.
//
// The tree is the stable contract between TYCL and any other tool, so a
// script can read the language without linking the Go library.
func astHandler(ctx *Ctx, args []string) *Result {
	in, err := ReadInput(args[0])
	if err != nil {
		return failFromError(err)
	}
	ctx.track(in)
	kind := "config"
	if len(args) > 1 {
		kind = args[1]
	}

	var root *stringParsing.ParsedNode
	switch kind {
	case "config", "conf":
		parser, perr := lcprocL.NewParser()
		if perr != nil {
			return Fail(diag.FromError(in.Source, perr)...)
		}
		nodes, perr := parser.Parse(in.Text)
		if perr != nil {
			return Fail(diag.FromError(in.Source, perr)...)
		}
		root = &nodes[0]
	case "contract", "cont":
		parser, perr := lcprocC.NewParser()
		if perr != nil {
			return Fail(diag.FromError(in.Source, perr)...)
		}
		nodes, perr := parser.Parse(in.Text)
		if perr != nil {
			return Fail(diag.FromError(in.Source, perr)...)
		}
		root = &nodes[0]
	default:
		return Usagef(
			"unknown kind %q: use 'config' or 'contract'", kind)
	}

	return Ok(map[string]any{
		"file": in.Name,
		"kind": kind,
		"ast":  buildAST(in.Source, root),
	}, "")
}

// buildAST converts a parsed node tree into the JSON representation.
func buildAST(src *diag.Source, node *stringParsing.ParsedNode) ASTNode {
	out := ASTNode{Type: node.Switch, Span: diag.NodeSpan(src, node)}
	if node.Raw != "" {
		out.Raw = node.Raw
	}
	if node.Switch == "COMMENT" {
		if value, ok := node.Metadata["value"].(string); ok {
			out.Comment = strings.TrimSpace(value)
		}
	}
	for _, child := range astools.GetChildren(node) {
		out.Children = append(out.Children, buildAST(src, &child))
	}
	return out
}

// typesHandler implements `tycl types`: list the type system of the language.
func typesHandler(ctx *Ctx, args []string) *Result {
	payload := map[string]any{
		"scalars": lang.ValidTypes()[:6],
		"arrays":  lang.ValidTypes()[6:],
	}
	if ctx.Format == FormatJSON {
		return Ok(payload, "")
	}
	var b strings.Builder
	b.WriteString("scalars:\n")
	for _, t := range payload["scalars"].([]string) {
		fmt.Fprintf(&b, "  %s\n", t)
	}
	b.WriteString("arrays:\n")
	for _, t := range payload["arrays"].([]string) {
		fmt.Fprintf(&b, "  %s\n", t)
	}
	return Document(b.String())
}

// langResolve renders a config path as "type: value" for terminal output.
func langResolve(cfg *shared.Config, path string) (string, bool) {
	entry, ok := lang.Resolve(cfg, path)
	if !ok {
		return "", false
	}
	switch value := entry.Value.(type) {
	case *shared.Config:
		return fmt.Sprintf("object: %d keys", len(ListPaths(value, ""))), true
	case []*shared.Config:
		return fmt.Sprintf("objects: %d items", len(value)), true
	}
	return entry.Type + ": " + scalarText(entry.Type, plainValue(entry)), true
}

// structureHandler implements `tycl structure`: the shape of a config.
func structureHandler(ctx *Ctx, args []string) *Result {
	in, cfg, res := loadConfig(ctx, args[0])
	if res != nil {
		return res
	}
	paths := ListPaths(cfg, "")
	if ctx.Format == FormatJSON {
		return Ok(map[string]any{
			"file":  in.Name,
			"paths": paths,
			"tree":  structureOf(cfg),
		}, "")
	}
	if len(paths) == 0 {
		return Document(in.Name + " is empty")
	}
	var b strings.Builder
	for _, p := range paths {
		entry, _ := langResolve(cfg, p)
		fmt.Fprintf(&b, "%-28s %s\n", p, entry)
	}
	return Document(b.String())
}

// docsHandler implements `tycl docs`: render config comments as documentation.
func docsHandler(ctx *Ctx, args []string) *Result {
	in, cfg, res := loadConfig(ctx, args[0])
	if res != nil {
		return res
	}
	docs := generation.GeneratePlainTextDocs(cfg)
	if ctx.Format == FormatJSON {
		return Ok(map[string]any{
			"file":     in.Name,
			"docs":     docs,
			"comments": cfg.Comments,
		}, "")
	}
	if strings.TrimSpace(docs) == "" {
		return Document(in.Name + " has no documentation comments")
	}
	return Document(docs)
}
