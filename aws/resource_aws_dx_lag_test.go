package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSDxLag_basic(t *testing.T) {
	lagName1 := fmt.Sprintf("tf-dx-lag-%s", acctest.RandString(5))
	lagName2 := fmt.Sprintf("tf-dx-lag-%s", acctest.RandString(5))
	resourceName := "aws_dx_lag.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDxLagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxLagConfig(lagName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxLagExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", lagName1),
					resource.TestCheckResourceAttr(resourceName, "connections_bandwidth", "1Gbps"),
					resource.TestCheckResourceAttr(resourceName, "location", "EqSe2-EQ"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			{
				Config: testAccDxLagConfig(lagName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxLagExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", lagName2),
					resource.TestCheckResourceAttr(resourceName, "connections_bandwidth", "1Gbps"),
					resource.TestCheckResourceAttr(resourceName, "location", "EqSe2-EQ"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAWSDxLag_tags(t *testing.T) {
	lagName := fmt.Sprintf("tf-dx-lag-%s", acctest.RandString(5))
	resourceName := "aws_dx_lag.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDxLagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxLagConfig_tags(lagName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxLagExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", lagName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Usage", "original"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			{
				Config: testAccDxLagConfig_tagsChanged(lagName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxLagExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", lagName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Usage", "changed"),
				),
			},
			{
				Config: testAccDxLagConfig(lagName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxLagExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", lagName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func testAccCheckAwsDxLagDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).dxconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dx_lag" {
			continue
		}

		input := &directconnect.DescribeLagsInput{
			LagId: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeLags(input)
		if err != nil {
			return err
		}
		for _, v := range resp.Lags {
			if *v.LagId == rs.Primary.ID && !(*v.LagState == directconnect.LagStateDeleted) {
				return fmt.Errorf("[DESTROY ERROR] Dx Lag (%s) found", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccCheckAwsDxLagExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccDxLagConfig(n string) string {
	return fmt.Sprintf(`
resource "aws_dx_lag" "test" {
  name                  = "%s"
  connections_bandwidth = "1Gbps"
  location              = "EqSe2-EQ"
  force_destroy         = true
}
`, n)
}

func testAccDxLagConfig_tags(n string) string {
	return fmt.Sprintf(`
resource "aws_dx_lag" "test" {
  name                  = "%s"
  connections_bandwidth = "1Gbps"
  location              = "EqSe2-EQ"
  force_destroy         = true

  tags = {
    Environment = "production"
    Usage       = "original"
  }
}
`, n)
}

func testAccDxLagConfig_tagsChanged(n string) string {
	return fmt.Sprintf(`
resource "aws_dx_lag" "test" {
  name                  = "%s"
  connections_bandwidth = "1Gbps"
  location              = "EqSe2-EQ"
  force_destroy         = true

  tags = {
    Usage = "changed"
  }
}
`, n)
}
