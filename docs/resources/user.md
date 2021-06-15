---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "rockset_user Resource - terraform-provider-rockset"
subcategory: ""
description: |-
  Manages a Rockset User.
---

# rockset_user (Resource)

Manages a Rockset User.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- **email** (String) Email address of the user. Also used to identify the user.
- **roles** (List of String) List of roles for the user. E.g. 'admin', 'member', 'read-only'.

### Optional

- **id** (String) The ID of this resource.

