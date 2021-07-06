package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/aws/aws-sdk-go/service/transfer"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/transfer/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func init() {
	RegisterServiceErrorCheckFunc(transfer.EndpointsID, testAccErrorCheckSkipTransfer)

	resource.AddTestSweepers("aws_transfer_server", &resource.Sweeper{
		Name: "aws_transfer_server",
		F:    testSweepTransferServers,
	})
}

func testSweepTransferServers(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).transferconn
	input := &transfer.ListServersInput{}
	sweepResources := make([]*testSweepResource, 0)

	err = conn.ListServersPages(input, func(page *transfer.ListServersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, server := range page.Servers {
			r := resourceAwsTransferServer()
			d := r.Data(nil)
			d.SetId(aws.StringValue(server.ServerId))
			d.Set("force_destroy", true) // In lieu of an aws_transfer_user sweeper.
			d.Set("identity_provider_type", server.IdentityProviderType)

			sweepResources = append(sweepResources, NewTestSweepResource(r, d, client))
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping Transfer Server sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Transfer Servers (%s): %w", region, err)
	}

	err = testSweepResourceOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Transfer Servers (%s): %w", region, err)
	}

	return nil
}

func testAccErrorCheckSkipTransfer(t *testing.T) resource.ErrorCheckFunc {
	return testAccErrorCheckSkipMessagesContaining(t,
		"Invalid server type: PUBLIC",
	)
}

func testAccAWSTransferServer_basic(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	iamRoleResourceName := "aws_iam_role.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSTransfer(t) },
		ErrorCheck:   testAccErrorCheck(t, transfer.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTransferServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferServerBasicConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "transfer", regexp.MustCompile(`server/.+`)),
					resource.TestCheckResourceAttr(resourceName, "certificate", ""),
					testAccMatchResourceAttrRegionalHostname(resourceName, "endpoint", "server.transfer", regexp.MustCompile(`s-[a-z0-9]+`)),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "PUBLIC"),
					resource.TestCheckResourceAttr(resourceName, "force_destroy", "false"),
					resource.TestCheckNoResourceAttr(resourceName, "host_key"),
					resource.TestCheckResourceAttrSet(resourceName, "host_key_fingerprint"),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_type", "SERVICE_MANAGED"),
					resource.TestCheckResourceAttr(resourceName, "invocation_role", ""),
					resource.TestCheckResourceAttr(resourceName, "logging_role", ""),
					resource.TestCheckResourceAttr(resourceName, "protocols.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "protocols.*", "SFTP"),
					resource.TestCheckResourceAttr(resourceName, "security_policy_name", "TransferSecurityPolicy-2018-11"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "url", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			{
				Config: testAccAWSTransferServerUpdatedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "transfer", regexp.MustCompile(`server/.+`)),
					resource.TestCheckResourceAttr(resourceName, "certificate", ""),
					resource.TestCheckResourceAttr(resourceName, "domain", "S3"),
					testAccMatchResourceAttrRegionalHostname(resourceName, "endpoint", "server.transfer", regexp.MustCompile(`s-[a-z0-9]+`)),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "PUBLIC"),
					resource.TestCheckResourceAttr(resourceName, "force_destroy", "false"),
					resource.TestCheckNoResourceAttr(resourceName, "host_key"),
					resource.TestCheckResourceAttrSet(resourceName, "host_key_fingerprint"),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_type", "SERVICE_MANAGED"),
					resource.TestCheckResourceAttr(resourceName, "invocation_role", ""),
					resource.TestCheckResourceAttrPair(resourceName, "logging_role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "protocols.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "protocols.*", "SFTP"),
					resource.TestCheckResourceAttr(resourceName, "security_policy_name", "TransferSecurityPolicy-2018-11"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "url", ""),
				),
			},
		},
	})
}

