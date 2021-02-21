package aws

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/service/costandusagereportservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAwsCurReportDefinition_basic(t *testing.T) {
	resourceName := "aws_cur_report_definition.test"
	s3BucketResourceName := "aws_s3_bucket.test"
	reportName := acctest.RandomWithPrefix("tf_acc_test")
	bucketName := fmt.Sprintf("tf-test-bucket-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckCur(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCurReportDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCurReportDefinitionConfig_basic(reportName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCurReportDefinitionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "report_name", reportName),
					resource.TestCheckResourceAttr(resourceName, "time_unit", "DAILY"),
					resource.TestCheckResourceAttr(resourceName, "compression", "GZIP"),
					resource.TestCheckResourceAttr(resourceName, "additional_schema_elements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "s3_bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "s3_prefix", ""),
					resource.TestCheckResourceAttrPair(resourceName, "s3_region", s3BucketResourceName, "region"),
					resource.TestCheckResourceAttr(resourceName, "additional_artifacts.#", "2"),
				),
			},
		},
	})
}

func TestAccAwsCurReportDefinition_textOrCsv(t *testing.T) {
	resourceName := "aws_cur_report_definition.test"
	s3BucketResourceName := "aws_s3_bucket.test"
	reportName := acctest.RandomWithPrefix("tf_acc_test")
	bucketName := fmt.Sprintf("tf-test-bucket-%d", acctest.RandInt())
	bucketPrefix := ""
	format := "textORcsv"
	compression := "GZIP"
	additionalArtifacts := []string{"REDSHIFT", "QUICKSIGHT"}
	refreshClosedReports := false
	reportVersioning := "CREATE_NEW_REPORT"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckCur(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCurReportDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCurReportDefinitionConfig_additional(reportName, bucketName, bucketPrefix, format, compression, additionalArtifacts, refreshClosedReports, reportVersioning),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCurReportDefinitionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "report_name", reportName),
					resource.TestCheckResourceAttr(resourceName, "time_unit", "DAILY"),
					resource.TestCheckResourceAttr(resourceName, "format", format),
					resource.TestCheckResourceAttr(resourceName, "compression", compression),
					resource.TestCheckResourceAttr(resourceName, "additional_schema_elements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "s3_bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "s3_prefix", bucketPrefix),
					resource.TestCheckResourceAttrPair(resourceName, "s3_region", s3BucketResourceName, "region"),
					resource.TestCheckResourceAttr(resourceName, "additional_artifacts.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "refresh_closed_reports", "false"),
					resource.TestCheckResourceAttr(resourceName, "report_versioning", reportVersioning),
				),
			},
		},
	})
}

func TestAccAwsCurReportDefinition_parquet(t *testing.T) {
	resourceName := "aws_cur_report_definition.test"
	s3BucketResourceName := "aws_s3_bucket.test"
	reportName := acctest.RandomWithPrefix("tf_acc_test")
	bucketName := fmt.Sprintf("tf-test-bucket-%d", acctest.RandInt())
	bucketPrefix := ""
	format := "Parquet"
	compression := "Parquet"
	additionalArtifacts := []string{}
	refreshClosedReports := false
	reportVersioning := "CREATE_NEW_REPORT"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckCur(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCurReportDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCurReportDefinitionConfig_additional(reportName, bucketName, bucketPrefix, format, compression, additionalArtifacts, refreshClosedReports, reportVersioning),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCurReportDefinitionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "report_name", reportName),
					resource.TestCheckResourceAttr(resourceName, "time_unit", "DAILY"),
					resource.TestCheckResourceAttr(resourceName, "format", format),
					resource.TestCheckResourceAttr(resourceName, "compression", compression),
					resource.TestCheckResourceAttr(resourceName, "additional_schema_elements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "s3_bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "s3_prefix", bucketPrefix),
					resource.TestCheckResourceAttrPair(resourceName, "s3_region", s3BucketResourceName, "region"),
					resource.TestCheckResourceAttr(resourceName, "refresh_closed_reports", "false"),
					resource.TestCheckResourceAttr(resourceName, "report_versioning", reportVersioning),
				),
			},
		},
	})
}

