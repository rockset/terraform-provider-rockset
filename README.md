# Terraform Provider for Rockset

This terraform provider implements a subset of the Rockset components available, see below for exact list.

## Installation

As it hasn't been published yet, it needs to be manually installed as follows:

#### 1. Build or Download
Either download the provider from a release or build the provider.

#### 2. Put the provider in the expected path
Untar/zip the provider. The executable should be named `terraform-provider-rockset`

Create the folder structure and move the provider to it.

The path will include the platform (e.g. `linux_amd64`, `windows_amd64`) and the version (e.g. `0.1.0`)
```
mkdir -p ~/.terraform.d/plugins/terraform.rockset.com/rockset/rockset/0.1.0/linux_amd64/
mv terraform-provider-rockset ~/.terraform.d/plugins/terraform.rockset.com/rockset/rockset/0.1.0/linux_amd64/
```

#### 3. Configure Terraform
Because this is not a provider coming from a Terraform repository 
Terraform will always assume it's from the default hashicorp repository.

Every place the provider is used (Both on the consuming end of a module, and within the module itself) 
the requirements must explicitly reference the above path.

E.g.
```
terraform {
  required_providers {
    rockset = {
      source  = "terraform.rockset.com/rockset/rockset"
    }
  }
}
```

The source string is structured:
<Repository URL>/<Org name>/<Provider name>

This is a stub for a future private repository `terraform.rockset.com`

With an org of `rockset`

Since this is the rockset provider, the provider is also named `rockset`.

#### 4. Terraform init
If the above is all done correctly you are now able to run `terraform init`.

It will see the provider in the folder and treat it as if it's already downloaded.

## Data sources

### `rockset_account`

The Rockset account data source allows you to access the account id and external id,
which are needed when creating integrations.

Example usage
```
data "rockset_account" "example" {}
```

## Resources

### `rockset_workspace`

