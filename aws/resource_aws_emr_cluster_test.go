package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/emr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_emr_cluster", &resource.Sweeper{
		Name: "aws_emr_cluster",
		F:    testSweepEmrClusters,
	})
}

func testSweepEmrClusters(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).emrconn

	input := &emr.ListClustersInput{
		ClusterStates: []*string{
			aws.String(emr.ClusterStateBootstrapping),
			aws.String(emr.ClusterStateRunning),
			aws.String(emr.ClusterStateStarting),
			aws.String(emr.ClusterStateWaiting),
		},
	}
	err = conn.ListClustersPages(input, func(page *emr.ListClustersOutput, isLast bool) bool {
		if page == nil {
			return !isLast
		}

		for _, cluster := range page.Clusters {
			describeClusterInput := &emr.DescribeClusterInput{
				ClusterId: cluster.Id,
			}
			terminateJobFlowsInput := &emr.TerminateJobFlowsInput{
				JobFlowIds: []*string{cluster.Id},
			}
			id := aws.StringValue(cluster.Id)

			log.Printf("[INFO] Deleting EMR Cluster: %s", id)
			_, err = conn.TerminateJobFlows(terminateJobFlowsInput)

			if err != nil {
				log.Printf("[ERROR] Error terminating EMR Cluster (%s): %s", id, err)
			}

			if err := conn.WaitUntilClusterTerminated(describeClusterInput); err != nil {
				log.Printf("[ERROR] Error waiting for EMR Cluster (%s) termination: %s", id, err)
			}
		}

		return !isLast
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping EMR Cluster sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("error retrieving EMR Clusters: %s", err)
	}

	return nil
}

func TestAccAWSEMRCluster_basic(t *testing.T) {
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.tf-test-cluster"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "scale_down_behavior", "TERMINATE_AT_TASK_COMPLETION"),
					resource.TestCheckResourceAttr(resourceName, "step.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckNoResourceAttr(resourceName, "additional_info"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
		},
	})
}

func TestAccAWSEMRCluster_additionalInfo(t *testing.T) {
	var cluster emr.Cluster
	expectedJSON := `
{
  "instanceAwsClientConfiguration": {
    "proxyPort": 8099,
    "proxyHost": "myproxy.example.com"
  }
}`

	resourceName := "aws_emr_cluster.tf-test-cluster"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfigAdditionalInfo(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "scale_down_behavior", "TERMINATE_AT_TASK_COMPLETION"),
					resource.TestCheckResourceAttr(resourceName, "step.#", "0"),
					testAccCheckResourceAttrEquivalentJSON(resourceName, "additional_info", expectedJSON),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"additional_info",
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
		},
	})
}

func TestAccAWSEMRCluster_disappears(t *testing.T) {
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.tf-test-cluster"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster),
					testAccCheckAWSEmrClusterDisappears(&cluster),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSEMRCluster_configurationsJson(t *testing.T) {
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.tf-test-cluster"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfigConfigurationsJson(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster),
					resource.TestMatchResourceAttr(resourceName, "configurations_json",
						regexp.MustCompile("{\"JAVA_HOME\":\"/usr/lib/jvm/java-1.8.0\".+")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
		},
	})
}

func TestAccAWSEMRCluster_CoreInstanceGroup_AutoscalingPolicy(t *testing.T) {
	var cluster1, cluster2, cluster3 emr.Cluster
	autoscalingPolicy1 := `
{
  "Constraints": {
    "MinCapacity": 1,
    "MaxCapacity": 2
  },
  "Rules": [
    {
      "Name": "ScaleOutMemoryPercentage",
      "Description": "Scale out if YARNMemoryAvailablePercentage is less than 15",
      "Action": {
        "SimpleScalingPolicyConfiguration": {
          "AdjustmentType": "CHANGE_IN_CAPACITY",
          "ScalingAdjustment": 1,
          "CoolDown": 300
        }
      },
      "Trigger": {
        "CloudWatchAlarmDefinition": {
          "ComparisonOperator": "LESS_THAN",
          "EvaluationPeriods": 1,
          "MetricName": "YARNMemoryAvailablePercentage",
          "Namespace": "AWS/ElasticMapReduce",
          "Period": 300,
          "Statistic": "AVERAGE",
          "Threshold": 15.0,
          "Unit": "PERCENT"
        }
      }
    }
  ]
}
`
	autoscalingPolicy2 := `
{
  "Constraints": {
    "MinCapacity": 1,
    "MaxCapacity": 3
  },
  "Rules": [
    {
      "Name": "ScaleOutMemoryPercentage",
      "Description": "Scale out if YARNMemoryAvailablePercentage is less than 15",
      "Action": {
        "SimpleScalingPolicyConfiguration": {
          "AdjustmentType": "CHANGE_IN_CAPACITY",
          "ScalingAdjustment": 1,
          "CoolDown": 300
        }
      },
      "Trigger": {
        "CloudWatchAlarmDefinition": {
          "ComparisonOperator": "LESS_THAN",
          "EvaluationPeriods": 1,
          "MetricName": "YARNMemoryAvailablePercentage",
          "Namespace": "AWS/ElasticMapReduce",
          "Period": 300,
          "Statistic": "AVERAGE",
          "Threshold": 15.0,
          "Unit": "PERCENT"
        }
      }
    }
  ]
}
`
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_emr_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfigCoreInstanceGroupAutoscalingPolicy(rName, autoscalingPolicy1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", "1"),
					testAccCheckResourceAttrEquivalentJSON(resourceName, "core_instance_group.0.autoscaling_policy", autoscalingPolicy1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
			{
				Config: testAccAWSEmrClusterConfigCoreInstanceGroupAutoscalingPolicy(rName, autoscalingPolicy2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster2),
					testAccCheckAWSEmrClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", "1"),
					testAccCheckResourceAttrEquivalentJSON(resourceName, "core_instance_group.0.autoscaling_policy", autoscalingPolicy2),
				),
			},
			{
				Config: testAccAWSEmrClusterConfigCoreInstanceGroupAutoscalingPolicyRemoved(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster3),
					testAccCheckAWSEmrClusterNotRecreated(&cluster2, &cluster3),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.0.autoscaling_policy", ""),
				),
			},
		},
	})
}

func TestAccAWSEMRCluster_CoreInstanceGroup_BidPrice(t *testing.T) {
	var cluster1, cluster2 emr.Cluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_emr_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfigCoreInstanceGroupBidPrice(rName, "0.50"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.0.bid_price", "0.50"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
			{
				Config: testAccAWSEmrClusterConfigCoreInstanceGroupBidPrice(rName, "0.51"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster2),
					testAccCheckAWSEmrClusterRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.0.bid_price", "0.51"),
				),
			},
		},
	})
}

func TestAccAWSEMRCluster_CoreInstanceGroup_InstanceCount(t *testing.T) {
	var cluster1, cluster2, cluster3 emr.Cluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_emr_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfigCoreInstanceGroupInstanceCount(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.0.instance_count", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
			{
				Config: testAccAWSEmrClusterConfigCoreInstanceGroupInstanceCount(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster2),
					testAccCheckAWSEmrClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.0.instance_count", "1"),
				),
			},
			{
				Config: testAccAWSEmrClusterConfigCoreInstanceGroupInstanceCount(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster3),
					testAccCheckAWSEmrClusterNotRecreated(&cluster2, &cluster3),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.0.instance_count", "2"),
				),
			},
		},
	})
}

func TestAccAWSEMRCluster_CoreInstanceGroup_InstanceType(t *testing.T) {
	var cluster1, cluster2 emr.Cluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_emr_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfigCoreInstanceGroupInstanceType(rName, "m4.large"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.0.instance_type", "m4.large"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
			{
				Config: testAccAWSEmrClusterConfigCoreInstanceGroupInstanceType(rName, "m4.xlarge"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster2),
					testAccCheckAWSEmrClusterRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.0.instance_type", "m4.xlarge"),
				),
			},
		},
	})
}

func TestAccAWSEMRCluster_CoreInstanceGroup_Name(t *testing.T) {
	var cluster1, cluster2 emr.Cluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_emr_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfigCoreInstanceGroupName(rName, "name1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.0.name", "name1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
			{
				Config: testAccAWSEmrClusterConfigCoreInstanceGroupName(rName, "name2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster2),
					testAccCheckAWSEmrClusterRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.0.name", "name2"),
				),
			},
		},
	})
}

func TestAccAWSEMRCluster_Ec2Attributes_DefaultManagedSecurityGroups(t *testing.T) {
	var cluster emr.Cluster
	var vpc ec2.Vpc

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_emr_cluster.tf-test-cluster"
	vpcResourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfigEc2AttributesDefaultManagedSecurityGroups(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster),
					testAccCheckVpcExists(vpcResourceName, &vpc),
				),
			},
			{
				Config:      testAccAWSEmrClusterConfigEc2AttributesDefaultManagedSecurityGroups(rName),
				Destroy:     true,
				ExpectError: regexp.MustCompile(`DependencyViolation`),
			},
			{
				PreConfig: func() {
					conn := testAccProvider.Meta().(*AWSClient).ec2conn

					err := testAccEmrDeleteManagedSecurityGroups(conn, &vpc)

					if err != nil {
						t.Fatal(err)
					}
				},
				Config:  testAccAWSEmrClusterConfigEc2AttributesDefaultManagedSecurityGroups(rName),
				Destroy: true,
			},
		},
	})
}

