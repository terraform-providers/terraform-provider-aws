package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/prometheusservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSAMPWorkspace_basic(t *testing.T) {
	workspaceAlias := acctest.RandomWithPrefix("tf_amp_workspace")
	resourceName := "aws_prometheus_workspace.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAMPWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAWSAMPWorkspaceConfigWithAlias(workspaceAlias),
				Check: resource.ComposeTestCheckFunc(
					testCheckAWSAMPWorkspaceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "alias", workspaceAlias),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAWSAMPWorkspaceConfigWithoutAlias(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "alias", ""),
				),
			},
			{
				Config: testAWSAMPWorkspaceConfigWithAlias(workspaceAlias),
				Check: resource.ComposeTestCheckFunc(
					testCheckAWSAMPWorkspaceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "alias", workspaceAlias),
				),
			},
		},
	})
}

func TestAccAWSAMPWorkspace_disappears(t *testing.T) {
	resourceName := "aws_prometheus_workspace.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAMPWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAWSAMPWorkspaceConfigWithoutAlias(),
				Check: resource.ComposeTestCheckFunc(
					testCheckAWSAMPWorkspaceExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsPrometheusWorkspace(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testCheckAWSAMPWorkspaceExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No AMP Workspace ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).prometheusserviceconn

		req := &prometheusservice.DescribeWorkspaceInput{
			WorkspaceId: aws.String(rs.Primary.ID),
		}
		describe, err := conn.DescribeWorkspace(req)
		if err != nil {
			return err
		}
		if describe == nil {
			return fmt.Errorf("Got nil account ?!")
		}

		return nil
	}
}

func testAccCheckAWSAMPWorkspaceDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).prometheusserviceconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_prometheus_workspace" {
			continue
		}

		_, err := conn.DescribeWorkspace(&prometheusservice.DescribeWorkspaceInput{
			WorkspaceId: aws.String(rs.Primary.ID),
		})
		if isAWSErr(err, prometheusservice.ErrCodeResourceNotFoundException, "") {
			continue
		}

		if err != nil {
			return fmt.Errorf("error reading Prometheus WorkSpace (%s): %w", rs.Primary.ID, err)
		}
	}

	return nil
}

func testAWSAMPWorkspaceConfigWithAlias(randName string) string {
	return fmt.Sprintf(`
resource "aws_prometheus_workspace" "test" {
  alias = %q
}
`, randName)
}

func testAWSAMPWorkspaceConfigWithoutAlias() string {
	return `
resource "aws_prometheus_workspace" "test" {
}
`
}
