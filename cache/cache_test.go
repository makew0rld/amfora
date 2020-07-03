package cache

import (
	"testing"

	"github.com/makeworld-the-better-one/amfora/structs"
	"github.com/stretchr/testify/assert"
)

var p = structs.Page{Url: "example.com"}
var p2 = structs.Page{Url: "example.org"}
var queryPage = structs.Page{Url: "gemini://example.com/test?query"}

func reset() {
	Clear()
	SetMaxPages(0)
	SetMaxSize(0)
}

func TestMaxPages(t *testing.T) {
	reset()
	SetMaxPages(1)
	Add(&p)
	Add(&p2)
	assert.Equal(t, 1, NumPages(), "there should only be one page")
}

func TestMaxSize(t *testing.T) {
	reset()
	assert := assert.New(t)
	SetMaxSize(p.Size())
	Add(&p)
	assert.Equal(1, NumPages(), "one page should be added")
	Add(&p2)
	assert.Equal(1, NumPages(), "there should still be just one page due to cache size limits")
	assert.Equal(p2.Url, urls[0], "the only page url should be the second page one")
}

func TestRemove(t *testing.T) {
	reset()
	Add(&p)
	Remove(p.Url)
	assert.Equal(t, 0, NumPages(), "there shouldn't be any pages after the removal")
}

func TestClearAndNumPages(t *testing.T) {
	reset()
	Add(&p)
	Clear()
	assert.Equal(t, 0, len(pages), "map should be empty")
	assert.Equal(t, 0, len(urls), "urls slice shoulde be empty")
	assert.Equal(t, 0, NumPages(), "NumPages should report empty too")
}

func TestSize(t *testing.T) {
	reset()
	Add(&p)
	assert.Equal(t, p.Size(), Size(), "sizes should match")
}

func TestGet(t *testing.T) {
	reset()
	Add(&p)
	Add(&p2)
	page, ok := Get(p.Url)
	if !ok {
		t.Fatal("Get should say that the page was found")
	}
	if page.Url != p.Url {
		t.Error("page urls don't match")
	}
}
