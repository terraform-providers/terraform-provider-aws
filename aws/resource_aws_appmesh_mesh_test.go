package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_appmesh_mesh", &resource.Sweeper{
		Name: "aws_appmesh_mesh",
		F:    testSweepAppmeshMeshes,
		Dependencies: []string{
			"aws_appmesh_virtual_service",
			"aws_appmesh_virtual_router",
			"aws_appmesh_virtual_node",
			"aws_appmesh_virtual_gateway",
		},
	})
}

func testSweepAppmeshMeshes(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).appmeshconn

	err = conn.ListMeshesPages(&appmesh.ListMeshesInput{}, func(page *appmesh.ListMeshesOutput, isLast bool) bool {
		if page == nil {
			return !isLast
		}

		for _, mesh := range page.Meshes {
			name := aws.StringValue(mesh.MeshName)

			input := &appmesh.DeleteMeshInput{
				MeshName: aws.String(name),
			}

			log.Printf("[INFO] Deleting Appmesh Mesh: %s", name)
			_, err := conn.DeleteMesh(input)

			if err != nil {
				log.Printf("[ERROR] Error deleting Appmesh Mesh (%s): %s", name, err)
			}
		}

		return !isLast
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Appmesh Mesh sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("error retrieving Appmesh Meshes: %s", err)
	}

	return nil
}

func testAccAwsAppmeshMesh_basic(t *testing.T) {
	var mesh appmesh.MeshData
	resourceName := "aws_appmesh_mesh.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appmesh.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppmeshMeshDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppmeshMeshConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshMeshExists(resourceName, &mesh),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					testAccCheckResourceAttrAccountID(resourceName, "mesh_owner"),
					testAccCheckResourceAttrAccountID(resourceName, "resource_owner"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "appmesh", regexp.MustCompile(`mesh/.+`)),
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

func testAccAwsAppmeshMesh_egressFilter(t *testing.T) {
	var mesh appmesh.MeshData
	resourceName := "aws_appmesh_mesh.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appmesh.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppmeshMeshDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppmeshMeshConfig_egressFilter(rName, "ALLOW_ALL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshMeshExists(resourceName, &mesh),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "spec.0.egress_filter.0.type", "ALLOW_ALL"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAppmeshMeshConfig_egressFilter(rName, "DROP_ALL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshMeshExists(resourceName, &mesh),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "spec.0.egress_filter.0.type", "DROP_ALL"),
				),
			},
			{
				PlanOnly: true,
				Config:   testAccAppmeshMeshConfig_basic(rName),
			},
		},
	})
}

func testAccAwsAppmeshMesh_tags(t *testing.T) {
	var mesh appmesh.MeshData
	resourceName := "aws_appmesh_mesh.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appmesh.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppmeshMeshDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppmeshMeshConfigWithTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshMeshExists(resourceName, &mesh),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "bar"),
					resource.TestCheckResourceAttr(resourceName, "tags.good", "bad"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAppmeshMeshConfigWithUpdateTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshMeshExists(resourceName, &mesh),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.good", "bad2"),
					resource.TestCheckResourceAttr(resourceName, "tags.fizz", "buzz"),
				),
			},
			{
				Config: testAccAppmeshMeshConfigWithRemoveTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshMeshExists(resourceName, &mesh),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
				),
			},
		},
	})
}

func testAccCheckAppmeshMeshDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).appmeshconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appmesh_mesh" {
			continue
		}

		_, err := conn.DescribeMesh(&appmesh.DescribeMeshInput{
			MeshName: aws.String(rs.Primary.Attributes["name"]),
		})
		if isAWSErr(err, "NotFoundException", "") {
			continue
		}
		if err != nil {
			return err
		}
		return fmt.Errorf("still exist.")
	}

	return nil
}

func testAccCheckAppmeshMeshExists(name string, v *appmesh.MeshData) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).appmeshconn

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		resp, err := conn.DescribeMesh(&appmesh.DescribeMeshInput{
			MeshName: aws.String(rs.Primary.Attributes["name"]),
		})
		if err != nil {
			return err
		}

		*v = *resp.Mesh

		return nil
	}
}

func testAccAppmeshMeshConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}
`, rName)
}

func testAccAppmeshMeshConfig_egressFilter(rName, egressFilterType string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q

  spec {
    egress_filter {
      type = %[2]q
    }
  }
}
`, rName, egressFilterType)
}

func testAccAppmeshMeshConfigWithTags(rName string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q

  tags = {
    foo  = "bar"
    good = "bad"
  }
}
`, rName)
}

func testAccAppmeshMeshConfigWithUpdateTags(rName string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q

  tags = {
    foo  = "bar"
    good = "bad2"
    fizz = "buzz"
  }
}
`, rName)
}

func testAccAppmeshMeshConfigWithRemoveTags(rName string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q

  tags = {
    foo = "bar"
  }
}
`, rName)
}