func TestAccAWSEMRCluster_Kerberos_ClusterDedicatedKdc(t *testing.T) {
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.tf-test-cluster"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	password := fmt.Sprintf("NeverKeepPasswordsInPlainText%s!", rName)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfig_Kerberos_ClusterDedicatedKdc(rName, password),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "kerberos_attributes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kerberos_attributes.0.kdc_admin_password", password),
					resource.TestCheckResourceAttr(resourceName, "kerberos_attributes.0.realm", "EC2.INTERNAL"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
					"kerberos_attributes.0.kdc_admin_password",
				},
			},
		},
	})
}

func TestAccAWSEMRCluster_MasterInstanceGroup_BidPrice(t *testing.T) {
	var cluster1, cluster2 emr.Cluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_emr_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfigMasterInstanceGroupBidPrice(rName, "0.50"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.0.bid_price", "0.50"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
			{
				Config: testAccAWSEmrClusterConfigMasterInstanceGroupBidPrice(rName, "0.51"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster2),
					testAccCheckAWSEmrClusterRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.0.bid_price", "0.51"),
				),
			},
		},
	})
}

func TestAccAWSEMRCluster_MasterInstanceGroup_InstanceCount(t *testing.T) {
	var cluster1, cluster2 emr.Cluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_emr_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfigMasterInstanceGroupInstanceCount(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.0.instance_count", "3"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
			{
				Config: testAccAWSEmrClusterConfigMasterInstanceGroupInstanceCount(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster2),
					testAccCheckAWSEmrClusterRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.0.instance_count", "1"),
				),
			},
		},
	})
}

func TestAccAWSEMRCluster_MasterInstanceGroup_InstanceType(t *testing.T) {
	var cluster1, cluster2 emr.Cluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_emr_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfigMasterInstanceGroupInstanceType(rName, "m4.large"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.0.instance_type", "m4.large"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
			{
				Config: testAccAWSEmrClusterConfigMasterInstanceGroupInstanceType(rName, "m4.xlarge"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster2),
					testAccCheckAWSEmrClusterRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.0.instance_type", "m4.xlarge"),
				),
			},
		},
	})
}

func TestAccAWSEMRCluster_MasterInstanceGroup_Name(t *testing.T) {
	var cluster1, cluster2 emr.Cluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_emr_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfigMasterInstanceGroupName(rName, "name1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.0.name", "name1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
			{
				Config: testAccAWSEmrClusterConfigMasterInstanceGroupName(rName, "name2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster2),
					testAccCheckAWSEmrClusterRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.0.name", "name2"),
				),
			},
		},
	})
}

func TestAccAWSEMRCluster_security_config(t *testing.T) {
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.tf-test-cluster"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfig_SecurityConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttrPair(resourceName, "security_configuration", "aws_emr_security_configuration.foo", "name"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
		},
	})
}

func TestAccAWSEMRCluster_Step_Basic(t *testing.T) {
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.tf-test-cluster"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfig_Step_Single(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "step.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "step.0.action_on_failure", "TERMINATE_CLUSTER"),
					resource.TestCheckResourceAttr(resourceName, "step.0.hadoop_jar_step.0.args.0", "state-pusher-script"),
					resource.TestCheckResourceAttr(resourceName, "step.0.hadoop_jar_step.0.jar", "command-runner.jar"),
					resource.TestCheckResourceAttr(resourceName, "step.0.hadoop_jar_step.0.main_class", ""),
					resource.TestCheckResourceAttr(resourceName, "step.0.hadoop_jar_step.0.properties.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "step.0.name", "Setup Hadoop Debugging"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
		},
	})
}

func TestAccAWSEMRCluster_Step_ConfigMode(t *testing.T) {
	var cluster1, cluster2, cluster3 emr.Cluster

	resourceName := "aws_emr_cluster.tf-test-cluster"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfig_Step_Single(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "step.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
			{
				Config: testAccAWSEmrClusterConfig_Step_NoBlocks(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "step.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
			{
				Config: testAccAWSEmrClusterConfig_Step_Zeroed(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster3),
					resource.TestCheckResourceAttr(resourceName, "step.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
		},
	})
}

func TestAccAWSEMRCluster_Step_Multiple(t *testing.T) {
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.tf-test-cluster"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfig_Step_Multiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "step.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "step.0.action_on_failure", "TERMINATE_CLUSTER"),
					resource.TestCheckResourceAttr(resourceName, "step.0.hadoop_jar_step.0.args.0", "state-pusher-script"),
					resource.TestCheckResourceAttr(resourceName, "step.0.hadoop_jar_step.0.jar", "command-runner.jar"),
					resource.TestCheckResourceAttr(resourceName, "step.0.name", "Setup Hadoop Debugging"),
					resource.TestCheckResourceAttr(resourceName, "step.1.action_on_failure", "CONTINUE"),
					resource.TestCheckResourceAttr(resourceName, "step.1.hadoop_jar_step.0.args.0", "spark-example"),
					resource.TestCheckResourceAttr(resourceName, "step.1.hadoop_jar_step.0.args.1", "SparkPi"),
					resource.TestCheckResourceAttr(resourceName, "step.1.hadoop_jar_step.0.args.2", "10"),
					resource.TestCheckResourceAttr(resourceName, "step.1.hadoop_jar_step.0.jar", "command-runner.jar"),
					resource.TestCheckResourceAttr(resourceName, "step.1.name", "Spark Step"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
		},
	})
}

func TestAccAWSEMRCluster_bootstrap_ordering(t *testing.T) {
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.test"
	rName := acctest.RandomWithPrefix("tf-emr-bootstrap")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfig_bootstrap(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.0.name", "runif"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.0.path", "s3://elasticmapreduce/bootstrap-actions/run-if"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.0.args.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.0.args.0", "instance.isMaster=true"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.0.args.1", "echo running on master node"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.name", "test"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.path", fmt.Sprintf("s3://%s/testscript.sh", rName)),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.args.#", "10"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
			{
				Config: testAccAWSEmrClusterConfig_bootstrapAdd(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.0.name", "runif"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.0.path", "s3://elasticmapreduce/bootstrap-actions/run-if"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.0.args.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.0.args.0", "instance.isMaster=true"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.0.args.1", "echo running on master node"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.name", "test"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.path", fmt.Sprintf("s3://%s/testscript.sh", rName)),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.args.#", "10"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.2.name", "runif-2"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.2.path", "s3://elasticmapreduce/bootstrap-actions/run-if"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.2.args.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.2.args.0", "instance.isMaster=true"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.2.args.1", "echo also running on master node"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
			{
				Config: testAccAWSEmrClusterConfig_bootstrapReorder(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.0.name", "runif"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.0.path", "s3://elasticmapreduce/bootstrap-actions/run-if"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.0.args.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.0.args.0", "instance.isMaster=true"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.0.args.1", "echo running on master node"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.2.name", "test"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.2.path", fmt.Sprintf("s3://%s/testscript.sh", rName)),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.2.args.#", "10"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.name", "runif-2"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.path", "s3://elasticmapreduce/bootstrap-actions/run-if"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.args.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.args.0", "instance.isMaster=true"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.args.1", "echo also running on master node"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
		},
	})
}

func TestAccAWSEMRCluster_terminationProtected(t *testing.T) {
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.tf-test-cluster"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfigTerminationPolicy(rName, "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),
				),
			},
			{
				Config: testAccAWSEmrClusterConfigTerminationPolicy(rName, "true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "termination_protection", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
			{
				//Need to turn off termination_protection to allow the job to be deleted
				Config: testAccAWSEmrClusterConfigTerminationPolicy(rName, "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
		},
	})
}

func TestAccAWSEMRCluster_keepJob(t *testing.T) {
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.tf-test-cluster"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfig_keepJob(rName, "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "keep_job_flow_alive_when_no_steps", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
		},
	})
}

func TestAccAWSEMRCluster_visibleToAllUsers(t *testing.T) {
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.tf-test-cluster"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "visible_to_all_users", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
			{
				Config: testAccAWSEmrClusterConfigVisibleToAllUsersUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "visible_to_all_users", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
		},
	})
}

func TestAccAWSEMRCluster_s3Logging(t *testing.T) {
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.tf-test-cluster"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	bucketName := fmt.Sprintf("s3n://%s/", rName)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfigS3Logging(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "log_uri", bucketName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
		},
	})
}

func TestAccAWSEMRCluster_tags(t *testing.T) {
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.tf-test-cluster"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "4"),
					resource.TestCheckResourceAttr(resourceName, "tags.role", "rolename"),
					resource.TestCheckResourceAttr(resourceName, "tags.dns_zone", "env_zone"),
					resource.TestCheckResourceAttr(resourceName, "tags.env", "env"),
					resource.TestCheckResourceAttr(resourceName, "tags.name", "name-env")),
			},
			{
				Config: testAccAWSEmrClusterConfigUpdatedTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.dns_zone", "new_zone"),
					resource.TestCheckResourceAttr(resourceName, "tags.Env", "production"),
					resource.TestCheckResourceAttr(resourceName, "tags.name", "name-env"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
		},
	})
}

