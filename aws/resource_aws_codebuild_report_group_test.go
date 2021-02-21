package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/codebuild/finder"
)

func init() {
	resource.AddTestSweepers("aws_codebuild_report_group", &resource.Sweeper{
		Name: "aws_codebuild_report_group",
		F:    testSweepCodeBuildReportGroups,
	})
}

func testSweepCodeBuildReportGroups(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*AWSClient).codebuildconn
	input := &codebuild.ListReportGroupsInput{}
	var sweeperErrs *multierror.Error

	err = conn.ListReportGroupsPages(input, func(page *codebuild.ListReportGroupsOutput, isLast bool) bool {
		if page == nil {
			return !isLast
		}

		for _, arn := range page.ReportGroups {
			id := aws.StringValue(arn)
			r := resourceAwsCodeBuildReportGroup()
			d := r.Data(nil)
			d.SetId(id)
			d.Set("delete_reports", true)

			err := r.Delete(d, client)
			if err != nil {
				sweeperErr := fmt.Errorf("error deleting CodeBuild Report Group (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !isLast
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping CodeBuild Report Group sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving CodeBuild ReportGroups: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSCodeBuildReportGroup_basic(t *testing.T) {
	var reportGroup codebuild.ReportGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_report_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuildReportGroup(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildReportGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildReportGroupBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildReportGroupExists(resourceName, &reportGroup),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "export_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "export_config.0.type", "NO_EXPORT"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "codebuild", fmt.Sprintf("report-group/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"delete_reports"},
			},
		},
	})
}

func TestAccAWSCodeBuildReportGroup_export_s3(t *testing.T) {
	var reportGroup codebuild.ReportGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_report_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuildReportGroup(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildReportGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildReportGroupS3ExportConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildReportGroupExists(resourceName, &reportGroup),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "export_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "export_config.0.type", "S3"),
					resource.TestCheckResourceAttr(resourceName, "export_config.0.s3_destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "export_config.0.s3_destination.0.packaging", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "export_config.0.s3_destination.0.encryption_disabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "export_config.0.s3_destination.0.path", "/some"),
					resource.TestCheckResourceAttrPair(resourceName, "export_config.0.s3_destination.0.encryption_key", "aws_kms_key.test", "arn"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "codebuild", fmt.Sprintf("report-group/%s", rName)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"delete_reports"},
			},
			{
				Config: testAccAWSCodeBuildReportGroupS3ExportUpdatedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildReportGroupExists(resourceName, &reportGroup),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "export_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "export_config.0.type", "S3"),
					resource.TestCheckResourceAttr(resourceName, "export_config.0.s3_destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "export_config.0.s3_destination.0.packaging", "ZIP"),
					resource.TestCheckResourceAttr(resourceName, "export_config.0.s3_destination.0.encryption_disabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "export_config.0.s3_destination.0.path", "/some2"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "codebuild", fmt.Sprintf("report-group/%s", rName)),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildReportGroup_tags(t *testing.T) {
	var reportGroup codebuild.ReportGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_report_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuildReportGroup(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildReportGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildReportGroupConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildReportGroupExists(resourceName, &reportGroup),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"delete_reports"},
			},
			{
				Config: testAccAWSCodeBuildReportGroupConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildReportGroupExists(resourceName, &reportGroup),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSCodeBuildReportGroupConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildReportGroupExists(resourceName, &reportGroup),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildReportGroup_deleteReports(t *testing.T) {
	var reportGroup codebuild.ReportGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_report_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuildReportGroup(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildReportGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildReportGroupDeleteReportsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildReportGroupExists(resourceName, &reportGroup),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"delete_reports"},
			},
		},
	})
}

func TestAccAWSCodeBuildReportGroup_disappears(t *testing.T) {
	var reportGroup codebuild.ReportGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_report_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuildReportGroup(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildReportGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildReportGroupBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildReportGroupExists(resourceName, &reportGroup),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCodeBuildReportGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccPreCheckAWSCodeBuildReportGroup(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).codebuildconn

	input := &codebuild.ListReportGroupsInput{}

	_, err := conn.ListReportGroups(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckAWSCodeBuildReportGroupDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).codebuildconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_codebuild_report_group" {
			continue
		}

		resp, err := finder.ReportGroupByArn(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if resp != nil {
			return fmt.Errorf("Found Report Group %s", rs.Primary.ID)
		}

	}
	return nil
}

func testAccCheckAWSCodeBuildReportGroupExists(name string, reportGroup *codebuild.ReportGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).codebuildconn

		resp, err := finder.ReportGroupByArn(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if resp == nil {
			return fmt.Errorf("Report Group %s not found", rs.Primary.ID)
		}

		*reportGroup = *resp

		return nil
	}
}

func testAccAWSCodeBuildReportGroupBasicConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_codebuild_report_group" "test" {
  name = %[1]q
  type = "TEST"

  export_config {
    type = "NO_EXPORT"
  }
}
`, rName)
}

func testAccAWSCodeBuildReportGroupBasicConfigS3ExportBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}
`, rName)
}

func testAccAWSCodeBuildReportGroupS3ExportConfig(rName string) string {
	return testAccAWSCodeBuildReportGroupBasicConfigS3ExportBase(rName) +
		fmt.Sprintf(`
resource "aws_codebuild_report_group" "test" {
  name = %[1]q
  type = "TEST"

  export_config {
    type = "S3"

    s3_destination {
      bucket              = aws_s3_bucket.test.id
      encryption_disabled = false
      encryption_key      = aws_kms_key.test.arn
      packaging           = "NONE"
      path                = "/some"
    }
  }
}
`, rName)
}

func testAccAWSCodeBuildReportGroupS3ExportUpdatedConfig(rName string) string {
	return testAccAWSCodeBuildReportGroupBasicConfigS3ExportBase(rName) +
		fmt.Sprintf(`
resource "aws_codebuild_report_group" "test" {
  name = %[1]q
  type = "TEST"

  export_config {
    type = "S3"

    s3_destination {
      bucket              = aws_s3_bucket.test.id
      encryption_key      = aws_kms_key.test.arn
      encryption_disabled = false
      packaging           = "ZIP"
      path                = "/some2"
    }
  }
}
`, rName)
}

func testAccAWSCodeBuildReportGroupConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_codebuild_report_group" "test" {
  name = %[1]q
  type = "TEST"

  export_config {
    type = "NO_EXPORT"
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSCodeBuildReportGroupConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_codebuild_report_group" "test" {
  name = %[1]q
  type = "TEST"

  export_config {
    type = "NO_EXPORT"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAWSCodeBuildReportGroupDeleteReportsConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_codebuild_report_group" "test" {
  name           = %[1]q
  type           = "TEST"
  delete_reports = true

  export_config {
    type = "NO_EXPORT"
  }
}
`, rName)
}
