variable "bucket" {
  type = string
}

variable "region" {
  type    = string
  default = "us-west-2"
}

variable "role-name" {
  type = string
  default = "rockset-integration-role"
}

variable "csv" {
  type    = string
  default = "cities.csv"
}

variable "retention" {
  type        = number
  description = "collection retention in seconds"
  default     = 3600
}