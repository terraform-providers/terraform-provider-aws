package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudhsmv2"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudhsmv2/finder"
)

func init() {
	resource.AddTestSweepers("aws_cloudhsm_v2_cluster", &resource.Sweeper{
		Name: "aws_cloudhsm_v2_cluster",
		F:    testSweepCloudhsmv2Clusters,
	})
}

func testSweepCloudhsmv2Clusters(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).cloudhsmv2conn
	sweepResources := make([]*testSweepResource, 0)
	var errors *multierror.Error

	input := &cloudhsmv2.DescribeClustersInput{}

	err = conn.DescribeClustersPages(input, func(page *cloudhsmv2.DescribeClustersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, cluster := range page.Clusters {
			if cluster == nil {
				continue
			}

			r := resourceAwsCloudHsmV2Cluster()
			d := r.Data(nil)
			d.SetId(aws.StringValue(cluster.ClusterId))
			sweepResources = append(sweepResources, NewTestSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errors = multierror.Append(errors, fmt.Errorf("error describing CloudHSMv2 Clusters: %w", err))
	}

	if len(sweepResources) > 0 {
		// any errors didn't prevent gathering of some work, so do it
		if err := testSweepResourceOrchestrator(sweepResources); err != nil {
			errors = multierror.Append(errors, fmt.Errorf("error sweeping CloudHSMv2 Clusters for %s: %w", region, err))
		}
	}

	if testSweepSkipSweepError(errors.ErrorOrNil()) {
		log.Printf("[WARN] Skipping CloudHSMv2 Cluster sweep for %s: %s", region, errors)
		return nil
	}

	return errors.ErrorOrNil()
}

func TestAccAWSCloudHsmV2Cluster_basic(t *testing.T) {
	resourceName := "aws_cloudhsm_v2_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, cloudhsmv2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudHsmV2ClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudHsmV2ClusterConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCloudHsmV2ClusterExists(resourceName),
					resource.TestMatchResourceAttr(resourceName, "cluster_id", regexp.MustCompile(`^cluster-.+`)),
					resource.TestCheckResourceAttr(resourceName, "cluster_state", cloudhsmv2.ClusterStateUninitialized),
					resource.TestCheckResourceAttr(resourceName, "hsm_type", "hsm1.medium"),
					resource.TestMatchResourceAttr(resourceName, "security_group_id", regexp.MustCompile(`^sg-.+`)),
					resource.TestCheckResourceAttr(resourceName, "source_backup_identifier", ""),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_ids.*", "aws_subnet.test.0", "id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_ids.*", "aws_subnet.test.1", "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", "aws_vpc.test", "id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"cluster_certificates"},
			},
		},
	})
}

func TestAccAWSCloudHsmV2Cluster_disappears(t *testing.T) {
	resourceName := "aws_cloudhsm_v2_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, cloudhsmv2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudHsmV2ClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudHsmV2ClusterConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloudHsmV2ClusterExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCloudHsmV2Cluster(), resourceName),
					// Verify Delete error handling
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCloudHsmV2Cluster(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSCloudHsmV2Cluster_Tags(t *testing.T) {
	resourceName := "aws_cloudhsm_v2_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, cloudhsmv2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudHsmV2ClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudHsmV2ClusterConfigTags2("key1", "value1", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloudHsmV2ClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"cluster_certificates"},
			},
			{
				Config: testAccAWSCloudHsmV2ClusterConfigTags1("key1", "value1updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloudHsmV2ClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
				),
			},
			{
				Config: testAccAWSCloudHsmV2ClusterConfigTags2("key1", "value1updated", "key3", "value3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloudHsmV2ClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key3", "value3"),
				),
			},
		},
	})
}

func testAccAWSCloudHsmV2ClusterConfigBase() string {
	return `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = element(data.aws_availability_zones.available.names, count.index)
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  vpc_id            = aws_vpc.test.id
}
`
}

func testAccAWSCloudHsmV2ClusterConfig() string {
	return composeConfig(testAccAWSCloudHsmV2ClusterConfigBase(), `
resource "aws_cloudhsm_v2_cluster" "test" {
  hsm_type   = "hsm1.medium"
  subnet_ids = aws_subnet.test[*].id
}
`)
}

func testAccAWSCloudHsmV2ClusterConfigTags1(tagKey1, tagValue1 string) string {
	return composeConfig(testAccAWSCloudHsmV2ClusterConfigBase(), fmt.Sprintf(`
resource "aws_cloudhsm_v2_cluster" "test" {
  hsm_type   = "hsm1.medium"
  subnet_ids = aws_subnet.test[*].id

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccAWSCloudHsmV2ClusterConfigTags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return composeConfig(testAccAWSCloudHsmV2ClusterConfigBase(), fmt.Sprintf(`
resource "aws_cloudhsm_v2_cluster" "test" {
  hsm_type   = "hsm1.medium"
  subnet_ids = aws_subnet.test[*].id

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccCheckAWSCloudHsmV2ClusterDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cloudhsmv2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudhsm_v2_cluster" {
			continue
		}
		cluster, err := finder.Cluster(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if cluster != nil && aws.StringValue(cluster.State) != cloudhsmv2.ClusterStateDeleted {
			return fmt.Errorf("CloudHSM cluster still exists %s", cluster)
		}
	}

	return nil
}

func testAccCheckAWSCloudHsmV2ClusterExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).cloudhsmv2conn
		it, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		_, err := finder.Cluster(conn, it.Primary.ID)

		if err != nil {
			return fmt.Errorf("CloudHSM cluster not found: %s", err)
		}

		return nil
	}
}
