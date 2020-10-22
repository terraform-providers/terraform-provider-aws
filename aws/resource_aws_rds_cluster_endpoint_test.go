package aws

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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
					resource.TestCheckResourceAttr(defaultResourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(readerResourceName, "tags.%", "0"),
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

func TestAccAWSRDSClusterEndpoint_tags(t *testing.T) {
	rInt := acctest.RandInt()
	var customReaderEndpoint rds.DBClusterEndpoint
	resourceName := "aws_rds_cluster_endpoint.reader"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterEndpointConfigTags1(rInt, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRDSClusterEndpointExists(resourceName, &customReaderEndpoint),
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
				Config: testAccAWSClusterEndpointConfigTags2(rInt, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRDSClusterEndpointExists(resourceName, &customReaderEndpoint),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSClusterEndpointConfigTags1(rInt, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRDSClusterEndpointExists(resourceName, &customReaderEndpoint),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
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

func testAccAWSClusterEndpointConfigBase(n int) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  engine                     = aws_rds_cluster.default.engine
  engine_version             = aws_rds_cluster.default.engine_version
  preferred_instance_classes = ["db.t3.small", "db.t2.small", "db.t3.medium"]
}

resource "aws_rds_cluster" "default" {
  cluster_identifier = "tf-aurora-cluster-%[1]d"
  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]
  database_name                   = "mydb"
  master_username                 = "foo"
  master_password                 = "mustbeeightcharaters"
  db_cluster_parameter_group_name = "default.aurora5.6"
  skip_final_snapshot             = true
}

resource "aws_rds_cluster_instance" "test1" {
  apply_immediately  = true
  cluster_identifier = aws_rds_cluster.default.id
  identifier         = "tf-aurora-cluster-instance-test1-%[1]d"
  instance_class     = data.aws_rds_orderable_db_instance.test.instance_class
}

resource "aws_rds_cluster_instance" "test2" {
  apply_immediately  = true
  cluster_identifier = aws_rds_cluster.default.id
  identifier         = "tf-aurora-cluster-instance-test2-%[1]d"
  instance_class     = data.aws_rds_orderable_db_instance.test.instance_class
}
`, n))
}

func testAccAWSClusterEndpointConfig(n int) string {
	return testAccAWSClusterEndpointConfigBase(n) + fmt.Sprintf(`
resource "aws_rds_cluster_endpoint" "reader" {
  cluster_identifier          = aws_rds_cluster.default.id
  cluster_endpoint_identifier = "reader-%[1]d"
  custom_endpoint_type        = "READER"

  static_members = [aws_rds_cluster_instance.test2.id]
}

resource "aws_rds_cluster_endpoint" "default" {
  cluster_identifier          = aws_rds_cluster.default.id
  cluster_endpoint_identifier = "default-%[1]d"
  custom_endpoint_type        = "ANY"

  excluded_members = [aws_rds_cluster_instance.test2.id]
}
`, n)
}

func testAccAWSClusterEndpointConfigTags1(n int, tagKey1, tagValue1 string) string {
	return testAccAWSClusterEndpointConfigBase(n) + fmt.Sprintf(`
resource "aws_rds_cluster_endpoint" "reader" {
  cluster_identifier          = aws_rds_cluster.default.id
  cluster_endpoint_identifier = "reader-%[1]d"
  custom_endpoint_type        = "READER"

  static_members = [aws_rds_cluster_instance.test2.id]

  tags = {
    %[2]q = %[3]q
  }
}
`, n, tagKey1, tagValue1)
}

func testAccAWSClusterEndpointConfigTags2(n int, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccAWSClusterEndpointConfigBase(n) + fmt.Sprintf(`
resource "aws_rds_cluster_endpoint" "reader" {
  cluster_identifier          = aws_rds_cluster.default.id
  cluster_endpoint_identifier = "reader-%[1]d"
  custom_endpoint_type        = "READER"

  static_members = [aws_rds_cluster_instance.test2.id]

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, n, tagKey1, tagValue1, tagKey2, tagValue2)
}