func TestAccAWSEMRCluster_root_volume_size(t *testing.T) {
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.tf-test-cluster"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "ebs_root_volume_size", "21"),
				),
			},
			{
				Config: testAccAWSEmrClusterConfigUpdatedRootVolumeSize(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "ebs_root_volume_size", "48"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
		},
	})
}

func TestAccAWSEMRCluster_step_concurrency_level(t *testing.T) {
	var cluster emr.Cluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_emr_cluster.tf-test-cluster"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfigStepConcurrencyLevel(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "step_concurrency_level", "2"),
				),
			},
			{
				Config: testAccAWSEmrClusterConfigStepConcurrencyLevel(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "step_concurrency_level", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
		},
	})
}

func TestAccAWSEMRCluster_ebs_config(t *testing.T) {
	var cluster emr.Cluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_emr_cluster.tf-test-cluster"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrEbsConfig(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.0.ebs_config.0.volumes_per_instance", "2"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.0.ebs_config.0.volumes_per_instance", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
		},
	})
}

func TestAccAWSEMRCluster_custom_ami_id(t *testing.T) {
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.tf-test-cluster"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfigCustomAmiID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttrSet(resourceName, "custom_ami_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
		},
	})
}

func TestAccAWSEMRCluster_instance_fleet(t *testing.T) {
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.tf-test-cluster"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfigInstanceFleets(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "master_instance_fleet.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_fleet.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSEMRCluster_instance_fleet_master_only(t *testing.T) {
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.tf-test-cluster"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfigInstanceFleetsMasterOnly(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "master_instance_fleet.#", "1"),
				),
			},
		},
	})
}

func testAccCheckAWSEmrDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).emrconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_emr_cluster" {
			continue
		}

		params := &emr.DescribeClusterInput{
			ClusterId: aws.String(rs.Primary.ID),
		}

		describe, err := conn.DescribeCluster(params)

		if err == nil {
			if describe.Cluster != nil &&
				*describe.Cluster.Status.State == "WAITING" {
				return fmt.Errorf("EMR Cluster still exists")
			}
		}

		providerErr, ok := err.(awserr.Error)
		if !ok {
			return err
		}

		log.Printf("[ERROR] %v", providerErr)
	}

	return nil
}

func testAccCheckAWSEmrClusterExists(n string, v *emr.Cluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No cluster id set")
		}
		conn := testAccProvider.Meta().(*AWSClient).emrconn
		describe, err := conn.DescribeCluster(&emr.DescribeClusterInput{
			ClusterId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return fmt.Errorf("EMR error: %v", err)
		}

		if describe.Cluster == nil || *describe.Cluster.Id != rs.Primary.ID {
			return fmt.Errorf("EMR cluster %q not found", rs.Primary.ID)
		}

		*v = *describe.Cluster

		if describe.Cluster.Status != nil {
			state := aws.StringValue(describe.Cluster.Status.State)
			if state != emr.ClusterStateRunning && state != emr.ClusterStateWaiting {
				return fmt.Errorf("EMR cluster %q is not RUNNING or WAITING, currently: %s", rs.Primary.ID, state)
			}
		}

		return nil
	}
}

func testAccCheckAWSEmrClusterDisappears(cluster *emr.Cluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).emrconn
		id := aws.StringValue(cluster.Id)

		terminateJobFlowsInput := &emr.TerminateJobFlowsInput{
			JobFlowIds: []*string{cluster.Id},
		}

		_, err := conn.TerminateJobFlows(terminateJobFlowsInput)

		if err != nil {
			return err
		}

		input := &emr.ListInstancesInput{
			ClusterId: cluster.Id,
		}
		var output *emr.ListInstancesOutput
		var instanceCount int

		err = resource.Retry(20*time.Minute, func() *resource.RetryError {
			var err error
			output, err = conn.ListInstances(input)

			if err != nil {
				return resource.NonRetryableError(err)
			}

			instanceCount = countEMRRemainingInstances(output, id)

			if instanceCount != 0 {
				return resource.RetryableError(fmt.Errorf("EMR Cluster (%s) has (%d) Instances remaining", id, instanceCount))
			}

			return nil
		})

		if isResourceTimeoutError(err) {
			output, err = conn.ListInstances(input)

			if err == nil {
				instanceCount = countEMRRemainingInstances(output, id)
			}
		}

		if instanceCount != 0 {
			return fmt.Errorf("EMR Cluster (%s) has (%d) Instances remaining", id, instanceCount)
		}

		if err != nil {
			return fmt.Errorf("error waiting for EMR Cluster (%s) Instances to drain: %s", id, err)
		}

		return nil
	}
}

func testAccCheckAWSEmrClusterNotRecreated(i, j *emr.Cluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.Id) != aws.StringValue(j.Id) {
			return fmt.Errorf("EMR Cluster recreated: %s -> %s", aws.StringValue(i.Id), aws.StringValue(j.Id))
		}

		return nil
	}
}

func testAccCheckAWSEmrClusterRecreated(i, j *emr.Cluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.Id) == aws.StringValue(j.Id) {
			return fmt.Errorf("EMR Cluster not recreated: %s", aws.StringValue(i.Id))
		}

		return nil
	}
}

func testAccEmrDeleteManagedSecurityGroups(conn *ec2.EC2, vpc *ec2.Vpc) error {
	// Reference: https://docs.aws.amazon.com/emr/latest/ManagementGuide/emr-man-sec-groups.html
	managedSecurityGroups := map[string]*ec2.SecurityGroup{
		"ElasticMapReduce-master": nil,
		"ElasticMapReduce-slave":  nil,
	}

	for groupName := range managedSecurityGroups {
		securityGroup, err := testAccEmrDescribeManagedSecurityGroup(conn, vpc, groupName)

		if err != nil {
			return fmt.Errorf("error describing EMR Managed Security Group (%s): %s", groupName, err)
		}

		managedSecurityGroups[groupName] = securityGroup
	}

	// EMR Managed Security Groups rules reference each other, so rules from all
	// groups must be revoked first.
	for groupName, securityGroup := range managedSecurityGroups {
		if securityGroup == nil {
			continue
		}

		err := testAccEmrRevokeManagedSecurityGroup(conn, securityGroup)

		if err != nil {
			return fmt.Errorf("error revoking EMR Managed Security Group (%s): %s", groupName, err)
		}
	}

	for groupName, securityGroup := range managedSecurityGroups {
		if securityGroup == nil {
			continue
		}

		err := testAccEmrDeleteManagedSecurityGroup(conn, securityGroup)

		if err != nil {
			return fmt.Errorf("error deleting EMR Managed Security Group (%s): %s", groupName, err)
		}
	}

	return nil
}

func testAccEmrDescribeManagedSecurityGroup(conn *ec2.EC2, vpc *ec2.Vpc, securityGroupName string) (*ec2.SecurityGroup, error) {
	input := &ec2.DescribeSecurityGroupsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("group-name"),
				Values: aws.StringSlice([]string{securityGroupName}),
			},
			{
				Name:   aws.String("vpc-id"),
				Values: []*string{vpc.VpcId},
			},
		},
	}

	output, err := conn.DescribeSecurityGroups(input)

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.SecurityGroups) != 1 {
		return nil, nil
	}

	return output.SecurityGroups[0], nil
}

func testAccEmrRevokeManagedSecurityGroup(conn *ec2.EC2, securityGroup *ec2.SecurityGroup) error {
	input := &ec2.RevokeSecurityGroupIngressInput{
		GroupId:       securityGroup.GroupId,
		IpPermissions: securityGroup.IpPermissions,
	}

	_, err := conn.RevokeSecurityGroupIngress(input)

	return err
}

func testAccEmrDeleteManagedSecurityGroup(conn *ec2.EC2, securityGroup *ec2.SecurityGroup) error {
	input := &ec2.DeleteSecurityGroupInput{
		GroupId: securityGroup.GroupId,
	}

	_, err := conn.DeleteSecurityGroup(input)

	return err
}

func testAccAWSEmrComposeConfig(mapPublicIPOnLaunch bool, config ...string) string {
	return composeConfig(append(config, testAccAWSEmrClusterConfigBaseVpc(mapPublicIPOnLaunch))...)
}

func testAccAWSEmrClusterConfigCurrentPartition() string {
	return `
data "aws_partition" "current" {}
`
}

