package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

// add sweeper to delete known test subnets
func init() {
	resource.AddTestSweepers("aws_subnet", &resource.Sweeper{
		Name: "aws_subnet",
		F:    testSweepSubnets,
		Dependencies: []string{
			"aws_autoscaling_group",
			"aws_batch_compute_environment",
			"aws_elastic_beanstalk_environment",
			"aws_cloudhsm_v2_cluster",
			"aws_db_subnet_group",
			"aws_directory_service_directory",
			"aws_ec2_client_vpn_endpoint",
			"aws_ec2_transit_gateway_vpc_attachment",
			"aws_efs_file_system",
			"aws_eks_cluster",
			"aws_elasticache_cluster",
			"aws_elasticache_replication_group",
			"aws_elasticsearch_domain",
			"aws_elb",
			"aws_emr_cluster",
			"aws_fsx_lustre_file_system",
			"aws_fsx_windows_file_system",
			"aws_lambda_function",
			"aws_lb",
			"aws_mq_broker",
			"aws_msk_cluster",
			"aws_network_interface",
			"aws_redshift_cluster",
			"aws_route53_resolver_endpoint",
			"aws_sagemaker_notebook_instance",
			"aws_spot_fleet_request",
			"aws_vpc_endpoint",
		},
	})
}

func testSweepSubnets(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*AWSClient).ec2conn
	sweepResources := make([]*testSweepResource, 0)
	var errs *multierror.Error

	input := &ec2.DescribeSubnetsInput{}

	err = conn.DescribeSubnetsPages(input, func(page *ec2.DescribeSubnetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, subnet := range page.Subnets {
			if subnet == nil {
				continue
			}

			id := aws.StringValue(subnet.SubnetId)

			if aws.BoolValue(subnet.DefaultForAz) {
				log.Printf("[DEBUG] Skipping default EC2 Subnet: %s", id)
				continue
			}

			r := resourceAwsSubnet()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, NewTestSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error describing EC2 Subnets for %s: %w", region, err))
	}

	if err = testSweepResourceOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping EC2 Subnets for %s: %w", region, err))
	}

	if testSweepSkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping EC2 Subnet sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func TestAccAWSSubnet_basic(t *testing.T) {
	var v ec2.Subnet
	resourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`subnet/subnet-.+`)),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone"),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone_id"),
					resource.TestCheckResourceAttr(resourceName, "cidr_block", "10.1.1.0/24"),
					resource.TestCheckResourceAttr(resourceName, "customer_owned_ipv4_pool", ""),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "map_customer_owned_ip_on_launch", "false"),
					resource.TestCheckResourceAttr(resourceName, "map_public_ip_on_launch", "false"),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccAWSSubnet_tags(t *testing.T) {
	var v ec2.Subnet
	resourceName := "aws_subnet.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetTagsConfig1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(resourceName, &v),
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
				Config: testAccSubnetTagsConfig2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccSubnetTagsConfig1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2")),
			},
		},
	})
}

func TestAccAWSSubnet_defaultTags_providerOnly(t *testing.T) {
	var providers []*schema.Provider
	var subnet ec2.Subnet
	resourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: testAccProviderFactoriesInternal(&providers),
		CheckDestroy:      testAccCheckSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: composeConfig(
					testAccAWSProviderConfigDefaultTags_Tags1("providerkey1", "providervalue1"),
					testAccSubnetConfig,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(resourceName, &subnet),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey1", "providervalue1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: composeConfig(
					testAccAWSProviderConfigDefaultTags_Tags2("providerkey1", "providervalue1", "providerkey2", "providervalue2"),
					testAccSubnetConfig,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(resourceName, &subnet),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey1", "providervalue1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey2", "providervalue2"),
				),
			},
			{
				Config: composeConfig(
					testAccAWSProviderConfigDefaultTags_Tags1("providerkey1", "value1"),
					testAccSubnetConfig,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(resourceName, &subnet),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey1", "value1"),
				),
			},
		},
	})
}

