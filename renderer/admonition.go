package renderer

import (
	"fmt"

	"github.com/kovetskiy/mark/parser"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

// HeadingAttributeFilter defines attribute names which heading elements can have
var AdmonitionAttributeFilter = html.GlobalAttributeFilter

// A Renderer struct is an implementation of renderer.NodeRenderer that renders
// nodes as (X)HTML.
type ConfluenceAdmonitionRenderer struct {
	html.Config
	LevelMap AdmonitionLevelMap
}

// NewConfluenceRenderer creates a new instance of the ConfluenceRenderer
func NewConfluenceAdmonitionRenderer(opts ...html.Option) renderer.NodeRenderer {
	return &ConfluenceAdmonitionRenderer{
		Config:   html.NewConfig(),
		LevelMap: nil,
	}
}

// RegisterFuncs implements NodeRenderer.RegisterFuncs .
func (r *ConfluenceAdmonitionRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(parser.KindAdmonition, r.renderAdmon)
}

// Define AdmonitionType enum
type AdmonitionType int

const (
	AInfo AdmonitionType = iota
	ANote
	AWarn
	ATip
	ANone
)

func (t AdmonitionType) String() string {
	return []string{"info", "note", "warning", "tip", "none"}[t]
}

type AdmonitionLevelMap map[ast.Node]int

func (m AdmonitionLevelMap) Level(node ast.Node) int {
	return m[node]
}

func ParseAdmonitionType(node ast.Node) AdmonitionType {
	n, ok := node.(*parser.Admonition)
	if !ok {
		return ANone
	}

	switch string(n.AdmonitionClass) {
	case "info":
		return AInfo
	case "note":
		return ANote
	case "warning":
		return AWarn
	case "tip":
		return ATip
	default:
		return ANone
	}
}

// GenerateAdmonitionLevel walks a given node and returns a map of blockquote levels
func GenerateAdmonitionLevel(someNode ast.Node) AdmonitionLevelMap {

	// We define state variable that track BlockQuote level while we walk the tree
	admonitionLevel := 0
	AdmonitionLevelMap := make(map[ast.Node]int)

	rootNode := someNode
	for rootNode.Parent() != nil {
		rootNode = rootNode.Parent()
	}
	_ = ast.Walk(rootNode, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if node.Kind() == ast.KindBlockquote && entering {
			AdmonitionLevelMap[node] = admonitionLevel
			admonitionLevel += 1
		}
		if node.Kind() == ast.KindBlockquote && !entering {
			admonitionLevel -= 1
		}
		return ast.WalkContinue, nil
	})
	return AdmonitionLevelMap
}

// renderBlockQuote will render a BlockQuote
func (r *ConfluenceAdmonitionRenderer) renderAdmon(writer util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	//	Initialize BlockQuote level map
	n := node.(*parser.Admonition)
	if r.LevelMap == nil {
		r.LevelMap = GenerateAdmonitionLevel(node)
	}

	admonitionType := ParseAdmonitionType(node)
	admonitionLevel := r.LevelMap.Level(node)

	if admonitionLevel == 0 && entering && admonitionType != ANone {
		prefix := fmt.Sprintf("<ac:structured-macro ac:name=\"%s\"><ac:parameter ac:name=\"icon\">true</ac:parameter><ac:rich-text-body>\n", admonitionType)
		if _, err := writer.Write([]byte(prefix)); err != nil {
			return ast.WalkStop, err
		}
		if string(n.Title) != "" {
			titleHTML := fmt.Sprintf("<p><strong>%s</strong></p>\n", string(n.Title))
			if _, err := writer.Write([]byte(titleHTML)); err != nil {
				return ast.WalkStop, err
			}
		}

		return ast.WalkContinue, nil
	}
	if admonitionLevel == 0 && !entering && admonitionType != ANone {
		suffix := "</ac:rich-text-body></ac:structured-macro>\n"
		if _, err := writer.Write([]byte(suffix)); err != nil {
			return ast.WalkStop, err
		}
		return ast.WalkContinue, nil
	}
	return r.renderAdmonition(writer, source, node, entering)
}

func (r *ConfluenceAdmonitionRenderer) renderAdmonition(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*parser.Admonition)
	if entering {
		if n.Attributes() != nil {
			_, _ = w.WriteString("<blockquote")
			html.RenderAttributes(w, n, AdmonitionAttributeFilter)
			_ = w.WriteByte('>')
		} else {
			_, _ = w.WriteString("<blockquote>\n")
		}
	} else {
		_, _ = w.WriteString("</blockquote>\n")
	}
	return ast.WalkContinue, nil
}
