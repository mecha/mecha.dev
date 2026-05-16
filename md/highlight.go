package md

import (
	"bytes"
	"html"
	"strings"

	"github.com/alecthomas/chroma/v2"
	chromaHtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/gomarkdown/markdown/ast"
)

var chromaFmtr = chromaHtml.New(
	chromaHtml.ClassPrefix("tok-"),
	chromaHtml.WithClasses(true),
)

var chromaFmtrWithLineNums = chromaHtml.New(
	chromaHtml.ClassPrefix("tok-"),
	chromaHtml.WithClasses(true),
	chromaHtml.WithLineNumbers(true),
	chromaHtml.WithLinkableLineNumbers(true, "L"),
)

func highlightCodeBlock(codeBlock *ast.CodeBlock) string {
	lang := codeBlockLang(codeBlock)
	code := string(codeBlock.Literal)

	lexer := lexerForCode(lang, code)
	it, err := chroma.Coalesce(lexer).Tokenise(nil, code)
	if err != nil {
		return html.EscapeString(code)
	}

	fmtr := chromaFmtr
	if lang != "" {
		fmtr = chromaFmtrWithLineNums
	}

	var htmlBuf bytes.Buffer
	err = fmtr.Format(&htmlBuf, styles.Fallback, it)
	if err != nil {
		return html.EscapeString(code)
	}

	return htmlBuf.String()
}

func lexerForCode(lang, code string) chroma.Lexer {
	if lang != "" {
		if lexer := lexers.Get(lang); lexer != nil {
			return lexer
		}
	}

	if lexer := lexers.Analyse(code); lexer != nil {
		return lexer
	}

	return lexers.Fallback
}

func codeBlockLang(codeBlock *ast.CodeBlock) string {
	fields := strings.Fields(string(codeBlock.Info))
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}
