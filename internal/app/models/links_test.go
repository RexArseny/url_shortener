package models

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLinks(t *testing.T) {
	expected := &Links{
		M:     sync.RWMutex{},
		Links: make(map[string]string),
	}

	actual := NewLinks()

	assert.Equal(t, expected, actual)
}
