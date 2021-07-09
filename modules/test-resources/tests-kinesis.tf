resource "aws_kinesis_stream" "rockset_kinesis_integration" {
  name             = "terraform-provider-rockset-tests-kinesis"
  shard_count      = 1
  retention_period = 48
}


resource "aws_iam_role" "rockset_kinesis_integration" {
  name               = "terraform-provider-rockset-tests-kinesis"
  assume_role_policy = data.aws_iam_policy_document.rockset_trust_policy.json
}

resource "aws_iam_policy" "rockset_kinesis_integration" {
  name   = "terraform-provider-rockset-tests-kinesis"
  policy = data.aws_iam_policy_document.rockset_kinesis_integration.json
}

data "aws_iam_policy_document" "rockset_kinesis_integration" {
  statement {
    sid = "RocksetKinesisIntegration"
    actions = [
      "kinesis:ListShards",
      "kinesis:DescribeStream",
      "kinesis:GetRecords",
      "kinesis:GetShardIterator"
    ]
    resources = [
      aws_kinesis_stream.rockset_kinesis_integration.arn
    ]
  }
}

resource "aws_iam_role_policy_attachment" "rockset_kinesis_integration" {
  role       = aws_iam_role.rockset_kinesis_integration.name
  policy_arn = aws_iam_policy.rockset_kinesis_integration.arn
}