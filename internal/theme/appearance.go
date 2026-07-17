package theme

import (
	"fmt"
	"strings"

	"github.com/invopop/jsonschema"
)

type Appearance string

const (
	Dark  Appearance = "dark"
	Light Appearance = "light"
)

func ParseAppearance(s string) (Appearance, error) {
	s = strings.TrimSpace(s)

	if strings.EqualFold(s, string(Dark)) {
		return Dark, nil
	} else if strings.EqualFold(s, string(Light)) {
		return Light, nil
	}
	return "", fmt.Errorf("invalid appearance %q: want %q or %q", s, Light, Dark)
}

func (Appearance) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type: "string",
		Enum: []any{string(Dark), string(Light)},
	}
}
