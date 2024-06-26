---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "rockset_user Data Source - rockset"
subcategory: ""
description: |-
  This data source can be used to fetch information about a specific user.
---

# rockset_user (Data Source)

This data source can be used to fetch information about a specific user.

## Example Usage

```terraform
data rockset_user pme {
  email = "pme@rockset.com"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `email` (String) User email. If absent or blank, it gets the current user.

### Read-Only

- `first_name` (String) User's first name.
- `id` (String) The user ID, in the form of the `email`.
- `last_name` (String) User's last name.
- `roles` (List of String) List of roles for the user. E.g. 'admin', 'member', 'read-only'.
- `state` (String) State of the user, either NEW or ACTIVE.
