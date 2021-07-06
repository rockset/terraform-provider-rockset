variable GCS_SERVICE_ACCOUNT_KEY {
  type = string
}

resource rockset_gcs_integration test {
  name = "terraform-provider-acceptance-tests-gcs-collection"
  description = "Terraform provider acceptance tests."
  service_account_key = var.GCS_SERVICE_ACCOUNT_KEY
}

resource rockset_gcs_collection test {
  name = "terraform-provider-acceptance-tests-gcs"
  workspace = "commons"
  description = "Terraform provider acceptance tests."
  retention_secs = 3600

  source {
    integration_name = rockset_gcs_integration.test.name
    bucket = "rockset-tf-test"
    prefix = "cities.csv"
    format = "csv"
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
  }

  field_mapping {
    name = "string to float"
    input_fields {
      field_name = "population"
      if_missing = "SKIP"
      is_drop = false
      param = "pop"
    }

    output_field {
      field_name = "pop"
      on_error = "FAIL"
      sql = "CAST(:pop as int)"
    }
  }
  field_mapping {
    name = "string to bool"
    input_fields {
      field_name = "visited"
      if_missing = "SKIP"
      is_drop = false
      param = "visited"
    }

    output_field {
      field_name = "been there"
      on_error = "SKIP"
      sql = "CAST(:visited as bool)"
    }
  }
}