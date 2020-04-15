# Terraform Provider for Rockset

This terraform provider implements a subset of the Rockset components available.

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

## Importing an existing integration

It is possible to import an existing S3 integration. First add a stub to the `main.tf` file,
e.g. if the integration name is `foobar`
```hcl-terraform
resource "rockset_s3" "foobar" {
}
```

Then run
```bash
$ terraform import rockset_s3.foobar foobar
```

Now you can print the existing state with
```bash
$ terraform show
# rockset_s3.foobar:
resource "rockset_s3" "foobar" {
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
