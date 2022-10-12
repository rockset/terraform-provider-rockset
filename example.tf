terraform {
  required_providers {
    rockset = {
      source  = "rockset/rockset"
      version = "~> 0.6"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "~> 3.0"
    }
  }
}

provider aws { region  = "us-west-2" }

// Use env vars ROCKSET_APIKEY and ROCKSET_APISERVER to configure provider.
provider rockset {}

// Buckets are universally unique.
// We generate universally random name so anyone running this will succeed.
resource random_uuid random_bucket_name {}

// This contains everything needed on the AWS side (buckets, policies) 
// and Rockset side (collections, lambdas, etc)
// Peruse the `example` folder to see the resources used.
module rockset {
  source = "./s3-example"
  bucket = "rockset-${random_uuid.random_bucket_name.result}"
}

output rockset {
  value = module.rockset.*
}