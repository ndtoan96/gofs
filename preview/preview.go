package preview

import (
	"archive/zip"
	"bufio"
	"bytes"
	"html/template"
	"path"

	htmlFmt "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/ndtoan96/gofs/io_helper"
)

const PREVIEW_LIMIT = 1024 * 10

func GetHtmlPreview(dir string, name string) template.HTML {
	if name == "" {
		return template.HTML("")
	}
	filePath := path.Join(dir, name)
	isText, err := io_helper.IsTextFile(filePath)
	if err != nil {
		return template.HTML("Cannot preview this file")
	}
	if isText {
		if path.Ext(name) == ".md" {
			md, err := renderMarkdown(filePath)
			if err != nil {
				return template.HTML("Cannot preview this file")
			}
			return template.HTML(md)
		} else if path.Ext(filePath) == ".svg" {
			return template.HTML("<img width=\"100%\" height=\"100%\" src=\"" + filePath + "\">")
		} else {
			lexer := lexers.Match(name)
			if lexer == nil {
				txt, err := renderText(filePath)
				if err != nil {
					return template.HTML("Cannot preview this file")
				}
				return template.HTML(txt)
			}
			code, err := renderCode(filePath)
			if err != nil {
				return template.HTML("Cannot preview this file")
			}
			return template.HTML(code)
		}
	} else {
		ext := path.Ext(filePath)
		if ext == ".jpg" || ext == ".png" || ext == ".gif" || ext == ".webp" || ext == ".jpeg" {
			return template.HTML("<img width=\"100%\" height=\"100%\" src=\"" + filePath + "\">")
		} else if ext == ".zip" {
			htmlZip, err := renderZip(filePath)
			if err != nil {
				return template.HTML("Cannot preview this file")
			}
			return template.HTML(htmlZip)
		}
		// TODO: preview PDF and epub
	}
	return template.HTML("Cannot preview this file")
}

func renderZip(filePath string) (string, error) {
	z, err := zip.OpenReader(filePath)
	if err != nil {
		return "", err
	}
	defer z.Close()
	files := "<ul>\n"
	for _, f := range z.File {
		files += "<li>" + f.Name + "</li>\n"
	}
	files += "</ul>"
	return files, nil
}

func renderCode(filePath string) (string, error) {
	lexer := lexers.Match(filePath)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	style := styles.Get("xcode")
	if style == nil {
		style = styles.Fallback
	}
	formatter := htmlFmt.New()
	content, err := io_helper.ReadPart(filePath, PREVIEW_LIMIT)
	if err != nil {
		return "", err
	}
	iterator, err := lexer.Tokenise(nil, string(content))
	if err != nil {
		return "", err
	}
	var buffer bytes.Buffer
	w := bufio.NewWriter(&buffer)
	err = formatter.Format(w, style, iterator)
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func renderText(filePath string) (string, error) {
	content, err := io_helper.ReadPart(filePath, PREVIEW_LIMIT)
	if err != nil {
		return "", err
	}
	return "<pre>\n" + string(content) + "\n</pre>", nil
}

func renderMarkdown(filePath string) (string, error) {
	content, err := io_helper.ReadPart(filePath, PREVIEW_LIMIT)
	if err != nil {
		return "", err
	}
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(content)

	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	md := string(markdown.Render(doc, renderer))
	endPreview := ""
	return md + endPreview, nil
}