func testAccAWSTransferServer_domain(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSTransfer(t) },
		ErrorCheck:   testAccErrorCheck(t, transfer.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTransferServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferServerDomainConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "domain", "EFS"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func testAccAWSTransferServer_disappears(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSTransfer(t) },
		ErrorCheck:   testAccErrorCheck(t, transfer.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTransferServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferServerBasicConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsTransferServer(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAWSTransferServer_securityPolicy(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSTransfer(t) },
		ErrorCheck:   testAccErrorCheck(t, transfer.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTransferServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferServerSecurityPolicyConfig("TransferSecurityPolicy-2020-06"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "security_policy_name", "TransferSecurityPolicy-2020-06"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			{
				Config: testAccAWSTransferServerSecurityPolicyConfig("TransferSecurityPolicy-2018-11"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "security_policy_name", "TransferSecurityPolicy-2018-11"),
				),
			},
		},
	})
}

func testAccAWSTransferServer_vpc(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	defaultSecurityGroupResourceName := "aws_default_security_group.test"
	subnetResourceName := "aws_subnet.test"
	vpcResourceName := "aws_vpc.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSTransfer(t) },
		ErrorCheck:   testAccErrorCheck(t, transfer.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTransferServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferServerVpcConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.address_allocation_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.security_group_ids.*", defaultSecurityGroupResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.subnet_ids.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_details.0.vpc_endpoint_id"),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint_details.0.vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			{
				Config: testAccAWSTransferServerVpcUpdateConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.address_allocation_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.security_group_ids.*", defaultSecurityGroupResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.subnet_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.subnet_ids.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_details.0.vpc_endpoint_id"),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint_details.0.vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
				),
			},
			{
				Config: testAccAWSTransferServerVpcConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.address_allocation_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.security_group_ids.*", defaultSecurityGroupResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.subnet_ids.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_details.0.vpc_endpoint_id"),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint_details.0.vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
				),
			},
		},
	})
}

func testAccAWSTransferServer_vpcAddressAllocationIds(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	eip1ResourceName := "aws_eip.test.0"
	eip2ResourceName := "aws_eip.test.1"
	defaultSecurityGroupResourceName := "aws_default_security_group.test"
	subnetResourceName := "aws_subnet.test"
	vpcResourceName := "aws_vpc.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSTransfer(t) },
		ErrorCheck:   testAccErrorCheck(t, transfer.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTransferServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferServerVpcAddressAllocationIdsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.address_allocation_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.address_allocation_ids.*", eip1ResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.security_group_ids.*", defaultSecurityGroupResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.subnet_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.subnet_ids.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_details.0.vpc_endpoint_id"),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint_details.0.vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			{
				Config: testAccAWSTransferServerVpcAddressAllocationIdsUpdateConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.address_allocation_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.address_allocation_ids.*", eip2ResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.security_group_ids.*", defaultSecurityGroupResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.subnet_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.subnet_ids.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_details.0.vpc_endpoint_id"),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint_details.0.vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
				),
			},
			{
				Config: testAccAWSTransferServerVpcConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.address_allocation_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.security_group_ids.*", defaultSecurityGroupResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.subnet_ids.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_details.0.vpc_endpoint_id"),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint_details.0.vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
				),
			},
		},
	})
}

func testAccAWSTransferServer_vpcSecurityGroupIds(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	securityGroup1ResourceName := "aws_security_group.test"
	securityGroup2ResourceName := "aws_security_group.test2"
	vpcResourceName := "aws_vpc.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSTransfer(t) },
		ErrorCheck:   testAccErrorCheck(t, transfer.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTransferServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferServerVpcSecurityGroupIdsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.address_allocation_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.security_group_ids.*", securityGroup1ResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.subnet_ids.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_details.0.vpc_endpoint_id"),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint_details.0.vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			{
				Config: testAccAWSTransferServerVpcSecurityGroupIdsUpdateConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.address_allocation_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.security_group_ids.*", securityGroup2ResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.subnet_ids.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_details.0.vpc_endpoint_id"),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint_details.0.vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
				),
			},
		},
	})
}

