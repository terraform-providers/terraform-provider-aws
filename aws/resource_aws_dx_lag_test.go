package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_dx_lag", &resource.Sweeper{
		Name:         "aws_dx_lag",
		F:            testSweepDxLags,
		Dependencies: []string{"aws_dx_connection"},
	})
}

func testSweepDxLags(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*AWSClient).dxconn

	var sweeperErrs *multierror.Error

	input := &directconnect.DescribeLagsInput{}

	// DescribeLags has no pagination support
	output, err := conn.DescribeLags(input)

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping Direct Connect LAG sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErr := fmt.Errorf("error listing Direct Connect LAGs for %s: %w", region, err)
		log.Printf("[ERROR] %s", sweeperErr)
		sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
		return sweeperErrs.ErrorOrNil()
	}

	if output == nil {
		log.Printf("[WARN] Skipping Direct Connect LAG sweep for %s: empty response", region)
		return sweeperErrs.ErrorOrNil()
	}

	for _, lag := range output.Lags {
		if lag == nil {
			continue
		}

		id := aws.StringValue(lag.LagId)

		r := resourceAwsDxLag()
		d := r.Data(nil)
		d.SetId(id)

		err = r.Delete(d, client)

		if err != nil {
			sweeperErr := fmt.Errorf("error deleting Direct Connect LAG (%s): %w", id, err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAwsDxLag_basic(t *testing.T) {
	var lag directconnect.Lag
	resourceName := "aws_dx_lag.test"
	rName1 := fmt.Sprintf("tf-testacc-dxlag-%s", acctest.RandString(15))
	rName2 := fmt.Sprintf("tf-testacc-dxlag-%s", acctest.RandString(15))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, directconnect.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDxLagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxLagConfig_basic(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxLagExists(resourceName, &lag),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "directconnect", regexp.MustCompile(`dxlag/.+`)),
					resource.TestCheckResourceAttr(resourceName, "connections_bandwidth", "1Gbps"),
					resource.TestCheckResourceAttr(resourceName, "force_destroy", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "location"),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccDxLagConfig_basic(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxLagExists(resourceName, &lag),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "directconnect", regexp.MustCompile(`dxlag/.+`)),
					resource.TestCheckResourceAttr(resourceName, "connections_bandwidth", "1Gbps"),
					resource.TestCheckResourceAttr(resourceName, "force_destroy", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "location"),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			// Test import.
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func TestAccAwsDxLag_Tags(t *testing.T) {
	var lag directconnect.Lag
	resourceName := "aws_dx_lag.test"
	rName := fmt.Sprintf("tf-testacc-dxlag-%s", acctest.RandString(15))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, directconnect.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDxLagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxLagConfig_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxLagExists(resourceName, &lag),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "directconnect", regexp.MustCompile(`dxlag/.+`)),
					resource.TestCheckResourceAttr(resourceName, "connections_bandwidth", "1Gbps"),
					resource.TestCheckResourceAttr(resourceName, "force_destroy", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "location"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2a"),
				),
			},
			{
				Config: testAccDxLagConfig_tagsUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxLagExists(resourceName, &lag),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "directconnect", regexp.MustCompile(`dxlag/.+`)),
					resource.TestCheckResourceAttr(resourceName, "force_destroy", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "location"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2b"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "Value3"),
				),
			},
			// Test import.
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
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
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		resp, err := conn.DescribeLags(&directconnect.DescribeLagsInput{
			LagId: aws.String(rs.Primary.ID),
		})
		if isAWSErr(err, directconnect.ErrCodeClientException, "does not exist") {
			continue
		}
		if err != nil {
			return err
		}

		for _, v := range resp.Lags {
			if aws.StringValue(v.LagId) == rs.Primary.ID && aws.StringValue(v.LagState) != directconnect.LagStateDeleted {
				return fmt.Errorf("[DESTROY ERROR] Direct Connect LAG (%s) not deleted", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccCheckAwsDxLagExists(name string, lag *directconnect.Lag) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).dxconn

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		resp, err := conn.DescribeLags(&directconnect.DescribeLagsInput{
			LagId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		for _, v := range resp.Lags {
			if aws.StringValue(v.LagId) == rs.Primary.ID {
				*lag = *v

				return nil
			}
		}

		return fmt.Errorf("Direct Connect LAG (%s) not found", rs.Primary.ID)
	}
}

func testAccDxLagConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_dx_locations" "test" {}

resource "aws_dx_lag" "test" {
  name                  = %[1]q
  connections_bandwidth = "1Gbps"
  location              = tolist(data.aws_dx_locations.test.location_codes)[0]
  force_destroy         = false
}
`, rName)
}

func testAccDxLagConfig_tags(rName string) string {
	return fmt.Sprintf(`
data "aws_dx_locations" "test" {}

resource "aws_dx_lag" "test" {
  name                  = %[1]q
  connections_bandwidth = "1Gbps"
  location              = tolist(data.aws_dx_locations.test.location_codes)[0]
  force_destroy         = true

  tags = {
    Name = %[1]q
    Key1 = "Value1"
    Key2 = "Value2a"
  }
}
`, rName)
}

func testAccDxLagConfig_tagsUpdated(rName string) string {
	return fmt.Sprintf(`
data "aws_dx_locations" "test" {}

resource "aws_dx_lag" "test" {
  name                  = %[1]q
  connections_bandwidth = "1Gbps"
  location              = tolist(data.aws_dx_locations.test.location_codes)[0]
  force_destroy         = true

  tags = {
    Name = %[1]q
    Key2 = "Value2b"
    Key3 = "Value3"
  }
}
`, rName)
}
