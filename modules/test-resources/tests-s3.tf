locals {
  csv_path = "${path.module}/files/cities.csv"
  xml_path = "${path.module}/files/note.xml"
}

resource "aws_s3_bucket" "provider_tests" {
  bucket = var.s3_bucket_name
}

resource "aws_s3_bucket_public_access_block" "provider_tests" {
  bucket = aws_s3_bucket.provider_tests.id

  block_public_acls   = true
  block_public_policy = true
}

resource "aws_s3_bucket_object" "csv" {
  bucket = aws_s3_bucket.provider_tests.bucket
  key    = "cities.csv"
  source = local.csv_path
  etag   = filemd5(local.csv_path)
}

resource "aws_s3_bucket_object" "xml" {
  bucket = aws_s3_bucket.provider_tests.bucket
  key    = "note.xml"
  source = local.xml_path
  etag   = filemd5(local.xml_path)
}

resource "aws_iam_role" "rockset_s3_integration" {
  name               = "terraform-provider-rockset-tests"
  assume_role_policy = data.aws_iam_policy_document.rockset_trust_policy.json
}

resource "aws_iam_policy" "rockset_s3_integration" {
  name   = "terraform-provider-rockset-tests"
  policy = data.aws_iam_policy_document.rockset_s3_integration.json
}

data "aws_iam_policy_document" "rockset_s3_integration" {
  statement {
    sid = "RocksetS3Integration"
    actions = [
      "s3:List*",
      "s3:GetObject"
    ]
    resources = [
      aws_s3_bucket.provider_tests.arn,
      "${aws_s3_bucket.provider_tests.arn}/*"
    ]
  }
}

resource "aws_iam_role_policy_attachment" "rockset_s3_integration" {
  role       = aws_iam_role.rockset_s3_integration.name
  policy_arn = aws_iam_policy.rockset_s3_integration.arn
}