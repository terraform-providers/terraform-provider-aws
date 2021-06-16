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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

// add sweeper to delete known test vpcs
func init() {
	resource.AddTestSweepers("aws_vpc", &resource.Sweeper{
		Name: "aws_vpc",
		Dependencies: []string{
			"aws_ec2_carrier_gateway",
			"aws_egress_only_internet_gateway",
			"aws_internet_gateway",
			"aws_nat_gateway",
			"aws_network_acl",
			"aws_route_table",
			"aws_security_group",
			"aws_subnet",
			"aws_vpc_peering_connection",
			"aws_vpn_gateway",
		},
		F: testSweepVPCs,
	})
}

func testSweepVPCs(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*AWSClient).ec2conn
	sweepResources := make([]*testSweepResource, 0)
	var errs *multierror.Error

	input := &ec2.DescribeVpcsInput{}

	err = conn.DescribeVpcsPages(input, func(page *ec2.DescribeVpcsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, vpc := range page.Vpcs {
			if vpc == nil {
				continue
			}

			id := aws.StringValue(vpc.VpcId)

			if aws.BoolValue(vpc.IsDefault) {
				log.Printf("[DEBUG] Skipping default EC2 VPC: %s", id)
				continue
			}

			r := resourceAwsVpc()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, NewTestSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error describing EC2 VPCs for %s: %w", region, err))
	}

	if err = testSweepResourceOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping EC2 VPCs for %s: %w", region, err))
	}

	if testSweepSkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping EC2 VPCs sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func TestAccAWSVpc_basic(t *testing.T) {
	var vpc ec2.Vpc
	resourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists(resourceName, &vpc),
					testAccCheckVpcCidr(&vpc, "10.1.0.0/16"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`vpc/vpc-.+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "assign_generated_ipv6_cidr_block", "false"),
					resource.TestMatchResourceAttr(resourceName, "default_route_table_id", regexp.MustCompile(`^rtb-.+`)),
					resource.TestCheckResourceAttr(resourceName, "cidr_block", "10.1.0.0/16"),
					resource.TestCheckResourceAttr(resourceName, "enable_dns_support", "true"),
					resource.TestCheckResourceAttr(resourceName, "instance_tenancy", "default"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_association_id", ""),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_block", ""),
					resource.TestMatchResourceAttr(resourceName, "main_route_table_id", regexp.MustCompile(`^rtb-.+`)),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
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

func TestAccAWSVpc_disappears(t *testing.T) {
	var vpc ec2.Vpc
	resourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists(resourceName, &vpc),
					testAccCheckVpcDisappears(&vpc),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSVpc_defaultTags_providerOnly(t *testing.T) {
	var providers []*schema.Provider
	var vpc ec2.Vpc
	resourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: testAccProviderFactoriesInternal(&providers),
		CheckDestroy:      testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: composeConfig(
					testAccAWSProviderConfigDefaultTags_Tags1("providerkey1", "providervalue1"),
					testAccVpcConfig,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists(resourceName, &vpc),
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
					testAccVpcConfig,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey1", "providervalue1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey2", "providervalue2"),
				),
			},
			{
				Config: composeConfig(
					testAccAWSProviderConfigDefaultTags_Tags1("providerkey1", "value1"),
					testAccVpcConfig,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey1", "value1"),
				),
			},
		},
	})
}

func TestAccAWSVpc_defaultTags_updateToProviderOnly(t *testing.T) {
	var providers []*schema.Provider
	var vpc ec2.Vpc
	resourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: testAccProviderFactoriesInternal(&providers),
		CheckDestroy:      testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSVPCConfigTags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", "value1"),
				),
			},
			{
				Config: composeConfig(
					testAccAWSProviderConfigDefaultTags_Tags1("key1", "value1"),
					testAccVpcConfig,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists(resourceName, &vpc),
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

func TestAccAWSVpc_defaultTags_updateToResourceOnly(t *testing.T) {
	var providers []*schema.Provider
	var vpc ec2.Vpc
	resourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: testAccProviderFactoriesInternal(&providers),
		CheckDestroy:      testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: composeConfig(
					testAccAWSProviderConfigDefaultTags_Tags1("key1", "value1"),
					testAccVpcConfig,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", "value1"),
				),
			},
			{
				Config: testAccAWSVPCConfigTags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists(resourceName, &vpc),
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

func TestAccAWSVpc_defaultTags_providerAndResource_nonOverlappingTag(t *testing.T) {
	var providers []*schema.Provider
	var vpc ec2.Vpc
	resourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: testAccProviderFactoriesInternal(&providers),
		CheckDestroy:      testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: composeConfig(
					testAccAWSProviderConfigDefaultTags_Tags1("providerkey1", "providervalue1"),
					testAccAWSVPCConfigTags1("resourcekey1", "resourcevalue1"),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists(resourceName, &vpc),
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
					testAccAWSVPCConfigTags2("resourcekey1", "resourcevalue1", "resourcekey2", "resourcevalue2"),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists(resourceName, &vpc),
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
					testAccAWSVPCConfigTags1("resourcekey3", "resourcevalue3"),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists(resourceName, &vpc),
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

func TestAccAWSVpc_defaultTags_providerAndResource_overlappingTag(t *testing.T) {
	var providers []*schema.Provider
	var vpc ec2.Vpc
	resourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: testAccProviderFactoriesInternal(&providers),
		CheckDestroy:      testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: composeConfig(
					testAccAWSProviderConfigDefaultTags_Tags1("overlapkey1", "providervalue1"),
					testAccAWSVPCConfigTags1("overlapkey1", "resourcevalue1"),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists(resourceName, &vpc),
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
					testAccAWSVPCConfigTags2("overlapkey1", "resourcevalue1", "overlapkey2", "resourcevalue2"),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists(resourceName, &vpc),
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
					testAccAWSVPCConfigTags1("overlapkey1", "resourcevalue2"),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.overlapkey1", "resourcevalue2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.overlapkey1", "resourcevalue2"),
				),
			},
		},
	})
}

func TestAccAWSVpc_defaultTags_providerAndResource_duplicateTag(t *testing.T) {
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: testAccProviderFactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: composeConfig(
					testAccAWSProviderConfigDefaultTags_Tags1("overlapkey", "overlapvalue"),
					testAccAWSVPCConfigTags1("overlapkey", "overlapvalue"),
				),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`"tags" are identical to those in the "default_tags" configuration block`),
			},
		},
	})
}

// TestAccAWSVpc_DynamicResourceTagsMergedWithLocals_IgnoreChanges ensures computed "tags_all"
// attributes are correctly determined when the provider-level default_tags block
// is left unused and resource tags (merged with local.tags) are only known at apply time,
// with additional lifecycle ignore_changes attributes, thereby eliminating "Inconsistent final plan" errors
// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/18366
func TestAccAWSVpc_DynamicResourceTagsMergedWithLocals_IgnoreChanges(t *testing.T) {
	var providers []*schema.Provider
	var vpc ec2.Vpc

	resourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: testAccProviderFactoriesInternal(&providers),
		CheckDestroy:      testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSVPCConfigWithIgnoreChanges_DynamicTagsMergedWithLocals("localkey", "localvalue"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.localkey", "localvalue"),
					resource.TestCheckResourceAttrSet(resourceName, "tags.created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "tags.updated_at"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.localkey", "localvalue"),
					resource.TestCheckResourceAttrSet(resourceName, "tags_all.created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "tags_all.updated_at"),
				),
				// Dynamic tag "updated_at" will cause a perpetual diff but that's OK for this test
				// as we want to ensure subsequent applies will not result in "inconsistent final plan errors"
				// for the attribute "tags_all"
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccAWSVPCConfigWithIgnoreChanges_DynamicTagsMergedWithLocals("localkey", "localvalue"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.localkey", "localvalue"),
					resource.TestCheckResourceAttrSet(resourceName, "tags.created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "tags.updated_at"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.localkey", "localvalue"),
					resource.TestCheckResourceAttrSet(resourceName, "tags_all.created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "tags_all.updated_at"),
				),
				// Dynamic tag "updated_at" will cause a perpetual diff but that's OK for this test
				// as we wanted to ensure this configuration applies successfully
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestAccAWSVpc_DynamicResourceTags_IgnoreChanges ensures computed "tags_all"
// attributes are correctly determined when the provider-level default_tags block
// is left unused and resource tags are only known at apply time,
// with additional lifecycle ignore_changes attributes, thereby eliminating "Inconsistent final plan" errors
// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/18366
func TestAccAWSVpc_DynamicResourceTags_IgnoreChanges(t *testing.T) {
	var providers []*schema.Provider
	var vpc ec2.Vpc

	resourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: testAccProviderFactoriesInternal(&providers),
		CheckDestroy:      testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSVPCConfigWithIgnoreChanges_DynamicTags,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttrSet(resourceName, "tags.created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "tags.updated_at"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "2"),
					resource.TestCheckResourceAttrSet(resourceName, "tags_all.created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "tags_all.updated_at"),
				),
				// Dynamic tag "updated_at" will cause a perpetual diff but that's OK for this test
				// as we want to ensure subsequent applies will not result in "inconsistent final plan errors"
				// for the attribute "tags_all"
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccAWSVPCConfigWithIgnoreChanges_DynamicTags,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttrSet(resourceName, "tags.created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "tags.updated_at"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "2"),
					resource.TestCheckResourceAttrSet(resourceName, "tags_all.created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "tags_all.updated_at"),
				),
				// Dynamic tag "updated_at" will cause a perpetual diff but that's OK for this test
				// as we wanted to ensure this configuration applies successfully
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSVpc_defaultAndIgnoreTags(t *testing.T) {
	var providers []*schema.Provider
	var vpc ec2.Vpc
	resourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: testAccProviderFactoriesInternal(&providers),
		CheckDestroy:      testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSVPCConfigTags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists(resourceName, &vpc),
					testAccCheckVpcUpdateTags(&vpc, nil, map[string]string{"defaultkey1": "defaultvalue1"}),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: composeConfig(
					testAccProviderConfigDefaultAndIgnoreTagsKeyPrefixes1("defaultkey1", "defaultvalue1", "defaultkey"),
					testAccAWSVPCConfigTags1("key1", "value1"),
				),
				PlanOnly: true,
			},
			{
				Config: composeConfig(
					testAccProviderConfigDefaultAndIgnoreTagsKeys1("defaultkey1", "defaultvalue1"),
					testAccAWSVPCConfigTags1("key1", "value1"),
				),
				PlanOnly: true,
			},
		},
	})
}

func TestAccAWSVpc_ignoreTags(t *testing.T) {
	var providers []*schema.Provider
	var vpc ec2.Vpc
	resourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: testAccProviderFactoriesInternal(&providers),
		CheckDestroy:      testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSVPCConfigTags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists(resourceName, &vpc),
					testAccCheckVpcUpdateTags(&vpc, nil, map[string]string{"ignorekey1": "ignorevalue1"}),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config:   testAccProviderConfigIgnoreTagsKeyPrefixes1("ignorekey") + testAccAWSVPCConfigTags1("key1", "value1"),
				PlanOnly: true,
			},
			{
				Config:   testAccProviderConfigIgnoreTagsKeys1("ignorekey1") + testAccAWSVPCConfigTags1("key1", "value1"),
				PlanOnly: true,
			},
		},
	})
}

func TestAccAWSVpc_AssignGeneratedIpv6CidrBlock(t *testing.T) {
	var vpc ec2.Vpc
	resourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcConfigAssignGeneratedIpv6CidrBlock(true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVpcExists(resourceName, &vpc),
					testAccCheckVpcCidr(&vpc, "10.1.0.0/16"),
					resource.TestCheckResourceAttr(resourceName, "assign_generated_ipv6_cidr_block", "true"),
					resource.TestCheckResourceAttr(resourceName, "cidr_block", "10.1.0.0/16"),
					resource.TestMatchResourceAttr(resourceName, "ipv6_association_id", regexp.MustCompile(`^vpc-cidr-assoc-.+`)),
					resource.TestMatchResourceAttr(resourceName, "ipv6_cidr_block", regexp.MustCompile(`/56$`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVpcConfigAssignGeneratedIpv6CidrBlock(false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVpcExists(resourceName, &vpc),
					testAccCheckVpcCidr(&vpc, "10.1.0.0/16"),
					resource.TestCheckResourceAttr(resourceName, "assign_generated_ipv6_cidr_block", "false"),
					resource.TestCheckResourceAttr(resourceName, "cidr_block", "10.1.0.0/16"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_association_id", ""),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_block", ""),
				),
			},
			{
				Config: testAccVpcConfigAssignGeneratedIpv6CidrBlock(true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVpcExists(resourceName, &vpc),
					testAccCheckVpcCidr(&vpc, "10.1.0.0/16"),
					resource.TestCheckResourceAttr(resourceName, "assign_generated_ipv6_cidr_block", "true"),
					resource.TestCheckResourceAttr(resourceName, "cidr_block", "10.1.0.0/16"),
					resource.TestMatchResourceAttr(resourceName, "ipv6_association_id", regexp.MustCompile(`^vpc-cidr-assoc-.+`)),
					resource.TestMatchResourceAttr(resourceName, "ipv6_cidr_block", regexp.MustCompile(`/56$`)),
				),
			},
		},
	})
}

func TestAccAWSVpc_Tenancy(t *testing.T) {
	var vpcDedicated ec2.Vpc
	var vpcDefault ec2.Vpc
	resourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcDedicatedConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists(resourceName, &vpcDedicated),
					resource.TestCheckResourceAttr(resourceName, "instance_tenancy", "dedicated"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVpcConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists(resourceName, &vpcDefault),
					resource.TestCheckResourceAttr(resourceName, "instance_tenancy", "default"),
					testAccCheckVpcIdsEqual(&vpcDedicated, &vpcDefault),
				),
			},
			{
				Config: testAccVpcDedicatedConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists(resourceName, &vpcDedicated),
					resource.TestCheckResourceAttr(resourceName, "instance_tenancy", "dedicated"),
					testAccCheckVpcIdsNotEqual(&vpcDedicated, &vpcDefault),
				),
			},
		},
	})
}

func TestAccAWSVpc_tags(t *testing.T) {
	var vpc ec2.Vpc
	resourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSVPCConfigTags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists(resourceName, &vpc),
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
				Config: testAccAWSVPCConfigTags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSVPCConfigTags1("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSVpc_update(t *testing.T) {
	var vpc ec2.Vpc
	resourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists(resourceName, &vpc),
					testAccCheckVpcCidr(&vpc, "10.1.0.0/16"),
					resource.TestCheckResourceAttr(resourceName, "cidr_block", "10.1.0.0/16"),
				),
			},
			{
				Config: testAccVpcConfigUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "enable_dns_hostnames", "true"),
				),
			},
		},
	})
}

func testAccCheckVpcDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpc" {
			continue
		}

		// Try to find the VPC
		DescribeVpcOpts := &ec2.DescribeVpcsInput{
			VpcIds: []*string{aws.String(rs.Primary.ID)},
		}
		resp, err := conn.DescribeVpcs(DescribeVpcOpts)
		if err == nil {
			if len(resp.Vpcs) > 0 {
				return fmt.Errorf("VPCs still exist.")
			}

			return nil
		}

		// Verify the error is what we want
		ec2err, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if ec2err.Code() != "InvalidVpcID.NotFound" {
			return err
		}
	}

	return nil
}

func testAccCheckVpcUpdateTags(vpc *ec2.Vpc, oldTags, newTags map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		return keyvaluetags.Ec2UpdateTags(conn, aws.StringValue(vpc.VpcId), oldTags, newTags)
	}
}

func testAccCheckVpcCidr(vpc *ec2.Vpc, expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(vpc.CidrBlock) != expected {
			return fmt.Errorf("Bad cidr: %s", aws.StringValue(vpc.CidrBlock))
		}

		return nil
	}
}

func testAccCheckVpcIdsEqual(vpc1, vpc2 *ec2.Vpc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(vpc1.VpcId) != aws.StringValue(vpc2.VpcId) {
			return fmt.Errorf("VPC IDs not equal")
		}

		return nil
	}
}

func testAccCheckVpcIdsNotEqual(vpc1, vpc2 *ec2.Vpc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(vpc1.VpcId) == aws.StringValue(vpc2.VpcId) {
			return fmt.Errorf("VPC IDs are equal")
		}

		return nil
	}
}

func testAccCheckVpcExists(n string, vpc *ec2.Vpc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No VPC ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		DescribeVpcOpts := &ec2.DescribeVpcsInput{
			VpcIds: []*string{aws.String(rs.Primary.ID)},
		}
		resp, err := conn.DescribeVpcs(DescribeVpcOpts)
		if err != nil {
			return err
		}
		if len(resp.Vpcs) == 0 || resp.Vpcs[0] == nil {
			return fmt.Errorf("VPC not found")
		}

		*vpc = *resp.Vpcs[0]

		return nil
	}
}

func testAccCheckVpcDisappears(vpc *ec2.Vpc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		input := &ec2.DeleteVpcInput{
			VpcId: vpc.VpcId,
		}

		_, err := conn.DeleteVpc(input)

		return err
	}
}

// https://github.com/hashicorp/terraform/issues/1301
func TestAccAWSVpc_bothDnsOptionsSet(t *testing.T) {
	var vpc ec2.Vpc
	resourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcConfig_BothDnsOptions,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "enable_dns_hostnames", "true"),
					resource.TestCheckResourceAttr(resourceName, "enable_dns_support", "true"),
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

// https://github.com/hashicorp/terraform/issues/10168
func TestAccAWSVpc_DisabledDnsSupport(t *testing.T) {
	var vpc ec2.Vpc
	resourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcConfig_DisabledDnsSupport,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "enable_dns_support", "false"),
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

func TestAccAWSVpc_classiclinkOptionSet(t *testing.T) {
	var vpc ec2.Vpc
	resourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcConfig_ClassiclinkOption,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "enable_classiclink", "true"),
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

func TestAccAWSVpc_classiclinkDnsSupportOptionSet(t *testing.T) {
	var vpc ec2.Vpc
	resourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcConfig_ClassiclinkDnsSupportOption,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "enable_classiclink_dns_support", "true"),
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

const testAccVpcConfig = `
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"
}
`

func testAccVpcConfigAssignGeneratedIpv6CidrBlock(assignGeneratedIpv6CidrBlock bool) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  assign_generated_ipv6_cidr_block = %t
  cidr_block                       = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-vpc-ipv6"
  }
}
`, assignGeneratedIpv6CidrBlock)
}