func testAccAWSEmrClusterConfig_bootstrap(r string) string {
	return testAccAWSEmrComposeConfig(false,
		testAccAWSEmrClusterConfigCurrentPartition(),
		testAccAWSEmrClusterConfigIAMServiceRoleBase(r),
		testAccAWSEmrClusterConfigIAMInstanceProfileBase(r),
		testAccAWSEmrClusterConfigBootstrapActionBucket(r),
		fmt.Sprintf(`
resource "aws_emr_cluster" "test" {
  name          = "%[1]s"
  release_label = "emr-5.0.0"
  applications  = ["Hadoop", "Hive"]
  log_uri       = "s3n://terraform/testlog/"

  master_instance_group {
    instance_type = "c4.large"
  }

  core_instance_group {
    instance_count = 1
    instance_type  = "c4.large"
  }

  service_role = aws_iam_role.emr_service.arn
  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
  ]

  ec2_attributes {
    subnet_id                         = aws_subnet.test.id
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = aws_iam_instance_profile.emr_instance_profile.arn
  }

  bootstrap_action {
    path = "s3://elasticmapreduce/bootstrap-actions/run-if"
    name = "runif"
    args = ["instance.isMaster=true", "echo running on master node"]
  }

  bootstrap_action {
    path = "s3://${aws_s3_bucket_object.testobject.bucket}/${aws_s3_bucket_object.testobject.key}"
    name = "test"

    args = ["1",
      "2",
      "3",
      "4",
      "5",
      "6",
      "7",
      "8",
      "9",
      "10",
    ]
  }
}
`, r),
	)
}

func testAccAWSEmrClusterConfig_bootstrapAdd(r string) string {
	return testAccAWSEmrComposeConfig(false,
		testAccAWSEmrClusterConfigCurrentPartition(),
		testAccAWSEmrClusterConfigIAMServiceRoleBase(r),
		testAccAWSEmrClusterConfigIAMInstanceProfileBase(r),
		testAccAWSEmrClusterConfigBootstrapActionBucket(r),
		fmt.Sprintf(`
resource "aws_emr_cluster" "test" {
  name          = "%[1]s"
  release_label = "emr-5.0.0"
  applications  = ["Hadoop", "Hive"]
  log_uri       = "s3n://terraform/testlog/"

  master_instance_group {
    instance_type = "c4.large"
  }

  core_instance_group {
    instance_count = 1
    instance_type  = "c4.large"
  }

  service_role = aws_iam_role.emr_service.arn
  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
  ]

  ec2_attributes {
    subnet_id                         = aws_subnet.test.id
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = aws_iam_instance_profile.emr_instance_profile.arn
  }

  bootstrap_action {
    path = "s3://elasticmapreduce/bootstrap-actions/run-if"
    name = "runif"
    args = ["instance.isMaster=true", "echo running on master node"]
  }

  bootstrap_action {
    path = "s3://${aws_s3_bucket_object.testobject.bucket}/${aws_s3_bucket_object.testobject.key}"
    name = "test"

    args = ["1",
      "2",
      "3",
      "4",
      "5",
      "6",
      "7",
      "8",
      "9",
      "10",
    ]
  }

  bootstrap_action {
    path = "s3://elasticmapreduce/bootstrap-actions/run-if"
    name = "runif-2"
    args = ["instance.isMaster=true", "echo also running on master node"]
  }
}
`, r),
	)
}

func testAccAWSEmrClusterConfig_bootstrapReorder(r string) string {
	return testAccAWSEmrComposeConfig(false,
		testAccAWSEmrClusterConfigCurrentPartition(),
		testAccAWSEmrClusterConfigIAMServiceRoleBase(r),
		testAccAWSEmrClusterConfigIAMInstanceProfileBase(r),
		testAccAWSEmrClusterConfigBootstrapActionBucket(r),
		fmt.Sprintf(`
resource "aws_emr_cluster" "test" {
  name          = "%[1]s"
  release_label = "emr-5.0.0"
  applications  = ["Hadoop", "Hive"]
  log_uri       = "s3n://terraform/testlog/"

  master_instance_group {
    instance_type = "c4.large"
  }

  core_instance_group {
    instance_count = 1
    instance_type  = "c4.large"
  }

  service_role = aws_iam_role.emr_service.arn
  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
  ]

  ec2_attributes {
    subnet_id                         = aws_subnet.test.id
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = aws_iam_instance_profile.emr_instance_profile.arn
  }

  bootstrap_action {
    path = "s3://elasticmapreduce/bootstrap-actions/run-if"
    name = "runif"
    args = ["instance.isMaster=true", "echo running on master node"]
  }

  bootstrap_action {
    path = "s3://elasticmapreduce/bootstrap-actions/run-if"
    name = "runif-2"
    args = ["instance.isMaster=true", "echo also running on master node"]
  }

  bootstrap_action {
    path = "s3://${aws_s3_bucket_object.testobject.bucket}/${aws_s3_bucket_object.testobject.key}"
    name = "test"

    args = ["1",
      "2",
      "3",
      "4",
      "5",
      "6",
      "7",
      "8",
      "9",
      "10",
    ]
  }
}
`, r),
	)
}

func testAccAWSEmrClusterConfig(r string) string {
	return testAccAWSEmrComposeConfig(false,
		testAccAWSEmrClusterConfigCurrentPartition(),
		testAccAWSEmrClusterConfigIAMServiceRoleBase(r),
		testAccAWSEmrClusterConfigIAMInstanceProfileBase(r),
		testAccAWSEmrClusterConfigIAMAutoscalingRole(r),
		fmt.Sprintf(`
resource "aws_emr_cluster" "tf-test-cluster" {
  name          = "%[1]s"
  release_label = "emr-4.6.0"
  applications  = ["Spark"]

  ec2_attributes {
    subnet_id                         = aws_subnet.test.id
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = aws_iam_instance_profile.emr_instance_profile.arn
  }

  master_instance_group {
    instance_type = "c4.large"
  }

  core_instance_group {
    instance_count = 1
    instance_type  = "c4.large"
  }

  tags = {
    role     = "rolename"
    dns_zone = "env_zone"
    env      = "env"
    name     = "name-env"
  }

  keep_job_flow_alive_when_no_steps = true
  termination_protection            = false

  scale_down_behavior = "TERMINATE_AT_TASK_COMPLETION"

  configurations = "test-fixtures/emr_configurations.json"

  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
    aws_iam_role_policy_attachment.emr_autoscaling_role,
  ]

  service_role         = aws_iam_role.emr_service.arn
  autoscaling_role     = aws_iam_role.emr_autoscaling_role.arn
  ebs_root_volume_size = 21
}
`, r),
	)
}

func testAccAWSEmrClusterConfigAdditionalInfo(r string) string {
	return testAccAWSEmrComposeConfig(false,
		testAccAWSEmrClusterConfigCurrentPartition(),
		testAccAWSEmrClusterConfigIAMServiceRoleBase(r),
		testAccAWSEmrClusterConfigIAMInstanceProfileBase(r),
		testAccAWSEmrClusterConfigIAMAutoscalingRole(r),
		fmt.Sprintf(`
resource "aws_emr_cluster" "tf-test-cluster" {
  name          = "%[1]s"
  release_label = "emr-4.6.0"
  applications  = ["Spark"]

  additional_info = <<EOF
{
	"instanceAwsClientConfiguration": {
		"proxyPort": 8099,
		"proxyHost": "myproxy.example.com"
	}
}
EOF

  ec2_attributes {
    subnet_id                         = aws_subnet.test.id
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = aws_iam_instance_profile.emr_instance_profile.arn
  }

  master_instance_group {
    instance_type = "c4.large"
  }

  core_instance_group {
    instance_count = 1
    instance_type  = "c4.large"
  }

  tags = {
    role     = "rolename"
    dns_zone = "env_zone"
    env      = "env"
    name     = "name-env"
  }

  keep_job_flow_alive_when_no_steps = true
  termination_protection            = false

  scale_down_behavior = "TERMINATE_AT_TASK_COMPLETION"

  configurations = "test-fixtures/emr_configurations.json"

  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
    aws_iam_role_policy_attachment.emr_autoscaling_role,
  ]

  service_role         = aws_iam_role.emr_service.arn
  autoscaling_role     = aws_iam_role.emr_autoscaling_role.arn
  ebs_root_volume_size = 21
}
`, r),
	)
}

func testAccAWSEmrClusterConfigConfigurationsJson(r string) string {
	return testAccAWSEmrComposeConfig(false,
		testAccAWSEmrClusterConfigCurrentPartition(),
		testAccAWSEmrClusterConfigIAMServiceRoleBase(r),
		testAccAWSEmrClusterConfigIAMInstanceProfileBase(r),
		fmt.Sprintf(`
resource "aws_emr_cluster" "tf-test-cluster" {
  name          = "%[1]s"
  release_label = "emr-4.6.0"
  applications  = ["Hadoop", "Spark"]

  ec2_attributes {
    subnet_id                         = aws_subnet.test.id
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = aws_iam_instance_profile.emr_instance_profile.arn
  }

  master_instance_group {
    instance_type = "c4.large"
  }

  core_instance_group {
    instance_count = 1
    instance_type  = "c4.large"
  }

  keep_job_flow_alive_when_no_steps = true
  termination_protection            = false

  configurations_json = <<EOF
[
   {
     "Classification": "hadoop-env",
     "Configurations": [
       {
         "Classification": "export",
         "Properties": {
           "JAVA_HOME": "/usr/lib/jvm/java-1.8.0"
         }
       }
     ],
     "Properties": {}
   },
   {
     "Classification": "spark-env",
     "Configurations": [
       {
         "Classification": "export",
         "Properties": {
           "JAVA_HOME": "/usr/lib/jvm/java-1.8.0"
         }
       }
     ],
     "Properties": {}
   }
]
EOF

  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
  ]

  service_role         = aws_iam_role.emr_service.arn
  ebs_root_volume_size = 21
}
`, r),
	)
}

