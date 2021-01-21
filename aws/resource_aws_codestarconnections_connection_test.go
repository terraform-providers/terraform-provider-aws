package aws

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codestarconnections"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSCodeStarConnectionsConnection_Basic(t *testing.T) {
	var v codestarconnections.Connection
	resourceName := "aws_codestarconnections_connection.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(codestarconnections.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeStarConnectionsConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeStarConnectionsConnectionConfigBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCodeStarConnectionsConnectionExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "id", "codestar-connections", regexp.MustCompile("connection/.+")),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "codestar-connections", regexp.MustCompile("connection/.+")),
					resource.TestCheckResourceAttr(resourceName, "provider_type", codestarconnections.ProviderTypeBitbucket),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "connection_status", codestarconnections.ConnectionStatusPending),
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

func TestAccAWSCodeStarConnectionsConnection_disappears(t *testing.T) {
	var v codestarconnections.Connection
	resourceName := "aws_codestarconnections_connection.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(codestarconnections.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeStarConnectionsConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeStarConnectionsConnectionConfigBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCodeStarConnectionsConnectionExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCodeStarConnectionsConnection(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSCodeStarConnectionsConnectionExists(n string, v *codestarconnections.Connection) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No CodeStar connection ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).codestarconnectionsconn

		resp, err := conn.GetConnection(&codestarconnections.GetConnectionInput{
			ConnectionArn: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		*v = *resp.Connection

		return nil
	}
}

func testAccCheckAWSCodeStarConnectionsConnectionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).codestarconnectionsconn

	for _, rs := range s.RootModule().Resources {
		switch rs.Type {
		case "aws_codestarconnections_connection":
			_, err := conn.GetConnection(&codestarconnections.GetConnectionInput{
				ConnectionArn: aws.String(rs.Primary.ID),
			})

			if err != nil && !isAWSErr(err, codestarconnections.ErrCodeResourceNotFoundException, "") {
				return err
			}
		}
	}

	return nil
}

func testAccAWSCodeStarConnectionsConnectionConfigBasic(rName string) string {
	return fmt.Sprintf(`
resource "aws_codestarconnections_connection" "test" {
  name          = %[1]q
  provider_type = "Bitbucket"
}
`, rName)
}
