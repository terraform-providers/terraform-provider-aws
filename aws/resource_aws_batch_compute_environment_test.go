package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/naming"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/batch/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func init() {
	resource.AddTestSweepers("aws_batch_compute_environment", &resource.Sweeper{
		Name: "aws_batch_compute_environment",
		Dependencies: []string{
			"aws_batch_job_queue",
		},
		F: testSweepBatchComputeEnvironments,
	})
}

func testSweepBatchComputeEnvironments(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*AWSClient).batchconn
	iamconn := client.(*AWSClient).iamconn

	var sweeperErrs *multierror.Error

	input := &batch.DescribeComputeEnvironmentsInput{}
	r := resourceAwsBatchComputeEnvironment()

	err = conn.DescribeComputeEnvironmentsPages(input, func(page *batch.DescribeComputeEnvironmentsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, computeEnvironment := range page.ComputeEnvironments {
			name := aws.StringValue(computeEnvironment.ComputeEnvironmentName)

			d := r.Data(nil)
			d.SetId(name)
			d.Set("compute_environment_name", name)

			// Reference: https://aws.amazon.com/premiumsupport/knowledge-center/batch-invalid-compute-environment/
			//
			// When a Compute Environment becomes INVALID, it is typically because the associated
			// IAM Role has disappeared. There is no automatic resolution via the API, except to
			// associate a new IAM Role that is valid, then delete the Compute Environment.
			//
			// We avoid doing this in the resource because it would be very unexpected behavior
			// for the resource and this issue should be fixed in the API (e.g. Service Linked Role).
			//
			// To save writing much more logic around IAM Role deletion, we allow the
			// aws_iam_role sweeper to handle cleaning these up.
			if aws.StringValue(computeEnvironment.Status) == batch.CEStatusInvalid {
				// Reusing the IAM Role name to prevent collisions and inventing a naming scheme
				serviceRoleARN, err := arn.Parse(aws.StringValue(computeEnvironment.ServiceRole))

				if err != nil {
					sweeperErr := fmt.Errorf("error parsing Batch Compute Environment (%s) Service Role ARN (%s): %w", name, aws.StringValue(computeEnvironment.ServiceRole), err)
					log.Printf("[ERROR] %s", sweeperErr)
					sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
					continue
				}

				servicePrincipal := fmt.Sprintf("%s.%s", batch.EndpointsID, testAccGetPartitionDNSSuffix())
				serviceRoleName := strings.TrimPrefix(serviceRoleARN.Resource, "role/")
				serviceRolePolicyARN := arn.ARN{
					AccountID: "aws",
					Partition: testAccGetPartition(),
					Resource:  "policy/service-role/AWSBatchServiceRole",
					Service:   iam.ServiceName,
				}.String()

				iamCreateRoleInput := &iam.CreateRoleInput{
					AssumeRolePolicyDocument: aws.String(fmt.Sprintf("{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"Service\": \"%s\"},\"Action\":\"sts:AssumeRole\"}]}", servicePrincipal)),
					RoleName:                 aws.String(serviceRoleName),
				}

				_, err = iamconn.CreateRole(iamCreateRoleInput)

				if err != nil {
					sweeperErr := fmt.Errorf("error creating IAM Role (%s) for INVALID Batch Compute Environment (%s): %w", serviceRoleName, name, err)
					log.Printf("[ERROR] %s", sweeperErr)
					sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
					continue
				}

				iamGetRoleInput := &iam.GetRoleInput{
					RoleName: aws.String(serviceRoleName),
				}

				err = iamconn.WaitUntilRoleExists(iamGetRoleInput)

				if err != nil {
					sweeperErr := fmt.Errorf("error waiting for IAM Role (%s) creation for INVALID Batch Compute Environment (%s): %w", serviceRoleName, name, err)
					log.Printf("[ERROR] %s", sweeperErr)
					sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
					continue
				}

				iamAttachRolePolicyInput := &iam.AttachRolePolicyInput{
					PolicyArn: aws.String(serviceRolePolicyARN),
					RoleName:  aws.String(serviceRoleName),
				}

				_, err = iamconn.AttachRolePolicy(iamAttachRolePolicyInput)

				if err != nil {
					sweeperErr := fmt.Errorf("error attaching Batch IAM Policy (%s) to IAM Role (%s) for INVALID Batch Compute Environment (%s): %w", serviceRolePolicyARN, serviceRoleName, name, err)
					log.Printf("[ERROR] %s", sweeperErr)
					sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
					continue
				}
			}

			err := r.Delete(d, client)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting Batch Compute Environment (%s): %w", name, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping Batch Compute Environment sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Batch Compute Environments: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSBatchComputeEnvironment_basic(t *testing.T) {
	var ce batch.ComputeEnvironmentDetail
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_batch_compute_environment.test"
	serviceRoleResourceName := "aws_iam_role.batch_service"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBatch(t) },
		ErrorCheck:   testAccErrorCheck(t, batch.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBatchComputeEnvironmentConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(resourceName, &ce),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "status_reason"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "UNMANAGED"),
				),
			},
		},
	})
}

