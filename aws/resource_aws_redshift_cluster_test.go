package aws

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_redshift_cluster", &resource.Sweeper{
		Name: "aws_redshift_cluster",
		F:    testSweepRedshiftClusters,
	})
}

func testSweepRedshiftClusters(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).redshiftconn

	err = conn.DescribeClustersPages(&redshift.DescribeClustersInput{}, func(resp *redshift.DescribeClustersOutput, isLast bool) bool {
		if len(resp.Clusters) == 0 {
			log.Print("[DEBUG] No Redshift clusters to sweep")
			return false
		}

		for _, c := range resp.Clusters {
			id := aws.StringValue(c.ClusterIdentifier)

			input := &redshift.DeleteClusterInput{
				ClusterIdentifier:        c.ClusterIdentifier,
				SkipFinalClusterSnapshot: aws.Bool(true),
			}
			_, err := conn.DeleteCluster(input)
			if err != nil {
				log.Printf("[ERROR] Failed deleting Redshift cluster (%s): %s", id, err)
			}
		}
		return !isLast
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Redshift Cluster sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving Redshift Clusters: %s", err)
	}
	return nil
}

func TestAccAWSRedshiftCluster_basic(t *testing.T) {
	var v redshift.Cluster

	ri := acctest.RandInt()
	config := testAccAWSRedshiftClusterConfig_basic(ri)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRedshiftClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftClusterExists("aws_redshift_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_redshift_cluster.default", "cluster_type", "single-node"),
					resource.TestCheckResourceAttr(
						"aws_redshift_cluster.default", "publicly_accessible", "true"),
					resource.TestMatchResourceAttr("aws_redshift_cluster.default", "dns_name", regexp.MustCompile(fmt.Sprintf("^tf-redshift-cluster-%d.*\\.redshift\\..*", ri))),
				),
			},
			{
				ResourceName:      "aws_redshift_cluster.default",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"final_snapshot_identifier",
					"master_password",
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccAWSRedshiftCluster_withFinalSnapshot(t *testing.T) {
	var v redshift.Cluster

	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRedshiftClusterSnapshot(rInt),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRedshiftClusterConfigWithFinalSnapshot(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftClusterExists("aws_redshift_cluster.default", &v),
				),
			},
			{
				ResourceName:      "aws_redshift_cluster.default",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"final_snapshot_identifier",
					"master_password",
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccAWSRedshiftCluster_kmsKey(t *testing.T) {
	var v redshift.Cluster

	resourceName := "aws_redshift_cluster.default"
	kmsResourceName := "aws_kms_key.foo"

	ri := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRedshiftClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRedshiftClusterConfig_kmsKey(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "cluster_type", "single-node"),
					resource.TestCheckResourceAttr(resourceName, "publicly_accessible", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", kmsResourceName, "arn"),
				),
			},
			{
				ResourceName:      "aws_redshift_cluster.default",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"final_snapshot_identifier",
					"master_password",
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccAWSRedshiftCluster_enhancedVpcRoutingEnabled(t *testing.T) {
	var v redshift.Cluster

	ri := acctest.RandInt()
	preConfig := testAccAWSRedshiftClusterConfig_enhancedVpcRoutingEnabled(ri)
	postConfig := testAccAWSRedshiftClusterConfig_enhancedVpcRoutingDisabled(ri)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRedshiftClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftClusterExists("aws_redshift_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_redshift_cluster.default", "enhanced_vpc_routing", "true"),
				),
			},
			{
				ResourceName:      "aws_redshift_cluster.default",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"final_snapshot_identifier",
					"master_password",
					"skip_final_snapshot",
				},
			},
			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftClusterExists("aws_redshift_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_redshift_cluster.default", "enhanced_vpc_routing", "false"),
				),
			},
		},
	})
}

