package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSEMRContainersVirtualClusterDataSource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceResourceName := "data.aws_emrcontainers_virtual_cluster.test"
	resourceName := "aws_emrcontainers_virtual_cluster.test"

	testExternalProviders := map[string]resource.ExternalProvider{
		"kubernetes": {
			Source:            "hashicorp/kubernetes",
			VersionConstraint: "~> 2.3",
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckAWSEks(t) },
		ErrorCheck:        testAccErrorCheck(t, eks.EndpointsID),
		Providers:         testAccProviders,
		ExternalProviders: testExternalProviders,
		CheckDestroy:      testAccCheckAWSEksClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEMRContainersVirtualClusterDataSourceConfig_Basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "id", dataSourceResourceName, "id"),
					resource.TestCheckResourceAttr(dataSourceResourceName, "container_provider.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "container_provider.0.id", dataSourceResourceName, "container_provider.0.id"),
					resource.TestCheckResourceAttr(dataSourceResourceName, "container_provider.0.info.#", "1"),
					resource.TestCheckResourceAttr(dataSourceResourceName, "container_provider.0.info.0.eks_info.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "container_provider.0.info.0.eks_info.0.namespace", dataSourceResourceName, "container_provider.0.info.0.eks_info.0.namespace"),
					resource.TestCheckResourceAttrPair(resourceName, "container_provider.0.type", dataSourceResourceName, "container_provider.0.type"),
					resource.TestCheckResourceAttrPair(resourceName, "created_at", dataSourceResourceName, "created_at"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceResourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "state", dataSourceResourceName, "state"),
				),
			},
		},
	})
}

func testAccAWSEMRContainersVirtualClusterDataSourceConfig_Basic(rName string) string {
	return composeConfig(testAccAwsEMRContainersVirtualClusterBasicConfig(rName), `
data "aws_emrcontainers_virtual_cluster" "test" {
  id = aws_emrcontainers_virtual_cluster.test.id
}
`)
}