func TestAccAWSBatchComputeEnvironment_disappears(t *testing.T) {
	var ce batch.ComputeEnvironmentDetail
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_batch_compute_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBatch(t) },
		ErrorCheck:   testAccErrorCheck(t, batch.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBatchComputeEnvironmentConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(resourceName, &ce),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsBatchComputeEnvironment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSBatchComputeEnvironment_NameGenerated(t *testing.T) {
	var ce batch.ComputeEnvironmentDetail
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_batch_compute_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBatch(t) },
		ErrorCheck:   testAccErrorCheck(t, batch.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBatchComputeEnvironmentConfigNameGenerated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(resourceName, &ce),
					naming.TestCheckResourceAttrNameGenerated(resourceName, "compute_environment_name"),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", "terraform-"),
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

func TestAccAWSBatchComputeEnvironment_NamePrefix(t *testing.T) {
	var ce batch.ComputeEnvironmentDetail
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_batch_compute_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBatch(t) },
		ErrorCheck:   testAccErrorCheck(t, batch.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBatchComputeEnvironmentConfigNamePrefix(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(resourceName, &ce),
					naming.TestCheckResourceAttrNameFromPrefix(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", rName),
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

func TestAccAWSBatchComputeEnvironment_createEc2(t *testing.T) {
	var ce batch.ComputeEnvironmentDetail
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_batch_compute_environment.test"
	instanceProfileResourceName := "aws_iam_instance_profile.ecs_instance"
	securityGroupResourceName := "aws_security_group.test"
	serviceRoleResourceName := "aws_iam_role.batch_service"
	subnetResourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBatch(t) },
		ErrorCheck:   testAccErrorCheck(t, batch.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBatchComputeEnvironmentConfigEC2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(resourceName, &ce),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.instance_role", instanceProfileResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "compute_resources.0.instance_type.*", "c4.large"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "16"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.spot_iam_fleet_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "EC2"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "status_reason"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "MANAGED"),
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

func TestAccAWSBatchComputeEnvironment_createEc2_DesiredVcpus_Ec2KeyPair_ImageId_ComputeResourcesTags(t *testing.T) {
	var ce batch.ComputeEnvironmentDetail
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_batch_compute_environment.test"
	amiDatasourceName := "data.aws_ami.amzn-ami-minimal-hvm-ebs"
	instanceProfileResourceName := "aws_iam_instance_profile.ecs_instance"
	keyPairResourceName := "aws_key_pair.test"
	securityGroupResourceName := "aws_security_group.test"
	serviceRoleResourceName := "aws_iam_role.batch_service"
	subnetResourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBatch(t) },
		ErrorCheck:   testAccErrorCheck(t, batch.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBatchComputeEnvironmentConfigEC2WithDesiredVcpusEc2KeyPairImageIdAndComputeResourcesTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(resourceName, &ce),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", "8"),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.ec2_key_pair", keyPairResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.image_id", amiDatasourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.instance_role", instanceProfileResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "compute_resources.0.instance_type.*", "c4.large"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "16"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", "4"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.spot_iam_fleet_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "EC2"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "status_reason"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "MANAGED"),
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

func TestAccAWSBatchComputeEnvironment_createSpot(t *testing.T) {
	var ce batch.ComputeEnvironmentDetail
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_batch_compute_environment.test"
	instanceProfileResourceName := "aws_iam_instance_profile.ecs_instance"
	securityGroupResourceName := "aws_security_group.test"
	serviceRoleResourceName := "aws_iam_role.batch_service"
	spotFleetRoleResourceName := "aws_iam_role.ec2_spot_fleet"
	subnetResourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBatch(t) },
		ErrorCheck:   testAccErrorCheck(t, batch.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBatchComputeEnvironmentConfigSpot(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(resourceName, &ce),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", "2"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.instance_role", instanceProfileResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "compute_resources.0.instance_type.*", "c4.large"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "16"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", "2"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.spot_iam_fleet_role", spotFleetRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "SPOT"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "status_reason"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "MANAGED"),
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

func TestAccAWSBatchComputeEnvironment_createSpot_AllocationStrategy_BidPercentage(t *testing.T) {
	var ce batch.ComputeEnvironmentDetail
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_batch_compute_environment.test"
	instanceProfileResourceName := "aws_iam_instance_profile.ecs_instance"
	securityGroupResourceName := "aws_security_group.test"
	serviceRoleResourceName := "aws_iam_role.batch_service"
	spotFleetRoleResourceName := "aws_iam_role.ec2_spot_fleet"
	subnetResourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBatch(t) },
		ErrorCheck:   testAccErrorCheck(t, batch.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBatchComputeEnvironmentConfigSpotWithAllocationStrategyAndBidPercentage(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(resourceName, &ce),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", "BEST_FIT"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", "60"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.instance_role", instanceProfileResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "compute_resources.0.instance_type.*", "c4.large"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "16"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.spot_iam_fleet_role", spotFleetRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "SPOT"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "status_reason"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "MANAGED"),
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

func TestAccAWSBatchComputeEnvironment_createFargate(t *testing.T) {
	var ce batch.ComputeEnvironmentDetail
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_batch_compute_environment.test"
	securityGroupResourceName := "aws_security_group.test"
	serviceRoleResourceName := "aws_iam_role.batch_service"
	subnetResourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBatch(t) },
		ErrorCheck:   testAccErrorCheck(t, batch.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBatchComputeEnvironmentConfigFargate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(resourceName, &ce),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "16"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.spot_iam_fleet_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "FARGATE"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "status_reason"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "MANAGED"),
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

func TestAccAWSBatchComputeEnvironment_createFargateSpot(t *testing.T) {
	var ce batch.ComputeEnvironmentDetail
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_batch_compute_environment.test"
	securityGroupResourceName := "aws_security_group.test"
	serviceRoleResourceName := "aws_iam_role.batch_service"
	subnetResourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBatch(t) },
		ErrorCheck:   testAccErrorCheck(t, batch.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBatchComputeEnvironmentConfigFargateSpot(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(resourceName, &ce),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "16"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.spot_iam_fleet_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "FARGATE_SPOT"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "status_reason"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "MANAGED"),
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

func TestAccAWSBatchComputeEnvironment_updateState(t *testing.T) {
	var ce batch.ComputeEnvironmentDetail
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_batch_compute_environment.test"
	serviceRoleResourceName := "aws_iam_role.batch_service"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBatch(t) },
		ErrorCheck:   testAccErrorCheck(t, batch.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBatchComputeEnvironmentConfigState(rName, "ENABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(resourceName, &ce),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "status_reason"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "UNMANAGED"),
				),
			},
			{
				Config: testAccAWSBatchComputeEnvironmentConfigState(rName, "disabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(resourceName, &ce),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "DISABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "status_reason"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "UNMANAGED"),
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

func TestAccAWSBatchComputeEnvironment_updateServiceRole(t *testing.T) {
	var ce batch.ComputeEnvironmentDetail
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_batch_compute_environment.test"
	serviceRoleResourceName1 := "aws_iam_role.batch_service"
	serviceRoleResourceName2 := "aws_iam_role.batch_service_2"
	securityGroupResourceName := "aws_security_group.test"
	subnetResourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBatch(t) },
		ErrorCheck:   testAccErrorCheck(t, batch.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBatchComputeEnvironmentConfigFargate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(resourceName, &ce),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "16"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.spot_iam_fleet_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "FARGATE"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", serviceRoleResourceName1, "arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "status_reason"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "MANAGED"),
				),
			},
			{
				Config: testAccAWSBatchComputeEnvironmentConfigFargateUpdatedServiceRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(resourceName, &ce),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "16"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.spot_iam_fleet_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "FARGATE"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", serviceRoleResourceName2, "arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "status_reason"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "MANAGED"),
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

func TestAccAWSBatchComputeEnvironment_defaultServiceRole(t *testing.T) {
	var ce batch.ComputeEnvironmentDetail
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_batch_compute_environment.test"
	securityGroupResourceName := "aws_security_group.test"
	subnetResourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSBatch(t)
			testAccPreCheckIamServiceLinkedRole(t, "/aws-service-role/batch")
		},
		ErrorCheck:   testAccErrorCheck(t, batch.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBatchComputeEnvironmentConfigFargateDefaultServiceRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(resourceName, &ce),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "16"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.spot_iam_fleet_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "FARGATE"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					testAccMatchResourceAttrGlobalARN(resourceName, "service_role", "iam", regexp.MustCompile(`role/aws-service-role/batch`)),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "status_reason"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "MANAGED"),
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

func TestAccAWSBatchComputeEnvironment_ComputeResources_MinVcpus(t *testing.T) {
	var ce batch.ComputeEnvironmentDetail
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_batch_compute_environment.test"
	instanceProfileResourceName := "aws_iam_instance_profile.ecs_instance"
	securityGroupResourceName := "aws_security_group.test"
	serviceRoleResourceName := "aws_iam_role.batch_service"
	subnetResourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBatch(t) },
		ErrorCheck:   testAccErrorCheck(t, batch.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBatchComputeEnvironmentConfigComputeResourcesMaxVcpusMinVcpus(rName, 4, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(resourceName, &ce),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.instance_role", instanceProfileResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "compute_resources.0.instance_type.*", "optimal"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "4"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.spot_iam_fleet_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "EC2"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "status_reason"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "MANAGED"),
				),
			},
			{
				Config: testAccAWSBatchComputeEnvironmentConfigComputeResourcesMaxVcpusMinVcpus(rName, 4, 4),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(resourceName, &ce),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", "4"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.instance_role", instanceProfileResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "compute_resources.0.instance_type.*", "optimal"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "4"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", "4"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.spot_iam_fleet_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "EC2"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "status_reason"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "MANAGED"),
				),
			},
			{
				Config: testAccAWSBatchComputeEnvironmentConfigComputeResourcesMaxVcpusMinVcpus(rName, 4, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(resourceName, &ce),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", "4"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.instance_role", instanceProfileResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "compute_resources.0.instance_type.*", "optimal"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "4"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", "2"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.spot_iam_fleet_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "EC2"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "status_reason"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "MANAGED"),
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

func TestAccAWSBatchComputeEnvironment_ComputeResources_MaxVcpus(t *testing.T) {
	var ce batch.ComputeEnvironmentDetail
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_batch_compute_environment.test"
	instanceProfileResourceName := "aws_iam_instance_profile.ecs_instance"
	securityGroupResourceName := "aws_security_group.test"
	serviceRoleResourceName := "aws_iam_role.batch_service"
	subnetResourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBatch(t) },
		ErrorCheck:   testAccErrorCheck(t, batch.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBatchComputeEnvironmentConfigComputeResourcesMaxVcpusMinVcpus(rName, 4, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(resourceName, &ce),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.instance_role", instanceProfileResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "compute_resources.0.instance_type.*", "optimal"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "4"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.spot_iam_fleet_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "EC2"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "status_reason"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "MANAGED"),
				),
			},
			{
				Config: testAccAWSBatchComputeEnvironmentConfigComputeResourcesMaxVcpusMinVcpus(rName, 8, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(resourceName, &ce),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.instance_role", instanceProfileResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "compute_resources.0.instance_type.*", "optimal"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "8"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.spot_iam_fleet_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "EC2"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "status_reason"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "MANAGED"),
				),
			},
			{
				Config: testAccAWSBatchComputeEnvironmentConfigComputeResourcesMaxVcpusMinVcpus(rName, 2, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(resourceName, &ce),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.instance_role", instanceProfileResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "compute_resources.0.instance_type.*", "optimal"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "2"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.spot_iam_fleet_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "EC2"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "status_reason"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "MANAGED"),
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

func TestAccAWSBatchComputeEnvironment_launchTemplate(t *testing.T) {
	var ce batch.ComputeEnvironmentDetail
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_batch_compute_environment.test"
	instanceProfileResourceName := "aws_iam_instance_profile.ecs_instance"
	launchTemplateResourceName := "aws_launch_template.test"
	securityGroupResourceName := "aws_security_group.test"
	serviceRoleResourceName := "aws_iam_role.batch_service"
	spotFleetRoleResourceName := "aws_iam_role.ec2_spot_fleet"
	subnetResourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBatch(t) },
		ErrorCheck:   testAccErrorCheck(t, batch.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBatchComputeEnvironmentConfigLaunchTemplate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(resourceName, &ce),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.instance_role", instanceProfileResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "compute_resources.0.instance_type.*", "c4.large"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.0.launch_template_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.launch_template.0.launch_template_name", launchTemplateResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.0.version", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "16"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.spot_iam_fleet_role", spotFleetRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "SPOT"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "status_reason"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "MANAGED"),
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

func TestAccAWSBatchComputeEnvironment_UpdateLaunchTemplate(t *testing.T) {
	var ce batch.ComputeEnvironmentDetail
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_batch_compute_environment.test"
	instanceProfileResourceName := "aws_iam_instance_profile.ecs_instance"
	launchTemplateResourceName := "aws_launch_template.test"
	securityGroupResourceName := "aws_security_group.test"
	serviceRoleResourceName := "aws_iam_role.batch_service"
	spotFleetRoleResourceName := "aws_iam_role.ec2_spot_fleet"
	subnetResourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBatch(t) },
		ErrorCheck:   testAccErrorCheck(t, batch.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBatchComputeEnvironmentUpdateLaunchTemplateInExistingComputeEnvironment(rName, "$Default"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(resourceName, &ce),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.instance_role", instanceProfileResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "compute_resources.0.instance_type.*", "c4.large"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.launch_template.0.launch_template_id", launchTemplateResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.0.launch_template_name", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.0.version", "$Default"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "16"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.spot_iam_fleet_role", spotFleetRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "SPOT"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "status_reason"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "MANAGED"),
				),
			},
			{
				Config: testAccAWSBatchComputeEnvironmentUpdateLaunchTemplateInExistingComputeEnvironment(rName, "$Latest"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(resourceName, &ce),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.instance_role", instanceProfileResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "compute_resources.0.instance_type.*", "c4.large"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.launch_template.0.launch_template_id", launchTemplateResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.0.launch_template_name", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.0.version", "$Latest"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "16"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.spot_iam_fleet_role", spotFleetRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "SPOT"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "status_reason"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "MANAGED"),
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

func TestAccAWSBatchComputeEnvironment_UpdateSecurityGroupsAndSubnets_Fargate(t *testing.T) {
	var ce batch.ComputeEnvironmentDetail
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_batch_compute_environment.test"
	securityGroupResourceName1 := "aws_security_group.test"
	securityGroupResourceName2 := "aws_security_group.test_2"
	securityGroupResourceName3 := "aws_security_group.test_3"
	serviceRoleResourceName := "aws_iam_role.batch_service"
	subnetResourceName1 := "aws_subnet.test"
	subnetResourceName2 := "aws_subnet.test_2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBatch(t) },
		ErrorCheck:   testAccErrorCheck(t, batch.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBatchComputeEnvironmentConfigFargate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(resourceName, &ce),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "16"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName1, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.spot_iam_fleet_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName1, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "FARGATE"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "status_reason"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "MANAGED"),
				),
			},
			{
				Config: testAccAWSBatchComputeEnvironmentConfigFargateUpdatedSecurityGroupsAndSubnets(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(resourceName, &ce),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "16"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName2, "id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName3, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.spot_iam_fleet_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName2, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "FARGATE"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "status_reason"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "MANAGED"),
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

func TestAccAWSBatchComputeEnvironment_Tags(t *testing.T) {
	var ce batch.ComputeEnvironmentDetail
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_batch_compute_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBatch(t) },
		ErrorCheck:   testAccErrorCheck(t, batch.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBatchComputeEnvironmentConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(resourceName, &ce),
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
				Config: testAccAWSBatchComputeEnvironmentConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(resourceName, &ce),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSBatchComputeEnvironmentConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(resourceName, &ce),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSBatchComputeEnvironment_createUnmanagedWithComputeResources(t *testing.T) {
	var ce batch.ComputeEnvironmentDetail
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_batch_compute_environment.test"
	serviceRoleResourceName := "aws_iam_role.batch_service"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBatch(t) },
		ErrorCheck:   testAccErrorCheck(t, batch.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBatchComputeEnvironmentConfigUnmanagedWithComputeResources(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(resourceName, &ce),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "status_reason"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "UNMANAGED"),
				),
			},
			// Can't import in this scenario.
		},
	})
}

// Test plan time errors...

func TestAccAWSBatchComputeEnvironment_createEc2WithoutComputeResources(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBatch(t) },
		ErrorCheck:   testAccErrorCheck(t, batch.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSBatchComputeEnvironmentConfigEC2WithoutComputeResources(rName),
				ExpectError: regexp.MustCompile(`computeResources must be provided for a MANAGED compute environment`),
			},
		},
	})
}

func TestAccAWSBatchComputeEnvironment_createSpotWithoutIamFleetRole(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBatch(t) },
		ErrorCheck:   testAccErrorCheck(t, batch.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSBatchComputeEnvironmentConfigSpotWithoutIamFleetRole(rName),
				ExpectError: regexp.MustCompile(`ComputeResources.spotIamFleetRole cannot not be null or empty`),
			},
		},
	})
}

func testAccCheckBatchComputeEnvironmentDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).batchconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_batch_compute_environment" {
			continue
		}

		_, err := finder.ComputeEnvironmentDetailByName(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Batch Compute Environment %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccCheckAwsBatchComputeEnvironmentExists(n string, v *batch.ComputeEnvironmentDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Batch Compute Environment ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).batchconn

		computeEnvironment, err := finder.ComputeEnvironmentDetailByName(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *computeEnvironment

		return nil
	}
}

func testAccPreCheckAWSBatch(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).batchconn

	input := &batch.DescribeComputeEnvironmentsInput{}

	_, err := conn.DescribeComputeEnvironments(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAWSBatchComputeEnvironmentConfigBase(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "ecs_instance" {
  name = "%[1]s_ecs_instance"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Effect": "Allow",
    "Principal": {
      "Service": "ec2.${data.aws_partition.current.dns_suffix}"
    }
  }]
}
EOF
}

