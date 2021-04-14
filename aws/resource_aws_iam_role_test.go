package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_iam_role", &resource.Sweeper{
		Name: "aws_iam_role",
		Dependencies: []string{
			"aws_batch_compute_environment",
			"aws_cloudformation_stack_set_instance",
			"aws_cognito_user_pool",
			"aws_config_configuration_aggregator",
			"aws_config_configuration_recorder",
			"aws_datasync_location_s3",
			"aws_dax_cluster",
			"aws_db_instance",
			"aws_db_option_group",
			"aws_eks_cluster",
			"aws_elastic_beanstalk_application",
			"aws_elastic_beanstalk_environment",
			"aws_elasticsearch_domain",
			"aws_glue_crawler",
			"aws_glue_job",
			"aws_instance",
			"aws_lambda_function",
			"aws_launch_configuration",
			"aws_redshift_cluster",
			"aws_spot_fleet_request",
		},
		F: testSweepIamRoles,
	})
}

func testSweepIamRoles(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).iamconn
	prefixes := []string{
		"another_rds",
		"batch_tf_acc_test",
		"codepipeline-",
		"cognito_authenticated_",
		"cognito_unauthenticated_",
		"CWLtoKinesisRole_",
		"ecs_instance_role",
		"ecs_tf",
		"EMR_AutoScaling_DefaultRole_",
		"enhanced-monitoring-role-",
		"es-domain-role-",
		"event_",
		"firehose",
		"foo_role",
		"foo-role",
		"foobar",
		"iam_emr",
		"iam_for_lambda",
		"iam_for_sfn",
		"rds",
		"role",
		"sns-delivery-status",
		"ssm_role",
		"ssm-role",
		"terraform-",
		"test",
		"tf",
	}
	// Some acceptance tests use acctest.RandString(10) rather than acctest.RandomWithPrefix()
	regex := regexp.MustCompile(`^[a-zA-Z0-9]{10}$`)
	roles := make([]*iam.Role, 0)

	err = conn.ListRolesPages(&iam.ListRolesInput{}, func(page *iam.ListRolesOutput, lastPage bool) bool {
		for _, role := range page.Roles {
			if regex.MatchString(aws.StringValue(role.RoleName)) {
				roles = append(roles, role)
				continue
			}

			for _, prefix := range prefixes {
				if strings.HasPrefix(aws.StringValue(role.RoleName), prefix) {
					roles = append(roles, role)
					break
				}
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping IAM Role sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error retrieving IAM Roles: %s", err)
	}

	if len(roles) == 0 {
		log.Print("[DEBUG] No IAM Roles to sweep")
		return nil
	}

	var sweeperErrs *multierror.Error

	for _, role := range roles {
		roleName := aws.StringValue(role.RoleName)
		log.Printf("[DEBUG] Deleting IAM Role (%s)", roleName)

		err := deleteIamRole(conn, roleName, true, true, true)
		if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			continue
		}
		if testSweepSkipResourceError(err) {
			log.Printf("[WARN] Skipping IAM Role (%s): %s", roleName, err)
			continue
		}
		if err != nil {
			sweeperErr := fmt.Errorf("error deleting IAM Role (%s): %w", roleName, err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSIAMRole_basic(t *testing.T) {
	var conf iam.GetRoleOutput
	rName := acctest.RandString(10)
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, iam.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIAMRoleConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "path", "/"),
					resource.TestCheckResourceAttrSet(resourceName, "create_date"),
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

func TestAccAWSIAMRole_basicWithDescription(t *testing.T) {
	var conf iam.GetRoleOutput
	rName := acctest.RandString(10)
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, iam.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIAMRoleConfigWithDescription(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "path", "/"),
					resource.TestCheckResourceAttr(resourceName, "description", "This 1s a D3scr!pti0n with weird content: &@90ë\"'{«¡Çø}"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSIAMRoleConfigWithUpdatedDescription(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "path", "/"),
					resource.TestCheckResourceAttr(resourceName, "description", "This 1s an Upd@ted D3scr!pti0n with weird content: &90ë\"'{«¡Çø}"),
				),
			},
			{
				Config: testAccAWSIAMRoleConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists(resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "create_date"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
				),
			},
		},
	})
}

func TestAccAWSIAMRole_namePrefix(t *testing.T) {
	var conf iam.GetRoleOutput
	rName := acctest.RandString(10)
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:        func() { testAccPreCheck(t) },
		ErrorCheck:      testAccErrorCheck(t, iam.EndpointsID),
		IDRefreshName:   resourceName,
		IDRefreshIgnore: []string{"name_prefix"},
		Providers:       testAccProviders,
		CheckDestroy:    testAccCheckAWSRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIAMRolePrefixNameConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists(resourceName, &conf),
					testAccCheckAWSRoleGeneratedNamePrefix(
						resourceName, "test-role-"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
		},
	})
}

