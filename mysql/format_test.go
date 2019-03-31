package mysql

import "testing"
import "github.com/stretchr/testify/assert"

func TestParseDate(t *testing.T) {
	// EN
	v, err := parseDate("2006-01-02", "2019-05-21")
	if assert.Nil(t, err) {
		assert.Equal(t, v, "2019-05-21")
	}

	// DE
	v, err = parseDate("02.01.2006", "01.12.2019")
	if assert.Nil(t, err) {
		assert.Equal(t, v, "2019-12-01")
	}

	// wrong date value
	v, err = parseDate("02.01.2006", "01.-12.2019")
	assert.NotNil(t, err)
}

func TestParseDateTime(t *testing.T) {
	// EN
	v, err := parseDateTime("2006-01-02 15:04:05", "2019-05-21 01:22:59")
	if assert.Nil(t, err) {
		assert.Equal(t, v, "2019-05-21 01:22:59")
	}

	// DE
	v, err = parseDateTime("02.01.2006 15:04:05", "21.05.2019 01:22:59")
	if assert.Nil(t, err) {
		assert.Equal(t, v, "2019-05-21 01:22:59")
	}

	// wrong date value
	v, err = parseDateTime("02.01.2006 15:04:05", "01.-12.2019 01.22.59")
	assert.NotNil(t, err)
}
func TestApplyNull(t *testing.T) {
	nullIf := []string{"nope", "yes", "somethign else", ""}
	assert.True(t, applyNull(nullIf, "yes"))
	assert.False(t, applyNull(nullIf, "yess"))
}

func TestParseFloat(t *testing.T) {
	assert.Equal(t, parseFloat("1.2", "1,500.50"), "1500.50")
	assert.Equal(t, parseFloat("1,2", "1500,50"), "1500.50")

	//todo
	// default float format must be 1,2
}