resource "aws_iam_role_policy_attachment" "ecs_instance" {
  role       = aws_iam_role.ecs_instance.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonEC2ContainerServiceforEC2Role"
}

resource "aws_iam_instance_profile" "ecs_instance" {
  name = %[1]q
  role = aws_iam_role_policy_attachment.ecs_instance.role
}

resource "aws_iam_role" "batch_service" {
  name = "%[1]s_batch_service"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Effect": "Allow",
    "Principal": {
      "Service": "batch.${data.aws_partition.current.dns_suffix}"
    }
  }]
}
EOF
}

resource "aws_iam_role_policy_attachment" "batch_service" {
  role       = aws_iam_role.batch_service.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSBatchServiceRole"
}

resource "aws_iam_role" "ec2_spot_fleet" {
  name = "%[1]s_ec2_spot_fleet"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Effect": "Allow",
    "Principal": {
      "Service": "spotfleet.${data.aws_partition.current.dns_suffix}"
    }
  }]
}
EOF
}

resource "aws_iam_role_policy_attachment" "ec2_spot_fleet" {
  role       = aws_iam_role.ec2_spot_fleet.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonEC2SpotFleetTaggingRole"
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.1.1.0/24"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccAWSBatchComputeEnvironmentConfigBasic(rName string) string {
	return composeConfig(
		testAccAWSBatchComputeEnvironmentConfigBase(rName),
		fmt.Sprintf(`
resource "aws_batch_compute_environment" "test" {
  compute_environment_name = %[1]q

  service_role = aws_iam_role.batch_service.arn
  type         = "UNMANAGED"
  depends_on   = [aws_iam_role_policy_attachment.batch_service]
}
`, rName))
}

