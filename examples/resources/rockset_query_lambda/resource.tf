variable "pinned-version" {
  description = "set this variable to pin the query lambda tag to a specific version"
}

resource "rockset_query_lambda" "top-movies" {
  name      = "top_movies"
  workspace = "commons"
  sql {
    query = file("${path.module}/data/top_movies.sql")
  }
}

resource "rockset_query_lambda_tag" "active" {
  name         = "active"
  query_lambda = rockset_query_lambda.top-movies.name
  workspace    = "commons"
  version      = var.pinned-version == null ? rockset_query_lambda.top-movies.version : var.pinned-version
}
