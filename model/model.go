package model

import (
	"fmt"
	"html/template"
	"path"
	"slices"
	"strings"
	"time"
)

type ISize interface {
	SizeStr() string
}

type FileSize int64
type DirSize uint64

func (s FileSize) SizeStr() string {
	switch {
	case s < 1024:
		return fmt.Sprintf("%v B", s)
	case s >= 1024 && s < 1024*1024:
		return fmt.Sprintf("%0.2f KB", float64(s)/1024)
	case s >= 1024*1024 && s < 1024*1024*1024:
		return fmt.Sprintf("%0.2f MB", float64(s)/(1024*1024))
	default:
		return fmt.Sprintf("%0.2f GB", float64(s)/(1024*1024*1024))
	}
}

func (s DirSize) SizeStr() string {
	return fmt.Sprintf("%v items", s)
}

type Item struct {
	Name         string
	LastModified time.Time
	Size         ISize
	IsDir        bool
	FileType     string
}

type FilesPageModel struct {
	Path        Path
	Items       []Item
	AllowWrite  bool
	SelectState string
	SortField   string
	Desc        bool
	Preview     template.HTML
	PreviewName string
}

type DeletePageModel struct {
	Path  Path
	Names []string
}

type ArchivePageModel struct {
	Path  Path
	Names []string
}

type NewFolderPageModel struct {
	Path Path
}

type RenamePageModel struct {
	Path     Path
	OldNames []string
}

type EditPageModel struct {
	Path     Path
	Names    []string
	Contents []string
}

type SearchPageModel struct {
	Path    Path
	Search  string
	Results []SearchResult
}

type ParentItem struct {
	Name string
	Path Path
}

type Path string

type SearchResult struct {
	Path  string
	IsDir bool
	Score int
}

func (r SearchResult) Dir() string {
	if r.IsDir {
		return r.Path
	} else {
		return path.Dir(r.Path)
	}
}

func (r SearchResult) RelativeDirFrom(p Path) string {
	tmp := strings.TrimPrefix(r.Path, string(p))
	tmp = strings.TrimPrefix(tmp, "/")
	if r.IsDir {
		return tmp
	} else {
		return path.Dir(tmp)
	}
}

func (r SearchResult) FileName() string {
	if r.IsDir {
		return ""
	} else {
		return path.Base(r.Path)
	}
}

func (p Path) Parents() []ParentItem {
	items := make([]ParentItem, 0)
	for parent := path.Dir(string(p)); parent != "."; parent = path.Dir(parent) {
		items = append(items, ParentItem{Path: Path(parent), Name: path.Base(parent)})
	}
	slices.Reverse(items)
	return items
}

func (p Path) CurrentDir() string {
	return path.Base(string(p))
}
