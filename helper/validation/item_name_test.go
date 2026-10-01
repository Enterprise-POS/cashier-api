package validation

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestItemNameRegex(t *testing.T) {
	valid := []string{
		"Apple",
		"Fish & Chips",
		"Coca-Cola 1.5L",
		"2L Aqua",
		"Kopi Susu (Large)",
		"Nasi Goreng, Spesial",
		"Tom's Burger",
		"Diskon 50%",
		"Paket A+B",
		"Café Latte",
		"抹茶ラテ",
		"Brand® Snack™",
		"a",
		strings.Repeat("a", 255),
	}
	for _, name := range valid {
		assert.True(t, ItemNameRegex.MatchString(name), "expected valid: %q", name)
	}

	invalid := []string{
		"",
		" Leading space",
		"-Dash first",
		"<script>alert(1)</script>",
		"Item <b>bold</b>",
		"A = B",
		"Line\nbreak",
		"Tab\tname",
		strings.Repeat("a", 256),
	}
	for _, name := range invalid {
		assert.False(t, ItemNameRegex.MatchString(name), "expected invalid: %q", name)
	}
}
