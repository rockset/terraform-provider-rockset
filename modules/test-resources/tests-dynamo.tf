resource "aws_dynamodb_table" "rockset_dynamo_integration" {
  name           = "terraform-provider-rockset-tests"
  read_capacity  = 10
  write_capacity = 10
  hash_key       = "exampleHashKey"

  attribute {
    name = "exampleHashKey"
    type = "S"
  }
}

resource "aws_dynamodb_table_item" "rockset_dynamo_integration" {
  table_name = aws_dynamodb_table.rockset_dynamo_integration.name
  hash_key   = aws_dynamodb_table.rockset_dynamo_integration.hash_key

  item = <<ITEM
{
  "exampleHashKey": {"S": "something"},
  "one": {"N": "11111"},
  "two": {"N": "22222"},
  "three": {"N": "33333"},
  "four": {"N": "44444"}
}
ITEM
}

resource "aws_iam_role" "rockset_dynamo_integration" {
  name               = "terraform-provider-rockset-tests-dynamo"
  assume_role_policy = data.aws_iam_policy_document.rockset_trust_policy.json
}

resource "aws_iam_policy" "rockset_dynamo_integration" {
  name   = "terraform-provider-rockset-tests-dynamo"
  policy = data.aws_iam_policy_document.rockset_dynamo_integration.json
}

data "aws_iam_policy_document" "rockset_dynamo_integration" {
  statement {
    sid = "RocksetDynamoIntegration"
    actions = [
      "dynamodb:Scan",
      "dynamodb:DescribeStream",
      "dynamodb:GetRecords",
      "dynamodb:GetShardIterator",
      "dynamodb:DescribeTable",
      "dynamodb:UpdateTable"
    ]
    resources = [
      "arn:aws:dynamodb:*:*:table/${aws_dynamodb_table.rockset_dynamo_integration.name}",
      "arn:aws:dynamodb:*:*:table/${aws_dynamodb_table.rockset_dynamo_integration.name}/stream/*"
    ]
  }
}

resource "aws_iam_role_policy_attachment" "rockset_dynamo_integration" {
  role       = aws_iam_role.rockset_dynamo_integration.name
  policy_arn = aws_iam_policy.rockset_dynamo_integration.arn
}