func testAccAWSBatchComputeEnvironmentConfigNameGenerated(rName string) string {
	return composeConfig(
		testAccAWSBatchComputeEnvironmentConfigBase(rName),
		`
resource "aws_batch_compute_environment" "test" {
  service_role = aws_iam_role.batch_service.arn
  type         = "UNMANAGED"
  depends_on   = [aws_iam_role_policy_attachment.batch_service]
}
`)
}

func testAccAWSBatchComputeEnvironmentConfigNamePrefix(rName string) string {
	return composeConfig(
		testAccAWSBatchComputeEnvironmentConfigBase(rName),
		fmt.Sprintf(`
resource "aws_batch_compute_environment" "test" {
  compute_environment_name_prefix = %[1]q

  service_role = aws_iam_role.batch_service.arn
  type         = "UNMANAGED"
  depends_on   = [aws_iam_role_policy_attachment.batch_service]
}
`, rName))
}

func testAccAWSBatchComputeEnvironmentConfigEC2(rName string) string {
	return composeConfig(
		testAccAWSBatchComputeEnvironmentConfigBase(rName),
		fmt.Sprintf(`
resource "aws_batch_compute_environment" "test" {
  compute_environment_name = %[1]q

  compute_resources {
    instance_role = aws_iam_instance_profile.ecs_instance.arn
    instance_type = [
      "c4.large",
    ]
    max_vcpus = 16
    security_group_ids = [
      aws_security_group.test.id
    ]
    subnets = [
      aws_subnet.test.id
    ]
    type = "EC2"
  }

  service_role = aws_iam_role.batch_service.arn
  type         = "MANAGED"
  depends_on   = [aws_iam_role_policy_attachment.batch_service]
}
`, rName))
}

