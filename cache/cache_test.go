package cache

import (
	"testing"

	"github.com/makeworld-the-better-one/amfora/structs"
	"github.com/stretchr/testify/assert"
)

var p = structs.Page{URL: "example.com"}
var p2 = structs.Page{URL: "example.org"}

func reset() {
	ClearPages()
	SetMaxPages(0)
	SetMaxSize(0)
}

func TestMaxPages(t *testing.T) {
	reset()
	SetMaxPages(1)
	AddPage(&p)
	AddPage(&p2)
	assert.Equal(t, 1, NumPages(), "there should only be one page")
}

func TestMaxSize(t *testing.T) {
	reset()
	assert := assert.New(t)
	SetMaxSize(p.Size())
	AddPage(&p)
	assert.Equal(1, NumPages(), "one page should be added")
	AddPage(&p2)
	assert.Equal(1, NumPages(), "there should still be just one page due to cache size limits")
	assert.Equal(p2.URL, urls[0], "the only page url should be the second page one")
}

func TestRemove(t *testing.T) {
	reset()
	AddPage(&p)
	RemovePage(p.URL)
	assert.Equal(t, 0, NumPages(), "there shouldn't be any pages after the removal")
}

func TestClearAndNumPages(t *testing.T) {
	reset()
	AddPage(&p)
	ClearPages()
	assert.Equal(t, 0, len(pages), "map should be empty")
	assert.Equal(t, 0, len(urls), "urls slice shoulde be empty")
	assert.Equal(t, 0, NumPages(), "NumPages should report empty too")
}

func TestSize(t *testing.T) {
	reset()
	AddPage(&p)
	assert.Equal(t, p.Size(), SizePages(), "sizes should match")
}

func TestGet(t *testing.T) {
	reset()
	AddPage(&p)
	AddPage(&p2)
	page, ok := GetPage(p.URL)
	if !ok {
		t.Fatal("Get should say that the page was found")
	}
	if page.URL != p.URL {
		t.Error("page urls don't match")
	}
}
