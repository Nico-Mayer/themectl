package integration

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func setTOMLString(config, key, value string) (string, error) {
	re := regexp.MustCompile(`(?m)^(\s*` + regexp.QuoteMeta(key) + `\s*=\s*).*$`)
	if !re.MatchString(config) {
		return "", fmt.Errorf("no `%s =` setting found in config", key)
	}
	repl := `${1}` + strings.ReplaceAll(strconv.Quote(value), "$", "$$")
	return re.ReplaceAllString(config, repl), nil
}
