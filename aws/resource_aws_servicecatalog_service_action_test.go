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

// add sweeper to delete known test servicecat service actions
func init() {
	resource.AddTestSweepers("aws_servicecatalog_service_action", &resource.Sweeper{
		Name:         "aws_servicecatalog_service_action",
		Dependencies: []string{},
		F:            testSweepServiceCatalogServiceActions,
	})
}

func testSweepServiceCatalogServiceActions(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*AWSClient).scconn
	sweepResources := make([]*testSweepResource, 0)
	var errs *multierror.Error

	input := &servicecatalog.ListServiceActionsInput{}

	err = conn.ListServiceActionsPages(input, func(page *servicecatalog.ListServiceActionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, sas := range page.ServiceActionSummaries {
			if sas == nil {
				continue
			}

			id := aws.StringValue(sas.Id)

			r := resourceAwsServiceCatalogServiceAction()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, NewTestSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error describing Service Catalog Service Actions for %s: %w", region, err))
	}

	if err = testSweepResourceOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Service Catalog Service Actions for %s: %w", region, err))
	}

	if testSweepSkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Service Catalog Service Actions sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func TestAccAWSServiceCatalogServiceAction_basic(t *testing.T) {
	resourceName := "aws_servicecatalog_service_action.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, servicecatalog.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogServiceActionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogServiceActionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogServiceActionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "accept_language", "en"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.name", "AWS-RestartEC2Instance"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.version", "1"),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"accept_language",
				},
			},
		},
	})
}

func TestAccAWSServiceCatalogServiceAction_disappears(t *testing.T) {
	resourceName := "aws_servicecatalog_service_action.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, servicecatalog.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogServiceActionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogServiceActionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogServiceActionExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsServiceCatalogServiceAction(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSServiceCatalogServiceAction_update(t *testing.T) {
	resourceName := "aws_servicecatalog_service_action.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	rName2 := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, servicecatalog.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogServiceActionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogServiceActionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogServiceActionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "accept_language", "en"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.name", "AWS-RestartEC2Instance"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.version", "1"),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				Config: testAccAWSServiceCatalogServiceActionConfig_update(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogServiceActionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "accept_language", "en"),
					resource.TestCheckResourceAttrPair(resourceName, "definition.0.assume_role", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "description", rName2),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
				),
			}},
	})
}

func testAccCheckAwsServiceCatalogServiceActionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).scconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_servicecatalog_service_action" {
			continue
		}

		input := &servicecatalog.DescribeServiceActionInput{
			Id: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeServiceAction(input)

		if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error getting Service Catalog Service Action (%s): %w", rs.Primary.ID, err)
		}

		if output != nil {
			return fmt.Errorf("Service Catalog Service Action (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAwsServiceCatalogServiceActionExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).scconn

		input := &servicecatalog.DescribeServiceActionInput{
			Id: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeServiceAction(input)

		if err != nil {
			return fmt.Errorf("error describing Service Catalog Service Action (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccAWSServiceCatalogServiceActionConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_servicecatalog_service_action" "test" {
  accept_language = "en"
  description     = %[1]q
  name            = %[1]q

  definition {
    name    = "AWS-RestartEC2Instance"
    version = "1"
  }
}
`, rName)
}

func testAccAWSServiceCatalogServiceActionConfig_update(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_partition" "current" {}

data "aws_iam_policy_document" "test" {
  statement {
    effect = "Allow"

    principals {
      type = "Service"

      identifiers = [
        "servicecatalog.${data.aws_region.current.name}.${data.aws_partition.current.dns_suffix}",
      ]
    }

    actions = [
      "sts:AssumeRole",
    ]
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test.json
}

resource "aws_servicecatalog_service_action" "test" {
  description = %[1]q
  name        = %[1]q

  definition {
    assume_role = aws_iam_role.test.arn
    name        = "AWS-RestartEC2Instance"
    version     = "1"
  }
}
`, rName)
}
