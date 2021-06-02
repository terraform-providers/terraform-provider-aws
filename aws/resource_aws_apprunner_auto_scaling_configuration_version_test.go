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
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/apprunner/waiter"
)

func init() {
	resource.AddTestSweepers("aws_apprunner_auto_scaling_configuration_version", &resource.Sweeper{
		Name:         "aws_apprunner_auto_scaling_configuration_version",
		F:            testSweepAppRunnerAutoScalingConfigurationVersions,
		Dependencies: []string{"aws_apprunner_service"},
	})
}

func testSweepAppRunnerAutoScalingConfigurationVersions(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*AWSClient).apprunnerconn
	sweepResources := make([]*testSweepResource, 0)
	ctx := context.Background()
	var errs *multierror.Error

	input := &apprunner.ListAutoScalingConfigurationsInput{}

	err = conn.ListAutoScalingConfigurationsPagesWithContext(ctx, input, func(page *apprunner.ListAutoScalingConfigurationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, summaryConfig := range page.AutoScalingConfigurationSummaryList {
			if summaryConfig == nil {
				continue
			}

			arn := aws.StringValue(summaryConfig.AutoScalingConfigurationArn)

			log.Printf("[INFO] Deleting App Runner AutoScaling Configuration Version (%s)", arn)
			r := resourceAwsAppRunnerAutoScalingConfigurationVersion()
			d := r.Data(nil)
			d.SetId(arn)

			sweepResources = append(sweepResources, NewTestSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing App Runner AutoScaling Configuration Versions: %w", err))
	}

	if err = testSweepResourceOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping App Runner AutoScaling Configuration Version for %s: %w", region, err))
	}

	if testSweepSkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping App Runner AutoScaling Configuration Versions sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func TestAccAwsAppRunnerAutoScalingConfigurationVersion_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_apprunner_auto_scaling_configuration_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAppRunner(t) },
		ErrorCheck:   testAccErrorCheck(t, apprunner.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppRunnerAutoScalingConfigurationVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppRunnerAutoScalingConfigurationVersionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppRunnerAutoScalingConfigurationVersionExists(resourceName),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "apprunner", regexp.MustCompile(fmt.Sprintf(`autoscalingconfiguration/%s/1/.+`, rName))),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_configuration_name", rName),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_configuration_revision", "1"),
					resource.TestCheckResourceAttr(resourceName, "latest", "true"),
					resource.TestCheckResourceAttr(resourceName, "max_concurrency", "100"),
					resource.TestCheckResourceAttr(resourceName, "max_size", "25"),
					resource.TestCheckResourceAttr(resourceName, "min_size", "1"),
					resource.TestCheckResourceAttr(resourceName, "status", waiter.AutoScalingConfigurationStatusActive),
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

func TestAccAwsAppRunnerAutoScalingConfigurationVersion_complex(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_apprunner_auto_scaling_configuration_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAppRunner(t) },
		ErrorCheck:   testAccErrorCheck(t, apprunner.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppRunnerAutoScalingConfigurationVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppRunnerAutoScalingConfigurationVersionConfig_withNonDefaults(rName, 50, 10, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppRunnerAutoScalingConfigurationVersionExists(resourceName),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "apprunner", regexp.MustCompile(fmt.Sprintf(`autoscalingconfiguration/%s/1/.+`, rName))),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_configuration_name", rName),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_configuration_revision", "1"),
					resource.TestCheckResourceAttr(resourceName, "latest", "true"),
					resource.TestCheckResourceAttr(resourceName, "max_concurrency", "50"),
					resource.TestCheckResourceAttr(resourceName, "max_size", "10"),
					resource.TestCheckResourceAttr(resourceName, "min_size", "2"),
					resource.TestCheckResourceAttr(resourceName, "status", waiter.AutoScalingConfigurationStatusActive),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Test resource recreation such that the revision number is still 1
				Config: testAccAppRunnerAutoScalingConfigurationVersionConfig_withNonDefaults(rName, 150, 20, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppRunnerAutoScalingConfigurationVersionExists(resourceName),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "apprunner", regexp.MustCompile(fmt.Sprintf(`autoscalingconfiguration/%s/1/.+`, rName))),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_configuration_name", rName),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_configuration_revision", "1"),
					resource.TestCheckResourceAttr(resourceName, "latest", "true"),
					resource.TestCheckResourceAttr(resourceName, "max_concurrency", "150"),
					resource.TestCheckResourceAttr(resourceName, "max_size", "20"),
					resource.TestCheckResourceAttr(resourceName, "min_size", "5"),
					resource.TestCheckResourceAttr(resourceName, "status", waiter.AutoScalingConfigurationStatusActive),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Test resource recreation such that the revision number is still 1
				Config: testAccAppRunnerAutoScalingConfigurationVersionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppRunnerAutoScalingConfigurationVersionExists(resourceName),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "apprunner", regexp.MustCompile(fmt.Sprintf(`autoscalingconfiguration/%s/1/.+`, rName))),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_configuration_name", rName),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_configuration_revision", "1"),
					resource.TestCheckResourceAttr(resourceName, "latest", "true"),
					resource.TestCheckResourceAttr(resourceName, "max_concurrency", "100"),
					resource.TestCheckResourceAttr(resourceName, "max_size", "25"),
					resource.TestCheckResourceAttr(resourceName, "min_size", "1"),
					resource.TestCheckResourceAttr(resourceName, "status", waiter.AutoScalingConfigurationStatusActive),
				),
			},
		},
	})
}

