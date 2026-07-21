package ui

import "charm.land/huh/v2"

func Confirm(title string) (bool, error) {
	var answer bool
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(title).
				Affirmative("Yes").
				Negative("No").
				Value(&answer),
		),
	)
	if err := form.Run(); err != nil {
		return false, err
	}
	return answer, nil
}