func testAccAWSTransferServer_vpcAddressAllocationIds_securityGroupIds(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	eip1ResourceName := "aws_eip.test.0"
	eip2ResourceName := "aws_eip.test.1"
	securityGroup1ResourceName := "aws_security_group.test"
	securityGroup2ResourceName := "aws_security_group.test2"
	subnetResourceName := "aws_subnet.test"
	vpcResourceName := "aws_vpc.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSTransfer(t) },
		ErrorCheck:   testAccErrorCheck(t, transfer.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTransferServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferServerVpcAddressAllocationIdsSecurityGroupIdsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.address_allocation_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.address_allocation_ids.*", eip1ResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.security_group_ids.*", securityGroup1ResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.subnet_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.subnet_ids.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_details.0.vpc_endpoint_id"),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint_details.0.vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			{
				Config: testAccAWSTransferServerVpcAddressAllocationIdsSecurityGroupIdsUpdateConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.address_allocation_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.address_allocation_ids.*", eip2ResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.security_group_ids.*", securityGroup2ResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.subnet_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_details.0.subnet_ids.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_details.0.vpc_endpoint_id"),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint_details.0.vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
				),
			},
		},
	})
}

func testAccAWSTransferServer_updateEndpointType_publicToVpc(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSTransfer(t) },
		ErrorCheck:   testAccErrorCheck(t, transfer.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTransferServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferServerBasicConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "PUBLIC"),
				),
			},
			{
				Config: testAccAWSTransferServerVpcConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func testAccAWSTransferServer_updateEndpointType_publicToVpc_addressAllocationIds(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSTransfer(t) },
		ErrorCheck:   testAccErrorCheck(t, transfer.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTransferServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferServerBasicConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "PUBLIC"),
				),
			},
			{
				Config: testAccAWSTransferServerVpcAddressAllocationIdsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func testAccAWSTransferServer_updateEndpointType_vpcEndpointToVpc(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSTransfer(t) },
		ErrorCheck:   testAccErrorCheck(t, transfer.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTransferServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferServerVpcEndpointConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC_ENDPOINT"),
				),
			},
			{
				Config: testAccAWSTransferServerVpcConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func testAccAWSTransferServer_updateEndpointType_vpcEndpointToVpc_addressAllocationIds(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSTransfer(t) },
		ErrorCheck:   testAccErrorCheck(t, transfer.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTransferServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferServerVpcEndpointConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC_ENDPOINT"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
				),
			},
			{
				Config: testAccAWSTransferServerVpcAddressAllocationIdsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func testAccAWSTransferServer_updateEndpointType_vpcEndpointToVpc_securityGroupIds(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSTransfer(t) },
		ErrorCheck:   testAccErrorCheck(t, transfer.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTransferServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferServerVpcEndpointConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC_ENDPOINT"),
				),
			},
			{
				Config: testAccAWSTransferServerVpcSecurityGroupIdsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func testAccAWSTransferServer_updateEndpointType_vpcToPublic(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSTransfer(t) },
		ErrorCheck:   testAccErrorCheck(t, transfer.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTransferServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferServerVpcConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
				),
			},
			{
				Config: testAccAWSTransferServerBasicConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "PUBLIC"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func testAccAWSTransferServer_protocols(t *testing.T) {
	var s transfer.DescribedServer
	var ca acmpca.CertificateAuthority
	resourceName := "aws_transfer_server.test"
	acmCAResourceName := "aws_acmpca_certificate_authority.test"
	acmCertificateResourceName := "aws_acm_certificate.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t); testAccPreCheckAWSTransfer(t) },
		ErrorCheck:   testAccErrorCheck(t, transfer.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTransferServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferServerProtocolsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &s),
					resource.TestCheckResourceAttr(resourceName, "certificate", ""),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_type", "API_GATEWAY"),
					resource.TestCheckResourceAttr(resourceName, "protocols.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "protocols.*", "FTP"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			// We need to create and activate the CA before issuing a certificate.
			{
				Config: testAccAWSTransferServerConfigRootCA(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAcmpcaCertificateAuthorityExists(acmCAResourceName, &ca),
					testAccCheckAwsAcmpcaCertificateAuthorityActivateCA(&ca),
				),
			},
			{
				Config: testAccAWSTransferServerProtocolsUpdateConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &s),
					resource.TestCheckResourceAttrPair(resourceName, "certificate", acmCertificateResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC"),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_type", "API_GATEWAY"),
					resource.TestCheckResourceAttr(resourceName, "protocols.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "protocols.*", "FTP"),
					resource.TestCheckTypeSetElemAttr(resourceName, "protocols.*", "FTPS"),
				),
			},
			{
				Config: testAccAWSTransferServerProtocolsUpdateConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					// CA must be DISABLED for deletion.
					testAccCheckAwsAcmpcaCertificateAuthorityDisableCA(&ca),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAWSTransferServer_apiGateway(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t); testAccPreCheckAWSTransfer(t) },
		ErrorCheck:   testAccErrorCheck(t, transfer.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTransferServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferServerApiGatewayIdentityProviderTypeConfig(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_type", "API_GATEWAY"),
					resource.TestCheckResourceAttrPair(resourceName, "invocation_role", "aws_iam_role.test", "arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func testAccAWSTransferServer_apiGateway_forceDestroy(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t); testAccPreCheckAWSTransfer(t) },
		ErrorCheck:   testAccErrorCheck(t, transfer.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTransferServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferServerApiGatewayIdentityProviderTypeConfig(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_type", "API_GATEWAY"),
					resource.TestCheckResourceAttrPair(resourceName, "invocation_role", "aws_iam_role.test", "arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func testAccAWSTransferServer_forceDestroy(t *testing.T) {
	var s transfer.DescribedServer
	var u transfer.DescribedUser
	var k transfer.SshPublicKey
	resourceName := "aws_transfer_server.test"
	userResourceName := "aws_transfer_user.test"
	sshKeyResourceName := "aws_transfer_ssh_key.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSTransfer(t) },
		ErrorCheck:   testAccErrorCheck(t, transfer.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTransferServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferServerForceDestroyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &s),
					testAccCheckAWSTransferUserExists(userResourceName, &u),
					testAccCheckAWSTransferSshKeyExists(sshKeyResourceName, &k),
					resource.TestCheckResourceAttr(resourceName, "force_destroy", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "host_key"},
			},
		},
	})
}

func testAccAWSTransferServer_hostKey(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	hostKey := "test-fixtures/transfer-ssh-rsa-key"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, transfer.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTransferServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferServerHostKeyConfig(hostKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "host_key_fingerprint", "SHA256:Z2pW9sPKDD/T34tVfCoolsRcECNTlekgaKvDn9t+9sg="),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "host_key"},
			},
		},
	})
}

