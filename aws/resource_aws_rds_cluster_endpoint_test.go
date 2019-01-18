package aws

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSRDSClusterEndpoint_basic(t *testing.T) {
	rInt := acctest.RandInt()
	var customReaderEndpoint rds.DBClusterEndpoint
	var customEndpoint rds.DBClusterEndpoint
	readerResourceName := "aws_rds_cluster_endpoint.reader"
	defaultResourceName := "aws_rds_cluster_endpoint.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterEndpointConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRDSClusterEndpointExists(readerResourceName, &customReaderEndpoint),
					testAccCheckAWSRDSClusterEndpointAttributes(&customReaderEndpoint),
					testAccCheckAWSRDSClusterEndpointExists(defaultResourceName, &customEndpoint),
					testAccCheckAWSRDSClusterEndpointAttributes(&customEndpoint),
					testAccMatchResourceAttrRegionalARN(readerResourceName, "arn", "rds", regexp.MustCompile(`cluster-endpoint:.+`)),
					resource.TestCheckResourceAttrSet(readerResourceName, "endpoint"),
					testAccMatchResourceAttrRegionalARN(defaultResourceName, "arn", "rds", regexp.MustCompile(`cluster-endpoint:.+`)),
					resource.TestCheckResourceAttrSet(defaultResourceName, "endpoint"),
				),
			},
			{
				ResourceName:      "aws_rds_cluster_endpoint.reader",
				ImportState:       true,
				ImportStateVerify: true,
			},

			{
				ResourceName:      "aws_rds_cluster_endpoint.default",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAWSRDSClusterEndpointAttributes(v *rds.DBClusterEndpoint) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if aws.StringValue(v.Endpoint) == "" {
			return fmt.Errorf("empty endpoint domain")
		}

		if aws.StringValue(v.CustomEndpointType) != "READER" &&
			aws.StringValue(v.CustomEndpointType) != "ANY" {
			return fmt.Errorf("Incorrect endpoint type: expected: READER or ANY, got: %s", aws.StringValue(v.CustomEndpointType))
		}

		if len(v.StaticMembers) == 0 && len(v.ExcludedMembers) == 0 {
			return fmt.Errorf("Empty members")
		}

		for _, m := range aws.StringValueSlice(v.StaticMembers) {
			if !strings.HasPrefix(m, "tf-aurora-cluster-instance") {
				return fmt.Errorf("Incorrect StaticMember Cluster Instance Identifier prefix:\nexpected: %s\ngot: %s", "tf-aurora-cluster-instance", m)
			}
		}

		for _, m := range aws.StringValueSlice(v.ExcludedMembers) {
			if !strings.HasPrefix(m, "tf-aurora-cluster-instance") {
				return fmt.Errorf("Incorrect ExcludeMember Cluster Instance Identifier prefix:\nexpected: %s\ngot: %s", "tf-aurora-cluster-instance", m)
			}
		}

		return nil
	}
}

func testAccCheckAWSClusterEndpointDestroy(s *terraform.State) error {
	return testAccCheckAWSClusterEndpointDestroyWithProvider(s, testAccProvider)
}

func testAccCheckAWSClusterEndpointDestroyWithProvider(s *terraform.State, provider *schema.Provider) error {
	conn := provider.Meta().(*AWSClient).rdsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_rds_cluster_endpoint" {
			continue
		}

		// Try to find the Group
		var err error
		resp, err := conn.DescribeDBClusterEndpoints(
			&rds.DescribeDBClusterEndpointsInput{
				DBClusterEndpointIdentifier: aws.String(rs.Primary.ID),
			})

		if err == nil {
			if len(resp.DBClusterEndpoints) != 0 &&
				*resp.DBClusterEndpoints[0].DBClusterEndpointIdentifier == rs.Primary.ID {
				return fmt.Errorf("DB Cluster Endpoint %s still exists", rs.Primary.ID)
			}
		}

		// Return nil if the cluster is already destroyed
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "DBClusterNotFoundFault" {
				return nil
			}
		}

		return err
	}

	return nil
}
func testAccCheckAWSRDSClusterEndpointExists(resourceName string, endpoint *rds.DBClusterEndpoint) resource.TestCheckFunc {
	return testAccCheckAWSRDSClusterEndpointExistsWithProvider(resourceName, endpoint, testAccProvider)
}

func testAccCheckAWSRDSClusterEndpointExistsWithProvider(resourceName string, endpoint *rds.DBClusterEndpoint, provider *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("DBClusterEndpoint ID is not set")
		}

		conn := provider.Meta().(*AWSClient).rdsconn

		response, err := conn.DescribeDBClusterEndpoints(&rds.DescribeDBClusterEndpointsInput{
			DBClusterEndpointIdentifier: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if len(response.DBClusterEndpoints) != 1 ||
			*response.DBClusterEndpoints[0].DBClusterEndpointIdentifier != rs.Primary.ID {
			return fmt.Errorf("DBClusterEndpoint not found")
		}

		*endpoint = *response.DBClusterEndpoints[0]
		return nil
	}
}

func testAccAWSClusterEndpointConfig(n int) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "default" {
  cluster_identifier = "tf-aurora-cluster-%d"
  availability_zones = ["us-west-2a","us-west-2b","us-west-2c"]
  database_name = "mydb"
  master_username = "foo"
  master_password = "mustbeeightcharaters"
  db_cluster_parameter_group_name = "default.aurora5.6"
  skip_final_snapshot = true
}

resource "aws_rds_cluster_instance" "test1" {
  apply_immediately = true
  cluster_identifier = "${aws_rds_cluster.default.id}"
  identifier = "tf-aurora-cluster-instance-test1-%d"
  instance_class = "db.t2.small"
}

resource "aws_rds_cluster_instance" "test2" {
  apply_immediately = true
  cluster_identifier = "${aws_rds_cluster.default.id}"
  identifier = "tf-aurora-cluster-instance-test2-%d"
  instance_class = "db.t2.small"
}

resource "aws_rds_cluster_endpoint" "reader" {
  cluster_identifier = "${aws_rds_cluster.default.id}"
  cluster_endpoint_identifier = "reader-%d"
  custom_endpoint_type = "READER"

  static_members = ["${aws_rds_cluster_instance.test2.id}"]
}

resource "aws_rds_cluster_endpoint" "default" {
  cluster_identifier = "${aws_rds_cluster.default.id}"
  cluster_endpoint_identifier = "default-%d"
  custom_endpoint_type = "ANY"

  excluded_members = ["${aws_rds_cluster_instance.test2.id}"]
}
`, n, n, n, n, n)
}
