package validators

import (
	"fmt"
	"regexp"
)

var nameRegexp = regexp.MustCompile("^[[:alnum:]][[:alnum:]-_]*$")

func NameValidator(val interface{}, key string) ([]string, []error) {
	s := val.(string)
	if nameRegexp.MatchString(s) {
		return nil, nil
	}
	return nil, []error{fmt.Errorf("%s must start with alphanumeric, the rest can be alphanumeric, -, or _", key)}
}
