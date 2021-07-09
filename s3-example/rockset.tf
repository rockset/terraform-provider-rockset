// This gets information about the rockset account
// The account id will vary by region 
// External id will be used to grant an IAM role s3 permissions
data rockset_account example {}

resource rockset_workspace example {
  name = local.bucket_string
}

resource rockset_s3_integration example {
  name         = "${local.bucket_string}-s3-integration"
  aws_role_arn = aws_iam_role.rockset.arn
  depends_on = [
    // Let's give AWS time for the role to be usable.
    // If we try too quickly, we get an IAM error.
    time_sleep.wait_for_role
  ]
}

resource rockset_alias example {
  name = "cities"
  workspace = rockset_workspace.example.name
  collections = [
    "${rockset_s3_collection.example.workspace}.${rockset_s3_collection.example.name}"
  ]
}

resource rockset_s3_collection example {
  workspace        = rockset_workspace.example.name
  // If we need to update the fields of this collection
  // we will bump the v# to a new number to force a new name
  // Collections cannot be updated, only replaced.
  name             = "cities-v1" 
  retention_secs   = 10000

  source {
    integration_name = rockset_s3_integration.example.name
    bucket           = aws_s3_bucket.rockset.bucket
    pattern          = "cities.csv"
    format = "csv"

    csv {
      first_line_as_column_names = false
      column_names               = [
        "country",
        "city",
        "population",
        "visited"
      ]
      column_types = [ 
        "STRING",
        "STRING",
        "STRING",
        "STRING",
      ]
    }
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

  // If we need to update the collection config,
  // we will create the new collection before destroying
  // The alias will only update upon successful creation
  // and only when the new collection is ready.
  lifecycle {
    create_before_destroy = true
  }  
}

data template_file sql {
  template = file("${path.module}/files/query.sql")
  vars = {
    workspace = rockset_s3_collection.example.workspace
    alias = rockset_alias.example.name
  }
}

resource rockset_query_lambda example {
  workspace = rockset_workspace.example.name
  name      = "cities"
  sql {
    query = data.template_file.sql.rendered
    default_parameter {
      name  = "country"
      type  = "string"
      value = "Sweden"
    }
  }
}