func testAccAWSEmrClusterConfig_Kerberos_ClusterDedicatedKdc(r string, password string) string {
	return testAccAWSEmrComposeConfig(false, fmt.Sprintf(`
resource "aws_emr_security_configuration" "foo" {
  configuration = <<EOF
{
  "AuthenticationConfiguration": {
    "KerberosConfiguration": {
      "Provider": "ClusterDedicatedKdc",
      "ClusterDedicatedKdcConfiguration": {
        "TicketLifetimeInHours": 24
      }
    }
  }
}
EOF
}

resource "aws_emr_cluster" "tf-test-cluster" {
  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  name                              = "%[1]s"
  release_label                     = "emr-5.12.0"
  security_configuration            = aws_emr_security_configuration.foo.name
  service_role                      = "EMR_DefaultRole"
  termination_protection            = false

  master_instance_group {
    instance_type = "c4.large"
  }

  core_instance_group {
    instance_count = 1
    instance_type  = "c4.large"
  }

  ec2_attributes {
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = "EMR_EC2_DefaultRole"
    subnet_id                         = aws_subnet.test.id
  }

  kerberos_attributes {
    kdc_admin_password = "%[2]s"
    realm              = "EC2.INTERNAL"
  }

  depends_on = [aws_route_table_association.test]
}
`, r, password))
}

func testAccAWSEmrClusterConfig_SecurityConfiguration(r string) string {
	return testAccAWSEmrComposeConfig(false,
		testAccAWSEmrClusterConfigCurrentPartition(),
		testAccAWSEmrClusterConfigIAMServiceRoleBase(r),
		testAccAWSEmrClusterConfigIAMInstanceProfileBase(r),
		testAccAWSEmrClusterConfigIAMAutoscalingRole(r),
		fmt.Sprintf(`
resource "aws_emr_cluster" "tf-test-cluster" {
  name          = "%[1]s"
  release_label = "emr-5.5.0"
  applications  = ["Spark"]

  ec2_attributes {
    subnet_id                         = aws_subnet.test.id
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = aws_iam_instance_profile.emr_instance_profile.arn
  }

  master_instance_group {
    instance_type = "c4.large"
  }

  core_instance_group {
    instance_count = 1
    instance_type  = "c4.large"
  }

  security_configuration = aws_emr_security_configuration.foo.name

  tags = {
    role     = "rolename"
    dns_zone = "env_zone"
    env      = "env"
    name     = "name-env"
  }

  keep_job_flow_alive_when_no_steps = true
  termination_protection            = false

  configurations = "test-fixtures/emr_configurations.json"

  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
    aws_iam_role_policy_attachment.emr_autoscaling_role,
  ]

  service_role     = aws_iam_role.emr_service.arn
  autoscaling_role = aws_iam_role.emr_autoscaling_role.arn
}

resource "aws_emr_security_configuration" "foo" {
  configuration = <<EOF
{
  "EncryptionConfiguration": {
    "AtRestEncryptionConfiguration": {
      "S3EncryptionConfiguration": {
        "EncryptionMode": "SSE-S3"
      },
      "LocalDiskEncryptionConfiguration": {
        "EncryptionKeyProviderType": "AwsKms",
        "AwsKmsKey": "${aws_kms_key.foo.arn}"
      }
    },
    "EnableInTransitEncryption": false,
    "EnableAtRestEncryption": true
  }
}
EOF
}

resource "aws_kms_key" "foo" {
  description             = "Terraform %[1]s"
  deletion_window_in_days = 7

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
`, r))
}

const testAccAWSEmrClusterConfig_Step_DebugLoggingStep = `
  # Example from: https://docs.aws.amazon.com/emr/latest/ManagementGuide/emr-plan-debugging.html
  step {
    action_on_failure = "TERMINATE_CLUSTER"
    name              = "Setup Hadoop Debugging"

    hadoop_jar_step {
      jar  = "command-runner.jar"
      args = ["state-pusher-script"]
    }
  }
`

const testAccAWSEmrClusterConfig_Step_SparkStep = `
  # Example from: https://docs.aws.amazon.com/emr/latest/ReleaseGuide/emr-spark-submit-step.html
  step {
    action_on_failure = "CONTINUE"
    name              = "Spark Step"

    hadoop_jar_step {
      jar  = "command-runner.jar"
      args = ["spark-example", "SparkPi", "10"]
    }
  }
`

func testAccAWSEmrClusterConfig_Step_Multiple(r string) string {
	stepConfig := testAccAWSEmrClusterConfig_Step_DebugLoggingStep + testAccAWSEmrClusterConfig_Step_SparkStep
	return testAccAWSEmrClusterConfig_Step(r, stepConfig)
}

func testAccAWSEmrClusterConfig_Step_NoBlocks(r string) string {
	return testAccAWSEmrClusterConfig_Step(r, "")
}

func testAccAWSEmrClusterConfig_Step_Single(r string) string {
	return testAccAWSEmrClusterConfig_Step(r, testAccAWSEmrClusterConfig_Step_DebugLoggingStep)
}

func testAccAWSEmrClusterConfig_Step_Zeroed(r string) string {
	return testAccAWSEmrClusterConfig_Step(r, "step = []")
}

func testAccAWSEmrClusterConfig_Step(r string, stepConfig string) string {
	return testAccAWSEmrComposeConfig(false, fmt.Sprintf(`
resource "aws_emr_cluster" "tf-test-cluster" {
  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  log_uri                           = "s3://${aws_s3_bucket.test.bucket}/"
  name                              = "%[1]s"
  release_label                     = "emr-5.12.0"
  service_role                      = "EMR_DefaultRole"
  termination_protection            = false

  master_instance_group {
    instance_type = "c4.large"
  }

  core_instance_group {
    instance_count = 1
    instance_type  = "c4.large"
  }

  ec2_attributes {
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = "EMR_EC2_DefaultRole"
    subnet_id                         = aws_subnet.test.id
  }

  %[2]s

  depends_on = [aws_route_table_association.test]
}

resource "aws_s3_bucket" "test" {
  bucket        = "%[1]s"
  force_destroy = true
}
`, r, stepConfig))
}

func testAccAWSEmrClusterConfigCoreInstanceGroupAutoscalingPolicy(rName, autoscalingPolicy string) string {
	return testAccAWSEmrComposeConfig(false,
		testAccAWSEmrClusterConfigCurrentPartition(),
		fmt.Sprintf(`
data "aws_iam_policy_document" "test" {
  statement {
    actions = ["sts:AssumeRole"]
    effect  = "Allow"

    principals {
      identifiers = [
        "application-autoscaling.${data.aws_partition.current.dns_suffix}",
        "elasticmapreduce.${data.aws_partition.current.dns_suffix}",
      ]
      type = "Service"
    }
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test.json
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonElasticMapReduceforAutoScalingRole"
}

resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  autoscaling_role                  = aws_iam_role.test.arn
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.12.0"
  service_role                      = "EMR_DefaultRole"

  ec2_attributes {
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = "EMR_EC2_DefaultRole"
    subnet_id                         = aws_subnet.test.id
  }

  master_instance_group {
    instance_type = "m4.large"
  }

  core_instance_group {
    autoscaling_policy = <<POLICY
%[2]s
POLICY
    instance_type      = "m4.large"
  }

  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.test,
  ]
}
`, rName, autoscalingPolicy),
	)
}

func testAccAWSEmrClusterConfigCoreInstanceGroupAutoscalingPolicyRemoved(rName string) string {
	return testAccAWSEmrComposeConfig(false,
		testAccAWSEmrClusterConfigCurrentPartition(),
		fmt.Sprintf(`
data "aws_iam_policy_document" "test" {
  statement {
    actions = ["sts:AssumeRole"]
    effect  = "Allow"

    principals {
      identifiers = [
        "application-autoscaling.${data.aws_partition.current.dns_suffix}",
        "elasticmapreduce.${data.aws_partition.current.dns_suffix}",
      ]
      type = "Service"
    }
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test.json
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonElasticMapReduceforAutoScalingRole"
}

resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  autoscaling_role                  = aws_iam_role.test.arn
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.12.0"
  service_role                      = "EMR_DefaultRole"

  ec2_attributes {
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = "EMR_EC2_DefaultRole"
    subnet_id                         = aws_subnet.test.id
  }

  master_instance_group {
    instance_type = "m4.large"
  }

  core_instance_group {
    instance_type = "m4.large"
  }

  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.test,
  ]
}
`, rName),
	)
}

func testAccAWSEmrClusterConfigCoreInstanceGroupBidPrice(rName, bidPrice string) string {
	return testAccAWSEmrComposeConfig(false, fmt.Sprintf(`
resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.12.0"
  service_role                      = "EMR_DefaultRole"

  ec2_attributes {
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = "EMR_EC2_DefaultRole"
    subnet_id                         = aws_subnet.test.id
  }

  master_instance_group {
    instance_type = "m4.large"
  }

  core_instance_group {
    bid_price     = %[2]q
    instance_type = "m4.large"
  }

  depends_on = [aws_route_table_association.test]
}
`, rName, bidPrice),
	)
}

