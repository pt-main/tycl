package cli

import (
	"strings"

	"github.com/pt-main/tap/go"
)

// Format names supported by the output layer.
const (
	FormatHuman = "human"
	FormatJSON  = "json"
)

// Ctx carries everything a handler needs: the parser flags, the output
// format and the source texts the command worked on.
type Ctx struct {
	Parser *tap.Parser
	Flags  map[string]string
	Format string
	Text   string
	// Sources maps a file name to its text, so an excerpt is always taken
	// from the file the diagnostic actually points into.
	Sources map[string]string
}

// track records the source text a diagnostic refers to, so the renderer can
// show the real line under the caret.
func (c *Ctx) track(in *Input) {
	if in == nil {
		return
	}
	c.Text = in.Text
	if c.Sources == nil {
		c.Sources = map[string]string{}
	}
	c.Sources[in.Name] = in.Text
}

// Flag reports whether a boolean flag was given.
func (c *Ctx) Flag(name string) bool {
	_, ok := c.Flags[name]
	return ok
}

// FlagValue returns a flag value and whether it was provided.
func (c *Ctx) FlagValue(name string) (string, bool) {
	v, ok := c.Flags[name]
	return v, ok
}

// StrictKeys reports whether strict key mode is enabled.
func (c *Ctx) StrictKeys() bool { return c.Flag("strict-keys") }

// OutputFormat resolves the requested output format, defaulting to human.
func (c *Ctx) OutputFormat() string {
	if c.Flag("json") {
		return FormatJSON
	}
	if v, ok := c.Flags["output"]; ok {
		if v == FormatJSON {
			return FormatJSON
		}
	}
	return FormatHuman
}

// newContext builds the handler context and applies colour settings.
func newContext(p *tap.Parser, text string) *Ctx {
	ctx := &Ctx{
		Parser:  p,
		Flags:   p.Flags,
		Text:    text,
		Sources: map[string]string{},
	}
	ctx.Format = ctx.OutputFormat()
	applyColorMode(ctx.Format, p.Flags)
	return ctx
}

// cleanOptional trims and drops empty optional arguments.
func cleanOptional(values []string) []string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}