func testAccAWSBatchComputeEnvironmentConfigEC2WithDesiredVcpusEc2KeyPairImageIdAndComputeResourcesTags(rName string) string {
	return composeConfig(
		testAccAWSBatchComputeEnvironmentConfigBase(rName),
		testAccLatestAmazonLinuxHvmEbsAmiConfig(),
		fmt.Sprintf(`
resource "aws_batch_compute_environment" "test" {
  compute_environment_name = %[1]q

  compute_resources {
    instance_role = aws_iam_instance_profile.ecs_instance.arn
    instance_type = [
      "c4.large",
    ]
    max_vcpus     = 16
    min_vcpus     = 4
    desired_vcpus = 8
    security_group_ids = [
      aws_security_group.test.id
    ]
    subnets = [
      aws_subnet.test.id
    ]
    type = "EC2"

    tags = {
      key1 = "value1"
    }

    ec2_key_pair = aws_key_pair.test.id
    image_id     = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  }

  service_role = aws_iam_role.batch_service.arn
  type         = "MANAGED"
  depends_on   = [aws_iam_role_policy_attachment.batch_service]
}

resource "aws_key_pair" "test" {
  key_name   = %[1]q
  public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 phodgson@thoughtworks.com"
}
`, rName))
}

func testAccAWSBatchComputeEnvironmentConfigFargate(rName string) string {
	return composeConfig(
		testAccAWSBatchComputeEnvironmentConfigBase(rName),
		fmt.Sprintf(`
resource "aws_batch_compute_environment" "test" {
  compute_environment_name = %[1]q

  compute_resources {
    max_vcpus = 16
    security_group_ids = [
      aws_security_group.test.id
    ]
    subnets = [
      aws_subnet.test.id
    ]
    type = "FARGATE"
  }

  service_role = aws_iam_role.batch_service.arn
  type         = "MANAGED"
  depends_on   = [aws_iam_role_policy_attachment.batch_service]
}
`, rName))
}

