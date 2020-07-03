package structs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSize(t *testing.T) {
	p := Page{
		Url:     "12345",
		Raw:     "12345",
		Content: "12345",
		Links:   []string{"1", "2", "3", "4", "5"},
	}
	assert.Equal(t, 20, p.Size(), "sizes should be equal")
}