func testAccAWSTransferServer_vpcEndpointId(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"
	vpcEndpointResourceName := "aws_vpc_endpoint.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	if testAccGetPartition() == "aws-us-gov" {
		t.Skip("Transfer Server VPC_ENDPOINT endpoint type is not supported in GovCloud partition")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSTransfer(t) },
		ErrorCheck:   testAccErrorCheck(t, transfer.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTransferServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferServerVpcEndpointConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.address_allocation_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.security_group_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.subnet_ids.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint_details.0.vpc_endpoint_id", vpcEndpointResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_details.0.vpc_id", ""),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "VPC_ENDPOINT"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "host_key"},
			},
		},
	})
}

func testAccCheckAWSTransferServerExists(n string, v *transfer.DescribedServer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Transfer Server ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).transferconn

		output, err := finder.ServerByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckAWSTransferServerDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).transferconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_transfer_server" {
			continue
		}

		_, err := finder.ServerByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Transfer Server %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccPreCheckAWSTransfer(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).transferconn

	input := &transfer.ListServersInput{}

	_, err := conn.ListServers(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAWSTransferServerConfigBaseVpc(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id                  = aws_vpc.test.id
  cidr_block              = "10.0.0.0/24"
  map_public_ip_on_launch = true

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_internet_gateway.test]
}

