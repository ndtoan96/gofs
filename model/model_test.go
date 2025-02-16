package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBreadCrumbItems(t *testing.T) {
	model := FilesPageModel{Path: "./a/b/c/d"}
	items := model.Parents()
	assert.Equal(t, items[0].Name, "a", "should be equal")
	assert.Equal(t, items[1].Name, "b", "should be equal")
	assert.Equal(t, items[2].Name, "c", "should be equal")
	assert.Equal(t, items[0].Path, "a", "should be equal")
	assert.Equal(t, items[1].Path, "a/b", "should be equal")
	assert.Equal(t, items[2].Path, "a/b/c", "should be equal")
}
