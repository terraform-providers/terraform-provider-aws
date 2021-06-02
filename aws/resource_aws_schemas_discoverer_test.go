package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/schemas"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/schemas/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func init() {
	resource.AddTestSweepers("aws_schemas_discoverer", &resource.Sweeper{
		Name: "aws_schemas_discoverer",
		F:    testSweepSchemasDiscoverers,
	})
}

func testSweepSchemasDiscoverers(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.(*AWSClient).schemasconn
	input := &schemas.ListDiscoverersInput{}
	var sweeperErrs *multierror.Error

	err = conn.ListDiscoverersPages(input, func(page *schemas.ListDiscoverersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, discoverer := range page.Discoverers {
			r := resourceAwsSchemasDiscoverer()
			d := r.Data(nil)
			d.SetId(aws.StringValue(discoverer.DiscovererId))
			err = r.Delete(d, client)

			if err != nil {
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping EventBridge Schemas Discoverer sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EventBridge Schemas Discoverers: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSSchemasDiscoverer_basic(t *testing.T) {
	var v schemas.DescribeDiscovererOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_schemas_discoverer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(schemas.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, schemas.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSchemasDiscovererDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSchemasDiscovererConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasDiscovererExists(resourceName, &v),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "schemas", fmt.Sprintf("discoverer/events-event-bus-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccAWSSchemasDiscoverer_disappears(t *testing.T) {
	var v schemas.DescribeDiscovererOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_schemas_discoverer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(schemas.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, schemas.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSchemasDiscovererDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSchemasDiscovererConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasDiscovererExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSchemasDiscoverer(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSSchemasDiscoverer_Description(t *testing.T) {
	var v schemas.DescribeDiscovererOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_schemas_discoverer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(schemas.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, schemas.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSchemasDiscovererDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSchemasDiscovererConfigDescription(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasDiscovererExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSchemasDiscovererConfigDescription(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasDiscovererExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
				),
			},
			{
				Config: testAccAWSSchemasDiscovererConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasDiscovererExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
				),
			},
		},
	})
}

func TestAccAWSSchemasDiscoverer_Tags(t *testing.T) {
	var v schemas.DescribeDiscovererOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_schemas_discoverer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(schemas.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, schemas.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSchemasDiscovererDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSchemasDiscovererConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasDiscovererExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSchemasDiscovererConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasDiscovererExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSSchemasDiscovererConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasDiscovererExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAWSSchemasDiscovererDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).schemasconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_schemas_discoverer" {
			continue
		}

		_, err := finder.DiscovererByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("EventBridge Schemas Discoverer %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckSchemasDiscovererExists(n string, v *schemas.DescribeDiscovererOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EventBridge Schemas Discoverer ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).schemasconn

		output, err := finder.DiscovererByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccAWSSchemasDiscovererConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name = %[1]q
}

resource "aws_schemas_discoverer" "test" {
  source_arn = aws_cloudwatch_event_bus.test.arn
}
`, rName)
}

func testAccAWSSchemasDiscovererConfigDescription(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name = %[1]q
}

resource "aws_schemas_discoverer" "test" {
  source_arn = aws_cloudwatch_event_bus.test.arn

  description = %[2]q
}
`, rName, description)
}

func testAccAWSSchemasDiscovererConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name = %[1]q
}

resource "aws_schemas_discoverer" "test" {
  source_arn = aws_cloudwatch_event_bus.test.arn

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSSchemasDiscovererConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name = %[1]q
}

resource "aws_schemas_discoverer" "test" {
  source_arn = aws_cloudwatch_event_bus.test.arn

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