func TestAccAWSSubnet_defaultTags_updateToProviderOnly(t *testing.T) {
	var providers []*schema.Provider
	var subnet ec2.Subnet
	resourceName := "aws_subnet.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: testAccProviderFactoriesInternal(&providers),
		CheckDestroy:      testAccCheckSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetTagsConfig1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(resourceName, &subnet),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", "value1"),
				),
			},
			{
				Config: composeConfig(
					testAccAWSProviderConfigDefaultTags_Tags1("key1", "value1"),
					testAccSubnetConfig,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(resourceName, &subnet),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", "value1"),
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

func TestAccAWSSubnet_defaultTags_updateToResourceOnly(t *testing.T) {
	var providers []*schema.Provider
	var subnet ec2.Subnet
	resourceName := "aws_subnet.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: testAccProviderFactoriesInternal(&providers),
		CheckDestroy:      testAccCheckSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: composeConfig(
					testAccAWSProviderConfigDefaultTags_Tags1("key1", "value1"),
					testAccSubnetConfig,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(resourceName, &subnet),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", "value1"),
				),
			},
			{
				Config: testAccSubnetTagsConfig1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(resourceName, &subnet),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", "value1"),
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

func TestAccAWSSubnet_defaultTags_providerAndResource_nonOverlappingTag(t *testing.T) {
	var providers []*schema.Provider
	var subnet ec2.Subnet
	resourceName := "aws_subnet.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: testAccProviderFactoriesInternal(&providers),
		CheckDestroy:      testAccCheckSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: composeConfig(
					testAccAWSProviderConfigDefaultTags_Tags1("providerkey1", "providervalue1"),
					testAccSubnetTagsConfig1(rName, "resourcekey1", "resourcevalue1"),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(resourceName, &subnet),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.resourcekey1", "resourcevalue1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey1", "providervalue1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.resourcekey1", "resourcevalue1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: composeConfig(
					testAccAWSProviderConfigDefaultTags_Tags1("providerkey1", "providervalue1"),
					testAccSubnetTagsConfig2(rName, "resourcekey1", "resourcevalue1", "resourcekey2", "resourcevalue2"),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(resourceName, &subnet),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.resourcekey1", "resourcevalue1"),
					resource.TestCheckResourceAttr(resourceName, "tags.resourcekey2", "resourcevalue2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey1", "providervalue1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.resourcekey1", "resourcevalue1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.resourcekey2", "resourcevalue2"),
				),
			},
			{
				Config: composeConfig(
					testAccAWSProviderConfigDefaultTags_Tags1("providerkey2", "providervalue2"),
					testAccSubnetTagsConfig1(rName, "resourcekey3", "resourcevalue3"),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(resourceName, &subnet),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.resourcekey3", "resourcevalue3"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey2", "providervalue2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.resourcekey3", "resourcevalue3"),
				),
			},
		},
	})
}

func TestAccAWSSubnet_defaultTags_providerAndResource_overlappingTag(t *testing.T) {
	var providers []*schema.Provider
	var subnet ec2.Subnet
	resourceName := "aws_subnet.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: testAccProviderFactoriesInternal(&providers),
		CheckDestroy:      testAccCheckSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: composeConfig(
					testAccAWSProviderConfigDefaultTags_Tags1("overlapkey1", "providervalue1"),
					testAccSubnetTagsConfig1(rName, "overlapkey1", "resourcevalue1"),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(resourceName, &subnet),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.overlapkey1", "resourcevalue1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: composeConfig(
					testAccAWSProviderConfigDefaultTags_Tags2("overlapkey1", "providervalue1", "overlapkey2", "providervalue2"),
					testAccSubnetTagsConfig2(rName, "overlapkey1", "resourcevalue1", "overlapkey2", "resourcevalue2"),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(resourceName, &subnet),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.overlapkey1", "resourcevalue1"),
					resource.TestCheckResourceAttr(resourceName, "tags.overlapkey2", "resourcevalue2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.overlapkey1", "resourcevalue1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.overlapkey2", "resourcevalue2"),
				),
			},
			{
				Config: composeConfig(
					testAccAWSProviderConfigDefaultTags_Tags1("overlapkey1", "providervalue1"),
					testAccSubnetTagsConfig1(rName, "overlapkey1", "resourcevalue2"),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(resourceName, &subnet),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.overlapkey1", "resourcevalue2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.overlapkey1", "resourcevalue2"),
				),
			},
		},
	})
}

