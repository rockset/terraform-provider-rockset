terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 3.47.0"
    }

    rockset = {
      source  = "rockset/rockset"
      version = ">= 0.1.1"
    }
  }
}