resource "aws_default_route_table" "test" {
  default_route_table_id = aws_vpc.test.default_route_table_id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_security_group" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_eip" "test" {
  count = 2

  vpc = true

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccAWSTransferServerConfigBaseLoggingRole(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {
      "Service": "transfer.amazonaws.com"
    },
    "Action": "sts:AssumeRole"
  }]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [{
    "Sid": "AllowFullAccesstoCloudWatchLogs",
    "Effect": "Allow",
    "Action": [
      "logs:*"
    ],
    "Resource": "*"
  }]
}
POLICY
}
`, rName)
}

func testAccAWSTransferServerConfigBaseApiGateway(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
}

resource "aws_api_gateway_resource" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  parent_id   = aws_api_gateway_rest_api.test.root_resource_id
  path_part   = "test"
}

resource "aws_api_gateway_method" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  resource_id   = aws_api_gateway_resource.test.id
  http_method   = "GET"
  authorization = "NONE"
}

resource "aws_api_gateway_method_response" "error" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method
  status_code = "400"
}

resource "aws_api_gateway_integration" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method

  type                    = "HTTP"
  uri                     = "https://www.google.de"
  integration_http_method = "GET"
}

resource "aws_api_gateway_integration_response" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_integration.test.http_method
  status_code = aws_api_gateway_method_response.error.status_code
}

resource "aws_api_gateway_deployment" "test" {
  depends_on = [aws_api_gateway_integration.test]

  rest_api_id       = aws_api_gateway_rest_api.test.id
  stage_name        = "test"
  description       = %[1]q
  stage_description = %[1]q

  variables = {
    "a" = "2"
  }
}
`, rName)
}

func testAccAWSTransferServerBasicConfig() string {
	return `
resource "aws_transfer_server" "test" {}
`
}

func testAccAWSTransferServerDomainConfig() string {
	return `
resource "aws_transfer_server" "test" {
  domain = "EFS"
}
`
}

func testAccAWSTransferServerSecurityPolicyConfig(policy string) string {
	return fmt.Sprintf(`
resource "aws_transfer_server" "test" {
  security_policy_name = %[1]q
}
`, policy)
}

func testAccAWSTransferServerUpdatedConfig(rName string) string {
	return composeConfig(
		testAccAWSTransferServerConfigBaseLoggingRole(rName),
		`
resource "aws_transfer_server" "test" {
  identity_provider_type = "SERVICE_MANAGED"
  logging_role           = aws_iam_role.test.arn
}
`)
}

func testAccAWSTransferServerApiGatewayIdentityProviderTypeConfig(rName string, forceDestroy bool) string {
	return composeConfig(
		testAccAWSTransferServerConfigBaseApiGateway(rName),
		testAccAWSTransferServerConfigBaseLoggingRole(rName),
		fmt.Sprintf(`
resource "aws_transfer_server" "test" {
  identity_provider_type = "API_GATEWAY"
  url                    = "${aws_api_gateway_deployment.test.invoke_url}${aws_api_gateway_resource.test.path}"
  invocation_role        = aws_iam_role.test.arn
  logging_role           = aws_iam_role.test.arn

  force_destroy = %[1]t
}
`, forceDestroy))
}

func testAccAWSTransferServerForceDestroyConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_transfer_server" "test" {
  force_destroy = true
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {
      "Service": "transfer.amazonaws.com"
    },
    "Action": "sts:AssumeRole"
  }]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [{
    "Sid": "AllowFullAccesstoS3",
    "Effect": "Allow",
    "Action": [
      "s3:*"
    ],
    "Resource": "*"
  }]
}
POLICY
}

resource "aws_transfer_user" "test" {
  server_id = aws_transfer_server.test.id
  user_name = %[1]q
  role      = aws_iam_role.test.arn
}

resource "aws_transfer_ssh_key" "test" {
  server_id = aws_transfer_server.test.id
  user_name = aws_transfer_user.test.user_name
  body      = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 phodgson@thoughtworks.com"
}
`, rName)
}

func testAccAWSTransferServerVpcEndpointConfig(rName string) string {
	return composeConfig(
		testAccAWSTransferServerConfigBaseVpc(rName),
		fmt.Sprintf(`
data "aws_vpc_endpoint_service" "test" {
  service = "transfer.server"
}

resource "aws_vpc_endpoint" "test" {
  vpc_id            = aws_vpc.test.id
  vpc_endpoint_type = "Interface"
  service_name      = data.aws_vpc_endpoint_service.test.service_name

  security_group_ids = [
    aws_security_group.test.id,
  ]

  tags = {
    Name = %[1]q
  }
}

resource "aws_transfer_server" "test" {
  endpoint_type = "VPC_ENDPOINT"

  endpoint_details {
    vpc_endpoint_id = aws_vpc_endpoint.test.id
  }
}
`, rName))
}

func testAccAWSTransferServerVpcConfig(rName string) string {
	return composeConfig(
		testAccAWSTransferServerConfigBaseVpc(rName),
		`
resource "aws_transfer_server" "test" {
  endpoint_type = "VPC"

  endpoint_details {
    vpc_id = aws_vpc.test.id
  }
}
`)
}

func testAccAWSTransferServerVpcUpdateConfig(rName string) string {
	return composeConfig(
		testAccAWSTransferServerConfigBaseVpc(rName),
		`
resource "aws_transfer_server" "test" {
  endpoint_type = "VPC"

  endpoint_details {
    subnet_ids = [aws_subnet.test.id]
    vpc_id     = aws_vpc.test.id
  }
}
`)
}

func testAccAWSTransferServerVpcAddressAllocationIdsConfig(rName string) string {
	return composeConfig(
		testAccAWSTransferServerConfigBaseVpc(rName),
		`
resource "aws_transfer_server" "test" {
  endpoint_type = "VPC"

  endpoint_details {
    address_allocation_ids = [aws_eip.test[0].id]
    subnet_ids             = [aws_subnet.test.id]
    vpc_id                 = aws_vpc.test.id
  }
}
`)
}

func testAccAWSTransferServerVpcAddressAllocationIdsUpdateConfig(rName string) string {
	return composeConfig(
		testAccAWSTransferServerConfigBaseVpc(rName),
		`
resource "aws_transfer_server" "test" {
  endpoint_type = "VPC"

  endpoint_details {
    address_allocation_ids = [aws_eip.test[1].id]
    subnet_ids             = [aws_subnet.test.id]
    vpc_id                 = aws_vpc.test.id
  }
}
`)
}

func testAccAWSTransferServerVpcAddressAllocationIdsSecurityGroupIdsConfig(rName string) string {
	return composeConfig(
		testAccAWSTransferServerConfigBaseVpc(rName),
		`
resource "aws_transfer_server" "test" {
  endpoint_type = "VPC"

  endpoint_details {
    address_allocation_ids = [aws_eip.test[0].id]
    security_group_ids     = [aws_security_group.test.id]
    subnet_ids             = [aws_subnet.test.id]
    vpc_id                 = aws_vpc.test.id
  }
}
`)
}

func testAccAWSTransferServerVpcAddressAllocationIdsSecurityGroupIdsUpdateConfig(rName string) string {
	return composeConfig(
		testAccAWSTransferServerConfigBaseVpc(rName),
		fmt.Sprintf(`