func TestAccAWSIAMRole_testNameChange(t *testing.T) {
	var conf iam.GetRoleOutput
	rName := acctest.RandString(10)
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, iam.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIAMRolePre(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists(resourceName, &conf),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSIAMRolePost(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists(resourceName, &conf),
				),
			},
		},
	})
}

func TestAccAWSIAMRole_badJSON(t *testing.T) {
	rName := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, iam.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSIAMRoleConfig_badJson(rName),
				ExpectError: regexp.MustCompile(`.*contains an invalid JSON:.*`),
			},
		},
	})
}

func TestAccAWSIAMRole_disappears(t *testing.T) {
	var role iam.GetRoleOutput

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, iam.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIAMRoleConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists(resourceName, &role),
					testAccCheckAWSRoleDisappears(&role),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSIAMRole_force_detach_policies(t *testing.T) {
	var conf iam.GetRoleOutput
	rName := acctest.RandString(10)
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, iam.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIAMRoleConfig_force_detach_policies(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists(resourceName, &conf),
					testAccAddAwsIAMRolePolicy(resourceName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_detach_policies"},
			},
		},
	})
}

func TestAccAWSIAMRole_MaxSessionDuration(t *testing.T) {
	var conf iam.GetRoleOutput
	rName := acctest.RandString(10)
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, iam.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckIAMRoleConfig_MaxSessionDuration(rName, 3599),
				ExpectError: regexp.MustCompile(`expected max_session_duration to be in the range`),
			},
			{
				Config:      testAccCheckIAMRoleConfig_MaxSessionDuration(rName, 43201),
				ExpectError: regexp.MustCompile(`expected max_session_duration to be in the range`),
			},
			{
				Config: testAccCheckIAMRoleConfig_MaxSessionDuration(rName, 3700),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "max_session_duration", "3700"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCheckIAMRoleConfig_MaxSessionDuration(rName, 3701),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "max_session_duration", "3701"),
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

func TestAccAWSIAMRole_PermissionsBoundary(t *testing.T) {
	var role iam.GetRoleOutput

	rName := acctest.RandString(10)
	resourceName := "aws_iam_role.test"

	permissionsBoundary1 := fmt.Sprintf("arn:%s:iam::aws:policy/AdministratorAccess", testAccGetPartition())
	permissionsBoundary2 := fmt.Sprintf("arn:%s:iam::aws:policy/ReadOnlyAccess", testAccGetPartition())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, iam.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSUserDestroy,
		Steps: []resource.TestStep{
			// Test creation
			{
				Config: testAccCheckIAMRoleConfig_PermissionsBoundary(rName, permissionsBoundary1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "permissions_boundary", permissionsBoundary1),
					testAccCheckAWSRolePermissionsBoundary(&role, permissionsBoundary1),
				),
			},
			// Test update
			{
				Config: testAccCheckIAMRoleConfig_PermissionsBoundary(rName, permissionsBoundary2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "permissions_boundary", permissionsBoundary2),
					testAccCheckAWSRolePermissionsBoundary(&role, permissionsBoundary2),
				),
			},
			// Test import
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy",
				},
			},
			// Test removal
			{
				Config: testAccAWSIAMRoleConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "permissions_boundary", ""),
					testAccCheckAWSRolePermissionsBoundary(&role, ""),
				),
			},
			// Test addition
			{
				Config: testAccCheckIAMRoleConfig_PermissionsBoundary(rName, permissionsBoundary1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "permissions_boundary", permissionsBoundary1),
					testAccCheckAWSRolePermissionsBoundary(&role, permissionsBoundary1),
				),
			},
			// Test empty value
			{
				Config: testAccCheckIAMRoleConfig_PermissionsBoundary(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "permissions_boundary", ""),
					testAccCheckAWSRolePermissionsBoundary(&role, ""),
				),
			},
		},
	})
}

