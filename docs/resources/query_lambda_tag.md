---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "rockset_query_lambda_tag Resource - rockset"
subcategory: ""
description: |-
  Manages a Rockset Query Lambda Tag.
---

# rockset_query_lambda_tag (Resource)

Manages a Rockset Query Lambda Tag.

## Example Usage

```terraform
resource "rockset_query_lambda_tag" "active" {
  name         = "active"
  query_lambda = "top-movies"
  workspace    = "commons"
  version      = "b22fb578b8106694"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Unique identifier for the tag. Can contain alphanumeric or dash characters.
- `query_lambda` (String) Unique identifier for the query lambda. Can contain alphanumeric or dash characters.
- `version` (String) Version of the query lambda this tag should point to.
- `workspace` (String) The name of the workspace the query lambda is in.

### Read-Only

- `id` (String) The ID of this resource.
