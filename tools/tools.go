// Description: This file is used to import tools that are used in the project.
// The tools are imported as blank imports, which means that they are imported for their side-effects only.
//
// Use with:
//  go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

//go:build tools

package tools

import (
	// Documentation generation
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"
)
