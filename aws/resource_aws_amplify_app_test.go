package aws

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/amplify"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/amplify/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/amplify/lister"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func init() {
	resource.AddTestSweepers("aws_amplify_app", &resource.Sweeper{
		Name: "aws_amplify_app",
		F:    testSweepAmplifyApps,
	})
}

func testSweepAmplifyApps(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).amplifyconn
	input := &amplify.ListAppsInput{}
	var sweeperErrs *multierror.Error

	err = lister.ListAppsPages(conn, input, func(page *amplify.ListAppsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, app := range page.Apps {
			r := resourceAwsAmplifyApp()
			d := r.Data(nil)
			d.SetId(aws.StringValue(app.AppId))
			err = r.Delete(d, client)

			if err != nil {
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping Amplify Apps sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Amplify Apps: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func testAccAWSAmplifyApp_basic(t *testing.T) {
	var app amplify.App
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_amplify_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSAmplify(t) },
		ErrorCheck:   testAccErrorCheck(t, amplify.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAmplifyAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAmplifyAppConfigName(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSAmplifyAppExists(resourceName, &app),
					resource.TestCheckNoResourceAttr(resourceName, "access_token"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "amplify", regexp.MustCompile(`apps/.+`)),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_patterns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "basic_auth_credentials", ""),
					resource.TestCheckResourceAttr(resourceName, "build_spec", ""),
					resource.TestCheckResourceAttr(resourceName, "custom_rule.#", "0"),
					resource.TestMatchResourceAttr(resourceName, "default_domain", regexp.MustCompile(`\.amplifyapp\.com$`)),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "enable_auto_branch_creation", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_basic_auth", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_branch_auto_build", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_branch_auto_deletion", "false"),
					resource.TestCheckResourceAttr(resourceName, "environment_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "iam_service_role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckNoResourceAttr(resourceName, "oauth_token"),
					resource.TestCheckResourceAttr(resourceName, "platform", "WEB"),
					resource.TestCheckResourceAttr(resourceName, "production_branch.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "repository", ""),
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

func testAccAWSAmplifyApp_disappears(t *testing.T) {
	var app amplify.App
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_amplify_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSAmplify(t) },
		ErrorCheck:   testAccErrorCheck(t, amplify.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAmplifyAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAmplifyAppConfigName(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSAmplifyAppExists(resourceName, &app),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsAmplifyApp(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAWSAmplifyApp_Tags(t *testing.T) {
	var app amplify.App
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_amplify_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSAmplify(t) },
		ErrorCheck:   testAccErrorCheck(t, amplify.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAmplifyAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAmplifyAppConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyAppExists(resourceName, &app),
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
				Config: testAccAWSAmplifyAppConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyAppExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSAmplifyAppConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyAppExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccAWSAmplifyApp_AutoBranchCreationConfig(t *testing.T) {
	var app amplify.App
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_amplify_app.test"

	credentials := base64.StdEncoding.EncodeToString([]byte("username1:password1"))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSAmplify(t) },
		ErrorCheck:   testAccErrorCheck(t, amplify.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAmplifyAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAmplifyAppConfigAutoBranchCreationConfigNoAutoBranchCreationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyAppExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.basic_auth_credentials", ""),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.build_spec", ""),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.enable_auto_build", "false"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.enable_basic_auth", "false"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.enable_performance_mode", "false"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.enable_pull_request_preview", "false"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.environment_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.framework", ""),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.pull_request_environment_name", ""),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.stage", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_patterns.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_patterns.0", "*"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_patterns.1", "*/**"),
					resource.TestCheckResourceAttr(resourceName, "enable_auto_branch_creation", "true"),
				),
			},
			{
				Config: testAccAWSAmplifyAppConfigAutoBranchCreationConfigAutoBranchCreationConfig(rName, credentials),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyAppExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.basic_auth_credentials", credentials),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.build_spec", "version: 0.1"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.enable_auto_build", "true"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.enable_basic_auth", "true"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.enable_performance_mode", "false"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.enable_pull_request_preview", "true"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.environment_variables.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.environment_variables.ENVVAR1", "1"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.framework", "React"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.pull_request_environment_name", "test1"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.stage", "DEVELOPMENT"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_patterns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_patterns.0", "feature/*"),
					resource.TestCheckResourceAttr(resourceName, "enable_auto_branch_creation", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAmplifyAppConfigAutoBranchCreationConfigAutoBranchCreationConfigUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyAppExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.#", "1"),
					// Clearing basic_auth_credentials not reflected in API.
					// resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.basic_auth_credentials", ""),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.build_spec", "version: 0.2"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.enable_auto_build", "false"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.enable_basic_auth", "false"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.enable_performance_mode", "false"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.enable_pull_request_preview", "false"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.environment_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.framework", "React"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.pull_request_environment_name", "test2"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.stage", "EXPERIMENTAL"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_patterns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_patterns.0", "feature/*"),
					resource.TestCheckResourceAttr(resourceName, "enable_auto_branch_creation", "true"),
				),
			},
			{
				Config: testAccAWSAmplifyAppConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyAppExists(resourceName, &app),
					// No change is reflected in API.
					// resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.#", "0"),
					// resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_patterns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "enable_auto_branch_creation", "false"),
				),
			},
		},
	})
}

func testAccAWSAmplifyApp_BasicAuthCredentials(t *testing.T) {
	var app amplify.App
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_amplify_app.test"

	credentials1 := base64.StdEncoding.EncodeToString([]byte("username1:password1"))
	credentials2 := base64.StdEncoding.EncodeToString([]byte("username2:password2"))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSAmplify(t) },
		ErrorCheck:   testAccErrorCheck(t, amplify.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAmplifyAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAmplifyAppConfigBasicAuthCredentials(rName, credentials1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyAppExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "basic_auth_credentials", credentials1),
					resource.TestCheckResourceAttr(resourceName, "enable_basic_auth", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAmplifyAppConfigBasicAuthCredentials(rName, credentials2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyAppExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "basic_auth_credentials", credentials2),
					resource.TestCheckResourceAttr(resourceName, "enable_basic_auth", "true"),
				),
			},
			{
				Config: testAccAWSAmplifyAppConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyAppExists(resourceName, &app),
					// Clearing basic_auth_credentials not reflected in API.
					// resource.TestCheckResourceAttr(resourceName, "basic_auth_credentials", ""),
					resource.TestCheckResourceAttr(resourceName, "enable_basic_auth", "false"),
				),
			},
		},
	})
}

func testAccAWSAmplifyApp_BuildSpec(t *testing.T) {
	var app amplify.App
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_amplify_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSAmplify(t) },
		ErrorCheck:   testAccErrorCheck(t, amplify.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAmplifyAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAmplifyAppConfigBuildSpec(rName, "version: 0.1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyAppExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "build_spec", "version: 0.1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAmplifyAppConfigBuildSpec(rName, "version: 0.2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyAppExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "build_spec", "version: 0.2"),
				),
			},
			{
				Config: testAccAWSAmplifyAppConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyAppExists(resourceName, &app),
					// build_spec is Computed.
					resource.TestCheckResourceAttr(resourceName, "build_spec", "version: 0.2"),
				),
			},
		},
	})
}

func testAccAWSAmplifyApp_CustomRules(t *testing.T) {
	var app amplify.App
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_amplify_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSAmplify(t) },
		ErrorCheck:   testAccErrorCheck(t, amplify.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAmplifyAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAmplifyAppConfigCustomRules(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyAppExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "custom_rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "custom_rule.0.source", "/<*>"),
					resource.TestCheckResourceAttr(resourceName, "custom_rule.0.status", "404"),
					resource.TestCheckResourceAttr(resourceName, "custom_rule.0.target", "/index.html"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAmplifyAppConfigCustomRulesUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "custom_rule.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "custom_rule.0.condition", "<US>"),
					resource.TestCheckResourceAttr(resourceName, "custom_rule.0.source", "/documents"),
					resource.TestCheckResourceAttr(resourceName, "custom_rule.0.status", "302"),
					resource.TestCheckResourceAttr(resourceName, "custom_rule.0.target", "/documents/us"),
					resource.TestCheckResourceAttr(resourceName, "custom_rule.1.source", "/<*>"),
					resource.TestCheckResourceAttr(resourceName, "custom_rule.1.status", "200"),
					resource.TestCheckResourceAttr(resourceName, "custom_rule.1.target", "/index.html"),
				),
			},
			{
				Config: testAccAWSAmplifyAppConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyAppExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "custom_rule.#", "0"),
				),
			},
		},
	})
}

func testAccAWSAmplifyApp_Description(t *testing.T) {
	var app1, app2, app3 amplify.App
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_amplify_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSAmplify(t) },
		ErrorCheck:   testAccErrorCheck(t, amplify.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAmplifyAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAmplifyAppConfigDescription(rName, "description 1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyAppExists(resourceName, &app1),
					resource.TestCheckResourceAttr(resourceName, "description", "description 1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAmplifyAppConfigDescription(rName, "description 2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyAppExists(resourceName, &app2),
					testAccCheckAWSAmplifyAppNotRecreated(&app1, &app2),
					resource.TestCheckResourceAttr(resourceName, "description", "description 2"),
				),
			},
			{
				Config: testAccAWSAmplifyAppConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyAppExists(resourceName, &app3),
					testAccCheckAWSAmplifyAppRecreated(&app2, &app3),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
				),
			},
		},
	})
}

func testAccAWSAmplifyApp_EnvironmentVariables(t *testing.T) {
	var app amplify.App
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_amplify_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSAmplify(t) },
		ErrorCheck:   testAccErrorCheck(t, amplify.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAmplifyAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAmplifyAppConfigEnvironmentVariables(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyAppExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "environment_variables.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "environment_variables.ENVVAR1", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAmplifyAppConfigEnvironmentVariablesUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyAppExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "environment_variables.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "environment_variables.ENVVAR1", "2"),
					resource.TestCheckResourceAttr(resourceName, "environment_variables.ENVVAR2", "2"),
				),
			},
			{
				Config: testAccAWSAmplifyAppConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyAppExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "environment_variables.%", "0"),
				),
			},
		},
	})
}

func testAccAWSAmplifyApp_IamServiceRole(t *testing.T) {
	var app1, app2, app3 amplify.App
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_amplify_app.test"
	iamRole1ResourceName := "aws_iam_role.test1"
	iamRole2ResourceName := "aws_iam_role.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSAmplify(t) },
		ErrorCheck:   testAccErrorCheck(t, amplify.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAmplifyAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAmplifyAppConfigIAMServiceRoleArn(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyAppExists(resourceName, &app1),
					resource.TestCheckResourceAttrPair(resourceName, "iam_service_role_arn", iamRole1ResourceName, "arn")),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAmplifyAppConfigIAMServiceRoleArnUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyAppExists(resourceName, &app2),
					testAccCheckAWSAmplifyAppNotRecreated(&app1, &app2),
					resource.TestCheckResourceAttrPair(resourceName, "iam_service_role_arn", iamRole2ResourceName, "arn"),
				),
			},
			{
				Config: testAccAWSAmplifyAppConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyAppExists(resourceName, &app3),
					testAccCheckAWSAmplifyAppRecreated(&app2, &app3),
					resource.TestCheckResourceAttr(resourceName, "iam_service_role_arn", ""),
				),
			},
		},
	})
}

func testAccAWSAmplifyApp_Name(t *testing.T) {
	var app amplify.App
	rName1 := acctest.RandomWithPrefix("tf-acc-test")
	rName2 := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_amplify_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSAmplify(t) },
		ErrorCheck:   testAccErrorCheck(t, amplify.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAmplifyAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAmplifyAppConfigName(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyAppExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAmplifyAppConfigName(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyAppExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
				),
			},
		},
	})
}

func testAccAWSAmplifyApp_Repository(t *testing.T) {
	key := "AMPLIFY_GITHUB_ACCESS_TOKEN"
	accessToken := os.Getenv(key)
	if accessToken == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	key = "AMPLIFY_GITHUB_REPOSITORY"
	repository := os.Getenv(key)
	if repository == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var app amplify.App
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_amplify_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSAmplify(t) },
		ErrorCheck:   testAccErrorCheck(t, amplify.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAmplifyAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAmplifyAppConfigRepository(rName, repository, accessToken),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyAppExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "access_token", accessToken),
					resource.TestCheckResourceAttr(resourceName, "repository", repository),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// access_token is ignored because AWS does not store access_token and oauth_token
				// See https://docs.aws.amazon.com/sdk-for-go/api/service/amplify/#CreateAppInput
				ImportStateVerifyIgnore: []string{"access_token"},
			},
		},
	})
}

func testAccCheckAWSAmplifyAppExists(n string, v *amplify.App) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Amplify App ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).amplifyconn

		output, err := finder.AppByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckAWSAmplifyAppDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).amplifyconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_amplify_app" {
			continue
		}

		_, err := finder.AppByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Amplify App %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccPreCheckAWSAmplify(t *testing.T) {
	if testAccGetPartition() == "aws-us-gov" {
		t.Skip("AWS Amplify is not supported in GovCloud partition")
	}
}

func testAccCheckAWSAmplifyAppNotRecreated(before, after *amplify.App) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.StringValue(before.AppId), aws.StringValue(after.AppId); before != after {
			return fmt.Errorf("Amplify App (%s/%s) recreated", before, after)
		}

		return nil
	}
}

func testAccCheckAWSAmplifyAppRecreated(before, after *amplify.App) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.StringValue(before.AppId), aws.StringValue(after.AppId); before == after {
			return fmt.Errorf("Amplify App (%s) not recreated", before)
		}

		return nil
	}
}

func testAccAWSAmplifyAppConfigName(rName string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q
}
`, rName)
}

func testAccAWSAmplifyAppConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSAmplifyAppConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAWSAmplifyAppConfigAutoBranchCreationConfigNoAutoBranchCreationConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q

  enable_auto_branch_creation = true

  auto_branch_creation_patterns = [
    "*",
    "*/**",
  ]
}
`, rName)
}

func testAccAWSAmplifyAppConfigAutoBranchCreationConfigAutoBranchCreationConfig(rName, basicAuthCredentials string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q

  enable_auto_branch_creation = true

  auto_branch_creation_patterns = [
    "feature/*",
  ]

  auto_branch_creation_config {
    build_spec = "version: 0.1"
    framework  = "React"
    stage      = "DEVELOPMENT"

    enable_basic_auth      = true
    basic_auth_credentials = %[2]q

    enable_auto_build             = true
    enable_pull_request_preview   = true
    pull_request_environment_name = "test1"

    environment_variables = {
      ENVVAR1 = "1"
    }
  }
}

`, rName, basicAuthCredentials)
}

func testAccAWSAmplifyAppConfigAutoBranchCreationConfigAutoBranchCreationConfigUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q

  enable_auto_branch_creation = true

  auto_branch_creation_patterns = [
    "feature/*",
  ]

  auto_branch_creation_config {
    build_spec = "version: 0.2"
    framework  = "React"
    stage      = "EXPERIMENTAL"

    enable_basic_auth = false

    enable_auto_build           = false
    enable_pull_request_preview = false

    pull_request_environment_name = "test2"
  }
}
`, rName)
}

func testAccAWSAmplifyAppConfigBasicAuthCredentials(rName, basicAuthCredentials string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q

  basic_auth_credentials = %[2]q
  enable_basic_auth      = true
}
`, rName, basicAuthCredentials)
}

func testAccAWSAmplifyAppConfigBuildSpec(rName, buildSpec string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q

  build_spec = %[2]q
}
`, rName, buildSpec)
}

func testAccAWSAmplifyAppConfigCustomRules(rName string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q

  custom_rule {
    source = "/<*>"
    status = "404"
    target = "/index.html"
  }
}
`, rName)
}

func testAccAWSAmplifyAppConfigCustomRulesUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q

  custom_rule {
    condition = "<US>"
    source    = "/documents"
    status    = "302"
    target    = "/documents/us"
  }

  custom_rule {
    source = "/<*>"
    status = "200"
    target = "/index.html"
  }
}
`, rName)
}

func testAccAWSAmplifyAppConfigDescription(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q

  description = %[2]q
}
`, rName, description)
}

