variable "partition_labels" {
  default = [
    "aws-cn",
    "aws-iso",
    "aws-iso-b",
    "aws-us-gov",
  ]
  description = "Set of AWS Partition labels"
  type        = set(string)
}

resource "github_issue_label" "partition" {
  for_each = var.partition_labels

  repository = "terraform-provider-aws"
  name       = "partition/${each.value}"
  color      = "bfd4f2"
}
