package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSDxLag_importBasic(t *testing.T) {
	resourceName := "aws_dx_lag.hoge"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDxLagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxLagConfig(acctest.RandString(5)),
			},

			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func TestAccAWSDxLag_basic(t *testing.T) {
	lagName1 := fmt.Sprintf("tf-dx-lag-%s", acctest.RandString(5))
	lagName2 := fmt.Sprintf("tf-dx-lag-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDxLagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxLagConfig(lagName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxLagExists("aws_dx_lag.hoge"),
					resource.TestCheckResourceAttr("aws_dx_lag.hoge", "name", lagName1),
					resource.TestCheckResourceAttr("aws_dx_lag.hoge", "connections_bandwidth", "1Gbps"),
					resource.TestCheckResourceAttr("aws_dx_lag.hoge", "location", "EqSe2"),
					resource.TestCheckResourceAttr("aws_dx_lag.hoge", "tags.%", "0"),
				),
			},
			{
				Config: testAccDxLagConfig(lagName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxLagExists("aws_dx_lag.hoge"),
					resource.TestCheckResourceAttr("aws_dx_lag.hoge", "name", lagName2),
					resource.TestCheckResourceAttr("aws_dx_lag.hoge", "connections_bandwidth", "1Gbps"),
					resource.TestCheckResourceAttr("aws_dx_lag.hoge", "location", "EqSe2"),
					resource.TestCheckResourceAttr("aws_dx_lag.hoge", "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAWSDxLag_tags(t *testing.T) {
	lagName := fmt.Sprintf("tf-dx-lag-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDxLagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxLagConfig_tags(lagName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxLagExists("aws_dx_lag.hoge"),
					resource.TestCheckResourceAttr("aws_dx_lag.hoge", "name", lagName),
					resource.TestCheckResourceAttr("aws_dx_lag.hoge", "tags.%", "2"),
					resource.TestCheckResourceAttr("aws_dx_lag.hoge", "tags.Usage", "original"),
				),
			},
			{
				Config: testAccDxLagConfig_tagsChanged(lagName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxLagExists("aws_dx_lag.hoge"),
					resource.TestCheckResourceAttr("aws_dx_lag.hoge", "name", lagName),
					resource.TestCheckResourceAttr("aws_dx_lag.hoge", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_dx_lag.hoge", "tags.Usage", "changed"),
				),
			},
			{
				Config: testAccDxLagConfig(lagName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxLagExists("aws_dx_lag.hoge"),
					resource.TestCheckResourceAttr("aws_dx_lag.hoge", "name", lagName),
					resource.TestCheckResourceAttr("aws_dx_lag.hoge", "tags.%", "0"),
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
resource "aws_dx_lag" "hoge" {
  name = "%s"
  connections_bandwidth = "1Gbps"
  location = "EqSe2"
  force_destroy = true
}
`, n)
}

func testAccDxLagConfig_tags(n string) string {
	return fmt.Sprintf(`
resource "aws_dx_lag" "hoge" {
  name = "%s"
  connections_bandwidth = "1Gbps"
  location = "EqSe2"
  force_destroy = true

  tags = {
    Environment = "production"
    Usage = "original"
  }
}
`, n)
}

func testAccDxLagConfig_tagsChanged(n string) string {
	return fmt.Sprintf(`
resource "aws_dx_lag" "hoge" {
  name = "%s"
  connections_bandwidth = "1Gbps"
  location = "EqSe2"
  force_destroy = true

  tags = {
    Usage = "changed"
  }
}
`, n)
}
