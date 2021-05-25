resource rockset_s3_integration test {
  name = "terraform-provider-acceptance-tests-s3-csv"
  description = "Terraform provider acceptance tests."
  aws_role_arn = "arn:aws:iam::469279130686:role/terraform-provider-rockset-tests"
}

resource rockset_s3_collection test {
  name              = "terraform-provider-acceptance-tests-s3-csv"
  workspace         = "commons"
  description       = "Terraform provider acceptance tests."
  retention_secs    = 3600
  integration_name  = rockset_s3_integration.test.name
  bucket            = "terraform-provider-rockset-tests"
  pattern           = "cities.csv"
  format            = "csv"
  csv {
    first_line_as_column_names = false
    column_names = [
      "country",
      "city",
      "population",
      "visited"
    ]
    column_types = [
      "STRING",
      "STRING",
      "STRING",
      "STRING"
    ]
  }
  field_mapping {
    name = "string to float"
    input_fields {
      field_name  = "population"
      if_missing  = "SKIP"
      is_drop      = false
      param        = "pop"
    }

    output_field {
      field_name = "pop"
      on_error   = "FAIL"
      sql        = "CAST(:pop as int)"
    }
  }
  field_mapping {
    name = "string to bool"
    input_fields {
      field_name = "visited"
      if_missing = "SKIP"
      is_drop    = false
      param      = "visited"
    }

    output_field {
      field_name = "been there"
      on_error   = "SKIP"
      sql        = "CAST(:visited as bool)"
    }
  }
}