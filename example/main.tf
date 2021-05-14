provider rockset {}

data rockset_account example {}

resource rockset_workspace example {
  name = "example"
}

resource rockset_s3_integration example {
  name         = "s3-integration"
  aws_role_arn = aws_iam_role.rockset.arn
}

resource rockset_s3_collection demo {
  workspace        = rockset_workspace.example.name
  name             = "cities"
  integration_name = rockset_s3_integration.example.name
  bucket           = aws_s3_bucket.rockset.bucket
  pattern          = var.csv
  retention        = 3600

  format = "csv"
  csv {
    first_line_as_column_names = false
    column_names               = [
      "country",
      "city",
      "population",
      "visited"]
  }

  field_mapping {
    name = "string to float"
    input_fields {
      field_name = "population"
      if_missing = "SKIP"
      is_drop    = false
      param      = "pop"
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

resource rockset_query_lambda test {
  workspace = rockset_workspace.example.name
  name      = "test"
  sql {
    query = file("${path.module}/query.sql")
    default_parameter {
      name  = "country"
      type  = "string"
      value = "Sweden"
    }
  }
}
