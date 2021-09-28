terraform {
  required_providers {
    rockset = {
      source  = "rockset/rockset"
      version = "~> 0.3.1"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "~> 3.0"
    }
  }
}