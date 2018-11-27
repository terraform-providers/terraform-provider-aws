package aws

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSDataSourceEcrLifecyclePolicyDocument_basic(t *testing.T) {
	// This really ought to be able to be a unit test rather than an
	// acceptance test, but just instantiating the AWS provider requires
	// some AWS API calls, and so this needs valid AWS credentials to work.
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcrLifecyclePolicyDocumentConfig,
				Check: resource.TestCheckResourceAttr(
					"data.aws_ecr_lifecycle_policy_document.test", "json",
					testAccAWSEcrLifecyclePolicyDocumentExpectedJSON,
				),
			},
		},
	})
}

func TestAccAWSDataSourceEcrLifecyclePolicyDocument_multipleRules(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcrLifecyclePolicyDocumentConfig_multipleRules,
				Check: resource.TestCheckResourceAttr(
					"data.aws_ecr_lifecycle_policy_document.test", "json",
					testAccAWSEcrLifecyclePolicyDocumentExpectedJSON_multipleRules,
				),
			},
		},
	})
}

func TestAccAWSDataSourceEcrLifecyclePolicyDocument_noDesc(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcrLifecyclePolicyDocumentConfig_noDesc,
				Check: resource.TestCheckResourceAttr(
					"data.aws_ecr_lifecycle_policy_document.test", "json",
					testAccAWSEcrLifecyclePolicyDocumentExpectedJSON_noDesc,
				),
			},
		},
	})
}

var testAccAWSEcrLifecyclePolicyDocumentConfig = `
data "aws_ecr_lifecycle_policy_document" "test" {
    rule {
      priority        = 1
      description     = "This is a test."
      selection       = {
        tag_status      = "tagged"
        tag_prefix_list = ["prod"]
        count_type      = "imageCountMoreThan"
        count_number    = 100
      }
    }
}
`

var testAccAWSEcrLifecyclePolicyDocumentExpectedJSON = `{
  "rules": [
    {
      "rulePriority": 1,
      "description": "This is a test.",
      "selection": {
        "tagStatus": "tagged",
        "tagPrefixList": [
          "prod"
        ],
        "countType": "imageCountMoreThan",
        "countNumber": 100
      },
      "action": {
        "type": "expire"
      }
    }
  ]
}`

var testAccAWSEcrLifecyclePolicyDocumentConfig_multipleRules = `
data "aws_ecr_lifecycle_policy_document" "test" {
    rule {
      priority        = 1
      description     = "This is a test."
      selection       = {
        tag_status      = "tagged"
        tag_prefix_list = ["prod"]
        count_type      = "imageCountMoreThan"
        count_number    = 100
      }
    }
    rule {
      priority        = 2
      description     = "This is another test."
      selection       = {
        tag_status      = "tagged"
        tag_prefix_list = ["dev"]
        count_type      = "imageCountMoreThan"
        count_number    = 25
      }
    }
}
`

var testAccAWSEcrLifecyclePolicyDocumentExpectedJSON_multipleRules = `{
  "rules": [
    {
      "rulePriority": 1,
      "description": "This is a test.",
      "selection": {
        "tagStatus": "tagged",
        "tagPrefixList": [
          "prod"
        ],
        "countType": "imageCountMoreThan",
        "countNumber": 100
      },
      "action": {
        "type": "expire"
      }
    },
    {
      "rulePriority": 2,
      "description": "This is another test.",
      "selection": {
        "tagStatus": "tagged",
        "tagPrefixList": [
          "dev"
        ],
        "countType": "imageCountMoreThan",
        "countNumber": 25
      },
      "action": {
        "type": "expire"
      }
    }
  ]
}`

var testAccAWSEcrLifecyclePolicyDocumentConfig_noDesc = `
data "aws_ecr_lifecycle_policy_document" "test" {
    rule {
      priority        = 1
      selection       = {
        tag_status      = "tagged"
        tag_prefix_list = ["prod"]
        count_type      = "imageCountMoreThan"
        count_number    = 100
      }
    }
}
`

var testAccAWSEcrLifecyclePolicyDocumentExpectedJSON_noDesc = `{
  "rules": [
    {
      "rulePriority": 1,
      "selection": {
        "tagStatus": "tagged",
        "tagPrefixList": [
          "prod"
        ],
        "countType": "imageCountMoreThan",
        "countNumber": 100
      },
      "action": {
        "type": "expire"
      }
    }
  ]
}`
