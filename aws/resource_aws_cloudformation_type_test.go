package aws

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAwsCloudformationType_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	typeName := fmt.Sprintf("HashiCorp::TerraformAwsProvider::TfAccTest%s", acctest.RandString(8))
	zipPath := testAccAwsCloudformationTypeZipGenerator(t, typeName)
	resourceName := "aws_cloudformation_type.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCloudformationTypeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudformationTypeConfigTypeName(rName, zipPath, typeName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsCloudformationTypeExists(resourceName),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "cloudformation", regexp.MustCompile(fmt.Sprintf("type/resource/%s/.+", strings.ReplaceAll(typeName, "::", "-")))),
					resource.TestCheckResourceAttr(resourceName, "default_version_id", ""),
					resource.TestCheckResourceAttr(resourceName, "deprecated_status", cloudformation.DeprecatedStatusLive),
					resource.TestCheckResourceAttr(resourceName, "description", "An example resource schema demonstrating some basic constructs and validation rules."),
					resource.TestCheckResourceAttr(resourceName, "execution_role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "is_default_version", "true"),
					resource.TestCheckResourceAttr(resourceName, "logging_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "provisioning_type", cloudformation.ProvisioningTypeFullyMutable),
					resource.TestCheckResourceAttr(resourceName, "schema_handler_package", fmt.Sprintf("s3://%s/test", rName)),
					resource.TestMatchResourceAttr(resourceName, "schema", regexp.MustCompile(`^\{.*`)),
					resource.TestCheckResourceAttr(resourceName, "source_url", "https://github.com/aws-cloudformation/aws-cloudformation-rpdk.git"),
					resource.TestCheckResourceAttr(resourceName, "type", cloudformation.RegistryTypeResource),
					testAccCheckResourceAttrRegionalARN(resourceName, "type_arn", "cloudformation", fmt.Sprintf("type/resource/%s", strings.ReplaceAll(typeName, "::", "-"))),
					resource.TestCheckResourceAttr(resourceName, "type_name", typeName),
					resource.TestCheckResourceAttr(resourceName, "visibility", cloudformation.VisibilityPrivate),
					resource.TestMatchResourceAttr(resourceName, "version_id", regexp.MustCompile(`.+`)),
				),
			},
		},
	})
}

func TestAccAwsCloudformationType_disappears(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	typeName := fmt.Sprintf("HashiCorp::TerraformAwsProvider::TfAccTest%s", acctest.RandString(8))
	zipPath := testAccAwsCloudformationTypeZipGenerator(t, typeName)
	resourceName := "aws_cloudformation_type.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCloudformationTypeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudformationTypeConfigTypeName(rName, zipPath, typeName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsCloudformationTypeExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCloudFormationType(), resourceName),
					// Verify Delete error handling
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCloudFormationType(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsCloudformationType_ExecutionRoleArn(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	typeName := fmt.Sprintf("HashiCorp::TerraformAwsProvider::TfAccTest%s", acctest.RandString(8))
	zipPath := testAccAwsCloudformationTypeZipGenerator(t, typeName)
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_cloudformation_type.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCloudformationTypeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudformationTypeConfigExecutionRoleArn(rName, zipPath, typeName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsCloudformationTypeExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "execution_role_arn", iamRoleResourceName, "arn"),
				),
			},
		},
	})
}

func TestAccAwsCloudformationType_LoggingConfig(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	typeName := fmt.Sprintf("HashiCorp::TerraformAwsProvider::TfAccTest%s", acctest.RandString(8))
	zipPath := testAccAwsCloudformationTypeZipGenerator(t, typeName)
	cloudwatchLogGroupResourceName := "aws_cloudwatch_log_group.test"
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_cloudformation_type.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCloudformationTypeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudformationTypeConfigLoggingConfig(rName, zipPath, typeName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsCloudformationTypeExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "logging_config.0.log_group_name", cloudwatchLogGroupResourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "logging_config.0.log_role_arn", iamRoleResourceName, "arn"),
				),
			},
		},
	})
}

func testAccCheckAwsCloudformationTypeExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).cfconn

		input := &cloudformation.DescribeTypeInput{
			Arn: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeType(input)

		if err != nil {
			return fmt.Errorf("error reading CloudFormation Type (%s): %w", rs.Primary.ID, err)
		}

		if output == nil {
			return fmt.Errorf("error reading CloudFormation Type (%s): empty response", rs.Primary.ID)
		}

		if aws.StringValue(output.DeprecatedStatus) != cloudformation.DeprecatedStatusLive {
			return fmt.Errorf("error reading CloudFormation Type (%s): unexpected deprecated status: %s", rs.Primary.ID, aws.StringValue(output.DeprecatedStatus))
		}

		return nil
	}
}

func testAccCheckAwsCloudformationTypeDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cfconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudformation_stack_set" {
			continue
		}

		input := &cloudformation.DescribeTypeInput{
			Arn: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeType(input)

		if tfawserr.ErrCodeEquals(err, cloudformation.ErrCodeTypeNotFoundException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error reading CloudFormation Type (%s): %w", rs.Primary.ID, err)
		}

		if output == nil {
			return fmt.Errorf("error reading CloudFormation Type (%s): empty response", rs.Primary.ID)
		}

		if aws.StringValue(output.DeprecatedStatus) != cloudformation.DeprecatedStatusDeprecated {
			return fmt.Errorf("error reading CloudFormation Type (%s): unexpected deprecated status: %s", rs.Primary.ID, aws.StringValue(output.DeprecatedStatus))
		}
	}

	return nil
}

// Since the CloudFormation resource schema type name must match the RegisterType TypeName
// and we randomize the latter to avoid testing race conditions, this function sets up a
// temporary directory and returns a string to the generated zip file.
//
// This is essentially the equivalent of running `cfn submit --dry-run` with the correct
// type names injected. Only .rpdk-config, handler.zip, and schema.json are required from
// the following zip structure that is created:
// .
// ├── .cfn_metadata.json
// ├── .rpdk-config
// ├── Makefile
// ├── cmd
// │   ├── main.go
// │   └── resource
// │       ├── model.go
// │       └── resource.go
// ├── handler.zip
// └── schema.json
//
// Where handler.zip has the single Go binary file from bin, handler, at the root. This
// binary file is large (near 10MB as of this writing), so it is in the .gitignore file.
// The test skip messaging includes directions for how to locally build the binary.
func testAccAwsCloudformationTypeZipGenerator(t *testing.T, typeName string) string {
	t.Helper()
	tempDir := t.TempDir()
	_, currentFilePath, _, _ := runtime.Caller(0)

	sourceDirectoryPath := filepath.Join(filepath.Dir(currentFilePath), "testdata", "service", "cloudformation", "examplecompany-exampleservice-exampleresource")
	targetDirectoryPath := filepath.Join(tempDir, "examplecompany-exampleservice-exampleresource")
	targetZipFilePath := filepath.Join(tempDir, "examplecompany-exampleservice-exampleresource.zip")

	err := os.Mkdir(targetDirectoryPath, 0777)

	if err != nil {
		t.Fatalf("error making directory %s: %s", targetDirectoryPath, err)
	}

	sourceRpdkConfigFilePath := filepath.Join(sourceDirectoryPath, ".rpdk-config")
	targetRpdkConfigFilePath := filepath.Join(targetDirectoryPath, ".rpdk-config")

	err = testAccAwsCloudformationTypeCopyFileWithTypeNameReplacement(sourceRpdkConfigFilePath, targetRpdkConfigFilePath, typeName)

	if err != nil {
		t.Fatal(err)
	}

	sourceBinDirectoryPath := filepath.Join(sourceDirectoryPath, "bin")
	sourceBinHandlerFilePath := filepath.Join(sourceBinDirectoryPath, "handler")
	targetHandlerZipFilePath := filepath.Join(targetDirectoryPath, "handler.zip")

	err = testAccAwsCloudformationTypeBuildHandlerZip(sourceBinHandlerFilePath, targetHandlerZipFilePath)

	if errors.Is(err, os.ErrNotExist) {
		t.Skipf("%s does not exist. To generate locally: cd %s && make build", sourceBinHandlerFilePath, sourceDirectoryPath)
	}

	if err != nil {
		t.Fatal(err)
	}

	sourceSchemaFilePath := filepath.Join(sourceDirectoryPath, "examplecompany-exampleservice-exampleresource.json")
	targetSchemaFilePath := filepath.Join(targetDirectoryPath, "schema.json")

	err = testAccAwsCloudformationTypeCopyFileWithTypeNameReplacement(sourceSchemaFilePath, targetSchemaFilePath, typeName)

	if err != nil {
		t.Fatal(err)
	}

	err = testAccAwsCloudformationTypeBuildDirectoryZip(targetDirectoryPath, targetZipFilePath)

	if err != nil {
		t.Fatal(err)
	}

	return targetZipFilePath
}