func TestAccAwsAppRunnerAutoScalingConfigurationVersion_MultipleVersions(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_apprunner_auto_scaling_configuration_version.test"
	otherResourceName := "aws_apprunner_auto_scaling_configuration_version.other"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAppRunner(t) },
		ErrorCheck:   testAccErrorCheck(t, apprunner.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppRunnerAutoScalingConfigurationVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppRunnerAutoScalingConfigurationVersionConfig_multipleVersions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppRunnerAutoScalingConfigurationVersionExists(resourceName),
					testAccCheckAwsAppRunnerAutoScalingConfigurationVersionExists(otherResourceName),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "apprunner", regexp.MustCompile(fmt.Sprintf(`autoscalingconfiguration/%s/1/.+`, rName))),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_configuration_name", rName),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_configuration_revision", "1"),
					resource.TestCheckResourceAttr(resourceName, "latest", "true"),
					resource.TestCheckResourceAttr(resourceName, "max_concurrency", "100"),
					resource.TestCheckResourceAttr(resourceName, "max_size", "25"),
					resource.TestCheckResourceAttr(resourceName, "min_size", "1"),
					resource.TestCheckResourceAttr(resourceName, "status", waiter.AutoScalingConfigurationStatusActive),
					testAccMatchResourceAttrRegionalARN(otherResourceName, "arn", "apprunner", regexp.MustCompile(fmt.Sprintf(`autoscalingconfiguration/%s/2/.+`, rName))),
					resource.TestCheckResourceAttr(otherResourceName, "auto_scaling_configuration_name", rName),
					resource.TestCheckResourceAttr(otherResourceName, "auto_scaling_configuration_revision", "2"),
					resource.TestCheckResourceAttr(otherResourceName, "latest", "true"),
					resource.TestCheckResourceAttr(otherResourceName, "max_concurrency", "100"),
					resource.TestCheckResourceAttr(otherResourceName, "max_size", "25"),
					resource.TestCheckResourceAttr(otherResourceName, "min_size", "1"),
					resource.TestCheckResourceAttr(otherResourceName, "status", waiter.AutoScalingConfigurationStatusActive),
				),
			},
			{
				// Test update of "latest" computed attribute after apply
				Config: testAccAppRunnerAutoScalingConfigurationVersionConfig_multipleVersions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppRunnerAutoScalingConfigurationVersionExists(resourceName),
					testAccCheckAwsAppRunnerAutoScalingConfigurationVersionExists(otherResourceName),
					resource.TestCheckResourceAttr(resourceName, "latest", "false"),
					resource.TestCheckResourceAttr(otherResourceName, "latest", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      otherResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsAppRunnerAutoScalingConfigurationVersion_UpdateMultipleVersions(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_apprunner_auto_scaling_configuration_version.test"
	otherResourceName := "aws_apprunner_auto_scaling_configuration_version.other"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAppRunner(t) },
		ErrorCheck:   testAccErrorCheck(t, apprunner.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppRunnerAutoScalingConfigurationVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppRunnerAutoScalingConfigurationVersionConfig_multipleVersions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppRunnerAutoScalingConfigurationVersionExists(resourceName),
					testAccCheckAwsAppRunnerAutoScalingConfigurationVersionExists(otherResourceName),
				),
			},
			{
				Config: testAccAppRunnerAutoScalingConfigurationVersionConfig_updateMultipleVersions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppRunnerAutoScalingConfigurationVersionExists(resourceName),
					testAccCheckAwsAppRunnerAutoScalingConfigurationVersionExists(otherResourceName),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "apprunner", regexp.MustCompile(fmt.Sprintf(`autoscalingconfiguration/%s/1/.+`, rName))),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_configuration_name", rName),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_configuration_revision", "1"),
					resource.TestCheckResourceAttr(resourceName, "latest", "false"),
					resource.TestCheckResourceAttr(resourceName, "max_concurrency", "100"),
					resource.TestCheckResourceAttr(resourceName, "max_size", "25"),
					resource.TestCheckResourceAttr(resourceName, "min_size", "1"),
					resource.TestCheckResourceAttr(resourceName, "status", waiter.AutoScalingConfigurationStatusActive),
					testAccMatchResourceAttrRegionalARN(otherResourceName, "arn", "apprunner", regexp.MustCompile(fmt.Sprintf(`autoscalingconfiguration/%s/2/.+`, rName))),
					resource.TestCheckResourceAttr(otherResourceName, "auto_scaling_configuration_name", rName),
					resource.TestCheckResourceAttr(otherResourceName, "auto_scaling_configuration_revision", "2"),
					resource.TestCheckResourceAttr(otherResourceName, "latest", "true"),
					resource.TestCheckResourceAttr(otherResourceName, "max_concurrency", "125"),
					resource.TestCheckResourceAttr(otherResourceName, "max_size", "20"),
					resource.TestCheckResourceAttr(otherResourceName, "min_size", "1"),
					resource.TestCheckResourceAttr(otherResourceName, "status", waiter.AutoScalingConfigurationStatusActive),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      otherResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsAppRunnerAutoScalingConfigurationVersion_disappears(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_apprunner_auto_scaling_configuration_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAppRunner(t) },
		ErrorCheck:   testAccErrorCheck(t, apprunner.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppRunnerAutoScalingConfigurationVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppRunnerAutoScalingConfigurationVersionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppRunnerAutoScalingConfigurationVersionExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsAppRunnerAutoScalingConfigurationVersion(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsAppRunnerAutoScalingConfigurationVersion_tags(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_apprunner_auto_scaling_configuration_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAppRunner(t) },
		ErrorCheck:   testAccErrorCheck(t, apprunner.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppRunnerAutoScalingConfigurationVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppRunnerAutoScalingConfigurationVersionConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppRunnerAutoScalingConfigurationVersionExists(resourceName),
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
				Config: testAccAppRunnerAutoScalingConfigurationVersionConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppRunnerAutoScalingConfigurationVersionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAppRunnerAutoScalingConfigurationVersionConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppRunnerAutoScalingConfigurationVersionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAwsAppRunnerAutoScalingConfigurationVersionDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_apprunner_auto_scaling_configuration_version" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).apprunnerconn

		input := &apprunner.DescribeAutoScalingConfigurationInput{
			AutoScalingConfigurationArn: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeAutoScalingConfigurationWithContext(context.Background(), input)

		if tfawserr.ErrCodeEquals(err, apprunner.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}

		if output != nil && output.AutoScalingConfiguration != nil && aws.StringValue(output.AutoScalingConfiguration.Status) != "inactive" {
			return fmt.Errorf("App Runner AutoScaling Configuration (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAwsAppRunnerAutoScalingConfigurationVersionExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No App Runner Service ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).apprunnerconn

		input := &apprunner.DescribeAutoScalingConfigurationInput{
			AutoScalingConfigurationArn: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeAutoScalingConfigurationWithContext(context.Background(), input)

		if err != nil {
			return err
		}

		if output == nil || output.AutoScalingConfiguration == nil {
			return fmt.Errorf("App Runner AutoScaling Configuration (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccAppRunnerAutoScalingConfigurationVersionConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_auto_scaling_configuration_version" "test" {
  auto_scaling_configuration_name = %[1]q
}
`, rName)
}

func testAccAppRunnerAutoScalingConfigurationVersionConfig_withNonDefaults(rName string, maxConcurrency, maxSize, minSize int) string {
	return fmt.Sprintf(`
resource "aws_apprunner_auto_scaling_configuration_version" "test" {
  auto_scaling_configuration_name = %[1]q

  max_concurrency = %[2]d
  max_size        = %[3]d
  min_size        = %[4]d
}
`, rName, maxConcurrency, maxSize, minSize)
}

func testAccAppRunnerAutoScalingConfigurationVersionConfig_multipleVersions(rName string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_auto_scaling_configuration_version" "test" {
  auto_scaling_configuration_name = %[1]q
}

resource "aws_apprunner_auto_scaling_configuration_version" "other" {
  auto_scaling_configuration_name = aws_apprunner_auto_scaling_configuration_version.test.auto_scaling_configuration_name
}
`, rName)
}

func testAccAppRunnerAutoScalingConfigurationVersionConfig_updateMultipleVersions(rName string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_auto_scaling_configuration_version" "test" {
  auto_scaling_configuration_name = %[1]q
}

resource "aws_apprunner_auto_scaling_configuration_version" "other" {
  auto_scaling_configuration_name = aws_apprunner_auto_scaling_configuration_version.test.auto_scaling_configuration_name

  max_concurrency = 125
  max_size        = 20
}
`, rName)
}

func testAccAppRunnerAutoScalingConfigurationVersionConfigTags1(rName string, tagKey1 string, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_auto_scaling_configuration_version" "test" {
  auto_scaling_configuration_name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAppRunnerAutoScalingConfigurationVersionConfigTags2(rName string, tagKey1 string, tagValue1 string, tagKey2 string, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_auto_scaling_configuration_version" "test" {
  auto_scaling_configuration_name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
