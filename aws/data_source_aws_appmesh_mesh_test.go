package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAWSAppmeshMesh_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appmesh_mesh.test"
	dataSourceName := "data.aws_appmesh_mesh.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appmesh.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, appmesh.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppmeshMeshDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsAppmeshMeshDataSourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "created_date", dataSourceName, "created_date"),
					resource.TestCheckResourceAttrPair(resourceName, "last_updated_date", dataSourceName, "last_updated_date"),
					resource.TestCheckResourceAttrPair(resourceName, "mesh_owner", dataSourceName, "mesh_owner"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_owner", dataSourceName, "resource_owner"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.egress_filter.0.type", dataSourceName, "spec.0.egress_filter.0.type"),
					resource.TestCheckResourceAttrPair(resourceName, "tags", dataSourceName, "tags"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSAppmeshMesh_meshOwner(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appmesh_mesh.test"
	dataSourceName := "data.aws_appmesh_mesh.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appmesh.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, appmesh.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppmeshMeshDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsAppmeshMeshDataSourceConfig_meshOwner(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "created_date", dataSourceName, "created_date"),
					resource.TestCheckResourceAttrPair(resourceName, "last_updated_date", dataSourceName, "last_updated_date"),
					resource.TestCheckResourceAttrPair(resourceName, "mesh_owner", dataSourceName, "mesh_owner"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_owner", dataSourceName, "resource_owner"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.egress_filter.0.type", dataSourceName, "spec.0.egress_filter.0.type"),
					resource.TestCheckResourceAttrPair(resourceName, "tags", dataSourceName, "tags"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSAppmeshMesh_specAndTagsSet(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appmesh_mesh.test"
	dataSourceName := "data.aws_appmesh_mesh.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appmesh.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, appmesh.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppmeshMeshDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsAppmeshMeshDataSourceConfig_specAndTagsSet(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "created_date", dataSourceName, "created_date"),
					resource.TestCheckResourceAttrPair(resourceName, "last_updated_date", dataSourceName, "last_updated_date"),
					resource.TestCheckResourceAttrPair(resourceName, "mesh_owner", dataSourceName, "mesh_owner"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_owner", dataSourceName, "resource_owner"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.egress_filter.0.type", dataSourceName, "spec.0.egress_filter.0.type"),
					resource.TestCheckResourceAttrPair(resourceName, "tags", dataSourceName, "tags"),
				),
			},
		},
	})
}

func testAccCheckAwsAppmeshMeshDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}

data "aws_appmesh_mesh" "test" {
  name = aws_appmesh_mesh.test.name
}
`, rName)
}

func testAccCheckAwsAppmeshMeshDataSourceConfig_meshOwner(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}

data "aws_appmesh_mesh" "test" {
  name       = aws_appmesh_mesh.test.name
  mesh_owner = data.aws_caller_identity.current.account_id
}
`, rName)
}

func testAccCheckAwsAppmeshMeshDataSourceConfig_specAndTagsSet(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_appmesh_mesh" "test" {
  name = %[1]q

  spec {
    egress_filter {
      type = "DROP_ALL"
    }
  }

  tags = {
    foo  = "bar"
    good = "bad"
  }
}

data "aws_appmesh_mesh" "test" {
  name = aws_appmesh_mesh.test.name
}
`, rName)
}