func testAccAWSBatchComputeEnvironmentConfigFargateDefaultServiceRole(rName string) string {
	return composeConfig(
		testAccAWSBatchComputeEnvironmentConfigBase(rName),
		fmt.Sprintf(`
resource "aws_batch_compute_environment" "test" {
  compute_environment_name = %[1]q

  compute_resources {
    max_vcpus = 16
    security_group_ids = [
      aws_security_group.test.id
    ]
    subnets = [
      aws_subnet.test.id
    ]
    type = "FARGATE"
  }

  type = "MANAGED"
}
`, rName))
}

func testAccAWSBatchComputeEnvironmentConfigFargateUpdatedServiceRole(rName string) string {
	return composeConfig(
		testAccAWSBatchComputeEnvironmentConfigBase(rName),
		fmt.Sprintf(`
resource "aws_batch_compute_environment" "test" {
  compute_environment_name = %[1]q

  compute_resources {
    max_vcpus = 16
    security_group_ids = [
      aws_security_group.test.id
    ]
    subnets = [
      aws_subnet.test.id
    ]
    type = "FARGATE"
  }

  service_role = aws_iam_role.batch_service_2.arn
  type         = "MANAGED"
  depends_on   = [aws_iam_role_policy_attachment.batch_service_2]
}

resource "aws_iam_role" "batch_service_2" {
  name = "%[1]s_batch_service_2"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Effect": "Allow",
    "Principal": {
      "Service": "batch.${data.aws_partition.current.dns_suffix}"
    }
  }]
}
EOF
}

resource "aws_iam_role_policy_attachment" "batch_service_2" {
  role       = aws_iam_role.batch_service_2.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSBatchServiceRole"
}
`, rName))
}

