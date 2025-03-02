package main

import (
	"bufio"
	"bytes"
	"html/template"
	"os"
	"path"
	"sort"
	"unicode/utf8"

	htmlFmt "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/ndtoan96/gofs/model"
)

const PREVIEW_LIMIT = 1024 * 10

func getHtmlPreview(dir string, name string) template.HTML {
	if name == "" {
		return template.HTML("")
	}
	filePath := path.Join(dir, name)
	isText, err := isTextFile(filePath)
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

	}
	return template.HTML("Cannot preview this file")
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
	content, err := readPart(filePath, PREVIEW_LIMIT)
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
	content, err := readPart(filePath, PREVIEW_LIMIT)
	if err != nil {
		return "", err
	}
	return "<pre>\n" + string(content) + "\n</pre>", nil
}

func renderMarkdown(filePath string) (string, error) {
	content, err := readPart(filePath, PREVIEW_LIMIT)
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

func readPart(filePath string, n int) ([]byte, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	buffer := make([]byte, n)
	bufReader := bufio.NewReader(f)
	numBytes, err := bufReader.Read(buffer)
	if err != nil {
		return nil, err
	}
	return buffer[:numBytes], nil
}

func isTextFile(filePath string) (bool, error) {
	content, err := readPart(filePath, 1024)
	if err != nil {
		return false, err
	}
	if utf8.Valid(content) {
		return true, nil
	} else {
		return false, nil
	}
}

func doSearch(dir string, text string) ([]model.SearchResult, error) {
	results := make([]model.SearchResult, 0)
	err := recursiveSearch(&results, dir, text)
	if err != nil {
		return nil, err
	}
	sort.Slice(results, func(i int, j int) bool {
		return results[i].Score > results[j].Score
	})
	return results, nil
}

func recursiveSearch(results *[]model.SearchResult, dir string, text string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		score := fuzzy.RankMatchFold(text, path.Join(dir, entry.Name()))
		if score != -1 {
			*results = append(*results, model.SearchResult{Path: path.Join(dir, entry.Name()), IsDir: entry.IsDir(), Score: score})
		}
		if entry.IsDir() {
			recursiveSearch(results, path.Join(dir, entry.Name()), text)
		}
	}
	return nil
}
