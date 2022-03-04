package outscale

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTagsSyntax(t *testing.T) {
	tags := map[string]bool{
		"=keyvalue":       false,
		"key=value":       true,
		"key=":            true,
		"key=value=false": false,
	}
	for tag, expected := range tags {
		res := validateExtraTagsFormat([]string{tag})

		assert.Equalf(t, expected, res, "The result is not the on expected for the tags '%v'", tag)
	}
}
