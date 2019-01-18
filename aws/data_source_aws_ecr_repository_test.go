package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSEcrDataSource_ecrRepository(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "data.aws_ecr_repository.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsEcrRepositoryDataSourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "arn", regexp.MustCompile(`^arn:aws:ecr:[a-zA-Z]+-[a-zA-Z]+-\d+:\d+:repository/tf-acc-test-\d+$`)),
					resource.TestCheckResourceAttrSet(resourceName, "registry_id"),
					resource.TestMatchResourceAttr(resourceName, "repository_url", regexp.MustCompile(`^\d+\.dkr\.ecr\.[a-zA-Z]+-[a-zA-Z]+-\d+\.amazonaws\.com/tf-acc-test-\d+$`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Usage", "original"),
				),
			},
		},
	})
}

func testAccCheckAwsEcrRepositoryDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository" "default" {
  name = %q

  tags = {
    Environment = "production"
    Usage = "original"
  }
}

data "aws_ecr_repository" "default" {
  name = "${aws_ecr_repository.default.name}"
}
`, rName)
}