func testAccAwsCloudformationTypeBuildDirectoryZip(sourceDirectoryPath string, targetZipFilePath string) error {
	targetZipFile, err := os.Create(targetZipFilePath)

	if err != nil {
		return fmt.Errorf("error creating file %s: %w", targetZipFilePath, err)
	}

	targetZipFileWriter := zip.NewWriter(targetZipFile)

	targetDirectoryFiles, err := os.ReadDir(sourceDirectoryPath)

	if err != nil {
		return fmt.Errorf("error listing %s directory files: %w", sourceDirectoryPath, err)
	}

	for _, targetDirectoryFile := range targetDirectoryFiles {
		sourceFilePath := filepath.Join(sourceDirectoryPath, targetDirectoryFile.Name())

		sourceFile, err := os.Open(sourceFilePath)

		if err != nil {
			return fmt.Errorf("error reading file %s: %w", sourceFilePath, err)
		}

		targetFile, err := targetZipFileWriter.Create(sourceFilePath)

		if err != nil {
			return fmt.Errorf("error creating file %s in zip: %w", sourceFilePath, err)
		}

		_, err = io.Copy(targetFile, sourceFile)

		if err != nil {
			return fmt.Errorf("error copying file %s into zip: %w", sourceFilePath, err)
		}
	}

	err = targetZipFileWriter.Close()

	if err != nil {
		return fmt.Errorf("error writing file %s: %w", targetZipFilePath, err)
	}

	return nil
}

func testAccAwsCloudformationTypeBuildHandlerZip(sourceBinHandlerFilePath string, targetZipFilePath string) error {
	sourceBinHandlerFile, err := os.Open(sourceBinHandlerFilePath)

	if err != nil {
		return fmt.Errorf("error reading file %s: %w", sourceBinHandlerFilePath, err)
	}

	targetZipFile, err := os.Create(targetZipFilePath)

	if err != nil {
		return fmt.Errorf("error creating file %s: %w", targetZipFilePath, err)
	}

	targetZipWriter := zip.NewWriter(targetZipFile)

	targetHandlerFile, err := targetZipWriter.Create("handler")

	if err != nil {
		return fmt.Errorf("error creating file %s in zip: %w", "handler", err)
	}

	_, err = io.Copy(targetHandlerFile, sourceBinHandlerFile)

	if err != nil {
		return fmt.Errorf("error copying file %s into zip: %w", sourceBinHandlerFilePath, err)
	}

	err = targetZipWriter.Close()

	if err != nil {
		return fmt.Errorf("error writing zip %s: %w", targetZipFilePath, err)
	}

	err = targetZipFile.Close()

	if err != nil {
		return fmt.Errorf("error writing file %s: %w", targetZipFilePath, err)
	}

	return nil
}

func testAccAwsCloudformationTypeCopyFileWithTypeNameReplacement(sourceFilePath string, targetFilePath string, targetTypeName string) error {
	sourceTypeName := "ExampleCompany::ExampleService::ExampleResource"
	sourceFile, err := os.ReadFile(sourceFilePath)

	if err != nil {
		return fmt.Errorf("error reading file %s: %w", sourceFilePath, err)
	}

	err = os.WriteFile(targetFilePath, []byte(strings.ReplaceAll(string(sourceFile), sourceTypeName, targetTypeName)), 0644)

	if err != nil {
		return fmt.Errorf("error writing file %s: %w", targetFilePath, err)
	}

	return nil
}

func testAccCloudformationTypeConfigBase(rName string, zipPath string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_bucket_object" "test" {
  bucket = aws_s3_bucket.test.bucket
  key    = "test"
  source = %[2]q
}
`, rName, zipPath)
}

func testAccCloudformationTypeConfigExecutionRoleArn(rName string, zipPath string, typeName string) string {
	return composeConfig(
		testAccCloudformationTypeConfigBase(rName, zipPath),
		fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "cloudformation.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_cloudformation_type" "test" {
  execution_role_arn     = aws_iam_role.test.arn
  schema_handler_package = "s3://${aws_s3_bucket_object.test.bucket}/${aws_s3_bucket_object.test.key}"
  type                   = "RESOURCE"
  type_name              = %[2]q
}
`, rName, typeName))
}

func testAccCloudformationTypeConfigLoggingConfig(rName string, zipPath string, typeName string) string {
	return composeConfig(
		testAccCloudformationTypeConfigBase(rName, zipPath),
		fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "cloudformation.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_cloudformation_type" "test" {
  schema_handler_package = "s3://${aws_s3_bucket_object.test.bucket}/${aws_s3_bucket_object.test.key}"
  type                   = "RESOURCE"
  type_name              = %[2]q

  logging_config {
    log_group_name = aws_cloudwatch_log_group.test.name
    log_role_arn   = aws_iam_role.test.arn
  }
}
`, rName, typeName))
}

func testAccCloudformationTypeConfigTypeName(rName string, zipPath string, typeName string) string {
	return composeConfig(
		testAccCloudformationTypeConfigBase(rName, zipPath),
		fmt.Sprintf(`
resource "aws_cloudformation_type" "test" {
  schema_handler_package = "s3://${aws_s3_bucket_object.test.bucket}/${aws_s3_bucket_object.test.key}"
  type                   = "RESOURCE"
  type_name              = %[1]q
}
`, typeName))
}
