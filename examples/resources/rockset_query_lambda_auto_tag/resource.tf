resource "rockset_query_lambda" "top-movies" {
  name      = "top_movies"
  workspace = "commons"
  sql {
    query = file("${path.module}/data/top_movies.sql")
  }
}

resource "rockset_query_lambda_auto_tag" "active" {
  template         = "v%s"
  max_tags         = 3
  query_lambda = rockset_query_lambda.top-movies.name
  workspace    = rockset_query_lambda.top-movies.workspace
  version      = rockset_query_lambda.top-movies.version
}