func TestAccAWSIAMRole_tags(t *testing.T) {
	var role iam.GetRoleOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, iam.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIAMRoleConfig_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag1", "test-value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag2", "test-value2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSIAMRoleConfig_tagsUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag2", "test-value"),
				),
			},
		},
	})
}

func TestAccAWSIAMRole_policyBasicInline(t *testing.T) {
	var role iam.GetRoleOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	policyName1 := acctest.RandomWithPrefix("tf-acc-test")
	policyName2 := acctest.RandomWithPrefix("tf-acc-test")
	policyName3 := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, iam.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRolePolicyInlineConfig(rName, policyName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "inline_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "managed_policy_arns.#", "0"),
				),
			},
			{
				Config: testAccAWSRolePolicyInlineConfigUpdate(rName, policyName2, policyName3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "inline_policy.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "managed_policy_arns.#", "0"),
				),
			},
			{
				Config: testAccAWSRolePolicyInlineConfigUpdateDown(rName, policyName3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "inline_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "managed_policy_arns.#", "0"),
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

func TestAccAWSIAMRole_policyBasicInlineEmpty(t *testing.T) {
	var role iam.GetRoleOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, iam.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRolePolicyEmptyInlineConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists(resourceName, &role),
				),
			},
		},
	})
}

func TestAccAWSIAMRole_policyBasicManaged(t *testing.T) {
	var role iam.GetRoleOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	policyName1 := acctest.RandomWithPrefix("tf-acc-test")
	policyName2 := acctest.RandomWithPrefix("tf-acc-test")
	policyName3 := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, iam.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRolePolicyManagedConfig(rName, policyName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "managed_policy_arns.#", "1"),
				),
			},
			{
				Config: testAccAWSRolePolicyManagedConfigUpdate(rName, policyName1, policyName2, policyName3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "managed_policy_arns.#", "2"),
				),
			},
			{
				Config: testAccAWSRolePolicyManagedConfigUpdateDown(rName, policyName1, policyName2, policyName3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "managed_policy_arns.#", "1"),
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

// TestAccAWSIAMRole_policyOutOfBandRemovalAddedBack_managedNonEmpty: if a policy is detached
// out of band, it should be reattached.
func TestAccAWSIAMRole_policyOutOfBandRemovalAddedBack_managedNonEmpty(t *testing.T) {
	var role iam.GetRoleOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	policyName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, iam.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRolePolicyManagedConfig(rName, policyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists(resourceName, &role),
					testAccCheckAWSRolePolicyDetachManagedPolicy(&role, policyName),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccAWSRolePolicyManagedConfig(rName, policyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "managed_policy_arns.#", "1"),
				),
			},
		},
	})
}

// TestAccAWSIAMRole_policyOutOfBandRemovalAddedBack_inlineNonEmpty: if a policy is removed
// out of band, it should be recreated.
func TestAccAWSIAMRole_policyOutOfBandRemovalAddedBack_inlineNonEmpty(t *testing.T) {
	var role iam.GetRoleOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	policyName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, iam.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRolePolicyInlineConfig(rName, policyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists(resourceName, &role),
					testAccCheckAWSRolePolicyRemoveInlinePolicy(&role, policyName),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccAWSRolePolicyInlineConfig(rName, policyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "inline_policy.#", "1"),
				),
			},
		},
	})
}

