# Uncomment the resources in this file to create the IAM resources required by the AWS WorkSpaces service.
# Also uncomment the `depends_on` meta-arguments on `aws_workspaces_directory.example` and `aws_workspaces_workspace.example`.
#
# resource "aws_iam_role" "workspaces-default" {
#   name               = "workspaces_DefaultRole"
#   assume_role_policy = "${data.aws_iam_policy_document.workspaces.json}"
# }
#
# data "aws_iam_policy_document" "workspaces" {
#   statement {
#     actions = ["sts:AssumeRole"]
#
#     principals {
#       type        = "Service"
#       identifiers = ["workspaces.amazonaws.com"]
#     }
#   }
# }
#
# resource "aws_iam_role_policy_attachment" "workspaces-default-service-access" {
#   role       = "${aws_iam_role.workspaces-default.name}"
#   policy_arn = "arn:aws:iam::aws:policy/AmazonWorkSpacesServiceAccess"
# }
#
# resource "aws_iam_role_policy_attachment" "workspaces-default-self-service-access" {
#   role       = "${aws_iam_role.workspaces-default.name}"
#   policy_arn = "arn:aws:iam::aws:policy/AmazonWorkSpacesSelfServiceAccess"
# }