resource "aws_security_group" "test2" {
  name   = "%[1]s-2"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = "%[1]s-2"
  }
}
resource "aws_transfer_server" "test" {
  endpoint_type = "VPC"

  endpoint_details {
    address_allocation_ids = [aws_eip.test[1].id]
    security_group_ids     = [aws_security_group.test2.id]
    subnet_ids             = [aws_subnet.test.id]
    vpc_id                 = aws_vpc.test.id
  }
}
`, rName))
}

func testAccAWSTransferServerVpcSecurityGroupIdsConfig(rName string) string {
	return composeConfig(
		testAccAWSTransferServerConfigBaseVpc(rName),
		`
resource "aws_transfer_server" "test" {
  endpoint_type = "VPC"

  endpoint_details {
    security_group_ids = [aws_security_group.test.id]
    vpc_id             = aws_vpc.test.id
  }
}
`)
}

func testAccAWSTransferServerVpcSecurityGroupIdsUpdateConfig(rName string) string {
	return composeConfig(
		testAccAWSTransferServerConfigBaseVpc(rName),
		fmt.Sprintf(`
resource "aws_security_group" "test2" {
  name   = "%[1]s-2"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = "%[1]s-2"
  }
}

resource "aws_transfer_server" "test" {
  endpoint_type = "VPC"

  endpoint_details {
    security_group_ids = [aws_security_group.test2.id]
    vpc_id             = aws_vpc.test.id
  }
}
`, rName))
}

func testAccAWSTransferServerHostKeyConfig(hostKey string) string {
	return fmt.Sprintf(`
resource "aws_transfer_server" "test" {
  host_key = file(%[1]q)
}
`, hostKey)
}

func testAccAWSTransferServerProtocolsConfig(rName string) string {
	return composeConfig(
		testAccAWSTransferServerConfigBaseVpc(rName),
		testAccAWSTransferServerConfigBaseApiGateway(rName),
		testAccAWSTransferServerConfigBaseLoggingRole(rName),
		`
resource "aws_transfer_server" "test" {
  identity_provider_type = "API_GATEWAY"
  url                    = "${aws_api_gateway_deployment.test.invoke_url}${aws_api_gateway_resource.test.path}"
  invocation_role        = aws_iam_role.test.arn
  logging_role           = aws_iam_role.test.arn
  protocols              = ["FTP"]

  endpoint_type = "VPC"
  endpoint_details {
    subnet_ids = [aws_subnet.test.id]
    vpc_id     = aws_vpc.test.id
  }
}
`)
}

func testAccAWSTransferServerConfigRootCA(rName string) string {
	return fmt.Sprintf(`
resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7
  type                            = "ROOT"

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = "%[1]s.com"
    }
  }
}
`, rName)
}

func testAccAWSTransferServerProtocolsUpdateConfig(rName string) string {
	return composeConfig(
		testAccAWSTransferServerConfigBaseVpc(rName),
		testAccAWSTransferServerConfigBaseApiGateway(rName),
		testAccAWSTransferServerConfigBaseLoggingRole(rName),
		testAccAWSTransferServerConfigRootCA(rName),
		fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  domain_name               = "test.%[1]s.com"
  certificate_authority_arn = aws_acmpca_certificate_authority.test.arn
}

resource "aws_transfer_server" "test" {
  identity_provider_type = "API_GATEWAY"
  url                    = "${aws_api_gateway_deployment.test.invoke_url}${aws_api_gateway_resource.test.path}"
  invocation_role        = aws_iam_role.test.arn
  logging_role           = aws_iam_role.test.arn
  protocols              = ["FTP", "FTPS"]
  certificate            = aws_acm_certificate.test.arn

  endpoint_type = "VPC"
  endpoint_details {
    subnet_ids = [aws_subnet.test.id]
    vpc_id     = aws_vpc.test.id
  }
}
`, rName))
}
