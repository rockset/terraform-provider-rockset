// this file contains the required configuration to create an S3 bucket and role
// configured to allow a Rockset integration to be created
// https://docs.rockset.com/amazon-s3

locals {
  csv_path = "${path.module}/files/cities.csv"
  bucket_string = replace(var.bucket, ".", "-") // DNS compatible to AWS name compatible
}

resource aws_s3_bucket rockset {
  bucket = var.bucket
}

// This uploads the file we've specified to the bucket
resource aws_s3_bucket_object cities {
  bucket = aws_s3_bucket.rockset.bucket
  key = "cities.csv"
  source = local.csv_path
  etag = filemd5(local.csv_path)
}

resource aws_iam_policy rockset_s3_integration {
  // Bucket is univerally unique, policy name will be too
  name   = "${local.bucket_string}-access" 
  policy = data.aws_iam_policy_document.s3_bucket_policy.json
}

data aws_iam_policy_document s3_bucket_policy {
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

resource aws_iam_role rockset {
  // Bucket is univerally unique, role name will be too
  name               = "${local.bucket_string}-role" 
  assume_role_policy = data.aws_iam_policy_document.rockset-trust-policy.json
}

data aws_iam_policy_document rockset-trust-policy {
  statement {
    sid     = ""
    effect  = "Allow"
    actions = [
      "sts:AssumeRole"
    ]
    principals {
      identifiers = [
        "arn:aws:iam::${data.rockset_account.example.account_id}:root"]
      type        = "AWS"
    }
    condition {
      test     = "StringEquals"
      values   = [
        data.rockset_account.example.external_id // From rockset data source
      ] 
      variable = "sts:ExternalId"
    }
  }
}

resource aws_iam_role_policy_attachment rockset_s3_integration {
  role       = aws_iam_role.rockset.name
  policy_arn = aws_iam_policy.rockset_s3_integration.arn
}

// We must give AWS a little time before we try to actually use
// the role that was created.
resource time_sleep wait_for_role {
  depends_on = [
    aws_iam_role_policy_attachment.rockset_s3_integration
  ]

  create_duration = "30s"
}