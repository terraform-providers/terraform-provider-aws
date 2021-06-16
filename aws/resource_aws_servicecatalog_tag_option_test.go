package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// add sweeper to delete known test servicecat tag options
func init() {
	resource.AddTestSweepers("aws_servicecatalog_tag_option", &resource.Sweeper{
		Name:         "aws_servicecatalog_tag_option",
		Dependencies: []string{},
		F:            testSweepServiceCatalogTagOptions,
	})
}

func testSweepServiceCatalogTagOptions(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*AWSClient).scconn
	sweepResources := make([]*testSweepResource, 0)
	var errs *multierror.Error

	input := &servicecatalog.ListTagOptionsInput{}

	err = conn.ListTagOptionsPages(input, func(page *servicecatalog.ListTagOptionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, tod := range page.TagOptionDetails {
			if tod == nil {
				continue
			}

			id := aws.StringValue(tod.Id)

			r := resourceAwsServiceCatalogTagOption()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, NewTestSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error describing Service Catalog Tag Options for %s: %w", region, err))
	}

	if err = testSweepResourceOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Service Catalog Tag Options for %s: %w", region, err))
	}

	if testSweepSkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Service Catalog Tag Options sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func TestAccAWSServiceCatalogTagOption_basic(t *testing.T) {
	resourceName := "aws_servicecatalog_tag_option.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, servicecatalog.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogTagOptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogTagOptionConfig_basic(rName, "värde", "active = true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogTagOptionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "active", "true"),
					resource.TestCheckResourceAttr(resourceName, "key", rName),
					resource.TestCheckResourceAttrSet(resourceName, "owner"),
					resource.TestCheckResourceAttr(resourceName, "value", "värde"),
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

func TestAccAWSServiceCatalogTagOption_disappears(t *testing.T) {
	resourceName := "aws_servicecatalog_tag_option.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, servicecatalog.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogTagOptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogTagOptionConfig_basic(rName, "värde", "active = true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogTagOptionExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsServiceCatalogTagOption(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSServiceCatalogTagOption_update(t *testing.T) {
	resourceName := "aws_servicecatalog_tag_option.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	rName2 := acctest.RandomWithPrefix("tf-acc-test")

	// UpdateTagOption() is very particular about what it receives. Only fields that change should
	// be included or it will throw servicecatalog.ErrCodeDuplicateResourceException, "already exists"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, servicecatalog.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogTagOptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogTagOptionConfig_basic(rName, "värde ett", ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "active", "true"),
					resource.TestCheckResourceAttr(resourceName, "key", rName),
					resource.TestCheckResourceAttrSet(resourceName, "owner"),
					resource.TestCheckResourceAttr(resourceName, "value", "värde ett"),
				),
			},
			{
				Config: testAccAWSServiceCatalogTagOptionConfig_basic(rName, "värde två", "active = true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "active", "true"),
					resource.TestCheckResourceAttr(resourceName, "key", rName),
					resource.TestCheckResourceAttrSet(resourceName, "owner"),
					resource.TestCheckResourceAttr(resourceName, "value", "värde två"),
				),
			},
			{
				Config: testAccAWSServiceCatalogTagOptionConfig_basic(rName, "värde två", "active = false"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "active", "false"),
					resource.TestCheckResourceAttr(resourceName, "key", rName), // cannot be updated in place
					resource.TestCheckResourceAttrSet(resourceName, "owner"),
					resource.TestCheckResourceAttr(resourceName, "value", "värde två"),
				),
			},
			{
				Config: testAccAWSServiceCatalogTagOptionConfig_basic(rName, "värde två", "active = true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "active", "true"),
					resource.TestCheckResourceAttr(resourceName, "key", rName), // cannot be updated in place
					resource.TestCheckResourceAttrSet(resourceName, "owner"),
					resource.TestCheckResourceAttr(resourceName, "value", "värde två"),
				),
			},
			{
				Config: testAccAWSServiceCatalogTagOptionConfig_basic(rName2, "värde ett", "active = true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "active", "true"),
					resource.TestCheckResourceAttr(resourceName, "key", rName2),
					resource.TestCheckResourceAttrSet(resourceName, "owner"),
					resource.TestCheckResourceAttr(resourceName, "value", "värde ett"),
				),
			},
		},
	})
}

func TestAccAWSServiceCatalogTagOption_notActive(t *testing.T) {
	resourceName := "aws_servicecatalog_tag_option.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, servicecatalog.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogTagOptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogTagOptionConfig_basic(rName, "värde ett", "active = false"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "active", "false"),
					resource.TestCheckResourceAttr(resourceName, "key", rName),
					resource.TestCheckResourceAttrSet(resourceName, "owner"),
					resource.TestCheckResourceAttr(resourceName, "value", "värde ett"),
				),
			},
		},
	})
}

func testAccCheckAwsServiceCatalogTagOptionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).scconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_servicecatalog_tag_option" {
			continue
		}

		input := &servicecatalog.DescribeTagOptionInput{
			Id: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeTagOption(input)

		if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error getting Service Catalog Tag Option (%s): %w", rs.Primary.ID, err)
		}

		if output != nil {
			return fmt.Errorf("Service Catalog Tag Option (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAwsServiceCatalogTagOptionExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).scconn

		input := &servicecatalog.DescribeTagOptionInput{
			Id: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeTagOption(input)

		if err != nil {
			return fmt.Errorf("error describing Service Catalog Tag Option (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccAWSServiceCatalogTagOptionConfig_basic(key, value, active string) string {
	return fmt.Sprintf(`
resource "aws_servicecatalog_tag_option" "test" {
  key   = %[1]q
  value = %[2]q
  %[3]s
}
`, key, value, active)
}
