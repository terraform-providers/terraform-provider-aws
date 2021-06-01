variable "workflow_labels" {
  default = {
    "examples"                       = "fef2c0",
    "hacktoberfest"                  = "2c0fad",
    "linter"                         = "cccccc",
    "needs-triage"                   = "e236d7",
    "terraform-plugin-sdk-migration" = "fad8c7",
    "terraform-plugin-sdk-v1"        = "fad8c7",
    "terraform-0.11"                 = "cccccc",
    "terraform-0.12"                 = "cccccc",    
    "terraform-0.13"                 = "cccccc",
    "terraform-0.14"                 = "cccccc",
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
