package mysql

import "testing"
import "github.com/stretchr/testify/assert"

func TestApplyNull(t *testing.T) {
	nullIf := []string{"nope", "yes", "somethign else", ""}
	assert.True(t, applyNull(nullIf, "yes"))
	assert.False(t, applyNull(nullIf, "yess"))
}
