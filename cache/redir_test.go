package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddRedir(t *testing.T) {
	ClearRedirs()
	AddRedir("A", "B")
	assert.Equal(t, "B", Redirect("A"), "A redirects to B")

	// Chain
	AddRedir("B", "C")
	assert.Equal(t, "C", Redirect("B"), "B redirects to C")
	assert.Equal(t, "C", Redirect("A"), "A now redirects to C too")

	// Loop
	ClearRedirs()
	AddRedir("A", "B")
	AddRedir("B", "A")
	assert.Equal(t, "A", Redirect("B"), "B redirects to A - most recent version of loop is used")
}