func TestAccAWSSubnet_defaultTags_providerAndResource_duplicateTag(t *testing.T) {
	var providers []*schema.Provider
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: testAccProviderFactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: composeConfig(
					testAccAWSProviderConfigDefaultTags_Tags1("overlapkey", "overlapvalue"),
					testAccSubnetTagsConfig1(rName, "overlapkey", "overlapvalue"),
				),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`"tags" are identical to those in the "default_tags" configuration block`),
			},
		},
	})
}

func TestAccAWSSubnet_defaultAndIgnoreTags(t *testing.T) {
	var providers []*schema.Provider
	var subnet ec2.Subnet
	resourceName := "aws_subnet.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: testAccProviderFactoriesInternal(&providers),
		CheckDestroy:      testAccCheckSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetTagsConfig1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(resourceName, &subnet),
					testAccCheckSubnetUpdateTags(&subnet, nil, map[string]string{"defaultkey1": "defaultvalue1"}),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: composeConfig(
					testAccProviderConfigDefaultAndIgnoreTagsKeyPrefixes1("defaultkey1", "defaultvalue1", "defaultkey"),
					testAccSubnetTagsConfig1(rName, "key1", "value1"),
				),
				PlanOnly: true,
			},
			{
				Config: composeConfig(
					testAccProviderConfigDefaultAndIgnoreTagsKeys1("defaultkey1", "defaultvalue1"),
					testAccSubnetTagsConfig1(rName, "key1", "value1"),
				),
				PlanOnly: true,
			},
		},
	})
}

// TestAccAWSSubnet_updateTagsKnownAtApply ensures computed "tags_all"
// attributes are correctly determined when the provider-level default_tags block
// is left unused and resource tags are only known at apply time, thereby
// eliminating "Inconsistent final plan" errors
// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/18366
func TestAccAWSSubnet_updateTagsKnownAtApply(t *testing.T) {
	var providers []*schema.Provider
	var subnet ec2.Subnet
	resourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: testAccProviderFactoriesInternal(&providers),
		CheckDestroy:      testAccCheckSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetConfig_tagsComputedFromVpcDataSource1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(resourceName, &subnet),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
				),
			},
			{
				Config: testAccSubnetConfig_tagsComputedFromVpcDataSource2("key1", "value1", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(resourceName, &subnet),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "2"),
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

func TestAccAWSSubnet_ignoreTags(t *testing.T) {
	var providers []*schema.Provider
	var subnet ec2.Subnet
	resourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: testAccProviderFactoriesInternal(&providers),
		CheckDestroy:      testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(resourceName, &subnet),
					testAccCheckSubnetUpdateTags(&subnet, nil, map[string]string{"ignorekey1": "ignorevalue1"}),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config:   testAccProviderConfigIgnoreTagsKeyPrefixes1("ignorekey") + testAccSubnetConfig,
				PlanOnly: true,
			},
			{
				Config:   testAccProviderConfigIgnoreTagsKeys1("ignorekey1") + testAccSubnetConfig,
				PlanOnly: true,
			},
		},
	})
}

func TestAccAWSSubnet_ipv6(t *testing.T) {
	var before, after ec2.Subnet
	resourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetConfigIpv6,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(resourceName, &before),
					testAccCheckAwsSubnetIpv6BeforeUpdate(&before),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSubnetConfigIpv6UpdateAssignIpv6OnCreation,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(resourceName, &after),
					testAccCheckAwsSubnetIpv6AfterUpdate(&after),
				),
			},
			{
				Config: testAccSubnetConfigIpv6UpdateIpv6Cidr,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(resourceName, &after),
					testAccCheckAwsSubnetNotRecreated(t, &before, &after),
				),
			},
		},
	})
}