func testAccAWSEmrClusterConfigCoreInstanceGroupInstanceCount(rName string, instanceCount int) string {
	return testAccAWSEmrComposeConfig(false, fmt.Sprintf(`
resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.12.0"
  service_role                      = "EMR_DefaultRole"

  ec2_attributes {
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = "EMR_EC2_DefaultRole"
    subnet_id                         = aws_subnet.test.id
  }

  master_instance_group {
    instance_type = "m4.large"
  }

  core_instance_group {
    instance_count = %[2]d
    instance_type  = "m4.large"
  }

  depends_on = [aws_route_table_association.test]
}
`, rName, instanceCount),
	)
}

func testAccAWSEmrClusterConfigCoreInstanceGroupInstanceType(rName, instanceType string) string {
	return testAccAWSEmrComposeConfig(false, fmt.Sprintf(`
resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.12.0"
  service_role                      = "EMR_DefaultRole"

  ec2_attributes {
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = "EMR_EC2_DefaultRole"
    subnet_id                         = aws_subnet.test.id
  }

  master_instance_group {
    instance_type = "m4.large"
  }

  core_instance_group {
    instance_type = %[2]q
  }

  depends_on = [aws_route_table_association.test]
}
`, rName, instanceType),
	)
}

func testAccAWSEmrClusterConfigCoreInstanceGroupName(rName, instanceGroupName string) string {
	return testAccAWSEmrComposeConfig(false, fmt.Sprintf(`
resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.12.0"
  service_role                      = "EMR_DefaultRole"

  ec2_attributes {
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = "EMR_EC2_DefaultRole"
    subnet_id                         = aws_subnet.test.id
  }

  master_instance_group {
    instance_type = "m4.large"
  }

  core_instance_group {
    instance_type = "m4.large"
    name          = %[2]q
  }

  depends_on = [aws_route_table_association.test]
}
`, rName, instanceGroupName),
	)
}

func testAccAWSEmrClusterConfigEc2AttributesDefaultManagedSecurityGroups(rName string) string {
	return testAccAWSEmrComposeConfig(false, fmt.Sprintf(`
resource "aws_emr_cluster" "tf-test-cluster" {
  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.28.0"
  service_role                      = "EMR_DefaultRole"

  ec2_attributes {
    instance_profile = "EMR_EC2_DefaultRole"
    subnet_id        = aws_subnet.test.id
  }

  master_instance_group {
    instance_type = "m4.large"
  }

  depends_on = [aws_route_table_association.test]
}
`, rName),
	)
}

func testAccAWSEmrClusterConfigMasterInstanceGroupBidPrice(rName, bidPrice string) string {
	return testAccAWSEmrComposeConfig(false, fmt.Sprintf(`
resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.12.0"
  service_role                      = "EMR_DefaultRole"

  ec2_attributes {
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = "EMR_EC2_DefaultRole"
    subnet_id                         = aws_subnet.test.id
  }

  master_instance_group {
    bid_price     = %[2]q
    instance_type = "m4.large"
  }

  depends_on = [aws_route_table_association.test]
}
`, rName, bidPrice),
	)
}

func testAccAWSEmrClusterConfigMasterInstanceGroupInstanceCount(rName string, instanceCount int) string {
	return testAccAWSEmrComposeConfig(true, fmt.Sprintf(`
resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.24.1"
  service_role                      = "EMR_DefaultRole"

  # Termination protection is automatically enabled for multiple master clusters
  termination_protection = false

  ec2_attributes {
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = "EMR_EC2_DefaultRole"
    subnet_id                         = aws_subnet.test.id
  }

  master_instance_group {
    instance_count = %[2]d
    instance_type  = "m4.large"
  }

  # core_instance_group is required with multiple masters
  core_instance_group {
    instance_type = "m4.large"
  }

  depends_on = [aws_route_table_association.test]
}
`, rName, instanceCount),
	)
}

func testAccAWSEmrClusterConfigMasterInstanceGroupInstanceType(rName, instanceType string) string {
	return testAccAWSEmrComposeConfig(false, fmt.Sprintf(`
resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.12.0"
  service_role                      = "EMR_DefaultRole"

  ec2_attributes {
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = "EMR_EC2_DefaultRole"
    subnet_id                         = aws_subnet.test.id
  }

  master_instance_group {
    instance_type = %[2]q
  }

  depends_on = [aws_route_table_association.test]
}
`, rName, instanceType),
	)
}

func testAccAWSEmrClusterConfigMasterInstanceGroupName(rName, instanceGroupName string) string {
	return testAccAWSEmrComposeConfig(false, fmt.Sprintf(`
resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.12.0"
  service_role                      = "EMR_DefaultRole"

  ec2_attributes {
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = "EMR_EC2_DefaultRole"
    subnet_id                         = aws_subnet.test.id
  }

  master_instance_group {
    instance_type = "m4.large"
    name          = %[2]q
  }

  depends_on = [aws_route_table_association.test]
}
`, rName, instanceGroupName),
	)
}

func testAccAWSEmrClusterConfigTerminationPolicy(r string, term string) string {
	return testAccAWSEmrComposeConfig(false,
		testAccAWSEmrClusterConfigCurrentPartition(),
		testAccAWSEmrClusterConfigIAMServiceRoleBase(r),
		testAccAWSEmrClusterConfigIAMInstanceProfileBase(r),
		testAccAWSEmrClusterConfigIAMAutoscalingRole(r),
		fmt.Sprintf(`
resource "aws_emr_cluster" "tf-test-cluster" {
  name          = "%[1]s"
  release_label = "emr-4.6.0"
  applications  = ["Spark"]

  ec2_attributes {
    subnet_id                         = aws_subnet.test.id
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = aws_iam_instance_profile.emr_instance_profile.arn
  }

  master_instance_group {
    instance_type = "c4.large"
  }

  core_instance_group {
    instance_count = 1
    instance_type  = "c4.large"
  }

  tags = {
    role     = "rolename"
    dns_zone = "env_zone"
    env      = "env"
    name     = "name-env"
  }

  keep_job_flow_alive_when_no_steps = true
  termination_protection            = %[2]s

  configurations = "test-fixtures/emr_configurations.json"

  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
    aws_iam_role_policy_attachment.emr_autoscaling_role,
  ]

  service_role     = aws_iam_role.emr_service.arn
  autoscaling_role = aws_iam_role.emr_autoscaling_role.arn
}
`, r, term),
	)
}

func testAccAWSEmrClusterConfig_keepJob(r string, keepJob string) string {
	return testAccAWSEmrComposeConfig(false,
		testAccAWSEmrClusterConfigCurrentPartition(),
		testAccAWSEmrClusterConfigIAMServiceRoleBase(r),
		testAccAWSEmrClusterConfigIAMInstanceProfileBase(r),
		testAccAWSEmrClusterConfigIAMAutoscalingRole(r),
		fmt.Sprintf(`
resource "aws_emr_cluster" "tf-test-cluster" {
  name          = "%[1]s"
  release_label = "emr-4.6.0"
  applications  = ["Spark"]

  ec2_attributes {
    subnet_id                         = aws_subnet.test.id
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = aws_iam_instance_profile.emr_instance_profile.arn
  }

  master_instance_group {
    instance_type = "c4.large"
  }

  core_instance_group {
    instance_count = 1
    instance_type  = "c4.large"
  }

  tags = {
    role     = "rolename"
    dns_zone = "env_zone"
    env      = "env"
    name     = "name-env"
  }

  keep_job_flow_alive_when_no_steps = %s
  termination_protection            = false

  step {
    action_on_failure = "CONTINUE"
    name              = "Sleep Step"

    hadoop_jar_step {
      jar  = "command-runner.jar"
      args = ["/bin/sleep", "60"]
    }
  }

  configurations = "test-fixtures/emr_configurations.json"

  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
    aws_iam_role_policy_attachment.emr_autoscaling_role,
  ]

  service_role     = aws_iam_role.emr_service.arn
  autoscaling_role = aws_iam_role.emr_autoscaling_role.arn
}
`, r, keepJob),
	)
}

func testAccAWSEmrClusterConfigVisibleToAllUsersUpdated(r string) string {
	return testAccAWSEmrComposeConfig(false,
		testAccAWSEmrClusterConfigCurrentPartition(),
		testAccAWSEmrClusterConfigIAMServiceRoleBase(r),
		testAccAWSEmrClusterConfigIAMInstanceProfileBase(r),
		testAccAWSEmrClusterConfigIAMAutoscalingRole(r),
		fmt.Sprintf(`
resource "aws_emr_cluster" "tf-test-cluster" {
  name          = "%[1]s"
  release_label = "emr-4.6.0"
  applications  = ["Spark"]

  ec2_attributes {
    subnet_id                         = aws_subnet.test.id
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = aws_iam_instance_profile.emr_instance_profile.arn
  }

  master_instance_group {
    instance_type = "c4.large"
  }

  core_instance_group {
    instance_count = 1
    instance_type  = "c4.large"
  }

  tags = {
    role     = "rolename"
    dns_zone = "env_zone"
    env      = "env"
    name     = "name-env"
  }

  keep_job_flow_alive_when_no_steps = true
  visible_to_all_users              = false

  configurations = "test-fixtures/emr_configurations.json"

  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
    aws_iam_role_policy_attachment.emr_autoscaling_role,
  ]

  service_role     = aws_iam_role.emr_service.arn
  autoscaling_role = aws_iam_role.emr_autoscaling_role.arn
}
`, r),
	)
}

