---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "rockset_workspace Resource - rockset"
subcategory: ""
description: |-
  Manages a Rockset workspace, which can hold collections, query lambdas and views.
---

# rockset_workspace (Resource)

Manages a Rockset workspace, which can hold collections, query lambdas and views.

## Example Usage

```terraform
resource "rockset_workspace" "demo" {
  name = "demo"
  description = "a workspace for demo collections"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Unique identifier for workspace. Can contain alphanumeric or dash characters.

### Optional

- `description` (String) Text describing the collection.

### Read-Only

- `collection_count` (Number) Number of collections in the workspace.
- `created_at` (String) Created at in ISO-8601.
- `created_by` (String) The user who created the workspace.
- `id` (String) The workspace ID, in the form of the workspace `name`.

## Import

Import is supported using the following syntax:

```shell
terraform import rockset_workspace.demo demo
```
