output "s3_role" {
  value = aws_iam_role.rockset_s3_integration.arn
}

output "dynamo_role" {
  value = aws_iam_role.rockset_dynamo_integration.arn
}

output "kinesis_role" {
  value = aws_iam_role.rockset_kinesis_integration.arn
}

output "kinesis_stream" {
  value = aws_kinesis_stream.rockset_kinesis_integration.*
}