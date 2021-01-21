package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/codecommit"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSCodeCommitTrigger_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codecommit_trigger.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCodeCommitTriggerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCodeCommitTrigger_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCodeCommitTriggerExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "trigger.#", "1"),
				),
			},
		},
	})
}

func testAccCheckCodeCommitTriggerDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).codecommitconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_codecommit_trigger" {
			continue
		}

		_, err := conn.GetRepositoryTriggers(&codecommit.GetRepositoryTriggersInput{
			RepositoryName: aws.String(rs.Primary.ID),
		})

		if ae, ok := err.(awserr.Error); ok && ae.Code() == "RepositoryDoesNotExistException" {
			continue
		}
		if err == nil {
			return fmt.Errorf("Trigger still exists: %s", rs.Primary.ID)
		}
		return err
	}

	return nil
}

func testAccCheckCodeCommitTriggerExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		codecommitconn := testAccProvider.Meta().(*AWSClient).codecommitconn
		out, err := codecommitconn.GetRepositoryTriggers(&codecommit.GetRepositoryTriggersInput{
			RepositoryName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if len(out.Triggers) == 0 {
			return fmt.Errorf("CodeCommit Trigger Failed: %q", out)
		}

		return nil
	}
}

func testAccCodeCommitTrigger_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_codecommit_repository" "test" {
  repository_name = %[1]q
}

resource "aws_codecommit_trigger" "test" {
  repository_name = aws_codecommit_repository.test.id

  trigger {
    name            = %[1]q
    events          = ["all"]
    destination_arn = aws_sns_topic.test.arn
  }
}
`, rName)
}