const testAccVpcConfigUpdate = `
resource "aws_vpc" "test" {
  cidr_block           = "10.1.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-vpc"
  }
}
`

func testAccAWSVPCConfigTags1(tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1)
}

func testAccAWSVPCConfigTags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAWSVPCConfigWithIgnoreChanges_DynamicTagsMergedWithLocals(localTagKey1, localTagValue1 string) string {
	return fmt.Sprintf(`
locals {
  tags = {
    %[1]q = %[2]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = merge(local.tags, {
    "created_at" = timestamp()
    "updated_at" = timestamp()
  })

  lifecycle {
    ignore_changes = [
      tags["created_at"],
    ]
  }
}
`, localTagKey1, localTagValue1)
}

const testAccAWSVPCConfigWithIgnoreChanges_DynamicTags = `
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    "created_at" = timestamp()
    "updated_at" = timestamp()
  }

  lifecycle {
    ignore_changes = [
      tags["created_at"],
    ]
  }
}
`

const testAccVpcDedicatedConfig = `
resource "aws_vpc" "test" {
  instance_tenancy = "dedicated"
  cidr_block       = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-vpc-dedicated"
  }
}
`

const testAccVpcConfig_BothDnsOptions = `
resource "aws_vpc" "test" {
  cidr_block           = "10.2.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = "terraform-testacc-vpc-both-dns-opts"
  }
}
`

const testAccVpcConfig_DisabledDnsSupport = `
resource "aws_vpc" "test" {
  cidr_block         = "10.2.0.0/16"
  enable_dns_support = false

  tags = {
    Name = "terraform-testacc-vpc-disabled-dns-support"
  }
}
`

const testAccVpcConfig_ClassiclinkOption = `
resource "aws_vpc" "test" {
  cidr_block         = "172.2.0.0/16"
  enable_classiclink = true

  tags = {
    Name = "terraform-testacc-vpc-classic-link"
  }
}
`

const testAccVpcConfig_ClassiclinkDnsSupportOption = `
resource "aws_vpc" "test" {
  cidr_block                     = "172.2.0.0/16"
  enable_classiclink             = true
  enable_classiclink_dns_support = true

  tags = {
    Name = "terraform-testacc-vpc-classic-link-support"
  }
}
`