func TestAccAWSSubnet_enableIpv6(t *testing.T) {
	var subnet ec2.Subnet
	resourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetConfigPreIpv6,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(resourceName, &subnet),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "assign_ipv6_address_on_creation", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSubnetConfigIpv6,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(resourceName, &subnet),
					resource.TestCheckResourceAttrSet(resourceName, "ipv6_cidr_block"),
					resource.TestCheckResourceAttr(resourceName, "assign_ipv6_address_on_creation", "true"),
				),
			},
			{
				Config: testAccSubnetConfigPreIpv6,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(resourceName, &subnet),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "assign_ipv6_address_on_creation", "false"),
				),
			},
		},
	})
}

func TestAccAWSSubnet_availabilityZoneId(t *testing.T) {
	var v ec2.Subnet
	resourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetConfigAvailabilityZoneId(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone"),
					resource.TestCheckResourceAttrPair(resourceName, "availability_zone_id", "data.aws_availability_zones.available", "zone_ids.0"),
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

func TestAccAWSSubnet_disappears(t *testing.T) {
	var v ec2.Subnet
	resourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSubnet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSSubnet_CustomerOwnedIpv4Pool(t *testing.T) {
	var subnet ec2.Subnet
	coipDataSourceName := "data.aws_ec2_coip_pool.test"
	resourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSOutpostsOutposts(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetConfigCustomerOwnedIpv4Pool(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(resourceName, &subnet),
					resource.TestCheckResourceAttrPair(resourceName, "customer_owned_ipv4_pool", coipDataSourceName, "pool_id"),
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

func TestAccAWSSubnet_MapCustomerOwnedIpOnLaunch(t *testing.T) {
	var subnet ec2.Subnet
	resourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSOutpostsOutposts(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetConfigMapCustomerOwnedIpOnLaunch(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(resourceName, &subnet),
					resource.TestCheckResourceAttr(resourceName, "map_customer_owned_ip_on_launch", "true"),
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

func TestAccAWSSubnet_MapPublicIpOnLaunch(t *testing.T) {
	var subnet ec2.Subnet
	resourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetConfigMapPublicIpOnLaunch(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(resourceName, &subnet),
					resource.TestCheckResourceAttr(resourceName, "map_public_ip_on_launch", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSubnetConfigMapPublicIpOnLaunch(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(resourceName, &subnet),
					resource.TestCheckResourceAttr(resourceName, "map_public_ip_on_launch", "false"),
				),
			},
			{
				Config: testAccSubnetConfigMapPublicIpOnLaunch(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(resourceName, &subnet),
					resource.TestCheckResourceAttr(resourceName, "map_public_ip_on_launch", "true"),
				),
			},
		},
	})
}

func TestAccAWSSubnet_outpost(t *testing.T) {
	var v ec2.Subnet
	outpostDataSourceName := "data.aws_outposts_outpost.test"
	resourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSOutpostsOutposts(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetConfigOutpost(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "outpost_arn", outpostDataSourceName, "arn"),
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

func testAccCheckAwsSubnetIpv6BeforeUpdate(subnet *ec2.Subnet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if subnet.Ipv6CidrBlockAssociationSet == nil {
			return fmt.Errorf("Expected IPV6 CIDR Block Association")
		}

		if !aws.BoolValue(subnet.AssignIpv6AddressOnCreation) {
			return fmt.Errorf("bad AssignIpv6AddressOnCreation: %t", aws.BoolValue(subnet.AssignIpv6AddressOnCreation))
		}

		return nil
	}
}

func testAccCheckAwsSubnetIpv6AfterUpdate(subnet *ec2.Subnet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.BoolValue(subnet.AssignIpv6AddressOnCreation) {
			return fmt.Errorf("bad AssignIpv6AddressOnCreation: %t", aws.BoolValue(subnet.AssignIpv6AddressOnCreation))
		}

		return nil
	}
}

func testAccCheckAwsSubnetNotRecreated(t *testing.T, before, after *ec2.Subnet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(before.SubnetId) != aws.StringValue(after.SubnetId) {
			t.Fatalf("Expected SubnetIDs not to change, but both got before: %s and after: %s",
				aws.StringValue(before.SubnetId), aws.StringValue(after.SubnetId))
		}
		return nil
	}
}

func testAccCheckSubnetDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_subnet" {
			continue
		}

		// Try to find the resource
		resp, err := conn.DescribeSubnets(&ec2.DescribeSubnetsInput{
			SubnetIds: []*string{aws.String(rs.Primary.ID)},
		})
		if err == nil {
			if len(resp.Subnets) > 0 {
				return fmt.Errorf("still exist.")
			}

			return nil
		}

		// Verify the error is what we want
		ec2err, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if ec2err.Code() != "InvalidSubnetID.NotFound" {
			return err
		}
	}

	return nil
}

func testAccCheckSubnetExists(n string, v *ec2.Subnet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		resp, err := conn.DescribeSubnets(&ec2.DescribeSubnetsInput{
			SubnetIds: []*string{aws.String(rs.Primary.ID)},
		})
		if err != nil {
			return err
		}
		if len(resp.Subnets) == 0 {
			return fmt.Errorf("Subnet not found")
		}

		*v = *resp.Subnets[0]

		return nil
	}
}

func testAccCheckSubnetUpdateTags(subnet *ec2.Subnet, oldTags, newTags map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		return keyvaluetags.Ec2UpdateTags(conn, aws.StringValue(subnet.SubnetId), oldTags, newTags)
	}
}

const testAccSubnetConfig = `
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-subnet"
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.1.1.0/24"
  vpc_id     = aws_vpc.test.id
}
`

func testAccSubnetTagsConfig1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.1.1.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccSubnetTagsConfig2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.1.1.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

const testAccSubnetComputedTagsBaseConfig = `
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
  tags       = local.tags
}

data "aws_vpc" "test" {
  id = aws_vpc.test.id
}

resource "aws_subnet" "test" {
  cidr_block = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  vpc_id     = aws_vpc.test.id
  tags       = data.aws_vpc.test.tags
}
`

func testAccSubnetConfig_tagsComputedFromVpcDataSource1(tagKey1, tagValue1 string) string {
	return composeConfig(
		testAccSubnetComputedTagsBaseConfig,
		fmt.Sprintf(`
locals {
  tags = {
    %q = %q
  }
}
`, tagKey1, tagValue1))
}

func testAccSubnetConfig_tagsComputedFromVpcDataSource2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return composeConfig(
		testAccSubnetComputedTagsBaseConfig,
		fmt.Sprintf(`
locals {
  tags = {
    %q = %q
    %q = %q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

const testAccSubnetConfigPreIpv6 = `
resource "aws_vpc" "test" {
  cidr_block                       = "10.10.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = "terraform-testacc-subnet-ipv6"
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.10.1.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = "tf-acc-subnet-ipv6"
  }
}
`

const testAccSubnetConfigIpv6 = `
resource "aws_vpc" "test" {
  cidr_block                       = "10.10.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = "terraform-testacc-subnet-ipv6"
  }
}

resource "aws_subnet" "test" {
  cidr_block                      = "10.10.1.0/24"
  vpc_id                          = aws_vpc.test.id
  ipv6_cidr_block                 = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, 1)
  assign_ipv6_address_on_creation = true

  tags = {
    Name = "tf-acc-subnet-ipv6"
  }
}
`

const testAccSubnetConfigIpv6UpdateAssignIpv6OnCreation = `
resource "aws_vpc" "test" {
  cidr_block                       = "10.10.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = "terraform-testacc-subnet-assign-ipv6-on-creation"
  }
}

