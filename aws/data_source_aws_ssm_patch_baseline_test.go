package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSSsmPatchBaselineDataSource_existingBaseline(t *testing.T) {
	resourceName := "data.aws_ssm_patch_baseline.test_existing"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsSsmPatchBaselineDataSourceConfig_existingBaseline(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "AWS-CentOSDefaultPatchBaseline"),
					resource.TestCheckResourceAttr(resourceName, "description", "Default Patch Baseline for CentOS Provided by AWS."),
				),
			},
		},
	})
}

func TestAccAWSSsmPatchBaselineDataSource_newBaseline(t *testing.T) {
	resourceName := "data.aws_ssm_patch_baseline.test_new"
	rName := acctest.RandomWithPrefix("tf-bl-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMPatchBaselineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsSsmPatchBaselineDataSourceConfig_newBaseline(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "name", "aws_ssm_patch_baseline.test_new", "name"),
					resource.TestCheckResourceAttrPair(resourceName, "description", "aws_ssm_patch_baseline.test_new", "description"),
					resource.TestCheckResourceAttrPair(resourceName, "operating_system", "aws_ssm_patch_baseline.test_new", "operating_system"),
				),
			},
		},
	})
}

// Test against one of the default baselines created by AWS
func testAccCheckAwsSsmPatchBaselineDataSourceConfig_existingBaseline() string {
	return `
data "aws_ssm_patch_baseline" "test_existing" {
  owner            = "AWS"
  name_prefix      = "AWS-"
  operating_system = "CENTOS"
}
`
}

// Create a new baseline and pull it back
func testAccCheckAwsSsmPatchBaselineDataSourceConfig_newBaseline(name string) string {
	return fmt.Sprintf(`
resource "aws_ssm_patch_baseline" "test_new" {
  name             = "%s"
  operating_system = "AMAZON_LINUX_2"
  description      = "Test"

  approval_rule {
    approve_after_days = 5
    patch_filter {
      key    = "CLASSIFICATION"
      values = ["*"]
    }
  }
}

data "aws_ssm_patch_baseline" "test_new" {
  owner            = "Self"
  name_prefix      = aws_ssm_patch_baseline.test_new.name
  operating_system = "AMAZON_LINUX_2"
}
`, name)
}
