# Terraform Provider for Rockset

To test, set the required environment variable `ROCKSET_APIKEY` and
edit `main.tf` and add your AWS external ID. Both of them are available in the
[Rockset Console](https://console.rockset.com/).

After that you can run:
```
$ go build && terraform init
$ terraform apply
data.rockset_account.demo: Refreshing state...
module.rockset.data.aws_iam_policy_document.rockset-trust-policy: Refreshing state...

An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  + create
 <= read (data resources)

Terraform will perform the following actions:

  # rockset_s3.demo will be created
  + resource "rockset_s3" "demo" {
      + aws_role_arn = (known after apply)
      + description  = "created by Rockset terraform provider"
      + id           = (known after apply)
      + name         = "provider integration"
    }

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
      + name   = "rockset-s3-integration"
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
                              + sts:ExternalId = "<external id>"
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
      + name                  = "provider-integration-test"
      + path                  = "/"
      + unique_id             = (known after apply)
    }

  # module.rockset.aws_iam_role_policy_attachment.rockset-s3-integration will be created
  + resource "aws_iam_role_policy_attachment" "rockset-s3-integration" {
      + id         = (known after apply)
      + policy_arn = (known after apply)
      + role       = "provider-integration-test"
    }

  # module.rockset.aws_s3_bucket.rockset will be created
  + resource "aws_s3_bucket" "rockset" {
      + acceleration_status         = (known after apply)
      + acl                         = "private"
      + arn                         = (known after apply)
      + bucket                      = "rockset-pme-provider-test-bucket"
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

Plan: 5 to add, 0 to change, 0 to destroy.

Do you want to perform these actions?
  Terraform will perform the actions described above.
  Only 'yes' will be accepted to approve.

  Enter a value:
```