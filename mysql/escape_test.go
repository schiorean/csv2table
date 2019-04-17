package mysql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEscapeString(t *testing.T) {
	assert.Equal(t, escapeString("Hi I'm \\ Sorin"), "Hi I\\'m \\\\ Sorin")
	assert.Equal(t, escapeString("And \"author\" is me\r\n"), "And \\\"author\\\" is me\\r\\n")
}
