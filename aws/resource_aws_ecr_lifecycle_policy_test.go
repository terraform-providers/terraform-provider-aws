package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSEcrLifecyclePolicy_basic(t *testing.T) {
	randString := acctest.RandString(10)
	rName := fmt.Sprintf("tf-acc-test-lifecycle-%s", randString)
	resourceName := "aws_ecr_lifecycle_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcrLifecyclePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEcrLifecyclePolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrLifecyclePolicyExists(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAWSEcrLifecyclePolicyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ecrconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ecr_lifecycle_policy" {
			continue
		}

		input := &ecr.GetLifecyclePolicyInput{
			RepositoryName: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetLifecyclePolicy(input)
		if err != nil {
			if isAWSErr(err, ecr.ErrCodeRepositoryNotFoundException, "") {
				return nil
			}
			if isAWSErr(err, ecr.ErrCodeLifecyclePolicyNotFoundException, "") {
				return nil
			}
			return err
		}
	}

	return nil
}

func testAccCheckAWSEcrLifecyclePolicyExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).ecrconn

		input := &ecr.GetLifecyclePolicyInput{
			RepositoryName: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetLifecyclePolicy(input)
		return err
	}
}

func testAccEcrLifecyclePolicyConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository" "test" {
  name = "%s"
}

resource "aws_ecr_lifecycle_policy" "test" {
  repository = aws_ecr_repository.test.name

  policy = <<EOF
{
  "rules": [
    {
      "rulePriority": 1,
      "description": "Expire images older than 14 days",
      "selection": {
        "tagStatus": "untagged",
        "countType": "sinceImagePushed",
        "countUnit": "days",
        "countNumber": 14
      },
      "action": {
        "type": "expire"
      }
    }
  ]
}
EOF
}
`, rName)
}
