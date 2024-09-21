package openssl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCertsDir(t *testing.T) {
	sut, err := GetCertsDir()

	assert.Nil(t, err)
	assert.NotNil(t, sut)
	assert.Contains(t, sut, "/home")
	assert.Contains(t, sut, "/.local/share/amfora")
}

func TestGetPageDir_InvalidURL(t *testing.T) {
    _, err := GetPageDir("/home/path", "gnu.org")
    assert.NotNil(t, err)

    _, err = GetPageDir("/home/path", "gemini://")
    assert.NotNil(t, err)

    _, err = GetPageDir("/home/path", "gemini")
    assert.NotNil(t, err)
}

func TestGetPageDir(t *testing.T) {
    sut, err := GetPageDir("/home/path", "gemini://gnu.org")

    assert.Nil(t, err)
    assert.Equal(t, "/home/path/gnu.org", sut)

}
