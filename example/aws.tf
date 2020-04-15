// this file contains the required configuration to create an S3 bucket and role
// configured to allow a Rockset integration to be created
// https://docs.rockset.com/amazon-s3
provider "aws" {
  region  = var.region
  version = "~> 2.47"
}

resource "aws_s3_bucket" "rockset" {
  bucket = var.bucket
}

resource "aws_s3_bucket_object" "cities" {
  bucket = aws_s3_bucket.rockset.bucket
  key = var.csv
  source = var.csv
  etag = filemd5(var.csv)
}

resource "aws_iam_policy" "rockset-s3-integration" {
  name   = "terraform-provider-rockset"
  policy = data.aws_iam_policy_document.s3-bucket-policy.json
}

data "aws_iam_policy_document" "s3-bucket-policy" {
  statement {
    sid       = "RocksetS3IntegrationPolicy"
    actions   = [
      "s3:List*",
      "s3:GetObject"
    ]
    resources = [
      aws_s3_bucket.rockset.arn,
      "${aws_s3_bucket.rockset.arn}/*"
    ]
  }
}

resource "aws_iam_role" "rockset" {
  name               = var.role-name
  assume_role_policy = data.aws_iam_policy_document.rockset-trust-policy.json
}

data "aws_iam_policy_document" "rockset-trust-policy" {
  statement {
    sid     = ""
    effect  = "Allow"
    actions = [
      "sts:AssumeRole"]
    principals {
      identifiers = [
        "arn:aws:iam::${data.rockset_account.example.account_id}:root"]
      type        = "AWS"
    }
    condition {
      test     = "StringEquals"
      values   = [
        data.rockset_account.example.external_id]
      variable = "sts:ExternalId"
    }
  }
}

resource "aws_iam_role_policy_attachment" "rockset-s3-integration" {
  role       = aws_iam_role.rockset.name
  policy_arn = aws_iam_policy.rockset-s3-integration.arn
}