// TestAccAWSIAMRole_policyOutOfBandAdditionRemoved_managedNonEmpty: if managed_policies arg
// exists and is non-empty, policy attached out of band should be removed
func TestAccAWSIAMRole_policyOutOfBandAdditionRemoved_managedNonEmpty(t *testing.T) {
	var role iam.GetRoleOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	policyName1 := acctest.RandomWithPrefix("tf-acc-test")
	policyName2 := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, iam.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRolePolicyExtraManagedConfig(rName, policyName1, policyName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists(resourceName, &role),
					testAccCheckAWSRolePolicyAttachManagedPolicy(&role, policyName2),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccAWSRolePolicyExtraManagedConfig(rName, policyName1, policyName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "managed_policy_arns.#", "1"),
				),
			},
		},
	})
}

// TestAccAWSIAMRole_policyOutOfBandAdditionRemoved_inlineNonEmpty: if inline_policy arg
// exists and is non-empty, policy added out of band should be removed
func TestAccAWSIAMRole_policyOutOfBandAdditionRemoved_inlineNonEmpty(t *testing.T) {
	var role iam.GetRoleOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	policyName1 := acctest.RandomWithPrefix("tf-acc-test")
	policyName2 := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, iam.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRolePolicyInlineConfig(rName, policyName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists(resourceName, &role),
					testAccCheckAWSRolePolicyAddInlinePolicy(&role, policyName2),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccAWSRolePolicyInlineConfig(rName, policyName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "inline_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "managed_policy_arns.#", "0"),
				),
			},
		},
	})
}

// TestAccAWSIAMRole_policyOutOfBandAdditionIgnored_inlineNonExistent: if there is no
// inline_policy attribute, out of band changes should be ignored.
func TestAccAWSIAMRole_policyOutOfBandAdditionIgnored_inlineNonExistent(t *testing.T) {
	var role iam.GetRoleOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	policyName1 := acctest.RandomWithPrefix("tf-acc-test")
	policyName2 := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, iam.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRolePolicyNoInlineConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists(resourceName, &role),
					testAccCheckAWSRolePolicyAddInlinePolicy(&role, policyName1),
				),
			},
			{
				Config: testAccAWSRolePolicyNoInlineConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists(resourceName, &role),
					testAccCheckAWSRolePolicyAddInlinePolicy(&role, policyName2),
				),
			},
			{
				Config: testAccAWSRolePolicyNoInlineConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists(resourceName, &role),
					testAccCheckAWSRolePolicyRemoveInlinePolicy(&role, policyName1),
					testAccCheckAWSRolePolicyRemoveInlinePolicy(&role, policyName2),
				),
			},
		},
	})
}

// TestAccAWSIAMRole_policyOutOfBandAdditionIgnored_managedNonExistent: if there is no
// managed_policies attribute, out of band changes should be ignored.
func TestAccAWSIAMRole_policyOutOfBandAdditionIgnored_managedNonExistent(t *testing.T) {
	var role iam.GetRoleOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	policyName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, iam.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRolePolicyNoManagedConfig(rName, policyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists(resourceName, &role),
					testAccCheckAWSRolePolicyAttachManagedPolicy(&role, policyName),
				),
			},
			{
				Config: testAccAWSRolePolicyNoManagedConfig(rName, policyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists(resourceName, &role),
					testAccCheckAWSRolePolicyDetachManagedPolicy(&role, policyName),
				),
			},
		},
	})
}

