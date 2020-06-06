variable "workflow_labels" {
  default = {
    "examples"                       = "fef2c0",
    "hacktoberfest"                  = "2c0fad",
    "needs-triage"                   = "e236d7",
    "terraform-plugin-sdk-migration" = "fad8c7",
  }
  description = "Name-color mapping of workflow issues"
  type        = map(string)
}

resource "github_issue_label" "workflow" {
  for_each = var.workflow_labels

  repository = "terraform-provider-aws"
  name       = each.key
  color      = each.value
}
