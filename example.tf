variable "bucket" {
  type = string
  default = "rockset-terraform-provider"
}

module "rockset" {
  source = "./example"
  bucket = var.bucket
}
