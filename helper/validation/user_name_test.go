package validation

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserNameRegex(t *testing.T) {
	valid := []string{
		"Transaction User",
		"User_123",
		"Abc",
		"A" + strings.Repeat("b", 255),
	}
	for _, name := range valid {
		assert.True(t, UserNameRegex.MatchString(name), "expected valid: %q", name)
	}

	invalid := []string{
		"",
		"Ab",
		"1User",
		" User",
		"User<script>",
		"A" + strings.Repeat("b", 256),
	}
	for _, name := range invalid {
		assert.False(t, UserNameRegex.MatchString(name), "expected invalid: %q", name)
	}
}
