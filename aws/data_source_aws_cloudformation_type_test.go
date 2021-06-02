package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAwsCloudformationTypeDataSource_Arn_Private(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	typeName := fmt.Sprintf("HashiCorp::TerraformAwsProvider::TfAccTest%s", acctest.RandString(8))
	zipPath := testAccAwsCloudformationTypeZipGenerator(t, typeName)
	resourceName := "aws_cloudformation_type.test"
	dataSourceName := "data.aws_cloudformation_type.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCloudformationTypeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCloudformationTypeDataSourceConfigArnPrivate(rName, zipPath, typeName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "deprecated_status", resourceName, "deprecated_status"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "documentation_url", resourceName, "documentation_url"),
					resource.TestCheckResourceAttrPair(dataSourceName, "execution_role_arn", resourceName, "execution_role_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "is_default_version", resourceName, "is_default_version"),
					resource.TestCheckResourceAttrPair(dataSourceName, "logging_config.#", resourceName, "logging_config.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "provisioning_type", resourceName, "provisioning_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "schema", resourceName, "schema"),
					resource.TestCheckResourceAttrPair(dataSourceName, "source_url", resourceName, "source_url"),
					resource.TestCheckResourceAttrPair(dataSourceName, "type", resourceName, "type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "type_name", resourceName, "type_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "visibility", resourceName, "visibility"),
				),
			},
		},
	})
}

func TestAccAwsCloudformationTypeDataSource_Arn_Public(t *testing.T) {
	dataSourceName := "data.aws_cloudformation_type.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCloudformationTypeDataSourceConfigArnPublic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceAttrRegionalARNNoAccount(dataSourceName, "arn", "cloudformation", "type/resource/AWS-Athena-WorkGroup"),
					resource.TestCheckResourceAttr(dataSourceName, "deprecated_status", cloudformation.DeprecatedStatusLive),
					resource.TestMatchResourceAttr(dataSourceName, "description", regexp.MustCompile(`.*`)),
					resource.TestCheckResourceAttr(dataSourceName, "documentation_url", ""),
					resource.TestCheckResourceAttr(dataSourceName, "is_default_version", "true"),
					resource.TestCheckResourceAttr(dataSourceName, "logging_config.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "provisioning_type", cloudformation.ProvisioningTypeFullyMutable),
					resource.TestMatchResourceAttr(dataSourceName, "schema", regexp.MustCompile(`^\{.*`)),
					resource.TestMatchResourceAttr(dataSourceName, "source_url", regexp.MustCompile(`^https://.+`)),
					resource.TestCheckResourceAttr(dataSourceName, "type", cloudformation.RegistryTypeResource),
					resource.TestCheckResourceAttr(dataSourceName, "type_name", "AWS::Athena::WorkGroup"),
					resource.TestCheckResourceAttr(dataSourceName, "visibility", cloudformation.VisibilityPublic),
				),
			},
		},
	})
}

func TestAccAwsCloudformationTypeDataSource_TypeName_Private(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	typeName := fmt.Sprintf("HashiCorp::TerraformAwsProvider::TfAccTest%s", acctest.RandString(8))
	zipPath := testAccAwsCloudformationTypeZipGenerator(t, typeName)
	resourceName := "aws_cloudformation_type.test"
	dataSourceName := "data.aws_cloudformation_type.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCloudformationTypeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCloudformationTypeDataSourceConfigTypeNamePrivate(rName, zipPath, typeName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "deprecated_status", resourceName, "deprecated_status"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "documentation_url", resourceName, "documentation_url"),
					resource.TestCheckResourceAttrPair(dataSourceName, "execution_role_arn", resourceName, "execution_role_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "is_default_version", resourceName, "is_default_version"),
					resource.TestCheckResourceAttrPair(dataSourceName, "logging_config.#", resourceName, "logging_config.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "provisioning_type", resourceName, "provisioning_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "schema", resourceName, "schema"),
					resource.TestCheckResourceAttrPair(dataSourceName, "source_url", resourceName, "source_url"),
					resource.TestCheckResourceAttrPair(dataSourceName, "type", resourceName, "type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "type_name", resourceName, "type_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "visibility", resourceName, "visibility"),
				),
			},
		},
	})
}

func TestAccAwsCloudformationTypeDataSource_TypeName_Public(t *testing.T) {
	dataSourceName := "data.aws_cloudformation_type.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCloudformationTypeDataSourceConfigTypeNamePublic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceAttrRegionalARNNoAccount(dataSourceName, "arn", "cloudformation", "type/resource/AWS-Athena-WorkGroup"),
					resource.TestCheckResourceAttr(dataSourceName, "deprecated_status", cloudformation.DeprecatedStatusLive),
					resource.TestMatchResourceAttr(dataSourceName, "description", regexp.MustCompile(`.*`)),
					resource.TestCheckResourceAttr(dataSourceName, "documentation_url", ""),
					resource.TestCheckResourceAttr(dataSourceName, "is_default_version", "true"),
					resource.TestCheckResourceAttr(dataSourceName, "logging_config.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "provisioning_type", cloudformation.ProvisioningTypeFullyMutable),
					resource.TestMatchResourceAttr(dataSourceName, "schema", regexp.MustCompile(`^\{.*`)),
					resource.TestMatchResourceAttr(dataSourceName, "source_url", regexp.MustCompile(`^https://.+`)),
					resource.TestCheckResourceAttr(dataSourceName, "type", cloudformation.RegistryTypeResource),
					resource.TestCheckResourceAttr(dataSourceName, "type_name", "AWS::Athena::WorkGroup"),
					resource.TestCheckResourceAttr(dataSourceName, "visibility", cloudformation.VisibilityPublic),
				),
			},
		},
	})
}

func testAccCloudformationTypeConfigPrivateBase(rName string, zipPath string, typeName string) string {
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

resource "aws_cloudformation_type" "test" {
  schema_handler_package = "s3://${aws_s3_bucket_object.test.bucket}/${aws_s3_bucket_object.test.key}"
  type                   = "RESOURCE"
  type_name              = %[3]q
}
`, rName, zipPath, typeName)
}

func testAccAwsCloudformationTypeDataSourceConfigArnPrivate(rName string, zipPath string, typeName string) string {
	return composeConfig(
		testAccCloudformationTypeConfigPrivateBase(rName, zipPath, typeName),
		`
data "aws_cloudformation_type" "test" {
  arn = aws_cloudformation_type.test.arn
}
`)
}

func testAccAwsCloudformationTypeDataSourceConfigArnPublic() string {
	return `
data "aws_partition" "current" {}

data "aws_region" "current" {}

data "aws_cloudformation_type" "test" {
  arn = "arn:${data.aws_partition.current.partition}:cloudformation:${data.aws_region.current.name}::type/resource/AWS-Athena-WorkGroup"
}
`
}

func testAccAwsCloudformationTypeDataSourceConfigTypeNamePrivate(rName string, zipPath string, typeName string) string {
	return composeConfig(
		testAccCloudformationTypeConfigPrivateBase(rName, zipPath, typeName),
		`
data "aws_cloudformation_type" "test" {
  type      = aws_cloudformation_type.test.type
  type_name = aws_cloudformation_type.test.type_name
}
`)
}

func testAccAwsCloudformationTypeDataSourceConfigTypeNamePublic() string {
	return `
data "aws_cloudformation_type" "test" {
  type      = "RESOURCE"
  type_name = "AWS::Athena::WorkGroup"
}
`
}
