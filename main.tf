provider "rockset" {}

variable "bucket" {
  default = "rockset-terraform-provider"
}

data "rockset_account" "demo" {}

resource "rockset_workspace" "demo" {
  name = "demo"
}

resource "rockset_s3_integration" "demo" {
  name         = "s3-integration"
  aws_role_arn = "arn:aws:iam::216690786812:role/provider-integration-test"
}

resource "rockset_s3_collection" "demo" {
  workspace        = rockset_workspace.demo.name
  name             = "s3-collection"
  integration_name = rockset_s3_integration.demo.name
  bucket           = var.bucket
  pattern          = "test.csv"
  retention        = 3600

  format = "csv"
  csv {
    first_line_as_column_names = false
    column_names = ["foo", "bar"]
  }

  field_mapping {
    name = "string to float"
    input_fields {
      field_name = "col2"
      if_missing = "SKIP"
      is_drop    = false
      param      = "val"
    }
    input_fields {
      field_name = "col3"
      if_missing = "PASS"
      is_drop    = false
      param      = "val2"
    }

    output_field {
      field_name = "col2_float"
      on_error   = "FAIL"
      sql        = "CAST(:val as float)"
    }
  }

  field_mapping {
    name = "string to bool"
    input_fields {
      field_name = "col3"
      if_missing = "SKIP"
      is_drop    = false
      param      = "val"
    }
    output_field {
      field_name = "col3_bool"
      on_error   = "SKIP"
      sql        = "CAST(:val as bool)"
    }
  }
}
