package md

import (
	"bufio"
	"html/template"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/gomarkdown/markdown"
	mdAst "github.com/gomarkdown/markdown/ast"
	mdHtml "github.com/gomarkdown/markdown/html"
	mdParser "github.com/gomarkdown/markdown/parser"
)

// A parsed markdown document
type ParsedDoc struct {
	// The front-matter data
	Head map[string]string
	// The converted HTML of the document
	Body template.HTML
}

// Parses a markdown file with front-matter support.
func ParseFile(filepath string) (*ParsedDoc, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	return Parse(file)
}

// Parses a markdown string with front-matter support.
func Parse(reader io.Reader) (*ParsedDoc, error) {
	head := make(map[string]string)

	scanner := bufio.NewScanner(reader)
	for lineNum := 0; scanner.Scan(); lineNum++ {
		lineStr := strings.TrimSpace(scanner.Text())

		if lineStr == "" {
			continue
		}
		if strings.HasPrefix(lineStr, "---") {
			break
		}

		name, value, found := strings.Cut(lineStr, ":")
		if !found {
			slog.Error("syntax error in markdown front matter", "line", lineNum)
			continue
		}
		name, value = strings.TrimSpace(name), strings.TrimSpace(value)
		head[name] = value
	}

	mdStr := ""
	for scanner.Scan() {
		mdStr += scanner.Text() + "\n"
	}
	body := ToHTML(strings.TrimSpace(mdStr))

	return &ParsedDoc{head, body}, nil
}

// Converts a markdown string into HTML
func ToHTML(md string) template.HTML {
	parser := mdParser.NewWithExtensions(
		mdParser.CommonExtensions | mdParser.AutoHeadingIDs | mdParser.NoEmptyLineBeforeBlock,
	)
	renderer := mdHtml.NewRenderer(mdHtml.RendererOptions{
		Flags: mdHtml.CommonFlags | mdHtml.HrefTargetBlank,
	})
	ast := parser.Parse([]byte(md))

	autoLinkHeadings(ast)

	htmlStr := string(markdown.Render(ast, renderer))
	return template.HTML(strings.TrimSpace(htmlStr))
}

// Adds an anchor link inside each level 2+ heading that links to itself
func autoLinkHeadings(ast mdAst.Node) {
	children := ast.GetChildren()

	for i, node := range children {
		heading, isHeading := node.(*mdAst.Heading)

		if !isHeading || heading.Level < 2 {
			continue
		}

		link := &mdAst.Link{
			Destination: []byte("#" + heading.HeadingID),
			Container: mdAst.Container{
				Children: heading.Children,
			},
		}

		newHeading := &mdAst.Heading{
			Level:     heading.Level,
			HeadingID: heading.HeadingID,
			Container: mdAst.Container{
				Children: []mdAst.Node{link},
			},
		}

		children[i] = newHeading
	}
}