func TestAccAWSRedshiftCluster_loggingEnabled(t *testing.T) {
	var v redshift.Cluster
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRedshiftClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRedshiftClusterConfig_loggingEnabled(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftClusterExists("aws_redshift_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_redshift_cluster.default", "logging.0.enable", "true"),
					resource.TestCheckResourceAttr(
						"aws_redshift_cluster.default", "logging.0.bucket_name", fmt.Sprintf("tf-test-redshift-logging-%d", rInt)),
				),
			},
			{
				ResourceName:      "aws_redshift_cluster.default",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"final_snapshot_identifier",
					"master_password",
					"skip_final_snapshot",
				},
			},
			{
				Config: testAccAWSRedshiftClusterConfig_loggingDisabled(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftClusterExists("aws_redshift_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_redshift_cluster.default", "logging.0.enable", "false"),
				),
			},
		},
	})
}

func TestAccAWSRedshiftCluster_snapshotCopy(t *testing.T) {
	var providers []*schema.Provider
	var v redshift.Cluster
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccMultipleRegionPreCheck(t, 2)
		},
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckAWSRedshiftClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRedshiftClusterConfig_snapshotCopyEnabled(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftClusterExists("aws_redshift_cluster.default", &v),
					resource.TestCheckResourceAttrPair("aws_redshift_cluster.default",
						"snapshot_copy.0.destination_region", "data.aws_region.alternate", "name"),
					resource.TestCheckResourceAttr(
						"aws_redshift_cluster.default", "snapshot_copy.0.retention_period", "1"),
				),
			},

			{
				Config: testAccAWSRedshiftClusterConfig_snapshotCopyDisabled(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftClusterExists("aws_redshift_cluster.default", &v),
					resource.TestCheckResourceAttr("aws_redshift_cluster.default", "snapshot_copy.#", "0"),
				),
			},
		},
	})
}

func TestAccAWSRedshiftCluster_iamRoles(t *testing.T) {
	var v redshift.Cluster

	ri := acctest.RandInt()
	preConfig := testAccAWSRedshiftClusterConfig_iamRoles(ri)
	postConfig := testAccAWSRedshiftClusterConfig_updateIamRoles(ri)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRedshiftClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftClusterExists("aws_redshift_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_redshift_cluster.default", "iam_roles.#", "2"),
				),
			},

			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftClusterExists("aws_redshift_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_redshift_cluster.default", "iam_roles.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSRedshiftCluster_publiclyAccessible(t *testing.T) {
	var v redshift.Cluster
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRedshiftClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRedshiftClusterConfig_notPubliclyAccessible(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftClusterExists("aws_redshift_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_redshift_cluster.default", "publicly_accessible", "false"),
				),
			},

			{
				Config: testAccAWSRedshiftClusterConfig_updatePubliclyAccessible(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftClusterExists("aws_redshift_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_redshift_cluster.default", "publicly_accessible", "true"),
				),
			},
		},
	})
}

func TestAccAWSRedshiftCluster_updateNodeCount(t *testing.T) {
	var v redshift.Cluster

	ri := acctest.RandInt()
	preConfig := testAccAWSRedshiftClusterConfig_basic(ri)
	postConfig := testAccAWSRedshiftClusterConfig_updateNodeCount(ri)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRedshiftClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftClusterExists("aws_redshift_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_redshift_cluster.default", "number_of_nodes", "1"),
				),
			},

			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftClusterExists("aws_redshift_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_redshift_cluster.default", "number_of_nodes", "2"),
					resource.TestCheckResourceAttr(
						"aws_redshift_cluster.default", "cluster_type", "multi-node"),
					resource.TestCheckResourceAttr(
						"aws_redshift_cluster.default", "node_type", "dc1.large"),
				),
			},
		},
	})
}

func TestAccAWSRedshiftCluster_updateNodeType(t *testing.T) {
	var v redshift.Cluster

	ri := acctest.RandInt()
	preConfig := testAccAWSRedshiftClusterConfig_basic(ri)
	postConfig := testAccAWSRedshiftClusterConfig_updateNodeType(ri)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRedshiftClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftClusterExists("aws_redshift_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_redshift_cluster.default", "node_type", "dc1.large"),
				),
			},

			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftClusterExists("aws_redshift_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_redshift_cluster.default", "number_of_nodes", "1"),
					resource.TestCheckResourceAttr(
						"aws_redshift_cluster.default", "cluster_type", "single-node"),
					resource.TestCheckResourceAttr(
						"aws_redshift_cluster.default", "node_type", "dc2.large"),
				),
			},
		},
	})
}