// TestAccAWSIAMRole_policyOutOfBandAdditionRemoved_inlineEmpty: if inline is added
// out of band with empty inline arg, should be removed
func TestAccAWSIAMRole_policyOutOfBandAdditionRemoved_inlineEmpty(t *testing.T) {
	var role iam.GetRoleOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	policyName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, iam.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRolePolicyEmptyInlineConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists(resourceName, &role),
					testAccCheckAWSRolePolicyAddInlinePolicy(&role, policyName),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccAWSRolePolicyEmptyInlineConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists(resourceName, &role),
				),
			},
		},
	})
}

// TestAccAWSIAMRole_policyOutOfBandAdditionRemoved_managedEmpty: if managed is attached
// out of band with empty managed arg, should be detached
func TestAccAWSIAMRole_policyOutOfBandAdditionRemoved_managedEmpty(t *testing.T) {
	var role iam.GetRoleOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	policyName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, iam.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRolePolicyEmptyManagedConfig(rName, policyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists(resourceName, &role),
					testAccCheckAWSRolePolicyAttachManagedPolicy(&role, policyName),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccAWSRolePolicyEmptyManagedConfig(rName, policyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists(resourceName, &role),
				),
			},
		},
	})
}

func testAccCheckAWSRoleDestroy(s *terraform.State) error {
	iamconn := testAccProvider.Meta().(*AWSClient).iamconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iam_role" {
			continue
		}

		// Try to get role
		_, err := iamconn.GetRole(&iam.GetRoleInput{
			RoleName: aws.String(rs.Primary.ID),
		})
		if err == nil {
			return fmt.Errorf("still exist.")
		}

		// Verify the error is what we want
		ec2err, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if ec2err.Code() != "NoSuchEntity" {
			return err
		}
	}

	return nil
}

func testAccCheckAWSRoleExists(n string, res *iam.GetRoleOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Role name is set")
		}

		iamconn := testAccProvider.Meta().(*AWSClient).iamconn

		resp, err := iamconn.GetRole(&iam.GetRoleInput{
			RoleName: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		*res = *resp

		return nil
	}
}

func testAccCheckAWSRoleDisappears(getRoleOutput *iam.GetRoleOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		iamconn := testAccProvider.Meta().(*AWSClient).iamconn

		roleName := aws.StringValue(getRoleOutput.Role.RoleName)

		_, err := iamconn.DeleteRole(&iam.DeleteRoleInput{
			RoleName: aws.String(roleName),
		})
		if err != nil {
			return fmt.Errorf("error deleting role %q: %s", roleName, err)
		}

		return nil
	}
}

func testAccCheckAWSRoleGeneratedNamePrefix(resource, prefix string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Resource not found")
		}
		name, ok := r.Primary.Attributes["name"]
		if !ok {
			return fmt.Errorf("Name attr not found: %#v", r.Primary.Attributes)
		}
		if !strings.HasPrefix(name, prefix) {
			return fmt.Errorf("Name: %q, does not have prefix: %q", name, prefix)
		}
		return nil
	}
}

// Attach inline policy out of band (outside of terraform)
func testAccAddAwsIAMRolePolicy(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Resource not found")
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Role name is set")
		}

		iamconn := testAccProvider.Meta().(*AWSClient).iamconn

		input := &iam.PutRolePolicyInput{
			RoleName: aws.String(rs.Primary.ID),
			PolicyDocument: aws.String(`{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  }
}`),
			PolicyName: aws.String(resource.UniqueId()),
		}

		_, err := iamconn.PutRolePolicy(input)
		return err
	}
}

func testAccCheckAWSRolePermissionsBoundary(getRoleOutput *iam.GetRoleOutput, expectedPermissionsBoundaryArn string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		actualPermissionsBoundaryArn := ""

		if getRoleOutput.Role.PermissionsBoundary != nil {
			actualPermissionsBoundaryArn = *getRoleOutput.Role.PermissionsBoundary.PermissionsBoundaryArn
		}

		if actualPermissionsBoundaryArn != expectedPermissionsBoundaryArn {
			return fmt.Errorf("PermissionsBoundary: '%q', expected '%q'.", actualPermissionsBoundaryArn, expectedPermissionsBoundaryArn)
		}

		return nil
	}
}

