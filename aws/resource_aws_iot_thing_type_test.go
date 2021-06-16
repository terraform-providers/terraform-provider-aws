package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_iot_thing_type", &resource.Sweeper{
		Name:         "aws_iot_thing_type",
		F:            testSweepIotThingTypes,
		Dependencies: []string{"aws_iot_thing"},
	})
}

func testSweepIotThingTypes(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*AWSClient).iotconn
	sweepResources := make([]*testSweepResource, 0)
	var errs *multierror.Error

	input := &iot.ListThingTypesInput{}

	err = conn.ListThingTypesPages(input, func(page *iot.ListThingTypesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, thingTypes := range page.ThingTypes {
			r := resourceAwsIotThingType()
			d := r.Data(nil)

			d.SetId(aws.StringValue(thingTypes.ThingTypeName))

			sweepResources = append(sweepResources, NewTestSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing IoT Thing Type for %s: %w", region, err))
	}

	if err := testSweepResourceOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping IoT Thing Type for %s: %w", region, err))
	}

	if testSweepSkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping IoT Thing Type sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func TestAccAWSIotThingType_basic(t *testing.T) {
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, iot.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIotThingTypeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIotThingTypeConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("aws_iot_thing_type.foo", "arn"),
					resource.TestCheckResourceAttr("aws_iot_thing_type.foo", "name", fmt.Sprintf("tf_acc_iot_thing_type_%d", rInt)),
				),
			},
			{
				ResourceName:      "aws_iot_thing_type.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSIotThingType_full(t *testing.T) {
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, iot.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIotThingTypeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIotThingTypeConfig_full(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("aws_iot_thing_type.foo", "arn"),
					resource.TestCheckResourceAttr("aws_iot_thing_type.foo", "properties.0.description", "MyDescription"),
					resource.TestCheckResourceAttr("aws_iot_thing_type.foo", "properties.0.searchable_attributes.#", "3"),
					resource.TestCheckResourceAttr("aws_iot_thing_type.foo", "deprecated", "true"),
				),
			},
			{
				ResourceName:      "aws_iot_thing_type.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSIotThingTypeConfig_fullUpdated(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_iot_thing_type.foo", "deprecated", "false"),
				),
			},
		},
	})
}

func testAccCheckAWSIotThingTypeDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).iotconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iot_thing_type" {
			continue
		}

		params := &iot.DescribeThingTypeInput{
			ThingTypeName: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeThingType(params)
		if err == nil {
			return fmt.Errorf("Expected IoT Thing Type to be destroyed, %s found", rs.Primary.ID)
		}

	}

	return nil
}

func testAccAWSIotThingTypeConfig_basic(rName int) string {
	return fmt.Sprintf(`
resource "aws_iot_thing_type" "foo" {
  name = "tf_acc_iot_thing_type_%d"
}
`, rName)
}

func testAccAWSIotThingTypeConfig_full(rName int) string {
	return fmt.Sprintf(`
resource "aws_iot_thing_type" "foo" {
  name       = "tf_acc_iot_thing_type_%d"
  deprecated = true

  properties {
    description           = "MyDescription"
    searchable_attributes = ["foo", "bar", "baz"]
  }
}
`, rName)
}

func testAccAWSIotThingTypeConfig_fullUpdated(rName int) string {
	return fmt.Sprintf(`
resource "aws_iot_thing_type" "foo" {
  name       = "tf_acc_iot_thing_type_%d"
  deprecated = false

  properties {
    description           = "MyDescription"
    searchable_attributes = ["foo", "bar", "baz"]
  }
}
`, rName)
}