func testAccAWSBatchComputeEnvironmentConfigFargateSpot(rName string) string {
	return composeConfig(
		testAccAWSBatchComputeEnvironmentConfigBase(rName),
		fmt.Sprintf(`
resource "aws_batch_compute_environment" "test" {
  compute_environment_name = %[1]q

  compute_resources {
    max_vcpus = 16
    security_group_ids = [
      aws_security_group.test.id
    ]
    subnets = [
      aws_subnet.test.id
    ]
    type = "FARGATE_SPOT"
  }

  service_role = aws_iam_role.batch_service.arn
  type         = "managed"
  depends_on   = [aws_iam_role_policy_attachment.batch_service]
}
`, rName))
}

func testAccAWSBatchComputeEnvironmentConfigSpot(rName string) string {
	return composeConfig(
		testAccAWSBatchComputeEnvironmentConfigBase(rName),
		fmt.Sprintf(`
resource "aws_batch_compute_environment" "test" {
  compute_environment_name = %[1]q

  compute_resources {
    instance_role = aws_iam_instance_profile.ecs_instance.arn
    instance_type = [
      "c4.large",
    ]
    max_vcpus = 16
    min_vcpus = 2
    security_group_ids = [
      aws_security_group.test.id
    ]
    spot_iam_fleet_role = aws_iam_role.ec2_spot_fleet.arn
    subnets = [
      aws_subnet.test.id
    ]
    type = "spot"
  }

  service_role = aws_iam_role.batch_service.arn
  type         = "MANAGED"
  depends_on   = [aws_iam_role_policy_attachment.batch_service]
}
`, rName))
}

func testAccAWSBatchComputeEnvironmentConfigSpotWithAllocationStrategyAndBidPercentage(rName string) string {
	return composeConfig(
		testAccAWSBatchComputeEnvironmentConfigBase(rName),
		fmt.Sprintf(`
resource "aws_batch_compute_environment" "test" {
  compute_environment_name = %[1]q

  compute_resources {
    allocation_strategy = "Best_Fit"
    bid_percentage      = 60
    instance_role       = aws_iam_instance_profile.ecs_instance.arn
    instance_type = [
      "c4.large",
    ]
    max_vcpus = 16
    min_vcpus = 0
    security_group_ids = [
      aws_security_group.test.id
    ]
    spot_iam_fleet_role = aws_iam_role.ec2_spot_fleet.arn
    subnets = [
      aws_subnet.test.id
    ]
    type = "SPOT"
  }

  service_role = aws_iam_role.batch_service.arn
  type         = "MANAGED"
  depends_on   = [aws_iam_role_policy_attachment.batch_service]
}
`, rName))
}

func testAccAWSBatchComputeEnvironmentConfigState(rName string, state string) string {
	return composeConfig(
		testAccAWSBatchComputeEnvironmentConfigBase(rName),
		fmt.Sprintf(`
resource "aws_batch_compute_environment" "test" {
  compute_environment_name = %[1]q

  service_role = aws_iam_role.batch_service.arn
  type         = "UNMANAGED"
  state        = %[2]q
  depends_on   = [aws_iam_role_policy_attachment.batch_service]
}
`, rName, state))
}

func testAccAWSBatchComputeEnvironmentConfigComputeResourcesMaxVcpusMinVcpus(rName string, maxVcpus int, minVcpus int) string {
	return composeConfig(
		testAccAWSBatchComputeEnvironmentConfigBase(rName),
		fmt.Sprintf(`
resource "aws_batch_compute_environment" "test" {
  compute_environment_name = %[1]q

  compute_resources {
    instance_role = aws_iam_instance_profile.ecs_instance.arn
    instance_type = ["optimal"]
    max_vcpus     = %[2]d
    min_vcpus     = %[3]d
    security_group_ids = [
      aws_security_group.test.id
    ]
    subnets = [
      aws_subnet.test.id
    ]
    type = "EC2"
  }

  service_role = aws_iam_role.batch_service.arn
  type         = "MANAGED"
  depends_on   = [aws_iam_role_policy_attachment.batch_service]
}
`, rName, maxVcpus, minVcpus))
}

func testAccAWSBatchComputeEnvironmentConfigFargateUpdatedSecurityGroupsAndSubnets(rName string) string {
	return composeConfig(
		testAccAWSBatchComputeEnvironmentConfigBase(rName),
		fmt.Sprintf(`
resource "aws_batch_compute_environment" "test" {
  compute_environment_name = %[1]q

  compute_resources {
    max_vcpus = 16
    security_group_ids = [
      aws_security_group.test_2.id,
      aws_security_group.test_3.id,
    ]
    subnets = [
      aws_subnet.test_2.id
    ]
    type = "FARGATE"
  }

  service_role = aws_iam_role.batch_service.arn
  type         = "MANAGED"
  depends_on   = [aws_iam_role_policy_attachment.batch_service]
}

resource "aws_security_group" "test_2" {
  name   = "%[1]s_2"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = "%[1]s_2"
  }
}

resource "aws_security_group" "test_3" {
  name   = "%[1]s_3"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = "%[1]s_3"
  }
}

resource "aws_subnet" "test_2" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.1.2.0/24"

  tags = {
    Name = "%[1]s_2"
  }
}
`, rName))
}