func TestAccAwsCurReportDefinition_athena(t *testing.T) {
	resourceName := "aws_cur_report_definition.test"
	s3BucketResourceName := "aws_s3_bucket.test"
	reportName := acctest.RandomWithPrefix("tf_acc_test")
	bucketName := fmt.Sprintf("tf-test-bucket-%d", acctest.RandInt())
	bucketPrefix := "data"
	format := "Parquet"
	compression := "Parquet"
	additionalArtifacts := []string{"ATHENA"}
	refreshClosedReports := false
	reportVersioning := "OVERWRITE_REPORT"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckCur(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCurReportDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCurReportDefinitionConfig_additional(reportName, bucketName, bucketPrefix, format, compression, additionalArtifacts, refreshClosedReports, reportVersioning),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCurReportDefinitionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "report_name", reportName),
					resource.TestCheckResourceAttr(resourceName, "time_unit", "DAILY"),
					resource.TestCheckResourceAttr(resourceName, "format", format),
					resource.TestCheckResourceAttr(resourceName, "compression", compression),
					resource.TestCheckResourceAttr(resourceName, "additional_schema_elements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "s3_bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "s3_prefix", bucketPrefix),
					resource.TestCheckResourceAttrPair(resourceName, "s3_region", s3BucketResourceName, "region"),
					resource.TestCheckResourceAttr(resourceName, "additional_artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "refresh_closed_reports", "false"),
					resource.TestCheckResourceAttr(resourceName, "report_versioning", reportVersioning),
				),
			},
		},
	})
}

func TestAccAwsCurReportDefinition_refresh(t *testing.T) {
	resourceName := "aws_cur_report_definition.test"
	s3BucketResourceName := "aws_s3_bucket.test"
	reportName := acctest.RandomWithPrefix("tf_acc_test")
	bucketName := fmt.Sprintf("tf-test-bucket-%d", acctest.RandInt())
	bucketPrefix := ""
	format := "textORcsv"
	compression := "GZIP"
	additionalArtifacts := []string{"REDSHIFT", "QUICKSIGHT"}
	refreshClosedReports := true
	reportVersioning := "CREATE_NEW_REPORT"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckCur(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCurReportDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCurReportDefinitionConfig_additional(reportName, bucketName, bucketPrefix, format, compression, additionalArtifacts, refreshClosedReports, reportVersioning),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCurReportDefinitionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "report_name", reportName),
					resource.TestCheckResourceAttr(resourceName, "time_unit", "DAILY"),
					resource.TestCheckResourceAttr(resourceName, "format", format),
					resource.TestCheckResourceAttr(resourceName, "compression", compression),
					resource.TestCheckResourceAttr(resourceName, "additional_schema_elements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "s3_bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "s3_prefix", bucketPrefix),
					resource.TestCheckResourceAttrPair(resourceName, "s3_region", s3BucketResourceName, "region"),
					resource.TestCheckResourceAttr(resourceName, "additional_artifacts.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "refresh_closed_reports", "true"),
					resource.TestCheckResourceAttr(resourceName, "report_versioning", reportVersioning),
				),
			},
		},
	})
}

func TestAccAwsCurReportDefinition_overwrite(t *testing.T) {
	resourceName := "aws_cur_report_definition.test"
	s3BucketResourceName := "aws_s3_bucket.test"
	reportName := acctest.RandomWithPrefix("tf_acc_test")
	bucketName := fmt.Sprintf("tf-test-bucket-%d", acctest.RandInt())
	bucketPrefix := ""
	format := "textORcsv"
	compression := "GZIP"
	additionalArtifacts := []string{"REDSHIFT", "QUICKSIGHT"}
	refreshClosedReports := false
	reportVersioning := "OVERWRITE_REPORT"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckCur(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCurReportDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCurReportDefinitionConfig_additional(reportName, bucketName, bucketPrefix, format, compression, additionalArtifacts, refreshClosedReports, reportVersioning),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCurReportDefinitionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "report_name", reportName),
					resource.TestCheckResourceAttr(resourceName, "time_unit", "DAILY"),
					resource.TestCheckResourceAttr(resourceName, "format", format),
					resource.TestCheckResourceAttr(resourceName, "compression", compression),
					resource.TestCheckResourceAttr(resourceName, "additional_schema_elements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "s3_bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "s3_prefix", ""),
					resource.TestCheckResourceAttrPair(resourceName, "s3_region", s3BucketResourceName, "region"),
					resource.TestCheckResourceAttr(resourceName, "additional_artifacts.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "refresh_closed_reports", "false"),
					resource.TestCheckResourceAttr(resourceName, "report_versioning", reportVersioning),
				),
			},
		},
	})
}

