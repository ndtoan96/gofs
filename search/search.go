package search

import (
	"os"
	"path"
	"sort"

	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/ndtoan96/gofs/model"
)

func Search(dir string, text string) ([]model.SearchResult, error) {
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