func testAccAWSAmplifyAppConfigEnvironmentVariables(rName string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q

  environment_variables = {
    ENVVAR1 = "1"
  }
}
`, rName)
}

func testAccAWSAmplifyAppConfigEnvironmentVariablesUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q

  environment_variables = {
    ENVVAR1 = "2",
    ENVVAR2 = "2"
  }
}
`, rName)
}

func testAccAWSAmplifyAppConfigIAMServiceRoleBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test1" {
  name = "%[1]s-1"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Principal": {
      "Service": "amplify.amazonaws.com"
    },
    "Effect": "Allow"
  }]
}
POLICY
}

resource "aws_iam_role" "test2" {
  name = "%[1]s-2"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Principal": {
      "Service": "amplify.amazonaws.com"
    },
    "Effect": "Allow"
  }]
}
POLICY
}
`, rName)
}

func testAccAWSAmplifyAppConfigIAMServiceRoleArn(rName string) string {
	return composeConfig(testAccAWSAmplifyAppConfigIAMServiceRoleBase(rName), fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q

  iam_service_role_arn = aws_iam_role.test1.arn
}
`, rName))
}

func testAccAWSAmplifyAppConfigIAMServiceRoleArnUpdated(rName string) string {
	return composeConfig(testAccAWSAmplifyAppConfigIAMServiceRoleBase(rName), fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q

  iam_service_role_arn = aws_iam_role.test2.arn
}
`, rName))
}

func testAccAWSAmplifyAppConfigRepository(rName, repository, accessToken string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q

  repository   = %[2]q
  access_token = %[3]q
}
`, rName, repository, accessToken)
}