func testAccCheckAWSRolePolicyDetachManagedPolicy(role *iam.GetRoleOutput, policyName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).iamconn

		var managedARN string
		input := &iam.ListAttachedRolePoliciesInput{
			RoleName: role.Role.RoleName,
		}

		err := conn.ListAttachedRolePoliciesPages(input, func(page *iam.ListAttachedRolePoliciesOutput, lastPage bool) bool {
			for _, v := range page.AttachedPolicies {
				if *v.PolicyName == policyName {
					managedARN = *v.PolicyArn
					break
				}
			}
			return !lastPage
		})
		if err != nil && !tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			return fmt.Errorf("finding managed policy (%s): %w", policyName, err)
		}
		if managedARN == "" {
			return fmt.Errorf("managed policy (%s) not found", policyName)
		}

		_, err = conn.DetachRolePolicy(&iam.DetachRolePolicyInput{
			PolicyArn: aws.String(managedARN),
			RoleName:  role.Role.RoleName,
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckAWSRolePolicyAttachManagedPolicy(role *iam.GetRoleOutput, policyName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).iamconn

		var managedARN string
		input := &iam.ListPoliciesInput{
			PathPrefix:        aws.String("/tf-testing/"),
			PolicyUsageFilter: aws.String("PermissionsPolicy"),
			Scope:             aws.String("Local"),
		}

		err := conn.ListPoliciesPages(input, func(page *iam.ListPoliciesOutput, lastPage bool) bool {
			for _, v := range page.Policies {
				if *v.PolicyName == policyName {
					managedARN = *v.Arn
					break
				}
			}
			return !lastPage
		})
		if err != nil && !tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			return fmt.Errorf("finding managed policy (%s): %w", policyName, err)
		}
		if managedARN == "" {
			return fmt.Errorf("managed policy (%s) not found", policyName)
		}

		_, err = conn.AttachRolePolicy(&iam.AttachRolePolicyInput{
			PolicyArn: aws.String(managedARN),
			RoleName:  role.Role.RoleName,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckAWSRolePolicyAddInlinePolicy(role *iam.GetRoleOutput, inlinePolicy string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).iamconn

		_, err := conn.PutRolePolicy(&iam.PutRolePolicyInput{
			PolicyDocument: aws.String(testAccAWSRolePolicyExtraInlineConfig()),
			PolicyName:     aws.String(inlinePolicy),
			RoleName:       aws.String(*role.Role.RoleName),
		})

		if err != nil {
			return err
		}
		return nil
	}
}

func testAccCheckAWSRolePolicyRemoveInlinePolicy(role *iam.GetRoleOutput, inlinePolicy string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).iamconn

		_, err := conn.DeleteRolePolicy(&iam.DeleteRolePolicyInput{
			PolicyName: aws.String(inlinePolicy),
			RoleName:   aws.String(*role.Role.RoleName),
		})

		if err != nil {
			return err
		}
		return nil
	}
}

func testAccCheckIAMRoleConfig_MaxSessionDuration(rName string, maxSessionDuration int) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name                 = "test-role-%s"
  path                 = "/"
  max_session_duration = %d

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.${data.aws_partition.current.dns_suffix}"
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
`, rName, maxSessionDuration)
}

func testAccCheckIAMRoleConfig_PermissionsBoundary(rName, permissionsBoundary string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.${data.aws_partition.current.dns_suffix}"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF

  name                 = "test-role-%s"
  path                 = "/"
  permissions_boundary = %q
}
`, rName, permissionsBoundary)
}

func testAccAWSIAMRoleConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = "test-role-%s"
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.${data.aws_partition.current.dns_suffix}"
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
`, rName)
}

func testAccAWSIAMRoleConfigWithDescription(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name        = "test-role-%s"
  description = "This 1s a D3scr!pti0n with weird content: &@90ë\"'{«¡Çø}"
  path        = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.${data.aws_partition.current.dns_suffix}"
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
`, rName)
}