func testAccAWSEmrClusterConfigUpdatedTags(r string) string {
	return testAccAWSEmrComposeConfig(false,
		testAccAWSEmrClusterConfigCurrentPartition(),
		testAccAWSEmrClusterConfigIAMServiceRoleBase(r),
		testAccAWSEmrClusterConfigIAMInstanceProfileBase(r),
		testAccAWSEmrClusterConfigIAMAutoscalingRole(r),
		fmt.Sprintf(`
resource "aws_emr_cluster" "tf-test-cluster" {
  name          = "%[1]s"
  release_label = "emr-4.6.0"
  applications  = ["Spark"]

  ec2_attributes {
    subnet_id                         = aws_subnet.test.id
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = aws_iam_instance_profile.emr_instance_profile.arn
  }

  master_instance_group {
    instance_type = "c4.large"
  }

  core_instance_group {
    instance_count = 1
    instance_type  = "c4.large"
  }

  tags = {
    dns_zone = "new_zone"
    Env      = "production"
    name     = "name-env"
  }

  keep_job_flow_alive_when_no_steps = true
  termination_protection            = false

  configurations = "test-fixtures/emr_configurations.json"

  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
    aws_iam_role_policy_attachment.emr_autoscaling_role,
  ]

  service_role     = aws_iam_role.emr_service.arn
  autoscaling_role = aws_iam_role.emr_autoscaling_role.arn
}
`, r),
	)
}

func testAccAWSEmrClusterConfigUpdatedRootVolumeSize(r string) string {
	return testAccAWSEmrComposeConfig(false,
		testAccAWSEmrClusterConfigCurrentPartition(),
		testAccAWSEmrClusterConfigIAMServiceRoleBase(r),
		testAccAWSEmrClusterConfigIAMInstanceProfileBase(r),
		testAccAWSEmrClusterConfigIAMAutoscalingRole(r),
		fmt.Sprintf(`
resource "aws_emr_cluster" "tf-test-cluster" {
  name          = "%[1]s"
  release_label = "emr-4.6.0"
  applications  = ["Spark"]

  ec2_attributes {
    subnet_id                         = aws_subnet.test.id
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = aws_iam_instance_profile.emr_instance_profile.arn
  }

  master_instance_group {
    instance_type = "c4.large"
  }

  core_instance_group {
    instance_count = 1
    instance_type  = "c4.large"
  }

  tags = {
    role     = "rolename"
    dns_zone = "env_zone"
    env      = "env"
    name     = "name-env"
  }

  keep_job_flow_alive_when_no_steps = true
  termination_protection            = false

  configurations = "test-fixtures/emr_configurations.json"

  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
    aws_iam_role_policy_attachment.emr_autoscaling_role,
  ]

  service_role         = aws_iam_role.emr_service.arn
  autoscaling_role     = aws_iam_role.emr_autoscaling_role.arn
  ebs_root_volume_size = 48
}
`, r),
	)
}

func testAccAWSEmrClusterConfigS3Logging(r string) string {
	return testAccAWSEmrComposeConfig(false,
		testAccAWSEmrClusterConfigCurrentPartition(),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = "%[1]s"
  force_destroy = true
}

resource "aws_emr_cluster" "tf-test-cluster" {
  name          = "%[1]s"
  release_label = "emr-4.6.0"
  applications  = ["Spark"]

  termination_protection            = false
  keep_job_flow_alive_when_no_steps = true

  master_instance_group {
    instance_type = "c4.large"
  }

  core_instance_group {
    instance_count = 1
    instance_type  = "c4.large"
  }

  log_uri = "s3://${aws_s3_bucket.test.bucket}/"

  ec2_attributes {
    instance_profile                  = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:instance-profile/EMR_EC2_DefaultRole"
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    subnet_id                         = aws_subnet.test.id
  }

  service_role = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:role/EMR_DefaultRole"
}

data "aws_caller_identity" "current" {}
`, r),
	)
}

func testAccAWSEmrClusterConfigCustomAmiID(r string) string {
	return testAccAWSEmrComposeConfig(false,
		testAccAWSEmrClusterConfigCurrentPartition(),
		testAccAWSEmrClusterConfigIAMServiceRoleCustomAmiID(r),
		testAccAWSEmrClusterConfigIAMInstanceProfileBase(r),
		testAccAWSEmrClusterConfigIAMAutoscalingRole(r),
		fmt.Sprintf(`
resource "aws_emr_cluster" "tf-test-cluster" {
  name          = "%[1]s"
  release_label = "emr-5.7.0"
  applications  = ["Spark"]

  ec2_attributes {
    subnet_id                         = aws_subnet.test.id
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = aws_iam_instance_profile.emr_instance_profile.arn
  }

  master_instance_group {
    instance_type = "c4.large"
  }

  core_instance_group {
    instance_count = 1
    instance_type  = "c4.large"
  }

  tags = {
    role     = "rolename"
    dns_zone = "env_zone"
    env      = "env"
    name     = "name-env"
  }

  keep_job_flow_alive_when_no_steps = true
  termination_protection            = false

  configurations = "test-fixtures/emr_configurations.json"

  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
    aws_iam_role_policy_attachment.emr_autoscaling_role,
  ]

  service_role         = aws_iam_role.emr_service.arn
  autoscaling_role     = aws_iam_role.emr_autoscaling_role.arn
  ebs_root_volume_size = 48
  custom_ami_id        = data.aws_ami.emr-custom-ami.id
}

data "aws_ami" "emr-custom-ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*"]
  }

  filter {
    name   = "architecture"
    values = ["x86_64"]
  }

  filter {
    name   = "root-device-type"
    values = ["ebs"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}
`, r),
	)
}

func testAccAWSEmrClusterConfigStepConcurrencyLevel(rName string, stepConcurrencyLevel int) string {
	return testAccAWSEmrComposeConfig(false, fmt.Sprintf(`
resource "aws_emr_cluster" "tf-test-cluster" {
  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.28.0"
  service_role                      = "EMR_DefaultRole"

  ec2_attributes {
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = "EMR_EC2_DefaultRole"
    subnet_id                         = aws_subnet.test.id
  }

  master_instance_group {
    instance_type = "m4.large"
  }

  step_concurrency_level = %[2]d

  depends_on = [aws_route_table_association.test]
}
`, rName, stepConcurrencyLevel),
	)
}

func testAccAWSEmrClusterConfigBaseVpc(mapPublicIPOnLaunch bool) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  # Many instance types are not available in this availability zone
  exclude_zone_ids = ["usw2-az4"]
  state            = "available"
  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "tf-acc-test-emr-cluster"
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-emr-cluster"
  }
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id

  ingress {
    from_port = 0
    protocol  = "-1"
    self      = true
    to_port   = 0
  }

  egress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 0
    protocol    = "-1"
    to_port     = 0
  }

  tags = {
    Name = "tf-acc-test-emr-cluster"
  }

  # EMR will modify ingress rules
  lifecycle {
    ignore_changes = [ingress]
  }
}

resource "aws_subnet" "test" {
  availability_zone       = data.aws_availability_zones.available.names[0]
  cidr_block              = "10.0.0.0/24"
  map_public_ip_on_launch = %[1]t
  vpc_id                  = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-emr-cluster"
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }
}

resource "aws_route_table_association" "test" {
  route_table_id = aws_route_table.test.id
  subnet_id      = aws_subnet.test.id
}
`, mapPublicIPOnLaunch)
}

