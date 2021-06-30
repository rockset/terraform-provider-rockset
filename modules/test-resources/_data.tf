data "rockset_account" "us_west_2" {}

data "aws_iam_policy_document" "rockset_trust_policy" {
  statement {
    sid    = ""
    effect = "Allow"
    actions = [
      "sts:AssumeRole"
    ]
    principals {
      identifiers = [
      "arn:aws:iam::${data.rockset_account.us_west_2.account_id}:root"]
      type = "AWS"
    }
    condition {
      test = "StringEquals"
      values = [
        data.rockset_account.us_west_2.external_id // From rockset data source
      ]
      variable = "sts:ExternalId"
    }
  }
}