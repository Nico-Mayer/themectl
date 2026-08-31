package ui

import (
	"charm.land/huh/v2"
)

func Select(title string, options []string) (string, error) {
	var selected string
	sel := newSelect(title, options, &selected)

	km := huh.NewDefaultKeyMap()
	km.Select.Prev.Unbind() // single-component form: no shift+tab
	km.Select.Next.Unbind()

	form := huh.NewForm(huh.NewGroup(sel)).WithKeyMap(km)

	if err := form.Run(); err != nil {
		return "", err
	}
	return selected, nil
}

func newSelect(title string, options []string, selected *string) *huh.Select[string] {
	opts := make([]huh.Option[string], len(options))
	for i, option := range options {
		opts[i] = huh.NewOption(option, option)
	}

	return huh.NewSelect[string]().
		Title(title).
		Options(opts...).
		Height(10).
		Filtering(false).
		Value(selected)
}