func TestAccAWSRedshiftCluster_tags(t *testing.T) {
	var v redshift.Cluster

	ri := acctest.RandInt()
	preConfig := testAccAWSRedshiftClusterConfig_tags(ri)
	postConfig := testAccAWSRedshiftClusterConfig_updatedTags(ri)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRedshiftClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftClusterExists("aws_redshift_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_redshift_cluster.default", "tags.%", "3"),
					resource.TestCheckResourceAttr("aws_redshift_cluster.default", "tags.environment", "Production"),
				),
			},

			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftClusterExists("aws_redshift_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_redshift_cluster.default", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_redshift_cluster.default", "tags.environment", "Production"),
				),
			},
		},
	})
}

func TestAccAWSRedshiftCluster_forceNewUsername(t *testing.T) {
	var first, second redshift.Cluster

	ri := acctest.RandInt()
	preConfig := testAccAWSRedshiftClusterConfig_basic(ri)
	postConfig := testAccAWSRedshiftClusterConfig_updatedUsername(ri)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRedshiftClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftClusterExists("aws_redshift_cluster.default", &first),
					testAccCheckAWSRedshiftClusterMasterUsername(&first, "foo_test"),
					resource.TestCheckResourceAttr("aws_redshift_cluster.default", "master_username", "foo_test"),
				),
			},

			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftClusterExists("aws_redshift_cluster.default", &second),
					testAccCheckAWSRedshiftClusterMasterUsername(&second, "new_username"),
					resource.TestCheckResourceAttr("aws_redshift_cluster.default", "master_username", "new_username"),
				),
			},
		},
	})
}

func TestAccAWSRedshiftCluster_changeAvailabilityZone(t *testing.T) {
	var first, second redshift.Cluster

	ri := acctest.RandInt()
	preConfig := testAccAWSRedshiftClusterConfig_basic(ri)
	postConfig := testAccAWSRedshiftClusterConfig_updatedAvailabilityZone(ri)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRedshiftClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftClusterExists("aws_redshift_cluster.default", &first),
					resource.TestCheckResourceAttrPair("aws_redshift_cluster.default", "availability_zone", "data.aws_availability_zones.available", "names.0"),
				),
			},

			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftClusterExists("aws_redshift_cluster.default", &second),
					resource.TestCheckResourceAttrPair("aws_redshift_cluster.default", "availability_zone", "data.aws_availability_zones.available", "names.1"),
				),
			},
		},
	})
}

func TestAccAWSRedshiftCluster_changeEncryption1(t *testing.T) {
	var cluster1, cluster2 redshift.Cluster

	ri := acctest.RandInt()
	preConfig := testAccAWSRedshiftClusterConfig_basic(ri)
	postConfig := testAccAWSRedshiftClusterConfig_encrypted(ri)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRedshiftClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftClusterExists("aws_redshift_cluster.default", &cluster1),
					resource.TestCheckResourceAttr("aws_redshift_cluster.default", "encrypted", "false"),
				),
			},

			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftClusterExists("aws_redshift_cluster.default", &cluster2),
					testAccCheckAWSRedshiftClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr("aws_redshift_cluster.default", "encrypted", "true"),
				),
			},
		},
	})
}

func TestAccAWSRedshiftCluster_changeEncryption2(t *testing.T) {
	var cluster1, cluster2 redshift.Cluster

	ri := acctest.RandInt()
	preConfig := testAccAWSRedshiftClusterConfig_encrypted(ri)
	postConfig := testAccAWSRedshiftClusterConfig_unencrypted(ri)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRedshiftClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftClusterExists("aws_redshift_cluster.default", &cluster1),
					resource.TestCheckResourceAttr("aws_redshift_cluster.default", "encrypted", "true"),
				),
			},
			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftClusterExists("aws_redshift_cluster.default", &cluster2),
					testAccCheckAWSRedshiftClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr("aws_redshift_cluster.default", "encrypted", "false"),
				),
			},
		},
	})
}

func testAccCheckAWSRedshiftClusterDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_redshift_cluster" {
			continue
		}

		// Try to find the Group
		conn := testAccProvider.Meta().(*AWSClient).redshiftconn
		var err error
		resp, err := conn.DescribeClusters(
			&redshift.DescribeClustersInput{
				ClusterIdentifier: aws.String(rs.Primary.ID),
			})

		if err == nil {
			if len(resp.Clusters) != 0 &&
				*resp.Clusters[0].ClusterIdentifier == rs.Primary.ID {
				return fmt.Errorf("Redshift Cluster %s still exists", rs.Primary.ID)
			}
		}

		// Return nil if the cluster is already destroyed
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "ClusterNotFound" {
				return nil
			}
		}

		return err
	}

	return nil
}

func testAccCheckAWSRedshiftClusterSnapshot(rInt int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_redshift_cluster" {
				continue
			}

			var err error

			// Try and delete the snapshot before we check for the cluster not found
			conn := testAccProvider.Meta().(*AWSClient).redshiftconn

			snapshot_identifier := fmt.Sprintf("tf-acctest-snapshot-%d", rInt)

			log.Printf("[INFO] Deleting the Snapshot %s", snapshot_identifier)
			_, snapDeleteErr := conn.DeleteClusterSnapshot(
				&redshift.DeleteClusterSnapshotInput{
					SnapshotIdentifier: aws.String(snapshot_identifier),
				})
			if snapDeleteErr != nil {
				return err
			}

			//lastly check that the Cluster is missing
			resp, err := conn.DescribeClusters(
				&redshift.DescribeClustersInput{
					ClusterIdentifier: aws.String(rs.Primary.ID),
				})

			if err == nil {
				if len(resp.Clusters) != 0 &&
					*resp.Clusters[0].ClusterIdentifier == rs.Primary.ID {
					return fmt.Errorf("Redshift Cluster %s still exists", rs.Primary.ID)
				}
			}

			// Return nil if the cluster is already destroyed
			if awsErr, ok := err.(awserr.Error); ok {
				if awsErr.Code() == "ClusterNotFound" {
					return nil
				}

				return err
			}

		}

		return nil
	}
}

func testAccCheckAWSRedshiftClusterExists(n string, v *redshift.Cluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Redshift Cluster Instance ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).redshiftconn
		resp, err := conn.DescribeClusters(&redshift.DescribeClustersInput{
			ClusterIdentifier: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		for _, c := range resp.Clusters {
			if *c.ClusterIdentifier == rs.Primary.ID {
				*v = *c
				return nil
			}
		}

		return fmt.Errorf("Redshift Cluster (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckAWSRedshiftClusterMasterUsername(c *redshift.Cluster, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *c.MasterUsername != value {
			return fmt.Errorf("Expected cluster's MasterUsername: %q, given: %q", value, *c.MasterUsername)
		}
		return nil
	}
}

func testAccCheckAWSRedshiftClusterNotRecreated(i, j *redshift.Cluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// In lieu of some other uniquely identifying attribute from the API that always changes
		// when a cluster is destroyed and recreated with the same identifier, we use the SSH key
		// as it will get regenerated when a cluster is destroyed.
		// Certain update operations (e.g KMS encrypting a cluster) will change ClusterCreateTime.
		// Clusters with the same identifier can/will have an overlapping Endpoint.Address.
		if aws.StringValue(i.ClusterPublicKey) != aws.StringValue(j.ClusterPublicKey) {
			return errors.New("Redshift Cluster was recreated")
		}

		return nil
	}
}