func testAccAWSIAMRoleConfigWithUpdatedDescription(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name        = "test-role-%s"
  description = "This 1s an Upd@ted D3scr!pti0n with weird content: &90ë\"'{«¡Çø}"
  path        = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.${data.aws_partition.current.dns_suffix}"
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
`, rName)
}

func testAccAWSIAMRolePrefixNameConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name_prefix = "test-role-%s"
  path        = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.${data.aws_partition.current.dns_suffix}"
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
`, rName)
}

func testAccAWSIAMRolePre(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = "tf_old_name_%[1]s"
  path = "/test/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "role_update_test" {
  name = "role_update_test_%[1]s"
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetBucketLocation",
        "s3:ListAllMyBuckets"
      ],
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::*"
    }
  ]
}
EOF
}

resource "aws_iam_instance_profile" "role_update_test" {
  name = "role_update_test_%[1]s"
  path = "/test/"
  role = aws_iam_role.test.name
}
`, rName)
}

func testAccAWSIAMRolePost(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = "tf_new_name_%[1]s"
  path = "/test/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "role_update_test" {
  name = "role_update_test_%[1]s"
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetBucketLocation",
        "s3:ListAllMyBuckets"
      ],
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::*"
    }
  ]
}
EOF
}

resource "aws_iam_instance_profile" "role_update_test" {
  name = "role_update_test_%[1]s"
  path = "/test/"
  role = aws_iam_role.test.name
}
`, rName)
}

func testAccAWSIAMRoleConfig_badJson(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = "test-role-%s"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
  {
    "Action": "sts:AssumeRole",
    "Principal": {
    "Service": "ec2.${data.aws_partition.current.dns_suffix}",
    },
    "Effect": "Allow",
    "Sid": ""
  }
  ]
}
POLICY
}
`, rName)
}

func testAccAWSIAMRoleConfig_force_detach_policies(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role_policy" "test" {
  name = "tf-iam-role-policy-%[1]s"
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "ec2:Describe*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_policy" "test" {
  name        = "tf-iam-policy-%[1]s"
  description = "A test policy"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "iam:ChangePassword"
      ],
      "Resource": "*",
      "Effect": "Allow"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = aws_iam_policy.test.arn
}

resource "aws_iam_role" "test" {
  name                  = "tf-iam-role-%[1]s"
  force_detach_policies = true

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}
`, rName)
}

func testAccAWSIAMRoleConfig_tags(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF

  tags = {
    tag1 = "test-value1"
    tag2 = "test-value2"
  }
}
`, rName)
}

func testAccAWSIAMRoleConfig_tagsUpdate(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF

  tags = {
    tag2 = "test-value"
  }
}
`, rName)
}

func testAccAWSRolePolicyInlineConfig(roleName, policyName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF

  inline_policy {
    name = %[2]q

    policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "ec2:Describe*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
  }
}
`, roleName, policyName)
}

func testAccAWSRolePolicyInlineConfigUpdate(roleName, policyName2, policyName3 string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF

  inline_policy {
    name = %[2]q

    policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "ec2:Describe*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
  }

  inline_policy {
    name = %[3]q

    policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "ec2:Describe*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
  }
}
`, roleName, policyName2, policyName3)
}

func testAccAWSRolePolicyInlineConfigUpdateDown(roleName, policyName3 string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF

  inline_policy {
    name = %[2]q

    policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "ec2:Describe*",
    "Resource": "*",
    "Condition": {
      "DateGreaterThan": {"aws:CurrentTime": "2017-07-01T00:00:00Z"},
      "DateLessThan": {"aws:CurrentTime": "2017-12-31T23:59:59Z"}
    }
  }
}
EOF
  }
}
`, roleName, policyName3)
}

func testAccAWSRolePolicyManagedConfig(roleName, policyName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_policy" "test" {
  name = %[1]q
  path = "/tf-testing/"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
    "Action": [
      "ec2:Describe*"
    ],
    "Effect": "Allow",
    "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_role" "test" {
  name                = %[2]q
  managed_policy_arns = [aws_iam_policy.test.arn]

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}
`, policyName, roleName)
}

