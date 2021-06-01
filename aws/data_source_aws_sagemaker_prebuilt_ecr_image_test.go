package aws

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSSageMakerPrebuiltECRImage_basic(t *testing.T) {
	expectedID := sageMakerPrebuiltECRImageIDByRegion_FactorMachines[testAccGetRegion()]

	dataSourceName := "data.aws_sagemaker_prebuilt_ecr_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsSageMakerPrebuiltECRImageConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "id", expectedID),
					resource.TestCheckResourceAttr(dataSourceName, "registry_id", expectedID),
					resource.TestCheckResourceAttr(dataSourceName, "registry_path", dataSourceAwsSageMakerPrebuiltECRImageCreatePath(expectedID, testAccGetRegion(), testAccGetPartitionDNSSuffix(), "kmeans", "1")),
				),
			},
		},
	})
}

func TestAccAWSSageMakerPrebuiltECRImage_region(t *testing.T) {
	expectedID := sageMakerPrebuiltECRImageIDByRegion_SparkML[testAccGetRegion()]

	dataSourceName := "data.aws_sagemaker_prebuilt_ecr_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsSageMakerPrebuiltECRImageExplicitRegionConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "id", expectedID),
					resource.TestCheckResourceAttr(dataSourceName, "registry_id", expectedID),
					resource.TestCheckResourceAttr(dataSourceName, "registry_path", dataSourceAwsSageMakerPrebuiltECRImageCreatePath(expectedID, testAccGetRegion(), testAccGetPartitionDNSSuffix(), "sagemaker-scikit-learn", "2.2-1.0.11.0")),
				),
			},
		},
	})
}

const testAccCheckAwsSageMakerPrebuiltECRImageConfig = `
data "aws_sagemaker_prebuilt_ecr_image" "test" {
  repository_name = "kmeans"
}
`

const testAccCheckAwsSageMakerPrebuiltECRImageExplicitRegionConfig = `
data "aws_region" "current" {}

data "aws_sagemaker_prebuilt_ecr_image" "test" {
  repository_name = "sagemaker-scikit-learn"
  image_tag       = "2.2-1.0.11.0"
  region          = data.aws_region.current.name
}
`