resource "aws_subnet" "test" {
  cidr_block                      = "10.10.1.0/24"
  vpc_id                          = aws_vpc.test.id
  ipv6_cidr_block                 = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, 1)
  assign_ipv6_address_on_creation = false

  tags = {
    Name = "tf-acc-subnet-assign-ipv6-on-creation"
  }
}
`

const testAccSubnetConfigIpv6UpdateIpv6Cidr = `
resource "aws_vpc" "test" {
  cidr_block                       = "10.10.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = "terraform-testacc-subnet-ipv6-update-cidr"
  }
}

resource "aws_subnet" "test" {
  cidr_block                      = "10.10.1.0/24"
  vpc_id                          = aws_vpc.test.id
  ipv6_cidr_block                 = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, 3)
  assign_ipv6_address_on_creation = false

  tags = {
    Name = "tf-acc-subnet-ipv6-update-cidr"
  }
}
`

func testAccSubnetConfigAvailabilityZoneId() string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), `
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-subnet"
  }
}

resource "aws_subnet" "test" {
  cidr_block           = "10.1.1.0/24"
  vpc_id               = aws_vpc.test.id
  availability_zone_id = data.aws_availability_zones.available.zone_ids[0]

  tags = {
    Name = "tf-acc-subnet"
  }
}
`)
}

func testAccSubnetConfigCustomerOwnedIpv4Pool() string {
	return `
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

