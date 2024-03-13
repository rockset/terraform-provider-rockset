# Terraform Provider for Rockset

This terraform provider implements the Rockset API. See the docs folder for which resources have been implemented.

## Setting and securing your Rockset API Key

We encourage you to keep your Rockset api key secure and *never* put it in plain text in terraform provider config or commit it to repositories. 
Using environment variables is the recommended way to configure the provider or run the acceptance tests.

### Securing your key
One potential way to secure your key and avoid committing it in plain text is by 
using another system that requires authentication, such as AWS SSM. 

E.g.
```bash
#!/bin/bash

export ROCKSET_APIKEY=$(aws ssm get-parameter --name '/terraform/rockset_api_key' --with-decryption --output text --query 'Parameter.Value')
export ROCKSET_APISERVER="https://api.usw2a1.rockset.com"
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
  api_server = "https://api.usw2a1.rockset.com"
}
```

## Example Module
A module is included which has an example s3 integration.

You can get the curl command for the created query lambda without string escaping using jq:
```
terraform output -json | jq -r .rockset.value[0].curl_command
```

## Doc generation

We use [terraform-docs](https://github.com/hashicorp/terraform-plugin-docs) for generating docs in the format the
provider repository expects.

The `examples` folder will render examples into resource doc pages.

The `templates` folder can be used to dictate how a page should render. Primarily used for the root page.

`tfplugindocs` will render documentation based on what's in all the implemented resources. You can preview the
documentation using the [terraform doc-preview](https://registry.terraform.io/tools/doc-preview) page

## Testing
Acceptance tests are written for all implemented resources and data sources. They can be run using `go test`. To run acceptance tests the environment variable `TF_ACC` must be set.

Additionally, `ROCKSET_APIKEY` and `ROCKSET_APISERVER` environment variables must be set. We encourage you to keep your api key safe and secure. 

Running acceptance tests creates real resources. Some acceptance tests may use features that require contacting Rockset support to enable for your org.

To run all tests:
```
TF_ACC=true go test -timeout 40m ./rockset/*
```

To run all tests with debug output:
```
TF_ACC=true go test -timeout 40m -v ./rockset/*
```

To run a specific test:
```
TF_ACC=true go test -timeout 40m -v ./rockset/* -run TestAccS3Collection_Basic
```

You may want to run tests with local changes in a dependency, such as in the [Rockset Go Client](https://github.com/rockset/rockset-go-client). (For example, you may be adding a field in both the Rockset Go client and the Terraform Provider) Use the `replace` keyword in the `go.mod` file to use the local version of the dependency instead.
```
module github.com/rockset/terraform-provider-rockset
replace github.com/rockset/rockset-go-client => ../rockset-go-client
```

### Pre-commit hooks

The repo comes with a `.pre-commit-config.yaml` file which can be enabled using `pre-commit install`,
which also runs through Circle-CI before any tests are run. 
It relies on having [`tfplugindocs`](https://github.com/hashicorp/terraform-plugin-docs) installed.

See [pre-commit](https://pre-commit.com/) for more details.

##  Local testing

If you want to test with a dev version of the provider, update your `~/.terraformrc` with

```
provider_installation {
  dev_overrides {
    "rockset/rockset" = "/Users/pme/src/rockset/terraform-provider-rockset"
  }
}
```

and you should see this warning when you run `terraform`

```
│ Warning: Provider development overrides are in effect
│
│ The following provider development overrides are set in the CLI configuration:
│  - rockset/rockset in /Users/pme/src/rockset/terraform-provider-rockset
│
│ The behavior may therefore not match any released version of the provider and applying changes may cause the state to become incompatible with published releases.
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
      source  = "rockset/rockset"
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

## Release Steps
- Merge branch to master
- On master branch create new tag: `git tag v{version_number}`
- Push tag: `git push origin v{version_number}`
Pushing the tag should trigger circle ci to create a draft release and build all necesary assets. Then you can navigate to github and publish the release. 