func testAccAWSRedshiftClusterConfig_updateNodeCount(rInt int) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_redshift_cluster" "default" {
  cluster_identifier                  = "tf-redshift-cluster-%d"
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc1.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  number_of_nodes                     = 2
  skip_final_snapshot                 = true
}
`, rInt))
}

func testAccAWSRedshiftClusterConfig_updateNodeType(rInt int) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_redshift_cluster" "default" {
  cluster_identifier                  = "tf-redshift-cluster-%d"
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc2.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  number_of_nodes                     = 1
  skip_final_snapshot                 = true
}
`, rInt))
}

func testAccAWSRedshiftClusterConfig_basic(rInt int) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_redshift_cluster" "default" {
  cluster_identifier                  = "tf-redshift-cluster-%d"
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc1.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  skip_final_snapshot                 = true
}
`, rInt))
}

func testAccAWSRedshiftClusterConfig_encrypted(rInt int) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_kms_key" "foo" {
  description = "Terraform acc test %d"

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
	POLICY

}

resource "aws_redshift_cluster" "default" {
  cluster_identifier                  = "tf-redshift-cluster-%d"
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc1.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  skip_final_snapshot                 = true
  encrypted                           = true
  kms_key_id                          = aws_kms_key.foo.arn
}
`, rInt, rInt))
}

func testAccAWSRedshiftClusterConfig_unencrypted(rInt int) string {
	// This is used along with the terraform config created testAccAWSRedshiftClusterConfig_encrypted, to test removal of encryption.
	//Removing the kms key here causes the key to be deleted before the redshift cluster is unencrypted, resulting in an unstable cluster. This is to be kept for the time-being unti we find a better way to handle this.
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_kms_key" "foo" {
  description = "Terraform acc test %d"

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
	POLICY

}

resource "aws_redshift_cluster" "default" {
  cluster_identifier                  = "tf-redshift-cluster-%d"
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc1.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  skip_final_snapshot                 = true
}
`, rInt, rInt))
}

func testAccAWSRedshiftClusterConfigWithFinalSnapshot(rInt int) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_redshift_cluster" "default" {
  cluster_identifier                  = "tf-redshift-cluster-%d"
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc1.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  skip_final_snapshot                 = false
  final_snapshot_identifier           = "tf-acctest-snapshot-%d"
}
`, rInt, rInt))
}

func testAccAWSRedshiftClusterConfig_kmsKey(rInt int) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_kms_key" "foo" {
  description = "Terraform acc test %d"

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_redshift_cluster" "default" {
  cluster_identifier                  = "tf-redshift-cluster-%d"
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc1.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  kms_key_id                          = aws_kms_key.foo.arn
  encrypted                           = true
  skip_final_snapshot                 = true
}
`, rInt, rInt))
}

func testAccAWSRedshiftClusterConfig_enhancedVpcRoutingEnabled(rInt int) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_redshift_cluster" "default" {
  cluster_identifier                  = "tf-redshift-cluster-%d"
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc1.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  enhanced_vpc_routing                = true
  skip_final_snapshot                 = true
}
`, rInt))
}

func testAccAWSRedshiftClusterConfig_enhancedVpcRoutingDisabled(rInt int) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_redshift_cluster" "default" {
  cluster_identifier                  = "tf-redshift-cluster-%d"
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc1.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  enhanced_vpc_routing                = false
  skip_final_snapshot                 = true
}
`, rInt))
}

func testAccAWSRedshiftClusterConfig_loggingDisabled(rInt int) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_redshift_cluster" "default" {
  cluster_identifier                  = "tf-redshift-cluster-%d"
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc1.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false

  logging {
    enable = false
  }

  skip_final_snapshot = true
}
`, rInt))
}

func testAccAWSRedshiftClusterConfig_loggingEnabled(rInt int) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_redshift_service_account" "main" {}

