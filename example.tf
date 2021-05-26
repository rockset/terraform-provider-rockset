terraform {
  required_providers {
    rockset = {
      // Must be in both the module and the consuming side's requirements
      source  = "terraform.rockset.com/rockset/rockset"
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

// Buckets are univerally unique. 
// We generate univerally random name so anyone running this will succeed.
resource random_uuid random_bucket_name {}

// This contains everything needed on the AWS side (buckets, policies) 
// and Rockset side (collections, lambdas, etc)
// Peruse the `example` folder to see the resources used.
module rockset {
  source = "./example"
  bucket = "rockset-${random_uuid.random_bucket_name.result}"
}