data "aws_ec2_local_gateway_route_tables" "test" {
  filter {
    name   = "outpost-arn"
    values = [data.aws_outposts_outpost.test.arn]
  }
}

data "aws_ec2_coip_pools" "test" {
  # Filtering by Local Gateway Route Table ID is documented but not working in EC2 API.
  # If there are multiple Outposts in the test account, this lookup can
  # be misaligned and cause downstream resource errors.
  #
  # filter {
  #   name   = "coip-pool.local-gateway-route-table-id"
  #   values = [tolist(data.aws_ec2_local_gateway_route_tables.test.ids)[0]]
  # }
}

data "aws_ec2_coip_pool" "test" {
  pool_id = tolist(data.aws_ec2_coip_pools.test.pool_ids)[0]
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-subnet-outpost"
  }
}

resource "aws_subnet" "test" {
  availability_zone               = data.aws_outposts_outpost.test.availability_zone
  cidr_block                      = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  customer_owned_ipv4_pool        = data.aws_ec2_coip_pool.test.pool_id
  map_customer_owned_ip_on_launch = true
  outpost_arn                     = data.aws_outposts_outpost.test.arn
  vpc_id                          = aws_vpc.test.id

  tags = {
    Name = "tf-acc-subnet-outpost"
  }
}
`
}

func testAccSubnetConfigMapCustomerOwnedIpOnLaunch(mapCustomerOwnedIpOnLaunch bool) string {
	return fmt.Sprintf(`
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

data "aws_ec2_local_gateway_route_tables" "test" {
  filter {
    name   = "outpost-arn"
    values = [data.aws_outposts_outpost.test.arn]
  }
}

data "aws_ec2_coip_pools" "test" {
  # Filtering by Local Gateway Route Table ID is documented but not working in EC2 API.
  # If there are multiple Outposts in the test account, this lookup can
  # be misaligned and cause downstream resource errors.
  #
  # filter {
  #   name   = "coip-pool.local-gateway-route-table-id"
  #   values = [tolist(data.aws_ec2_local_gateway_route_tables.test.ids)[0]]
  # }
}

data "aws_ec2_coip_pool" "test" {
  pool_id = tolist(data.aws_ec2_coip_pools.test.pool_ids)[0]
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-subnet-outpost"
  }
}

resource "aws_subnet" "test" {
  availability_zone               = data.aws_outposts_outpost.test.availability_zone
  cidr_block                      = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  customer_owned_ipv4_pool        = data.aws_ec2_coip_pool.test.pool_id
  map_customer_owned_ip_on_launch = %[1]t
  outpost_arn                     = data.aws_outposts_outpost.test.arn
  vpc_id                          = aws_vpc.test.id

  tags = {
    Name = "tf-acc-subnet-outpost"
  }
}
`, mapCustomerOwnedIpOnLaunch)
}

func testAccSubnetConfigMapPublicIpOnLaunch(mapPublicIpOnLaunch bool) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-subnet-map-public-ip-on-launch"
  }
}

resource "aws_subnet" "test" {
  cidr_block              = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  map_public_ip_on_launch = %[1]t
  vpc_id                  = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-subnet-map-public-ip-on-launch"
  }
}
`, mapPublicIpOnLaunch)
}

func testAccSubnetConfigOutpost() string {
	return `
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-subnet-outpost"
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_outposts_outpost.test.availability_zone
  cidr_block        = "10.1.1.0/24"
  outpost_arn       = data.aws_outposts_outpost.test.arn
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-subnet-outpost"
  }
}
`
}
