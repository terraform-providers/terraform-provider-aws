package aws

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apprunner"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/apprunner/finder"
)

func init() {
	resource.AddTestSweepers("aws_apprunner_connection", &resource.Sweeper{
		Name:         "aws_apprunner_connection",
		F:            testSweepAppRunnerConnections,
		Dependencies: []string{"aws_apprunner_service"},
	})
}

func testSweepAppRunnerConnections(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*AWSClient).apprunnerconn
	sweepResources := make([]*testSweepResource, 0)
	ctx := context.Background()

	var errs *multierror.Error

	input := &apprunner.ListConnectionsInput{}

	err = conn.ListConnectionsPagesWithContext(ctx, input, func(page *apprunner.ListConnectionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, c := range page.ConnectionSummaryList {
			if c == nil {
				continue
			}

			name := aws.StringValue(c.ConnectionName)

			log.Printf("[INFO] Deleting App Runner Connection: %s", name)

			r := resourceAwsAppRunnerConnection()
			d := r.Data(nil)
			d.SetId(name)
			d.Set("arn", c.ConnectionArn)

			sweepResources = append(sweepResources, NewTestSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing App Runner Connections: %w", err))
	}

	if err = testSweepResourceOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping App Runner Connections for %s: %w", region, err))
	}

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping App Runner Connections sweep for %s: %s", region, err)
		return nil // In case we have completed some pages, but had errors
	}

	return errs.ErrorOrNil()
}

func TestAccAwsAppRunnerConnection_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_apprunner_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAppRunner(t) },
		ErrorCheck:   testAccErrorCheck(t, apprunner.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppRunnerConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppRunnerConnection_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppRunnerConnectionExists(resourceName),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "apprunner", regexp.MustCompile(fmt.Sprintf(`connection/%s/.+`, rName))),
					resource.TestCheckResourceAttr(resourceName, "connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "provider_type", apprunner.ProviderTypeGithub),
					resource.TestCheckResourceAttr(resourceName, "status", apprunner.ConnectionStatusPendingHandshake),
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

func TestAccAwsAppRunnerConnection_disappears(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_apprunner_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAppRunner(t) },
		ErrorCheck:   testAccErrorCheck(t, apprunner.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppRunnerConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppRunnerConnection_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppRunnerConnectionExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsAppRunnerConnection(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsAppRunnerConnection_tags(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_apprunner_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAppRunner(t) },
		ErrorCheck:   testAccErrorCheck(t, apprunner.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppRunnerConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppRunnerConnectionConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppRunnerConnectionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAppRunnerConnectionConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppRunnerConnectionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key2", "value2"),
				),
			},
			{
				Config: testAccAppRunnerConnectionConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppRunnerConnectionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAwsAppRunnerConnectionDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_apprunner_connection" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).apprunnerconn

		connection, err := finder.ConnectionSummaryByName(context.Background(), conn, rs.Primary.ID)

		if tfawserr.ErrCodeEquals(err, apprunner.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}

		if connection != nil {
			return fmt.Errorf("App Runner Connection (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAwsAppRunnerConnectionExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No App Runner Connection ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).apprunnerconn

		connection, err := finder.ConnectionSummaryByName(context.Background(), conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if connection == nil {
			return fmt.Errorf("App Runner Connection (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccAppRunnerConnection_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_connection" "test" {
  connection_name = %q
  provider_type   = "GITHUB"
}
`, rName)
}

func testAccAppRunnerConnectionConfigTags1(rName string, tagKey1 string, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_connection" "test" {
  connection_name = %[1]q
  provider_type   = "GITHUB"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAppRunnerConnectionConfigTags2(rName string, tagKey1 string, tagValue1 string, tagKey2 string, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_connection" "test" {
  connection_name = %[1]q
  provider_type   = "GITHUB"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
