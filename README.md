# [BETA] Terraform Provider for Rockset

**This is a work in progress and is considered a beta at this point in time. No official support provided until GA.**

This terraform provider implements the Rockset API. See the docs folder for what has been implemented so far.

## Setting and securing your Rockset api key

We encourage you to keep your rockset api key secure and *never put it in plain text in terraform provider config or commit it to repositories*. 
Using environment variables is the recommended way to configure the provider or run the acceptance tests.

### Securing your key
One potential way to secure your key and avoid committing it in plain text is by 
using another system that requires authentication, such as AWS SSM. 

E.g.
```bash
#!/bin/bash

export ROCKSET_APIKEY=$(aws ssm get-parameter --name '/terraform/rockset_api_key' --with-decryption --output text | awk '{print $7}')
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

## Example Module
A module is included which has an example s3 integration.

You can get the curl command for the created query lambda without string escaping using jq:
```
terraform output -json | jq -r .rockset.value[0].curl_command
```

## Doc generation

We use [terraform-docs](https://github.com/terraform-docs/terraform-docs) for generating docs in the format the
provider repository expects.

The `examples` folder will render examples into resource doc pages.

The `templates` folder can be used to dictate how a page should render. Primarily used for the root page.

`terraform-docs` will render documentation based on what's in all the implemented resources.

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

## Manual Installation

To manually install this provider from a build or release 
follow these steps to place the executable in the folder Terraform expects.

This should not be necessary for general use of the provider as you can fetch it from the repository.
But this can be useful for local testing as you're building new features of the provider.

### Move the provider

Untar/zip the provider. The executable should be named `terraform-provider-rockset`. Rename it to match if necessary.

Create the folder structure and move the provider to it.

The path will include the platform (e.g. `linux_amd64`, `windows_amd64`) and the version (e.g. `0.1.0`)
```
mkdir -p ~/.terraform.d/plugins/terraform.rockset.com/rockset/rockset/0.1.0/linux_amd64/
mv terraform-provider-rockset ~/.terraform.d/plugins/terraform.rockset.com/rockset/rockset/0.1.0/linux_amd64/
```

### Configure Terraform
Terraform will always assume unknown providers are from the default hashicorp repository.

Every place the provider is used (Both on the consuming end of a module, and within the module itself) 
the requirements must explicitly reference the path.

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

This is a stub for a hypothetical repository `terraform.rockset.com`

With an org of `rockset`

Since this is the rockset provider, the provider is also named `rockset`.

If the above is all done correctly you are now able to run `terraform init`.

It will see the provider in the folder and treat it as if it's already downloaded.
