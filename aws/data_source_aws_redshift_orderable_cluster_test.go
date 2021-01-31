package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSRedshiftOrderableClusterDataSource_ClusterType(t *testing.T) {
	dataSourceName := "data.aws_redshift_orderable_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAWSRedshiftOrderableClusterPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRedshiftOrderableClusterDataSourceConfig_ClusterType("multi-node"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "cluster_type", "multi-node"),
				),
			},
		},
	})
}

func TestAccAWSRedshiftOrderableClusterDataSource_ClusterVersion(t *testing.T) {
	dataSourceName := "data.aws_redshift_orderable_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAWSRedshiftOrderableClusterPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRedshiftOrderableClusterDataSourceConfig_ClusterVersion("1.0"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "cluster_version", "1.0"),
				),
			},
		},
	})
}

func TestAccAWSRedshiftOrderableClusterDataSource_NodeType(t *testing.T) {
	dataSourceName := "data.aws_redshift_orderable_cluster.test"
	nodeType := "dc2.8xlarge"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAWSRedshiftOrderableClusterPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRedshiftOrderableClusterDataSourceConfig_NodeType(nodeType),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "node_type", nodeType),
				),
			},
		},
	})
}

func TestAccAWSRedshiftOrderableClusterDataSource_PreferredNodeTypes(t *testing.T) {
	dataSourceName := "data.aws_redshift_orderable_cluster.test"
	preferredNodeType := "dc2.8xlarge"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAWSRedshiftOrderableClusterPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRedshiftOrderableClusterDataSourceConfig_PreferredNodeTypes(preferredNodeType),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "node_type", preferredNodeType),
				),
			},
		},
	})
}

func testAccAWSRedshiftOrderableClusterPreCheck(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).redshiftconn

	input := &redshift.DescribeOrderableClusterOptionsInput{
		MaxRecords: aws.Int64(20),
	}

	_, err := conn.DescribeOrderableClusterOptions(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAWSRedshiftOrderableClusterDataSourceConfig_ClusterType(clusterType string) string {
	return fmt.Sprintf(`
data "aws_redshift_orderable_cluster" "test" {
  cluster_type         = %[1]q
  preferred_node_types = ["dc2.large", "ds2.xlarge"]
}
`, clusterType)
}

func testAccAWSRedshiftOrderableClusterDataSourceConfig_ClusterVersion(clusterVersion string) string {
	return fmt.Sprintf(`
data "aws_redshift_orderable_cluster" "test" {
  cluster_version      = %[1]q
  preferred_node_types = ["dc2.8xlarge", "ds2.8xlarge"]
}
`, clusterVersion)
}

func testAccAWSRedshiftOrderableClusterDataSourceConfig_NodeType(nodeType string) string {
	return fmt.Sprintf(`
data "aws_redshift_orderable_cluster" "test" {
  node_type            = %[1]q
  preferred_node_types = ["dc2.8xlarge", "ds2.8xlarge"]
}
`, nodeType)
}

func testAccAWSRedshiftOrderableClusterDataSourceConfig_PreferredNodeTypes(preferredNodeType string) string {
	return fmt.Sprintf(`
data "aws_redshift_orderable_cluster" "test" {
  preferred_node_types = [
    "non-existent",
    %[1]q,
    "try-again",
  ]
}
`, preferredNodeType)
}
