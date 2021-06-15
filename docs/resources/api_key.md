---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "rockset_api_key Resource - terraform-provider-rockset"
subcategory: ""
description: |-
  Manage a Rockset Api Key.
---

# rockset_api_key (Resource)

Manage a Rockset Api Key.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- **name** (String) Name of the api key.

### Optional

- **id** (String) The ID of this resource.
- **user** (String) User to create the key for. If not set, defaults to authenticated user.

### Read-Only

- **key** (String, Sensitive)

