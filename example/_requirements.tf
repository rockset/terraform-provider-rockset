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