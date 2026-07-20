package store

import "os"

func ReadCurrent(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func WriteCurrent(path, themeId string) error {
	err := os.WriteFile(path, []byte(themeId), 0o644)
	return err
}
