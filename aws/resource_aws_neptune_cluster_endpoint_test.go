package aws

import (
	//"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/neptune/finder"
)

func TestAccAWSNeptuneClusterEndpoint_basic(t *testing.T) {
	var dbCluster neptune.DBClusterEndpoint
	rName := acctest.RandomWithPrefix("tf-acc")
	resourceName := "aws_neptune_cluster_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, neptune.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterEndpointConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterEndpointExists(resourceName, &dbCluster),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "rds", regexp.MustCompile(`cluster-endpoint:.+`)),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "READER"),
					resource.TestCheckResourceAttr(resourceName, "cluster_endpoint_identifier", rName),
					resource.TestCheckResourceAttrPair(resourceName, "cluster_identifier", "aws_neptune_cluster.test", "cluster_identifier"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "static_members.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "excluded_members.#", "0"),
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

func TestAccAWSNeptuneClusterEndpoint_tags(t *testing.T) {
	if testAccGetPartition() == "aws-us-gov" {
		t.Skip("Neptune Cluster Endpoint tags are not supported in GovCloud partition")
	}

	var v neptune.DBClusterEndpoint
	rName := acctest.RandomWithPrefix("tf-acc")
	resourceName := "aws_neptune_cluster_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, neptune.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterEndpointConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterEndpointExists(resourceName, &v),
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
				Config: testAccAWSNeptuneClusterEndpointConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterEndpointExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSNeptuneClusterEndpointConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterEndpointExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSNeptuneClusterEndpoint_disappears(t *testing.T) {
	var dbCluster neptune.DBClusterEndpoint
	rName := acctest.RandomWithPrefix("tf-acc")
	resourceName := "aws_neptune_cluster_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, neptune.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterEndpointConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterEndpointExists(resourceName, &dbCluster),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsNeptuneClusterEndpoint(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSNeptuneClusterEndpoint_disappears_cluster(t *testing.T) {
	var dbCluster neptune.DBClusterEndpoint
	rName := acctest.RandomWithPrefix("tf-acc")
	resourceName := "aws_neptune_cluster_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, neptune.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterEndpointConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterEndpointExists(resourceName, &dbCluster),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsNeptuneCluster(), "aws_neptune_cluster.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSNeptuneClusterEndpointDestroy(s *terraform.State) error {
	return testAccCheckAWSNeptuneClusterEndpointDestroyWithProvider(s, testAccProvider)
}

func testAccCheckAWSNeptuneClusterEndpointDestroyWithProvider(s *terraform.State, provider *schema.Provider) error {
	conn := provider.Meta().(*AWSClient).neptuneconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_neptune_cluster_endpoint" {
			continue
		}

		_, err := finder.EndpointById(conn, rs.Primary.ID)
		// Return nil if the cluster is already destroyed
		if err != nil {
			if isAWSErr(err, neptune.ErrCodeDBClusterNotFoundFault, "") {
				return nil
			}
		}

		return err
	}

	return nil
}

func testAccCheckAWSNeptuneClusterEndpointExists(n string, v *neptune.DBClusterEndpoint) resource.TestCheckFunc {
	return testAccCheckAWSNeptuneClusterEndpointExistsWithProvider(n, v, func() *schema.Provider { return testAccProvider })
}

func testAccCheckAWSNeptuneClusterEndpointExistsWithProvider(n string, v *neptune.DBClusterEndpoint, providerF func() *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Neptune Instance ID is set")
		}

		provider := providerF()
		conn := provider.Meta().(*AWSClient).neptuneconn
		resp, err := finder.EndpointById(conn, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Neptune Cluster Endpoint (%s) not found: %w", rs.Primary.ID, err)
		}

		*v = *resp

		return nil
	}
}

func testAccAWSNeptuneClusterEndpointConfigBase(rName string) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
locals {
  availability_zone_names = slice(data.aws_availability_zones.available.names, 0, min(3, length(data.aws_availability_zones.available.names)))
}

resource "aws_neptune_cluster" "test" {
  cluster_identifier                   = %[1]q
  availability_zones                   = local.availability_zone_names
  engine                               = "neptune"
  neptune_cluster_parameter_group_name = "default.neptune1"
  skip_final_snapshot                  = true
}
`, rName))
}

func testAccAWSNeptuneClusterEndpointConfig(rName string) string {
	return composeConfig(testAccAWSNeptuneClusterEndpointConfigBase(rName), fmt.Sprintf(`
resource "aws_neptune_cluster_endpoint" "test" {
  cluster_identifier          = aws_neptune_cluster.test.cluster_identifier
  cluster_endpoint_identifier = %[1]q
  endpoint_type               = "READER"
}
`, rName))
}

func testAccAWSNeptuneClusterEndpointConfigTags1(rName, tagKey1, tagValue1 string) string {
	return composeConfig(testAccAWSNeptuneClusterEndpointConfigBase(rName), fmt.Sprintf(`
resource "aws_neptune_cluster_endpoint" "test" {
  cluster_identifier          = aws_neptune_cluster.test.cluster_identifier
  cluster_endpoint_identifier = %[1]q
  endpoint_type               = "READER"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccAWSNeptuneClusterEndpointConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return composeConfig(testAccAWSNeptuneClusterEndpointConfigBase(rName), fmt.Sprintf(`
resource "aws_neptune_cluster_endpoint" "test" {
  cluster_identifier          = aws_neptune_cluster.test.cluster_identifier
  cluster_endpoint_identifier = %[1]q
  endpoint_type               = "READER"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
