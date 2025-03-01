package model

import (
	"fmt"
	"path"
	"slices"
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
}

type FilesPageModel struct {
	Path        Path
	Items       []Item
	AllowWrite  bool
	SelectState string
	SortField   string
	Desc        bool
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

type ParentItem struct {
	Name string
	Path Path
}

type Path string

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
