package provider

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
)

var (
	rocksetNameRegexp    = regexp.MustCompile(`^[[:alnum:]][[:alnum:]-_]*$`)
	rocksetNameValidator = stringvalidator.RegexMatches(rocksetNameRegexp, "alphanumeric, dash, or underscore characters")
)