func testAccAWSEmrClusterConfigIAMServiceRoleBase(r string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "emr_service" {
  name = "%[1]s_default_role"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "elasticmapreduce.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_role_policy_attachment" "emr_service" {
  role       = aws_iam_role.emr_service.id
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonElasticMapReduceRole"
}

`, r)
}

func testAccAWSEmrClusterConfigIAMServiceRoleCustomAmiID(r string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "emr_service" {
  name = "%[1]s_default_role"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "elasticmapreduce.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_role_policy_attachment" "emr_service" {
  role       = aws_iam_role.emr_service.id
  policy_arn = aws_iam_policy.emr_service.arn
}

resource "aws_iam_policy" "emr_service" {
  name = "%[1]s_emr"

  policy = <<EOT
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Resource": "*",
        "Action": [
            "ec2:AuthorizeSecurityGroupEgress",
            "ec2:AuthorizeSecurityGroupIngress",
            "ec2:CancelSpotInstanceRequests",
            "ec2:CreateNetworkInterface",
            "ec2:CreateSecurityGroup",
            "ec2:CreateTags",
            "ec2:DeleteNetworkInterface",
            "ec2:DeleteSecurityGroup",
            "ec2:DeleteTags",
            "ec2:DescribeAvailabilityZones",
            "ec2:DescribeAccountAttributes",
            "ec2:DescribeDhcpOptions",
            "ec2:DescribeImages",
            "ec2:DescribeInstanceStatus",
            "ec2:DescribeInstances",
            "ec2:DescribeKeyPairs",
            "ec2:DescribeNetworkAcls",
            "ec2:DescribeNetworkInterfaces",
            "ec2:DescribePrefixLists",
            "ec2:DescribeRouteTables",
            "ec2:DescribeSecurityGroups",
            "ec2:DescribeSpotInstanceRequests",
            "ec2:DescribeSpotPriceHistory",
            "ec2:DescribeSubnets",
            "ec2:DescribeVpcAttribute",
            "ec2:DescribeVpcEndpoints",
            "ec2:DescribeVpcEndpointServices",
            "ec2:DescribeVpcs",
            "ec2:DetachNetworkInterface",
            "ec2:ModifyImageAttribute",
            "ec2:ModifyInstanceAttribute",
            "ec2:RequestSpotInstances",
            "ec2:RevokeSecurityGroupEgress",
            "ec2:RunInstances",
            "ec2:TerminateInstances",
            "ec2:DeleteVolume",
            "ec2:DescribeVolumeStatus",
            "ec2:DescribeVolumes",
            "ec2:DetachVolume",
            "iam:GetRole",
            "iam:GetRolePolicy",
            "iam:ListInstanceProfiles",
            "iam:ListRolePolicies",
            "iam:PassRole",
            "s3:CreateBucket",
            "s3:Get*",
            "s3:List*",
            "sdb:BatchPutAttributes",
            "sdb:Select",
            "sqs:CreateQueue",
            "sqs:Delete*",
            "sqs:GetQueue*",
            "sqs:PurgeQueue",
            "sqs:ReceiveMessage"
        ]
    }]
}
EOT
}
`, r)
}

func testAccAWSEmrClusterConfigIAMInstanceProfileBase(r string) string {
	return fmt.Sprintf(`
resource "aws_iam_instance_profile" "emr_instance_profile" {
  name = "%[1]s_profile"
  role = aws_iam_role.emr_instance_profile.name
}

resource "aws_iam_role" "emr_instance_profile" {
  name = "%[1]s_profile_role"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_role_policy_attachment" "emr_instance_profile" {
  role       = aws_iam_role.emr_instance_profile.id
  policy_arn = aws_iam_policy.emr_instance_profile.arn
}

resource "aws_iam_policy" "emr_instance_profile" {
  name = "%[1]s_profile"

  policy = <<EOT
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Resource": "*",
        "Action": [
            "cloudwatch:*",
            "dynamodb:*",
            "ec2:Describe*",
            "elasticmapreduce:Describe*",
            "elasticmapreduce:ListBootstrapActions",
            "elasticmapreduce:ListClusters",
            "elasticmapreduce:ListInstanceGroups",
            "elasticmapreduce:ListInstances",
            "elasticmapreduce:ListSteps",
            "kinesis:CreateStream",
            "kinesis:DeleteStream",
            "kinesis:DescribeStream",
            "kinesis:GetRecords",
            "kinesis:GetShardIterator",
            "kinesis:MergeShards",
            "kinesis:PutRecord",
            "kinesis:SplitShard",
            "rds:Describe*",
            "s3:*",
            "sdb:*",
            "sns:*",
            "sqs:*"
        ]
    }]
}
EOT
}
`, r)
}

func testAccAWSEmrClusterConfigIAMAutoscalingRole(r string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "emr_autoscaling_role" {
  name               = "%[1]s_autoscaling_role"
  assume_role_policy = data.aws_iam_policy_document.emr_autoscaling_role.json
}

data "aws_iam_policy_document" "emr_autoscaling_role" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["elasticmapreduce.${data.aws_partition.current.dns_suffix}", "application-autoscaling.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

resource "aws_iam_role_policy_attachment" "emr_autoscaling_role" {
  role       = aws_iam_role.emr_autoscaling_role.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonElasticMapReduceforAutoScalingRole"
}
`, r)
}

func testAccAWSEmrClusterConfigBootstrapActionBucket(r string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "tester" {
  bucket = "%[1]s"
  acl    = "public-read"
}

resource "aws_s3_bucket_object" "testobject" {
  bucket  = aws_s3_bucket.tester.bucket
  key     = "testscript.sh"
  content = <<EOF
#!/bin/bash
echo $@
EOF


  acl = "public-read"
}
`, r)
}

func testAccAWSEmrEbsConfig(rName string, volumesPerInstance int) string {
	return testAccAWSEmrComposeConfig(false, fmt.Sprintf(`
resource "aws_emr_cluster" "tf-test-cluster" {
  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  name                              = "%[1]s"
  release_label                     = "emr-5.28.0"
  service_role                      = "EMR_DefaultRole"

  ec2_attributes {
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = "EMR_EC2_DefaultRole"
    subnet_id                         = aws_subnet.test.id
  }

  master_instance_group {
    instance_type = "m4.large"
    ebs_config {
      size                 = 32
      type                 = "gp2"
      volumes_per_instance = %[2]d
    }
    ebs_config {
      size                 = 50
      type                 = "gp2"
      volumes_per_instance = %[2]d
    }
  }
  core_instance_group {
    instance_count = 1
    instance_type  = "m4.large"
    ebs_config {
      size                 = 32
      type                 = "gp2"
      volumes_per_instance = %[2]d
    }
  }

  depends_on = [aws_route_table_association.test]
}
`, rName, volumesPerInstance),
	)
}

func testAccAWSEmrClusterConfigInstanceFleets(r string) string {
	return testAccAWSEmrComposeConfig(false,
		testAccAWSEmrClusterConfigCurrentPartition(),
		testAccAWSEmrClusterConfigIAMServiceRoleBase(r),
		testAccAWSEmrClusterConfigIAMInstanceProfileBase(r),
		testAccAWSEmrClusterConfigBootstrapActionBucket(r),
		fmt.Sprintf(`
resource "aws_emr_cluster" "tf-test-cluster" {
  name          = "%[1]s"
  release_label = "emr-5.30.1"
  applications  = ["Hadoop", "Hive"]
  log_uri       = "s3n://terraform/testlog/"

  master_instance_fleet {
    instance_type_configs {
      instance_type = "m3.xlarge"
    }

    target_on_demand_capacity = 1
  }
  core_instance_fleet {
    instance_type_configs {
      bid_price_as_percentage_of_on_demand_price = 80
      ebs_config {
        size                 = 100
        type                 = "gp2"
        volumes_per_instance = 1
      }
      instance_type     = "m3.xlarge"
      weighted_capacity = 1
    }
    instance_type_configs {
      bid_price_as_percentage_of_on_demand_price = 100
      ebs_config {
        size                 = 100
        type                 = "gp2"
        volumes_per_instance = 1
      }
      instance_type     = "m4.xlarge"
      weighted_capacity = 1
    }
    instance_type_configs {
      bid_price_as_percentage_of_on_demand_price = 100
      ebs_config {
        size                 = 100
        type                 = "gp2"
        volumes_per_instance = 1
      }
      instance_type     = "m4.2xlarge"
      weighted_capacity = 2
    }
    launch_specifications {
      spot_specification {
        allocation_strategy      = "capacity-optimized"
        block_duration_minutes   = 0
        timeout_action           = "SWITCH_TO_ON_DEMAND"
        timeout_duration_minutes = 10
      }
    }
    name                      = "core fleet"
    target_on_demand_capacity = 0
    target_spot_capacity      = 2
  }
  service_role = aws_iam_role.emr_service.arn
  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
  ]

  ec2_attributes {
    subnet_id                         = aws_subnet.test.id
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = aws_iam_instance_profile.emr_instance_profile.arn
  }

  bootstrap_action {
    path = "s3://elasticmapreduce/bootstrap-actions/run-if"
    name = "runif"
    args = ["instance.isMaster=true", "echo running on master node"]
  }
}
`, r),
	)
}

func testAccAWSEmrClusterConfigInstanceFleetsMasterOnly(r string) string {
	return testAccAWSEmrComposeConfig(false,
		testAccAWSEmrClusterConfigCurrentPartition(),
		testAccAWSEmrClusterConfigIAMServiceRoleBase(r),
		testAccAWSEmrClusterConfigIAMInstanceProfileBase(r),
		testAccAWSEmrClusterConfigBootstrapActionBucket(r),
		fmt.Sprintf(`
resource "aws_emr_cluster" "tf-test-cluster" {
  name          = "%[1]s"
  release_label = "emr-5.30.1"
  applications  = ["Hadoop", "Hive"]
  log_uri       = "s3n://terraform/testlog/"

  master_instance_fleet {
    instance_type_configs {
      instance_type = "m3.xlarge"
    }

    target_on_demand_capacity = 1
  }
  service_role = aws_iam_role.emr_service.arn
  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
  ]

  ec2_attributes {
    subnet_id                         = aws_subnet.test.id
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = aws_iam_instance_profile.emr_instance_profile.arn
  }

  bootstrap_action {
    path = "s3://elasticmapreduce/bootstrap-actions/run-if"
    name = "runif"
    args = ["instance.isMaster=true", "echo running on master node"]
  }
}
`, r),
	)
}