func testAccCheckAwsCurReportDefinitionDestroy(s *terraform.State) error {
	conn := testAccProviderCur.Meta().(*AWSClient).costandusagereportconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cur_report_definition" {
			continue
		}
		reportName := rs.Primary.ID
		matchingReportDefinition, err := describeCurReportDefinition(conn, reportName)
		if err != nil {
			return err
		}
		if matchingReportDefinition != nil {
			return fmt.Errorf("Report Definition still exists: %q", rs.Primary.ID)
		}
	}
	return nil

}

func testAccCheckAwsCurReportDefinitionExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProviderCur.Meta().(*AWSClient).costandusagereportconn

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Resource not found: %s", resourceName)
		}
		reportName := rs.Primary.ID
		matchingReportDefinition, err := describeCurReportDefinition(conn, reportName)
		if err != nil {
			return err
		}
		if matchingReportDefinition == nil {
			return fmt.Errorf("Report Definition does not exist: %q", rs.Primary.ID)
		}
		return nil
	}
}

func testAccAwsCurReportDefinitionConfig_basic(reportName string, bucketName string) string {
	return composeConfig(
		testAccCurRegionProviderConfig(),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = "%[2]s"
  acl           = "private"
  force_destroy = true
}

data "aws_partition" "current" {}

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.id

  policy = <<POLICY
{
  "Version": "2008-10-17",
  "Id": "s3policy",
  "Statement": [
    {
      "Sid": "AllowCURBillingACLPolicy",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:${data.aws_partition.current.partition}:iam::386209384616:root"
      },
      "Action": [
        "s3:GetBucketAcl",
        "s3:GetBucketPolicy"
      ],
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::${aws_s3_bucket.test.id}"
    },
    {
      "Sid": "AllowCURPutObject",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:${data.aws_partition.current.partition}:iam::386209384616:root"
      },
      "Action": "s3:PutObject",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::${aws_s3_bucket.test.id}/*"
    }
  ]
}
POLICY
}

resource "aws_cur_report_definition" "test" {
  depends_on = [aws_s3_bucket_policy.test] # needed to avoid "ValidationException: Failed to verify customer bucket permission."

  report_name                = "%[1]s"
  time_unit                  = "DAILY"
  format                     = "textORcsv"
  compression                = "GZIP"
  additional_schema_elements = ["RESOURCES"]
  s3_bucket                  = aws_s3_bucket.test.id
  s3_prefix                  = ""
  s3_region                  = aws_s3_bucket.test.region
  additional_artifacts       = ["REDSHIFT", "QUICKSIGHT"]
}
`, reportName, bucketName))
}

func testAccAwsCurReportDefinitionConfig_additional(reportName string, bucketName string, bucketPrefix string, format string, compression string, additionalArtifacts []string, refreshClosedReports bool, reportVersioning string) string {
	artifactsStr := strings.Join(additionalArtifacts, "\", \"")

	if len(additionalArtifacts) > 0 {
		artifactsStr = fmt.Sprintf("additional_artifacts       = [\"%s\"]", artifactsStr)
	} else {
		artifactsStr = ""
	}

	return composeConfig(
		testAccCurRegionProviderConfig(),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = "%[2]s"
  acl           = "private"
  force_destroy = true
}

data "aws_partition" "current" {}

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.id

  policy = <<POLICY
{
  "Version": "2008-10-17",
  "Id": "s3policy",
  "Statement": [
    {
      "Sid": "AllowCURBillingACLPolicy",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:${data.aws_partition.current.partition}:iam::386209384616:root"
      },
      "Action": [
        "s3:GetBucketAcl",
        "s3:GetBucketPolicy"
      ],
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::${aws_s3_bucket.test.id}"
    },
    {
      "Sid": "AllowCURPutObject",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:${data.aws_partition.current.partition}:iam::386209384616:root"
      },
      "Action": "s3:PutObject",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::${aws_s3_bucket.test.id}/*"
    }
  ]
}
POLICY
}

resource "aws_cur_report_definition" "test" {
  depends_on = [aws_s3_bucket_policy.test] # needed to avoid "ValidationException: Failed to verify customer bucket permission."

  report_name                = "%[1]s"
  time_unit                  = "DAILY"
  format                     = "%[4]s"
  compression                = "%[5]s"
  additional_schema_elements = ["RESOURCES"]
  s3_bucket                  = aws_s3_bucket.test.id
  s3_prefix                  = "%[3]s"
  s3_region                  = aws_s3_bucket.test.region
	%[6]s
  refresh_closed_reports = %[7]t
  report_versioning      = "%[8]s"
}
`, reportName, bucketName, bucketPrefix, format, compression, artifactsStr, refreshClosedReports, reportVersioning))
}

func TestCheckAwsCurReportDefinitionPropertyCombination(t *testing.T) {
	type propertyCombinationTestCase struct {
		additionalArtifacts []string
		compression         string
		format              string
		prefix              string
		reportVersioning    string
		shouldError         bool
	}

	testCases := map[string]propertyCombinationTestCase{
		"TestAthenaAndAdditionalArtifacts": {
			additionalArtifacts: []string{
				costandusagereportservice.AdditionalArtifactAthena,
				costandusagereportservice.AdditionalArtifactRedshift,
			},
			compression:      costandusagereportservice.CompressionFormatZip,
			format:           costandusagereportservice.ReportFormatTextOrcsv,
			prefix:           "prefix/",
			reportVersioning: costandusagereportservice.ReportVersioningCreateNewReport,
			shouldError:      true,
		},
		"TestAthenaAndEmptyPrefix": {
			additionalArtifacts: []string{
				costandusagereportservice.AdditionalArtifactAthena,
			},
			compression:      costandusagereportservice.CompressionFormatZip,
			format:           costandusagereportservice.ReportFormatTextOrcsv,
			prefix:           "",
			reportVersioning: costandusagereportservice.ReportVersioningCreateNewReport,
			shouldError:      true,
		},
		"TestAthenaAndOverrideReport": {
			additionalArtifacts: []string{
				costandusagereportservice.AdditionalArtifactAthena,
			},
			compression:      costandusagereportservice.CompressionFormatZip,
			format:           costandusagereportservice.ReportFormatTextOrcsv,
			prefix:           "prefix/",
			reportVersioning: costandusagereportservice.ReportVersioningCreateNewReport,
			shouldError:      true,
		},
		"TestAthenaWithNonParquetFormat": {
			additionalArtifacts: []string{
				costandusagereportservice.AdditionalArtifactAthena,
			},
			compression:      costandusagereportservice.CompressionFormatParquet,
			format:           costandusagereportservice.ReportFormatTextOrcsv,
			prefix:           "prefix/",
			reportVersioning: costandusagereportservice.ReportVersioningCreateNewReport,
			shouldError:      true,
		},
		"TestParquetFormatWithoutParquetCompression": {
			additionalArtifacts: []string{
				costandusagereportservice.AdditionalArtifactAthena,
			},
			compression:      costandusagereportservice.CompressionFormatZip,
			format:           costandusagereportservice.ReportFormatParquet,
			prefix:           "prefix/",
			reportVersioning: costandusagereportservice.ReportVersioningCreateNewReport,
			shouldError:      true,
		},
		"TestCSVFormatWithParquetCompression": {
			additionalArtifacts: []string{
				costandusagereportservice.AdditionalArtifactAthena,
			},
			compression:      costandusagereportservice.CompressionFormatParquet,
			format:           costandusagereportservice.ReportFormatTextOrcsv,
			prefix:           "prefix/",
			reportVersioning: costandusagereportservice.ReportVersioningCreateNewReport,
			shouldError:      true,
		},
		"TestRedshiftWithParquetformat": {
			additionalArtifacts: []string{
				costandusagereportservice.AdditionalArtifactRedshift,
			},
			compression:      costandusagereportservice.CompressionFormatParquet,
			format:           costandusagereportservice.ReportFormatParquet,
			prefix:           "prefix/",
			reportVersioning: costandusagereportservice.ReportVersioningCreateNewReport,
			shouldError:      true,
		},
		"TestQuicksightWithParquetformat": {
			additionalArtifacts: []string{
				costandusagereportservice.AdditionalArtifactQuicksight,
			},
			compression:      costandusagereportservice.CompressionFormatParquet,
			format:           costandusagereportservice.ReportFormatParquet,
			prefix:           "prefix/",
			reportVersioning: costandusagereportservice.ReportVersioningCreateNewReport,
			shouldError:      true,
		},
		"TestAthenaValidCombination": {
			additionalArtifacts: []string{
				costandusagereportservice.AdditionalArtifactAthena,
			},
			compression:      costandusagereportservice.CompressionFormatParquet,
			format:           costandusagereportservice.ReportFormatParquet,
			prefix:           "prefix/",
			reportVersioning: costandusagereportservice.ReportVersioningOverwriteReport,
			shouldError:      false,
		},
		"TestRedshiftWithGzipedOverwrite": {
			additionalArtifacts: []string{
				costandusagereportservice.AdditionalArtifactRedshift,
			},
			compression:      costandusagereportservice.CompressionFormatGzip,
			format:           costandusagereportservice.ReportFormatTextOrcsv,
			prefix:           "prefix/",
			reportVersioning: costandusagereportservice.ReportVersioningOverwriteReport,
			shouldError:      false,
		},
		"TestRedshiftWithZippedOverwrite": {
			additionalArtifacts: []string{
				costandusagereportservice.AdditionalArtifactRedshift,
			},
			compression:      costandusagereportservice.CompressionFormatZip,
			format:           costandusagereportservice.ReportFormatTextOrcsv,
			prefix:           "",
			reportVersioning: costandusagereportservice.ReportVersioningOverwriteReport,
			shouldError:      false,
		},
		"TestRedshiftWithGzipedCreateNew": {
			additionalArtifacts: []string{
				costandusagereportservice.AdditionalArtifactRedshift,
			},
			compression:      costandusagereportservice.CompressionFormatGzip,
			format:           costandusagereportservice.ReportFormatTextOrcsv,
			prefix:           "",
			reportVersioning: costandusagereportservice.ReportVersioningCreateNewReport,
			shouldError:      false,
		},
		"TestRedshiftWithZippedCreateNew": {
			additionalArtifacts: []string{
				costandusagereportservice.AdditionalArtifactRedshift,
			},
			compression:      costandusagereportservice.CompressionFormatZip,
			format:           costandusagereportservice.ReportFormatTextOrcsv,
			prefix:           "",
			reportVersioning: costandusagereportservice.ReportVersioningCreateNewReport,
			shouldError:      false,
		},
		"TestQuicksightWithGzipedOverwrite": {
			additionalArtifacts: []string{
				costandusagereportservice.AdditionalArtifactQuicksight,
			},
			compression:      costandusagereportservice.CompressionFormatGzip,
			format:           costandusagereportservice.ReportFormatTextOrcsv,
			prefix:           "",
			reportVersioning: costandusagereportservice.ReportVersioningOverwriteReport,
			shouldError:      false,
		},
		"TestQuicksightWithZippedOverwrite": {
			additionalArtifacts: []string{
				costandusagereportservice.AdditionalArtifactQuicksight,
			},
			compression:      costandusagereportservice.CompressionFormatZip,
			format:           costandusagereportservice.ReportFormatTextOrcsv,
			prefix:           "",
			reportVersioning: costandusagereportservice.ReportVersioningOverwriteReport,
			shouldError:      false,
		},
		"TestQuicksightWithGzipedCreateNew": {
			additionalArtifacts: []string{
				costandusagereportservice.AdditionalArtifactQuicksight,
			},
			compression:      costandusagereportservice.CompressionFormatGzip,
			format:           costandusagereportservice.ReportFormatTextOrcsv,
			prefix:           "",
			reportVersioning: costandusagereportservice.ReportVersioningCreateNewReport,
			shouldError:      false,
		},
		"TestQuicksightWithZippedCreateNew": {
			additionalArtifacts: []string{
				costandusagereportservice.AdditionalArtifactQuicksight,
			},
			compression:      costandusagereportservice.CompressionFormatZip,
			format:           costandusagereportservice.ReportFormatTextOrcsv,
			prefix:           "",
			reportVersioning: costandusagereportservice.ReportVersioningCreateNewReport,
			shouldError:      false,
		},
		"TestMultipleArtifacts": {
			additionalArtifacts: []string{
				costandusagereportservice.AdditionalArtifactRedshift,
				costandusagereportservice.AdditionalArtifactQuicksight,
			},
			compression:      costandusagereportservice.CompressionFormatGzip,
			format:           costandusagereportservice.ReportFormatTextOrcsv,
			prefix:           "prefix/",
			reportVersioning: costandusagereportservice.ReportVersioningOverwriteReport,
			shouldError:      false,
		},
	}

	for name, tCase := range testCases {
		t.Run(name, func(t *testing.T) {
			err := checkAwsCurReportDefinitionPropertyCombination(
				tCase.additionalArtifacts,
				tCase.compression,
				tCase.format,
				tCase.prefix,
				tCase.reportVersioning,
			)

			if tCase.shouldError && err == nil {
				t.Fail()
			} else if !tCase.shouldError && err != nil {
				t.Fail()
			}
		})
	}
}