resource "aws_s3_bucket" "bucket" {
  bucket        = "tf-test-redshift-logging-%d"
  force_destroy = true

  policy = <<EOF
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "Stmt1376526643067",
      "Effect": "Allow",
      "Principal": {
        "AWS": "${data.aws_redshift_service_account.main.arn}"
      },
      "Action": "s3:PutObject",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::tf-test-redshift-logging-%d/*"
    },
    {
      "Sid": "Stmt137652664067",
      "Effect": "Allow",
      "Principal": {
        "AWS": "${data.aws_redshift_service_account.main.arn}"
      },
      "Action": "s3:GetBucketAcl",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::tf-test-redshift-logging-%d"
    }
  ]
}
EOF
}

resource "aws_redshift_cluster" "default" {
  cluster_identifier                  = "tf-redshift-cluster-%d"
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc1.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false

  logging {
    enable      = true
    bucket_name = aws_s3_bucket.bucket.bucket
  }

  skip_final_snapshot = true
}
`, rInt, rInt, rInt, rInt))
}

func testAccAWSRedshiftClusterConfig_snapshotCopyDisabled(rInt int) string {
	return composeConfig(
		testAccMultipleRegionProviderConfig(2),
		testAccAvailableAZsNoOptInConfig(),
		fmt.Sprintf(`
resource "aws_redshift_cluster" "default" {
  cluster_identifier                  = "tf-redshift-cluster-%d"
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc1.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  skip_final_snapshot                 = true
}
`, rInt))
}

func testAccAWSRedshiftClusterConfig_snapshotCopyEnabled(rInt int) string {
	return composeConfig(
		testAccMultipleRegionProviderConfig(2),
		testAccAvailableAZsNoOptInConfig(),
		fmt.Sprintf(`
data "aws_region" "alternate" {
  provider = "awsalternate"
}

resource "aws_redshift_cluster" "default" {
  cluster_identifier                  = "tf-redshift-cluster-%d"
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc1.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false

  snapshot_copy {
    destination_region = data.aws_region.alternate.name
    retention_period   = 1
  }

  skip_final_snapshot = true
}
`, rInt))
}

func testAccAWSRedshiftClusterConfig_tags(rInt int) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_redshift_cluster" "default" {
  cluster_identifier                  = "tf-redshift-cluster-%d"
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc1.large"
  automated_snapshot_retention_period = 7
  allow_version_upgrade               = false
  skip_final_snapshot                 = true

  tags = {
    environment = "Production"
    cluster     = "reader"
    Type        = "master"
  }
}
`, rInt))
}

func testAccAWSRedshiftClusterConfig_updatedTags(rInt int) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_redshift_cluster" "default" {
  cluster_identifier                  = "tf-redshift-cluster-%d"
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc1.large"
  automated_snapshot_retention_period = 7
  allow_version_upgrade               = false
  skip_final_snapshot                 = true

  tags = {
    environment = "Production"
  }
}
`, rInt))
}

func testAccAWSRedshiftClusterConfig_notPubliclyAccessible(rInt int) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-redshift-cluster-not-publicly-accessible"
  }
}

resource "aws_internet_gateway" "foo" {
  vpc_id = aws_vpc.foo.id

  tags = {
    foo = "bar"
  }
}

resource "aws_subnet" "foo" {
  cidr_block        = "10.1.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.foo.id

  tags = {
    Name = "tf-acc-redshift-cluster-not-publicly-accessible-foo"
  }
}

resource "aws_subnet" "bar" {
  cidr_block        = "10.1.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]
  vpc_id            = aws_vpc.foo.id

  tags = {
    Name = "tf-acc-redshift-cluster-not-publicly-accessible-bar"
  }
}

resource "aws_subnet" "foobar" {
  cidr_block        = "10.1.3.0/24"
  availability_zone = data.aws_availability_zones.available.names[2]
  vpc_id            = aws_vpc.foo.id

  tags = {
    Name = "tf-acc-redshift-cluster-not-publicly-accessible-foobar"
  }
}

resource "aws_redshift_subnet_group" "foo" {
  name        = "foo-%d"
  description = "foo description"
  subnet_ids  = [aws_subnet.foo.id, aws_subnet.bar.id, aws_subnet.foobar.id]
}

resource "aws_redshift_cluster" "default" {
  cluster_identifier                  = "tf-redshift-cluster-%d"
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc1.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  cluster_subnet_group_name           = aws_redshift_subnet_group.foo.name
  publicly_accessible                 = false
  skip_final_snapshot                 = true

  depends_on = [aws_internet_gateway.foo]
}
`, rInt, rInt))
}

func testAccAWSRedshiftClusterConfig_updatePubliclyAccessible(rInt int) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-redshift-cluster-upd-publicly-accessible"
  }
}

resource "aws_internet_gateway" "foo" {
  vpc_id = aws_vpc.foo.id

  tags = {
    foo = "bar"
  }
}

resource "aws_subnet" "foo" {
  cidr_block        = "10.1.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.foo.id

  tags = {
    Name = "tf-acc-redshift-cluster-upd-publicly-accessible-foo"
  }
}

resource "aws_subnet" "bar" {
  cidr_block        = "10.1.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]
  vpc_id            = aws_vpc.foo.id

  tags = {
    Name = "tf-acc-redshift-cluster-upd-publicly-accessible-bar"
  }
}

resource "aws_subnet" "foobar" {
  cidr_block        = "10.1.3.0/24"
  availability_zone = data.aws_availability_zones.available.names[2]
  vpc_id            = aws_vpc.foo.id

  tags = {
    Name = "tf-acc-redshift-cluster-upd-publicly-accessible-foobar"
  }
}

resource "aws_redshift_subnet_group" "foo" {
  name        = "foo-%d"
  description = "foo description"
  subnet_ids  = [aws_subnet.foo.id, aws_subnet.bar.id, aws_subnet.foobar.id]
}

resource "aws_redshift_cluster" "default" {
  cluster_identifier                  = "tf-redshift-cluster-%d"
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc1.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  cluster_subnet_group_name           = aws_redshift_subnet_group.foo.name
  publicly_accessible                 = true
  skip_final_snapshot                 = true

  depends_on = [aws_internet_gateway.foo]
}
`, rInt, rInt))
}

func testAccAWSRedshiftClusterConfig_iamRoles(rInt int) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_iam_role" "ec2-role" {
  name = "test-role-ec2-%d"
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.amazonaws.com"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}

resource "aws_iam_role" "lambda-role" {
  name = "test-role-lambda-%d"
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "lambda.amazonaws.com"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}

resource "aws_redshift_cluster" "default" {
  cluster_identifier                  = "tf-redshift-cluster-%d"
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc1.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  iam_roles                           = [aws_iam_role.ec2-role.arn, aws_iam_role.lambda-role.arn]
  skip_final_snapshot                 = true
}
`, rInt, rInt, rInt))
}

func testAccAWSRedshiftClusterConfig_updateIamRoles(rInt int) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_iam_role" "ec2-role" {
  name = "test-role-ec2-%d"
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.amazonaws.com"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}

resource "aws_iam_role" "lambda-role" {
  name = "test-role-lambda-%d"
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "lambda.amazonaws.com"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}

resource "aws_redshift_cluster" "default" {
  cluster_identifier                  = "tf-redshift-cluster-%d"
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc1.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  iam_roles                           = [aws_iam_role.ec2-role.arn]
  skip_final_snapshot                 = true
}
`, rInt, rInt, rInt))
}

func testAccAWSRedshiftClusterConfig_updatedUsername(rInt int) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_redshift_cluster" "default" {
  cluster_identifier                  = "tf-redshift-cluster-%d"
  availability_zone                   = data.aws_availability_zones.available.names[0]
  database_name                       = "mydb"
  master_username                     = "new_username"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc1.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  skip_final_snapshot                 = true
}
`, rInt))
}

func testAccAWSRedshiftClusterConfig_updatedAvailabilityZone(rInt int) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_redshift_cluster" "default" {
  cluster_identifier                  = "tf-redshift-cluster-%d"
  availability_zone                   = data.aws_availability_zones.available.names[1]
  database_name                       = "mydb"
  master_username                     = "foo_test"
  master_password                     = "Mustbe8characters"
  node_type                           = "dc1.large"
  automated_snapshot_retention_period = 0
  allow_version_upgrade               = false
  skip_final_snapshot                 = true
}
`, rInt))
}