func testAccAWSBatchComputeEnvironmentConfigEC2WithoutComputeResources(rName string) string {
	return composeConfig(
		testAccAWSBatchComputeEnvironmentConfigBase(rName),
		fmt.Sprintf(`
resource "aws_batch_compute_environment" "test" {
  compute_environment_name = %[1]q

  service_role = aws_iam_role.batch_service.arn
  type         = "MANAGED"
  depends_on   = [aws_iam_role_policy_attachment.batch_service]
}
`, rName))
}

func testAccAWSBatchComputeEnvironmentConfigUnmanagedWithComputeResources(rName string) string {
	return composeConfig(
		testAccAWSBatchComputeEnvironmentConfigBase(rName),
		fmt.Sprintf(`
resource "aws_batch_compute_environment" "test" {
  compute_environment_name = %[1]q

  compute_resources {
    instance_role = aws_iam_instance_profile.ecs_instance.arn
    instance_type = [
      "c4.large",
    ]
    max_vcpus = 16
    min_vcpus = 0
    security_group_ids = [
      aws_security_group.test.id
    ]
    subnets = [
      aws_subnet.test.id
    ]
    type = "EC2"
  }

  service_role = aws_iam_role.batch_service.arn
  type         = "UNMANAGED"
  depends_on   = [aws_iam_role_policy_attachment.batch_service]
}
`, rName))
}

func testAccAWSBatchComputeEnvironmentConfigSpotWithoutIamFleetRole(rName string) string {
	return composeConfig(
		testAccAWSBatchComputeEnvironmentConfigBase(rName),
		fmt.Sprintf(`
resource "aws_batch_compute_environment" "test" {
  compute_environment_name = %[1]q

  compute_resources {
    instance_role = aws_iam_instance_profile.ecs_instance.arn
    instance_type = [
      "c4.large",
    ]
    max_vcpus = 16
    min_vcpus = 0
    security_group_ids = [
      aws_security_group.test.id
    ]
    subnets = [
      aws_subnet.test.id
    ]
    type = "SPOT"
  }

  service_role = aws_iam_role.batch_service.arn
  type         = "MANAGED"
  depends_on   = [aws_iam_role_policy_attachment.batch_service]
}
`, rName))
}

func testAccAWSBatchComputeEnvironmentConfigLaunchTemplate(rName string) string {
	return composeConfig(
		testAccAWSBatchComputeEnvironmentConfigBase(rName),
		fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q
}

resource "aws_batch_compute_environment" "test" {
  compute_environment_name = %[1]q

  compute_resources {
    instance_role = aws_iam_instance_profile.ecs_instance.arn
    instance_type = [
      "c4.large",
    ]

    launch_template {
      launch_template_name = aws_launch_template.test.name
    }

    max_vcpus = 16
    min_vcpus = 0
    security_group_ids = [
      aws_security_group.test.id
    ]
    spot_iam_fleet_role = aws_iam_role.ec2_spot_fleet.arn
    subnets = [
      aws_subnet.test.id
    ]
    type = "SPOT"
  }

  service_role = aws_iam_role.batch_service.arn
  type         = "MANAGED"
  depends_on   = [aws_iam_role_policy_attachment.batch_service]
}
`, rName))
}

func testAccAWSBatchComputeEnvironmentUpdateLaunchTemplateInExistingComputeEnvironment(rName string, version string) string {
	return composeConfig(
		testAccAWSBatchComputeEnvironmentConfigBase(rName),
		fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q
}

resource "aws_batch_compute_environment" "test" {
  compute_environment_name = %[1]q

  compute_resources {
    instance_role = aws_iam_instance_profile.ecs_instance.arn
    instance_type = [
      "c4.large",
    ]

    launch_template {
      launch_template_id = aws_launch_template.test.id
      version            = %[2]q
    }

    max_vcpus = 16
    min_vcpus = 0
    security_group_ids = [
      aws_security_group.test.id
    ]
    spot_iam_fleet_role = aws_iam_role.ec2_spot_fleet.arn
    subnets = [
      aws_subnet.test.id
    ]
    type = "SPOT"
  }

  service_role = aws_iam_role.batch_service.arn
  type         = "MANAGED"
  depends_on   = [aws_iam_role_policy_attachment.batch_service]
}
`, rName, version))
}

func testAccAWSBatchComputeEnvironmentConfigTags1(rName string, tagKey1, tagValue1 string) string {
	return composeConfig(
		testAccAWSBatchComputeEnvironmentConfigBase(rName),
		fmt.Sprintf(`
resource "aws_batch_compute_environment" "test" {
  depends_on = [aws_iam_role_policy_attachment.batch_service]

  compute_environment_name = %[1]q
  service_role             = aws_iam_role.batch_service.arn
  type                     = "UNMANAGED"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccAWSBatchComputeEnvironmentConfigTags2(rName string, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return composeConfig(
		testAccAWSBatchComputeEnvironmentConfigBase(rName),
		fmt.Sprintf(`
resource "aws_batch_compute_environment" "test" {
  depends_on = [aws_iam_role_policy_attachment.batch_service]

  compute_environment_name = %[1]q
  service_role             = aws_iam_role.batch_service.arn
  type                     = "UNMANAGED"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
