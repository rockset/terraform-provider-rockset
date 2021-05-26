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

## Setting and securing your Rockset api key

We encourage you to keep your rockset api key secure and never put it in plain text in terraform provider config or commit it to repositories. 
Using environment variables is the recommended way to configure the provider or run the acceptance tests.

### Securing your key
One potential way to secure your key and avoid committing it in plain text is by 
using another system that requires authentication, such as AWS SSM. 

E.g.
```bash
#!/bin/bash

export ROCKSET_APIKEY=$(aws ssm get-parameter --name '/john/rockset_api_key' --with-decryption --output text | awk '{print $7}')
export ROCKSET_APISERVER="api.rs2.usw2.rockset.com"
```

### Environment variables
An empty provider config will read from the environment variables `ROCKSET_APIKEY` and `ROCKSET_APISERVER`.
```terraform
provider rockset {}
```

### Terraform Variables
If you want to explicitly set the api key and/or server using terraform variables, 
you can set those terraform variables using environment variables.

Any terraform variable can be set using `TF_VAR_` prefixing the variable name. E.g. `TF_VAR_ROCKSET_APIKEY`

For the below config you would `export TF_VAR_ROCKSET_APIKEY="your apikey"`
```
provider rockset {
  api_key = "var.ROCKSET_APIKEY"
  api_server = "api.rs2.usw2.rockset.com"
}
```

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
resource rockset_s3_collection example {
  workspace        = rockset_workspace.example.name
  name             = "cities"
  integration_name = rockset_s3_integration.example.name
  bucket           = aws_s3_bucket.rockset.bucket
  pattern          = "cities.csv"
  retention_secs   = 3600

  format = "csv"
  csv {
    first_line_as_column_names = false
    column_names               = [
      "country",
      "city",
      "population",
      "visited"
    ]
    column_types = [ 
      "STRING",
      "STRING",
      "STRING",
      "STRING",
    ]
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
Acceptance tests are written for all implemented resources and data sources. They can be run using `go test`. To run acceptance tests the environment variable `TF_ACC` must be set.

Additionally, `ROCKSET_APIKEY` and `ROCKSET_APISERVER` environment variables must be set. We encourage you to keep your api key safe and secure. 

Running acceptance tests creates real resources. Some acceptance tests may use features that require contacting Rockset support to enable for your org.

To run all tests:
```
TF_ACC=true go test ./rockset/*
```

To run all tests with debug output:
```
TF_ACC=true go test -v ./rockset/*
```

To run a specific test:
```
TF_ACC=true go test -v ./rockset/* -run TestAccS3Collection_Basic
```