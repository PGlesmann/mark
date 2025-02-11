package renderer

import (
	"fmt"
	"regexp"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

type ConfluenceAdmonitionRenderer struct {
	html.Config
}

// NewConfluenceAdmonitionRenderer creates a new instance of the ConfluenceAdmonitionRenderer
func NewConfluenceAdmonitionRenderer(opts ...html.Option) renderer.NodeRenderer {
	return &ConfluenceAdmonitionRenderer{
		Config: html.NewConfig(),
	}
}

// RegisterFuncs implements NodeRenderer.RegisterFuncs
func (r *ConfluenceAdmonitionRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindParagraph, r.renderAdmonition)
}

var admonitionPattern = regexp.MustCompile(`(?i)^!!!\s*(info|note|warning|tip)\s*(.*)$`)

func (r *ConfluenceAdmonitionRenderer) renderAdmonition(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		textNode, ok := n.FirstChild().(*ast.Text)
		if !ok {
			return ast.WalkContinue, nil
		}
		matches := admonitionPattern.FindStringSubmatch(string(textNode.Text(source)))
		if matches == nil {
			return ast.WalkContinue, nil
		}

		admonitionType := matches[1]
		content := matches[2]

		prefix := fmt.Sprintf("<ac:structured-macro ac:name=\"%s\"><ac:parameter ac:name=\"icon\">true</ac:parameter><ac:rich-text-body>", admonitionType)
		if _, err := w.Write([]byte(prefix)); err != nil {
			return ast.WalkStop, err
		}
		if _, err := w.Write([]byte(content)); err != nil {
			return ast.WalkStop, err
		}
	} else {
		suffix := "</ac:rich-text-body></ac:structured-macro>\n"
		if _, err := w.Write([]byte(suffix)); err != nil {
			return ast.WalkStop, err
		}
	}
	return ast.WalkContinue, nil
}