func testAccAWSRolePolicyManagedConfigUpdate(roleName, policyName1, policyName2, policyName3 string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_policy" "test" {
  name = %[1]q
  path = "/tf-testing/"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
    "Action": [
      "ec2:Describe*"
    ],
    "Effect": "Allow",
    "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_policy" "test2" {
  name = %[2]q
  path = "/tf-testing/"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
    "Action": [
      "ec2:Describe*"
    ],
    "Effect": "Allow",
    "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_policy" "test3" {
  name = %[3]q
  path = "/tf-testing/"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
    "Action": [
      "ec2:Describe*"
    ],
    "Effect": "Allow",
    "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_role" "test" {
  name                = %[4]q
  managed_policy_arns = [aws_iam_policy.test2.arn, aws_iam_policy.test3.arn]

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}
`, policyName1, policyName2, policyName3, roleName)
}

func testAccAWSRolePolicyManagedConfigUpdateDown(roleName, policyName1, policyName2, policyName3 string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_policy" "test" {
  name = %[1]q
  path = "/tf-testing/"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
    "Action": [
      "ec2:Describe*"
    ],
    "Effect": "Allow",
    "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_policy" "test2" {
  name = %[2]q
  path = "/tf-testing/"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
    "Action": [
      "ec2:Describe*"
    ],
    "Effect": "Allow",
    "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_policy" "test3" {
  name = %[3]q
  path = "/tf-testing/"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
    "Action": [
      "ec2:Describe*"
    ],
    "Effect": "Allow",
    "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_role" "test" {
  name                = %[4]q
  managed_policy_arns = [aws_iam_policy.test3.arn]

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}
`, policyName1, policyName2, policyName3, roleName)
}

func testAccAWSRolePolicyExtraManagedConfig(roleName, policyName1, policyName2 string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_policy" "test" {
  name = %[1]q
  path = "/tf-testing/"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
    "Action": [
      "ec2:Describe*"
    ],
    "Effect": "Allow",
    "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_policy" "test2" {
  name = %[2]q
  path = "/tf-testing/"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
    "Action": [
      "ec2:Describe*"
    ],
    "Effect": "Allow",
    "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_role" "test" {
  name                = %[3]q
  managed_policy_arns = [aws_iam_policy.test.arn]

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}
`, policyName1, policyName2, roleName)
}

func testAccAWSRolePolicyExtraInlineConfig() string {
	return `{
	"Version": "2012-10-17",
	"Statement": [
		{
		"Action": [
			"ec2:Describe*"
		],
		"Effect": "Allow",
		"Resource": "*"
		}
	]
}`
}

func testAccAWSRolePolicyNoInlineConfig(roleName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}
`, roleName)
}

func testAccAWSRolePolicyNoManagedConfig(roleName, policyName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_policy" "managed-policy1" {
  name = %[2]q
  path = "/tf-testing/"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
    "Action": [
      "ec2:Describe*"
    ],
    "Effect": "Allow",
    "Resource": "*"
    }
  ]
}
EOF
}
`, roleName, policyName)
}

func testAccAWSRolePolicyEmptyInlineConfig(roleName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF

  inline_policy {}
}
`, roleName)
}

func testAccAWSRolePolicyEmptyManagedConfig(roleName, policyName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF

  managed_policy_arns = []
}

resource "aws_iam_policy" "managed-policy1" {
  name = %[2]q
  path = "/tf-testing/"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
    "Action": [
      "ec2:Describe*"
    ],
    "Effect": "Allow",
    "Resource": "*"
    }
  ]
}
EOF
}
`, roleName, policyName)
}
