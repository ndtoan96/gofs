package search

import (
	"os"
	"path"
	"sort"
	"strings"

	"github.com/gobwas/glob"
	"github.com/ndtoan96/gofs/model"
)

func Search(dir string, text string) ([]model.SearchResult, error) {
	results := make([]model.SearchResult, 0)
	lowerText := strings.ToLower(text)
	pattern, err := glob.Compile(lowerText)
	var patternP *glob.Glob
	if err != nil {
		patternP = nil
	} else {
		patternP = &pattern
	}
	err = recursiveSearch(&results, dir, lowerText, patternP)
	if err != nil {
		return nil, err
	}
	sort.Slice(results, func(i int, j int) bool {
		return results[i].Path < results[j].Path
	})
	return results, nil
}

func recursiveSearch(results *[]model.SearchResult, dir string, text string, pattern *glob.Glob) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		name := strings.ToLower(entry.Name())
		p := strings.ToLower(path.Join(dir, name))
		if strings.Contains(name, text) {
			*results = append(*results, model.SearchResult{Path: path.Join(dir, entry.Name()), IsDir: entry.IsDir()})
		} else if pattern != nil && ((*pattern).Match(name) || (*pattern).Match(p)) {
			*results = append(*results, model.SearchResult{Path: path.Join(dir, entry.Name()), IsDir: entry.IsDir()})
		}
		if entry.IsDir() {
			recursiveSearch(results, path.Join(dir, entry.Name()), text, pattern)
		}
	}
	return nil
}