Provides a Rockset [workspace](https://docs.rockset.com/workspaces)

Example usage
```
resource "rockset_workspace" "example" {
  name = "example"
}
```

### `rockset_s3_integration`

Provides an [S3 integration](https://docs.rockset.com/amazon-s3#create-an-s3-integration)

Example usage
```
resource "rockset_s3_integration" "example" {
  name         = "s3-integration"
  aws_role_arn = aws_iam_role.rockset.arn
}
```

### `rockset_s3_collection`

Provides an [S3 collection](https://docs.rockset.com/amazon-s3#create-a-collection)

```
resource "rockset_s3_collection" "demo" {
  workspace        = rockset_workspace.example.name
  name             = "cities"
  integration_name = rockset_s3_integration.example.name
  bucket           = aws_s3_bucket.rockset.bucket
  pattern          = var.csv
  retention        = 3600

  format = "csv"
  csv {
    first_line_as_column_names = false
    column_names               = [
      "country",
      "city",
      "population",
      "visited"]
  }

  field_mapping {
    name = "string to float"
    input_fields {
      field_name = "population"
      if_missing = "SKIP"
      is_drop    = false
      param      = "pop"
    }

    output_field {
      field_name = "pop"
      on_error   = "FAIL"
      sql        = "CAST(:pop as int)"
    }
  }

  field_mapping {
    name = "string to bool"
    input_fields {
      field_name = "visited"
      if_missing = "SKIP"
      is_drop    = false
      param      = "visited"
    }

    output_field {
      field_name = "been there"
      on_error   = "SKIP"
      sql        = "CAST(:visited as bool)"
    }
  }
}
```

### `rockset_collection`

Provides a basic empty collection with no sources. Intended for use with the [write api](https://docs.rockset.com/write-api/).

```
resource "rockset_collection" "demo" {
  workspace        = rockset_workspace.example.name
  name             = "cities"
  description      = "write api collection"
}
```


### `rockset_query_lambda`

```
resource "rockset_query_lambda" "test" {
  workspace = rockset_workspace.example.name
  name      = "test"
  sql {
    query = file("${path.module}/query.sql")
    default_parameter {
      name  = "country"
      type  = "string"
      value = "Sweden"
    }
  }
}
```

## Sample usage

## Importing an existing collection
Collection ids are of the format `<workspace>:<collection name>`

Given:
```
resource "rockset_collection" "demo" {
  workspace        = "foo"
  name             = "demo"
  description      = "write api collection"
}
```
You could import with:
`terraform import rockset_collection.demo foo:demo`

## Importing an existing integration

It is possible to import an existing S3 integration. First add a stub to the `main.tf` file,
e.g. if the integration name is `foobar`
```hcl-terraform
resource "rockset_s3_integration" "foobar" {
}
```

Then run
```bash
$ terraform import rockset_s3_integration.foobar foobar
```

Now you can print the existing state with
```bash
$ terraform show
# rockset_s3_integration.foobar:
resource "rockset_s3_integration" "foobar" {
    aws_role_arn = "<arn>"
    id           = "foobar"
    name         = "foobar"
}
```

And finally update `main.tf` with it.

## Testing

To test, set the required environment variable `ROCKSET_APIKEY` which is available in the
[Rockset Console](https://console.rockset.com/).

```hcl-terraform
variable "bucket" {
  type = string
  default = "rockset-terraform-provider"
}

module "rockset" {
  source = "./example"
  bucket = var.bucket
}
```

After that you can build and run
```
$ go build && terraform init
$ terraform apply
module.rockset.data.rockset_account.example: Refreshing state...
module.rockset.data.aws_iam_policy_document.rockset-trust-policy: Refreshing state...

An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  + create
 <= read (data resources)

Terraform will perform the following actions:

  # module.rockset.data.aws_iam_policy_document.s3-bucket-policy will be read during apply
  # (config refers to values not yet known)
 <= data "aws_iam_policy_document" "s3-bucket-policy"  {
      + id   = (known after apply)
      + json = (known after apply)

      + statement {
          + actions   = [
              + "s3:GetObject",
              + "s3:List*",
            ]
          + resources = [
              + (known after apply),
              + (known after apply),
            ]
          + sid       = "RocksetS3IntegrationPolicy"
        }
    }

  # module.rockset.aws_iam_policy.rockset-s3-integration will be created
  + resource "aws_iam_policy" "rockset-s3-integration" {
      + arn    = (known after apply)
      + id     = (known after apply)
      + name   = "terraform-provider-rockset"
      + path   = "/"
      + policy = (known after apply)
    }

  # module.rockset.aws_iam_role.rockset will be created
  + resource "aws_iam_role" "rockset" {
      + arn                   = (known after apply)
      + assume_role_policy    = jsonencode(
            {
              + Statement = [
                  + {
                      + Action    = "sts:AssumeRole"
                      + Condition = {
                          + StringEquals = {
                              + sts:ExternalId = "4ef12b664c96efcd836edfef7b3e085e908f46052b4dcaec3faac57a5b08048e"
                            }
                        }
                      + Effect    = "Allow"
                      + Principal = {
                          + AWS = "arn:aws:iam::318212636800:root"
                        }
                      + Sid       = ""
                    },
                ]
              + Version   = "2012-10-17"
            }
        )
      + create_date           = (known after apply)
      + force_detach_policies = false
      + id                    = (known after apply)
      + max_session_duration  = 3600
      + name                  = "rockset-integration-role"
      + path                  = "/"
      + unique_id             = (known after apply)
    }

  # module.rockset.aws_iam_role_policy_attachment.rockset-s3-integration will be created
  + resource "aws_iam_role_policy_attachment" "rockset-s3-integration" {
      + id         = (known after apply)
      + policy_arn = (known after apply)
      + role       = "rockset-integration-role"
    }

  # module.rockset.aws_s3_bucket.rockset will be created
  + resource "aws_s3_bucket" "rockset" {
      + acceleration_status         = (known after apply)
      + acl                         = "private"
      + arn                         = (known after apply)
      + bucket                      = "rockset-terraform-provider"
      + bucket_domain_name          = (known after apply)
      + bucket_regional_domain_name = (known after apply)
      + force_destroy               = false
      + hosted_zone_id              = (known after apply)
      + id                          = (known after apply)
      + region                      = (known after apply)
      + request_payer               = (known after apply)
      + website_domain              = (known after apply)
      + website_endpoint            = (known after apply)

      + versioning {
          + enabled    = (known after apply)
          + mfa_delete = (known after apply)
        }
    }

  # module.rockset.aws_s3_bucket_object.cities will be created
  + resource "aws_s3_bucket_object" "cities" {
      + acl                    = "private"
      + bucket                 = "rockset-terraform-provider"
      + content_type           = (known after apply)
      + etag                   = "5d56d40f55acf46f7e6c2c118d706b6d"
      + force_destroy          = false
      + id                     = (known after apply)
      + key                    = "cities.csv"
      + server_side_encryption = (known after apply)
      + source                 = "cities.csv"
      + storage_class          = (known after apply)
      + version_id             = (known after apply)
    }

  # module.rockset.rockset_query_lambda.test will be created
  + resource "rockset_query_lambda" "test" {
      + description = "created by Rockset terraform provider"
      + id          = (known after apply)
      + name        = "test"
      + state       = (known after apply)
      + version     = (known after apply)
      + workspace   = "example"

      + sql {
          + query = <<~EOT
                SELECT *
                FROM example.cities
                WHERE cities.country = :country
            EOT

          + default_parameter {
              + name  = "country"
              + type  = "string"
              + value = "Sweden"
            }
        }
    }

  # module.rockset.rockset_s3_collection.demo will be created
  + resource "rockset_s3_collection" "demo" {
      + bucket           = "rockset-terraform-provider"
      + description      = "created by Rockset terraform provider"
      + format           = "csv"
      + id               = (known after apply)
      + integration_name = "s3-integration"
      + name             = "cities"
      + pattern          = "cities.csv"
      + retention        = 3600
      + workspace        = "example"

      + csv {
          + column_names               = [
              + "country",
              + "city",
              + "population",
              + "visited",
            ]
          + column_types               = []
          + encoding                   = "UTF-8"
          + escape_char                = "\\"
          + first_line_as_column_names = false
          + quote_char                 = "\""
          + separator                  = ","
        }

      + field_mapping {
          + name = "string to float"

          + input_fields {
              + field_name = "population"
              + if_missing = "SKIP"
              + is_drop    = false
              + param      = "pop"
            }

          + output_field {
              + field_name = "pop"
              + on_error   = "FAIL"
              + sql        = "CAST(:pop as int)"
            }
        }
      + field_mapping {
          + name = "string to bool"

          + input_fields {
              + field_name = "visited"
              + if_missing = "SKIP"
              + is_drop    = false
              + param      = "visited"
            }

          + output_field {
              + field_name = "been there"
              + on_error   = "SKIP"
              + sql        = "CAST(:visited as bool)"
            }
        }
    }

  # module.rockset.rockset_s3_integration.example will be created
  + resource "rockset_s3_integration" "example" {
      + aws_role_arn = (known after apply)
      + description  = "created by Rockset terraform provider"
      + id           = (known after apply)
      + name         = "s3-integration"
    }

  # module.rockset.rockset_workspace.example will be created
  + resource "rockset_workspace" "example" {
      + created_by  = (known after apply)
      + description = "created by Rockset terraform provider"
      + id          = (known after apply)
      + name        = "example"
    }

Plan: 9 to add, 0 to change, 0 to destroy.

Do you want to perform these actions?
  Terraform will perform the actions described above.
  Only 'yes' will be accepted to approve.

  Enter a value:
```
