package aws

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/elasticbeanstalk"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

// initialize sweeper
func init() {
	resource.AddTestSweepers("aws_beanstalk_environment", &resource.Sweeper{
		Name: "aws_beanstalk_environment",
		F:    testSweepBeanstalkEnvironments,
	})
}

func testSweepBeanstalkEnvironments(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	beanstalkconn := client.(*AWSClient).elasticbeanstalkconn

	resp, err := beanstalkconn.DescribeEnvironments(&elasticbeanstalk.DescribeEnvironmentsInput{
		IncludeDeleted: aws.Bool(false),
	})

	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Elastic Beanstalk Environment sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving beanstalk environment: %s", err)
	}

	if len(resp.Environments) == 0 {
		log.Print("[DEBUG] No aws beanstalk environments to sweep")
		return nil
	}

	for _, bse := range resp.Environments {
		var testOptGroup bool
		for _, testName := range []string{
			"terraform-",
			"tf-test-",
			"tf_acc_",
			"tf-acc-",
		} {
			if strings.HasPrefix(*bse.EnvironmentName, testName) {
				testOptGroup = true
			}
		}

		if !testOptGroup {
			log.Printf("Skipping (%s) (%s)", *bse.EnvironmentName, *bse.EnvironmentId)
			continue
		}

		log.Printf("Trying to terminate (%s) (%s)", *bse.EnvironmentName, *bse.EnvironmentId)

		_, err := beanstalkconn.TerminateEnvironment(
			&elasticbeanstalk.TerminateEnvironmentInput{
				EnvironmentId:      bse.EnvironmentId,
				TerminateResources: aws.Bool(true),
			})

		if err != nil {
			elasticbeanstalkerr, ok := err.(awserr.Error)
			if ok && (elasticbeanstalkerr.Code() == "InvalidConfiguration.NotFound" || elasticbeanstalkerr.Code() == "ValidationError") {
				log.Printf("[DEBUG] beanstalk environment (%s) not found", *bse.EnvironmentName)
				return nil
			}

			return err
		}

		waitForReadyTimeOut, _ := time.ParseDuration("5m")
		pollInterval, _ := time.ParseDuration("10s")

		// poll for deletion
		t := time.Now()
		stateConf := &resource.StateChangeConf{
			Pending:      []string{"Terminating"},
			Target:       []string{"Terminated"},
			Refresh:      environmentStateRefreshFunc(beanstalkconn, *bse.EnvironmentId, t),
			Timeout:      waitForReadyTimeOut,
			Delay:        10 * time.Second,
			PollInterval: pollInterval,
			MinTimeout:   3 * time.Second,
		}

		_, err = stateConf.WaitForState()
		if err != nil {
			return fmt.Errorf(
				"Error waiting for Elastic Beanstalk Environment (%s) to become terminated: %s",
				*bse.EnvironmentId, err)
		}
		log.Printf("> Terminated (%s) (%s)", *bse.EnvironmentName, *bse.EnvironmentId)
	}

	return nil
}

func TestAWSElasticBeanstalkEnvironment_importBasic(t *testing.T) {
	resourceName := "aws_elastic_beanstalk_application.tftest"

	applicationName := fmt.Sprintf("tf-test-name-%d", acctest.RandInt())
	environmentName := fmt.Sprintf("tf-test-env-name-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBeanstalkAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBeanstalkEnvImportConfig(applicationName, environmentName),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccBeanstalkEnvImportConfig(appName, envName string) string {
	return fmt.Sprintf(`resource "aws_elastic_beanstalk_application" "tftest" {
	  name = "%s"
	  description = "tf-test-desc"
	}

	resource "aws_elastic_beanstalk_environment" "tfenvtest" {
	  name = "%s"
	  application = "${aws_elastic_beanstalk_application.tftest.name}"
	  solution_stack_name = "64bit Amazon Linux running Python"
	}`, appName, envName)
}

func TestAccAWSBeanstalkEnv_basic(t *testing.T) {
	var app elasticbeanstalk.EnvironmentDescription

	rString := acctest.RandString(8)
	appName := fmt.Sprintf("tf_acc_app_env_basic_%s", rString)
	envName := fmt.Sprintf("tf-acc-env-basic-%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBeanstalkEnvDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBeanstalkEnvConfig(appName, envName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBeanstalkEnvExists("aws_elastic_beanstalk_environment.tfenvtest", &app),
					resource.TestMatchResourceAttr(
						"aws_elastic_beanstalk_environment.tfenvtest", "arn",
						regexp.MustCompile(fmt.Sprintf("^arn:[^:]+:elasticbeanstalk:[^:]+:[^:]+:environment/%s/%s$", appName, envName))),
				),
			},
		},
	})
}

func TestAccAWSBeanstalkEnv_tier(t *testing.T) {
	var app elasticbeanstalk.EnvironmentDescription
	beanstalkQueuesNameRegexp := regexp.MustCompile("https://sqs.+?awseb[^,]+")

	rString := acctest.RandString(8)
	instanceProfileName := fmt.Sprintf("tf_acc_profile_beanstalk_env_tier_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_beanstalk_env_tier_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_beanstalk_env_tier_%s", rString)
	appName := fmt.Sprintf("tf_acc_app_env_tier_%s", rString)
	envName := fmt.Sprintf("tf-acc-env-tier-%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBeanstalkEnvDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBeanstalkWorkerEnvConfig(instanceProfileName, roleName, policyName, appName, envName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBeanstalkEnvTier("aws_elastic_beanstalk_environment.tfenvtest", &app),
					resource.TestMatchResourceAttr(
						"aws_elastic_beanstalk_environment.tfenvtest", "queues.0", beanstalkQueuesNameRegexp),
				),
			},
		},
	})
}

func TestAccAWSBeanstalkEnv_outputs(t *testing.T) {
	var app elasticbeanstalk.EnvironmentDescription

	beanstalkAsgNameRegexp := regexp.MustCompile("awseb.+?AutoScalingGroup[^,]+")
	beanstalkElbNameRegexp := regexp.MustCompile("awseb.+?EBLoa[^,]+")
	beanstalkInstancesNameRegexp := regexp.MustCompile("i-([0-9a-fA-F]{8}|[0-9a-fA-F]{17})")
	beanstalkLcNameRegexp := regexp.MustCompile("awseb.+?AutoScalingLaunch[^,]+")

	rString := acctest.RandString(8)
	appName := fmt.Sprintf("tf_acc_app_env_outputs_%s", rString)
	envName := fmt.Sprintf("tf-acc-env-outputs-%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBeanstalkEnvDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBeanstalkEnvConfig(appName, envName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBeanstalkEnvExists("aws_elastic_beanstalk_environment.tfenvtest", &app),
					resource.TestMatchResourceAttr(
						"aws_elastic_beanstalk_environment.tfenvtest", "autoscaling_groups.0", beanstalkAsgNameRegexp),
					resource.TestMatchResourceAttr(
						"aws_elastic_beanstalk_environment.tfenvtest", "load_balancers.0", beanstalkElbNameRegexp),
					resource.TestMatchResourceAttr(
						"aws_elastic_beanstalk_environment.tfenvtest", "instances.0", beanstalkInstancesNameRegexp),
					resource.TestMatchResourceAttr(
						"aws_elastic_beanstalk_environment.tfenvtest", "launch_configurations.0", beanstalkLcNameRegexp),
				),
			},
		},
	})
}

func TestAccAWSBeanstalkEnv_cname_prefix(t *testing.T) {
	var app elasticbeanstalk.EnvironmentDescription

	rString := acctest.RandString(8)
	appName := fmt.Sprintf("tf_acc_app_env_cname_prefix_%s", rString)
	envName := fmt.Sprintf("tf-acc-env-cname-prefix-%s", rString)
	cnamePrefix := fmt.Sprintf("tf-acc-cname-%s", rString)

	beanstalkCnameRegexp := regexp.MustCompile("^" + cnamePrefix + ".+?elasticbeanstalk.com$")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBeanstalkEnvDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBeanstalkEnvCnamePrefixConfig(appName, envName, cnamePrefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBeanstalkEnvExists("aws_elastic_beanstalk_environment.tfenvtest", &app),
					resource.TestMatchResourceAttr(
						"aws_elastic_beanstalk_environment.tfenvtest", "cname", beanstalkCnameRegexp),
				),
			},
		},
	})
}

func TestAccAWSBeanstalkEnv_config(t *testing.T) {
	var app elasticbeanstalk.EnvironmentDescription

	rString := acctest.RandString(8)
	appName := fmt.Sprintf("tf_acc_app_env_config_%s", rString)
	envName := fmt.Sprintf("tf-acc-env-config-%s", rString)
	cfgTplName := fmt.Sprintf("tf_acc_cfg_tpl_config_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBeanstalkEnvDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBeanstalkConfigTemplate(appName, envName, cfgTplName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBeanstalkEnvExists("aws_elastic_beanstalk_environment.tftest", &app),
					testAccCheckBeanstalkEnvConfigValue("aws_elastic_beanstalk_environment.tftest", "1"),
				),
			},

			{
				Config: testAccBeanstalkConfigTemplate(appName, envName, cfgTplName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBeanstalkEnvExists("aws_elastic_beanstalk_environment.tftest", &app),
					testAccCheckBeanstalkEnvConfigValue("aws_elastic_beanstalk_environment.tftest", "2"),
				),
			},

			{
				Config: testAccBeanstalkConfigTemplate(appName, envName, cfgTplName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBeanstalkEnvExists("aws_elastic_beanstalk_environment.tftest", &app),
					testAccCheckBeanstalkEnvConfigValue("aws_elastic_beanstalk_environment.tftest", "3"),
				),
			},
		},
	})
}

func TestAccAWSBeanstalkEnv_resource(t *testing.T) {
	var app elasticbeanstalk.EnvironmentDescription

	rString := acctest.RandString(8)
	appName := fmt.Sprintf("tf_acc_app_env_resource_%s", rString)
	envName := fmt.Sprintf("tf-acc-env-resource-%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBeanstalkEnvDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBeanstalkResourceOptionSetting(appName, envName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBeanstalkEnvExists("aws_elastic_beanstalk_environment.tfenvtest", &app),
				),
			},
		},
	})
}

func TestAccAWSBeanstalkEnv_tags(t *testing.T) {
	var app elasticbeanstalk.EnvironmentDescription

	rString := acctest.RandString(8)
	appName := fmt.Sprintf("tf_acc_app_env_resource_%s", rString)
	envName := fmt.Sprintf("tf-acc-env-resource-%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBeanstalkEnvDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBeanstalkEnvConfig_empty_settings(appName, envName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBeanstalkEnvExists("aws_elastic_beanstalk_environment.tfenvtest", &app),
					testAccCheckBeanstalkEnvTagsMatch(&app, map[string]string{}),
				),
			},

			{
				Config: testAccBeanstalkTagsTemplate(appName, envName, "test1", "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBeanstalkEnvExists("aws_elastic_beanstalk_environment.tfenvtest", &app),
					testAccCheckBeanstalkEnvTagsMatch(&app, map[string]string{"firstTag": "test1", "secondTag": "test2"}),
				),
			},

			{
				Config: testAccBeanstalkTagsTemplate(appName, envName, "test2", "test1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBeanstalkEnvExists("aws_elastic_beanstalk_environment.tfenvtest", &app),
					testAccCheckBeanstalkEnvTagsMatch(&app, map[string]string{"firstTag": "test2", "secondTag": "test1"}),
				),
			},

			{
				Config: testAccBeanstalkEnvConfig_empty_settings(appName, envName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBeanstalkEnvExists("aws_elastic_beanstalk_environment.tfenvtest", &app),
					testAccCheckBeanstalkEnvTagsMatch(&app, map[string]string{}),
				),
			},
		},
	})
}

func TestAccAWSBeanstalkEnv_vpc(t *testing.T) {
	var app elasticbeanstalk.EnvironmentDescription

	rString := acctest.RandString(8)
	sgName := fmt.Sprintf("tf_acc_sg_beanstalk_env_vpc_%s", rString)
	appName := fmt.Sprintf("tf_acc_app_env_vpc_%s", rString)
	envName := fmt.Sprintf("tf-acc-env-vpc-%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBeanstalkEnvDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBeanstalkEnv_VPC(sgName, appName, envName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBeanstalkEnvExists("aws_elastic_beanstalk_environment.default", &app),
				),
			},
		},
	})
}

func TestAccAWSBeanstalkEnv_template_change(t *testing.T) {
	var app elasticbeanstalk.EnvironmentDescription

	rString := acctest.RandString(8)
	appName := fmt.Sprintf("tf_acc_app_env_tpl_change_%s", rString)
	envName := fmt.Sprintf("tf-acc-env-tpl-change-%s", rString)
	cfgTplName := fmt.Sprintf("tf_acc_tpl_env_tpl_change_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBeanstalkEnvDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBeanstalkEnv_TemplateChange_stack(appName, envName, cfgTplName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBeanstalkEnvExists("aws_elastic_beanstalk_environment.environment", &app),
				),
			},
			{
				Config: testAccBeanstalkEnv_TemplateChange_temp(appName, envName, cfgTplName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBeanstalkEnvExists("aws_elastic_beanstalk_environment.environment", &app),
				),
			},
			{
				Config: testAccBeanstalkEnv_TemplateChange_stack(appName, envName, cfgTplName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBeanstalkEnvExists("aws_elastic_beanstalk_environment.environment", &app),
				),
			},
		},
	})
}

func TestAccAWSBeanstalkEnv_basic_settings_update(t *testing.T) {
	var app elasticbeanstalk.EnvironmentDescription

	rString := acctest.RandString(8)
	appName := fmt.Sprintf("tf_acc_app_env_basic_settings_upd_%s", rString)
	envName := fmt.Sprintf("tf-acc-env-basic-settings-upd-%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBeanstalkEnvDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBeanstalkEnvConfig_empty_settings(appName, envName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBeanstalkEnvExists("aws_elastic_beanstalk_environment.tfenvtest", &app),
					testAccVerifyBeanstalkConfig(&app, []string{}),
				),
			},
			{
				Config: testAccBeanstalkEnvConfig_settings(appName, envName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBeanstalkEnvExists("aws_elastic_beanstalk_environment.tfenvtest", &app),
					testAccVerifyBeanstalkConfig(&app, []string{"ENV_STATIC", "ENV_UPDATE"}),
				),
			},
			{
				Config: testAccBeanstalkEnvConfig_settings(appName, envName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBeanstalkEnvExists("aws_elastic_beanstalk_environment.tfenvtest", &app),
					testAccVerifyBeanstalkConfig(&app, []string{"ENV_STATIC", "ENV_UPDATE"}),
				),
			},
			{
				Config: testAccBeanstalkEnvConfig_empty_settings(appName, envName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBeanstalkEnvExists("aws_elastic_beanstalk_environment.tfenvtest", &app),
					testAccVerifyBeanstalkConfig(&app, []string{}),
				),
			},
		},
	})
}

func TestAccAWSBeanstalkEnv_version_label(t *testing.T) {
	var app elasticbeanstalk.EnvironmentDescription

	rString := acctest.RandString(8)
	bucketName := fmt.Sprintf("tf-acc-bucket-beanstalk-env-version-label-%s", rString)
	appName := fmt.Sprintf("tf_acc_app_env_version_label_%s", rString)
	appVersionName := fmt.Sprintf("tf_acc_version_env_version_label_%s", rString)
	uAppVersionName := fmt.Sprintf("tf_acc_version_env_version_label_v2_%s", rString)
	envName := fmt.Sprintf("tf-acc-env-version-label-%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBeanstalkEnvDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBeanstalkEnvApplicationVersionConfig(bucketName, appName, appVersionName, envName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBeanstalkApplicationVersionDeployed("aws_elastic_beanstalk_environment.default", &app),
				),
			},
			{
				Config: testAccBeanstalkEnvApplicationVersionConfig(bucketName, appName, uAppVersionName, envName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBeanstalkApplicationVersionDeployed("aws_elastic_beanstalk_environment.default", &app),
				),
			},
		},
	})
}

func TestAccAWSBeanstalkEnv_settingWithJsonValue(t *testing.T) {
	var app elasticbeanstalk.EnvironmentDescription

	rString := acctest.RandString(8)
	appName := fmt.Sprintf("tf_acc_app_env_setting_w_json_value_%s", rString)
	queueName := fmt.Sprintf("tf_acc_queue_beanstalk_env_setting_w_json_value_%s", rString)
	keyPairName := fmt.Sprintf("tf_acc_keypair_beanstalk_env_setting_w_json_value_%s", rString)
	instanceProfileName := fmt.Sprintf("tf_acc_profile_beanstalk_env_setting_w_json_value_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_beanstalk_env_setting_w_json_value_%s", rString)
	policyName := fmt.Sprintf("tf-acc-policy-beanstalk-env-setting-w-json-value-%s", rString)
	envName := fmt.Sprintf("tf-acc-env-setting-w-json-value-%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBeanstalkEnvDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBeanstalkEnvSettingJsonValue(appName, queueName, keyPairName, instanceProfileName, roleName, policyName, envName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBeanstalkEnvExists("aws_elastic_beanstalk_environment.default", &app),
				),
			},
		},
	})
}

func TestAccAWSBeanstalkEnv_platformArn(t *testing.T) {
	var app elasticbeanstalk.EnvironmentDescription

	rString := acctest.RandString(8)
	appName := fmt.Sprintf("tf_acc_app_env_platform_arn_%s", rString)
	envName := fmt.Sprintf("tf-acc-env-platform-arn-%s", rString)
	platformArn := "arn:aws:elasticbeanstalk:us-east-1::platform/Go 1 running on 64bit Amazon Linux/2.9.0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBeanstalkEnvDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBeanstalkEnvConfig_platform_arn(appName, envName, platformArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBeanstalkEnvExists("aws_elastic_beanstalk_environment.tfenvtest", &app),
					resource.TestCheckResourceAttr("aws_elastic_beanstalk_environment.tfenvtest", "platform_arn", platformArn),
				),
			},
		},
	})
}

func testAccVerifyBeanstalkConfig(env *elasticbeanstalk.EnvironmentDescription, expected []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if env == nil {
			return fmt.Errorf("Nil environment in testAccVerifyBeanstalkConfig")
		}
		conn := testAccProvider.Meta().(*AWSClient).elasticbeanstalkconn

		resp, err := conn.DescribeConfigurationSettings(&elasticbeanstalk.DescribeConfigurationSettingsInput{
			ApplicationName: env.ApplicationName,
			EnvironmentName: env.EnvironmentName,
		})

		if err != nil {
			return fmt.Errorf("Error describing config settings in testAccVerifyBeanstalkConfig: %s", err)
		}

		// should only be 1 environment
		if len(resp.ConfigurationSettings) != 1 {
			return fmt.Errorf("Expected only 1 set of Configuration Settings in testAccVerifyBeanstalkConfig, got (%d)", len(resp.ConfigurationSettings))
		}

		cs := resp.ConfigurationSettings[0]

		var foundEnvs []string
		testStrings := []string{"ENV_STATIC", "ENV_UPDATE"}
		for _, os := range cs.OptionSettings {
			for _, k := range testStrings {
				if *os.OptionName == k {
					foundEnvs = append(foundEnvs, k)
				}
			}
		}

		// if expected is zero, then we should not have found any of the predefined
		// env vars
		if len(expected) == 0 {
			if len(foundEnvs) > 0 {
				return fmt.Errorf("Found configs we should not have: %#v", foundEnvs)
			}
			return nil
		}

		sort.Strings(testStrings)
		sort.Strings(expected)
		if !reflect.DeepEqual(testStrings, expected) {
			return fmt.Errorf("Error matching strings, expected:\n\n%#v\n\ngot:\n\n%#v\n", testStrings, foundEnvs)
		}

		return nil
	}
}

func testAccCheckBeanstalkEnvDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).elasticbeanstalkconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_elastic_beanstalk_environment" {
			continue
		}

		// Try to find the environment
		describeBeanstalkEnvOpts := &elasticbeanstalk.DescribeEnvironmentsInput{
			EnvironmentIds: []*string{aws.String(rs.Primary.ID)},
		}
		resp, err := conn.DescribeEnvironments(describeBeanstalkEnvOpts)
		if err == nil {
			switch {
			case len(resp.Environments) > 1:
				return fmt.Errorf("Error %d environments match, expected 1", len(resp.Environments))
			case len(resp.Environments) == 1:
				if *resp.Environments[0].Status == "Terminated" {
					return nil
				}
				return fmt.Errorf("Elastic Beanstalk ENV still exists.")
			default:
				return nil
			}
		}

		// Verify the error is what we want
		ec2err, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if ec2err.Code() != "InvalidBeanstalkEnvID.NotFound" {
			return err
		}
	}

	return nil
}

func testAccCheckBeanstalkEnvExists(n string, app *elasticbeanstalk.EnvironmentDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Elastic Beanstalk ENV is not set")
		}

		env, err := describeBeanstalkEnv(testAccProvider.Meta().(*AWSClient).elasticbeanstalkconn, aws.String(rs.Primary.ID))
		if err != nil {
			return err
		}

		*app = *env

		return nil
	}
}

func testAccCheckBeanstalkEnvTier(n string, app *elasticbeanstalk.EnvironmentDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Elastic Beanstalk ENV is not set")
		}

		env, err := describeBeanstalkEnv(testAccProvider.Meta().(*AWSClient).elasticbeanstalkconn, aws.String(rs.Primary.ID))
		if err != nil {
			return err
		}
		if *env.Tier.Name != "Worker" {
			return fmt.Errorf("Beanstalk Environment tier is %s, expected Worker", *env.Tier.Name)
		}

		*app = *env

		return nil
	}
}

func testAccCheckBeanstalkEnvConfigValue(n string, expectedValue string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).elasticbeanstalkconn

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Elastic Beanstalk ENV is not set")
		}

		resp, err := conn.DescribeConfigurationOptions(&elasticbeanstalk.DescribeConfigurationOptionsInput{
			ApplicationName: aws.String(rs.Primary.Attributes["application"]),
			EnvironmentName: aws.String(rs.Primary.Attributes["name"]),
			Options: []*elasticbeanstalk.OptionSpecification{
				{
					Namespace:  aws.String("aws:elasticbeanstalk:application:environment"),
					OptionName: aws.String("TEMPLATE"),
				},
			},
		})
		if err != nil {
			return err
		}

		if len(resp.Options) != 1 {
			return fmt.Errorf("Found %d options, expected 1.", len(resp.Options))
		}

		log.Printf("[DEBUG] %d Elastic Beanstalk Option values returned.", len(resp.Options[0].ValueOptions))

		for _, value := range resp.Options[0].ValueOptions {
			if *value != expectedValue {
				return fmt.Errorf("Option setting value: %s. Expected %s", *value, expectedValue)
			}
		}

		return nil
	}
}

func testAccCheckBeanstalkEnvTagsMatch(env *elasticbeanstalk.EnvironmentDescription, expectedValue map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if env == nil {
			return fmt.Errorf("Nil environment in testAccCheckBeanstalkEnvTagsMatch")
		}

		conn := testAccProvider.Meta().(*AWSClient).elasticbeanstalkconn

		tags, err := conn.ListTagsForResource(&elasticbeanstalk.ListTagsForResourceInput{
			ResourceArn: env.EnvironmentArn,
		})

		if err != nil {
			return err
		}

		foundTags := tagsToMapBeanstalk(tags.ResourceTags)

		if !reflect.DeepEqual(foundTags, expectedValue) {
			return fmt.Errorf("Tag value: %s.  Expected %s", foundTags, expectedValue)
		}

		return nil
	}
}

func testAccCheckBeanstalkApplicationVersionDeployed(n string, app *elasticbeanstalk.EnvironmentDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Elastic Beanstalk ENV is not set")
		}

		env, err := describeBeanstalkEnv(testAccProvider.Meta().(*AWSClient).elasticbeanstalkconn, aws.String(rs.Primary.ID))
		if err != nil {
			return err
		}

		if *env.VersionLabel != rs.Primary.Attributes["version_label"] {
			return fmt.Errorf("Elastic Beanstalk version deployed %s. Expected %s", *env.VersionLabel, rs.Primary.Attributes["version_label"])
		}

		*app = *env

		return nil
	}
}

func describeBeanstalkEnv(conn *elasticbeanstalk.ElasticBeanstalk,
	envID *string) (*elasticbeanstalk.EnvironmentDescription, error) {
	describeBeanstalkEnvOpts := &elasticbeanstalk.DescribeEnvironmentsInput{
		EnvironmentIds: []*string{envID},
	}

	log.Printf("[DEBUG] Elastic Beanstalk Environment TEST describe opts: %s", describeBeanstalkEnvOpts)

	resp, err := conn.DescribeEnvironments(describeBeanstalkEnvOpts)
	if err != nil {
		return &elasticbeanstalk.EnvironmentDescription{}, err
	}
	if len(resp.Environments) == 0 {
		return &elasticbeanstalk.EnvironmentDescription{}, fmt.Errorf("Elastic Beanstalk ENV not found.")
	}
	if len(resp.Environments) > 1 {
		return &elasticbeanstalk.EnvironmentDescription{}, fmt.Errorf("Found %d environments, expected 1.", len(resp.Environments))
	}
	return resp.Environments[0], nil
}

func testAccBeanstalkEnvConfig(appName, envName string) string {
	return fmt.Sprintf(`
 resource "aws_elastic_beanstalk_application" "tftest" {
	 name = "%s"
	 description = "tf-test-desc"
 }

 resource "aws_elastic_beanstalk_environment" "tfenvtest" {
	 name = "%s"
	 application = "${aws_elastic_beanstalk_application.tftest.name}"
	 solution_stack_name = "64bit Amazon Linux running Python"
	 depends_on = ["aws_elastic_beanstalk_application.tftest"]
 }
 `, appName, envName)
}

func testAccBeanstalkEnvConfig_platform_arn(appName, envName, platformArn string) string {
	return fmt.Sprintf(`
provider "aws" {
  region = "us-east-1"
}

resource "aws_elastic_beanstalk_application" "tftest" {
  name = "%s"
  description = "tf-test-desc"
}

resource "aws_elastic_beanstalk_environment" "tfenvtest" {
  name = "%s"
  application = "${aws_elastic_beanstalk_application.tftest.name}"
  platform_arn = "%s"
}
`, appName, envName, platformArn)
}

func testAccBeanstalkEnvConfig_empty_settings(appName, envName string) string {
	return fmt.Sprintf(`
resource "aws_elastic_beanstalk_application" "tftest" {
  name = "%s"
  description = "tf-test-desc"
}

resource "aws_elastic_beanstalk_environment" "tfenvtest" {
  name = "%s"
  application = "${aws_elastic_beanstalk_application.tftest.name}"
  solution_stack_name = "64bit Amazon Linux running Python"

  wait_for_ready_timeout = "15m"
}`, appName, envName)
}

func testAccBeanstalkEnvConfig_settings(appName, envName string) string {
	return fmt.Sprintf(`
resource "aws_elastic_beanstalk_application" "tftest" {
  name = "%s"
  description = "tf-test-desc"
}

resource "aws_elastic_beanstalk_environment" "tfenvtest" {
  name                = "%s"
  application         = "${aws_elastic_beanstalk_application.tftest.name}"
  solution_stack_name = "64bit Amazon Linux running Python"

  wait_for_ready_timeout = "15m"

  setting {
    namespace = "aws:elasticbeanstalk:application:environment"
    name      = "ENV_STATIC"
    value     = "true"
  }

  setting {
    namespace = "aws:elasticbeanstalk:application:environment"
    name      = "ENV_UPDATE"
    value     = "true"
  }

  setting {
    namespace = "aws:elasticbeanstalk:application:environment"
    name      = "ENV_REMOVE"
    value     = "true"
  }

  setting {
    namespace = "aws:autoscaling:scheduledaction"
    resource  = "ScheduledAction01"
    name      = "MinSize"
    value     = 2
  }

  setting {
    namespace = "aws:autoscaling:scheduledaction"
    resource  = "ScheduledAction01"
    name      = "MaxSize"
    value     = 3
  }

  setting {
    namespace = "aws:autoscaling:scheduledaction"
    resource  = "ScheduledAction01"
    name      = "StartTime"
    value     = "2016-07-28T04:07:02Z"
  }
}`, appName, envName)
}

func testAccBeanstalkWorkerEnvConfig(instanceProfileName, roleName, policyName, appName, envName string) string {
	return fmt.Sprintf(`
 resource "aws_iam_instance_profile" "tftest" {
	 name = "%s"
	 roles = ["${aws_iam_role.tftest.name}"]
 }

 resource "aws_iam_role" "tftest" {
	 name = "%s"
	 path = "/"
	 assume_role_policy = "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Action\":\"sts:AssumeRole\",\"Principal\":{\"Service\":\"ec2.amazonaws.com\"},\"Effect\":\"Allow\",\"Sid\":\"\"}]}"
 }

 resource "aws_iam_role_policy" "tftest" {
	 name = "%s"
	 role = "${aws_iam_role.tftest.id}"
	 policy = "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Sid\":\"QueueAccess\",\"Action\":[\"sqs:ChangeMessageVisibility\",\"sqs:DeleteMessage\",\"sqs:ReceiveMessage\"],\"Effect\":\"Allow\",\"Resource\":\"*\"}]}"
 }

 resource "aws_elastic_beanstalk_application" "tftest" {
	 name = "%s"
	 description = "tf-test-desc"
 }

 resource "aws_elastic_beanstalk_environment" "tfenvtest" {
	 name = "%s"
	 application = "${aws_elastic_beanstalk_application.tftest.name}"
	 tier = "Worker"
	 solution_stack_name = "64bit Amazon Linux running Python"

	 setting {
		 namespace = "aws:autoscaling:launchconfiguration"
		 name      = "IamInstanceProfile"
		 value     = "${aws_iam_instance_profile.tftest.name}"
	 }
 }`, instanceProfileName, roleName, policyName, appName, envName)
}

func testAccBeanstalkEnvCnamePrefixConfig(appName, envName, cnamePrefix string) string {
	return fmt.Sprintf(`
resource "aws_elastic_beanstalk_application" "tftest" {
  name = "%s"
  description = "tf-test-desc"
}

resource "aws_elastic_beanstalk_environment" "tfenvtest" {
  name = "%s"
  application = "${aws_elastic_beanstalk_application.tftest.name}"
  cname_prefix = "%s"
  solution_stack_name = "64bit Amazon Linux running Python"
}
`, appName, envName, cnamePrefix)
}

func testAccBeanstalkConfigTemplate(appName, envName, cfgTplName string, cfgTplValue int) string {
	return fmt.Sprintf(`
resource "aws_elastic_beanstalk_application" "tftest" {
  name = "%s"
  description = "tf-test-desc"
}

resource "aws_elastic_beanstalk_environment" "tftest" {
  name = "%s"
  application = "${aws_elastic_beanstalk_application.tftest.name}"
  template_name = "${aws_elastic_beanstalk_configuration_template.tftest.name}"
}

resource "aws_elastic_beanstalk_configuration_template" "tftest" {
  name        = "%s"
  application = "${aws_elastic_beanstalk_application.tftest.name}"
  solution_stack_name = "64bit Amazon Linux running Python"

  setting {
    namespace = "aws:elasticbeanstalk:application:environment"
    name      = "TEMPLATE"
    value     = "%d"
  }
}`, appName, envName, cfgTplName, cfgTplValue)
}

func testAccBeanstalkResourceOptionSetting(appName, envName string) string {
	return fmt.Sprintf(`
resource "aws_elastic_beanstalk_application" "tftest" {
  name = "%s"
  description = "tf-test-desc"
}

resource "aws_elastic_beanstalk_environment" "tfenvtest" {
  name = "%s"
  application = "${aws_elastic_beanstalk_application.tftest.name}"
  solution_stack_name = "64bit Amazon Linux running Python"

  setting {
    namespace = "aws:autoscaling:scheduledaction"
    resource = "ScheduledAction01"
    name = "MinSize"
    value = "2"
  }

  setting {
    namespace = "aws:autoscaling:scheduledaction"
    resource = "ScheduledAction01"
    name = "MaxSize"
    value = "6"
  }

  setting {
    namespace = "aws:autoscaling:scheduledaction"
    resource = "ScheduledAction01"
    name = "Recurrence"
    value = "0 8 * * *"
  }
}`, appName, envName)
}

func testAccBeanstalkTagsTemplate(appName, envName, firstTag, secondTag string) string {
	return fmt.Sprintf(`
resource "aws_elastic_beanstalk_application" "tftest" {
  name = "%s"
  description = "tf-test-desc"
}

resource "aws_elastic_beanstalk_environment" "tfenvtest" {
  name = "%s"
  application = "${aws_elastic_beanstalk_application.tftest.name}"
  solution_stack_name = "64bit Amazon Linux running Python"

  wait_for_ready_timeout = "15m"

  tags = {
    firstTag = "%s"
    secondTag = "%s"
  }
}`, appName, envName, firstTag, secondTag)
}

func testAccBeanstalkEnv_VPC(sgName, appName, envName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "tf_b_test" {
  cidr_block = "10.0.0.0/16"
	tags = {
		Name = "terraform-testacc-elastic-beanstalk-env-vpc"
	}
}

resource "aws_internet_gateway" "tf_b_test" {
  vpc_id = "${aws_vpc.tf_b_test.id}"
}

resource "aws_route" "r" {
  route_table_id = "${aws_vpc.tf_b_test.main_route_table_id}"
  destination_cidr_block = "0.0.0.0/0"
  gateway_id = "${aws_internet_gateway.tf_b_test.id}"
}

resource "aws_subnet" "main" {
  vpc_id     = "${aws_vpc.tf_b_test.id}"
  cidr_block = "10.0.0.0/24"
  tags = {
    Name = "tf-acc-elastic-beanstalk-env-vpc"
  }
}

resource "aws_security_group" "default" {
  name = "%s"
  vpc_id = "${aws_vpc.tf_b_test.id}"
}

resource "aws_elastic_beanstalk_application" "default" {
  name = "%s"
  description = "tf-test-desc"
}

resource "aws_elastic_beanstalk_environment" "default" {
  name = "%s"
  application = "${aws_elastic_beanstalk_application.default.name}"
  solution_stack_name = "64bit Amazon Linux running Python"

  setting {
    namespace = "aws:ec2:vpc"
    name      = "VPCId"
    value     = "${aws_vpc.tf_b_test.id}"
  }

  setting {
    namespace = "aws:ec2:vpc"
    name      = "Subnets"
    value     = "${aws_subnet.main.id}"
  }

  setting {
    namespace = "aws:ec2:vpc"
    name      = "AssociatePublicIpAddress"
    value     = "true"
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "SecurityGroups"
    value     = "${aws_security_group.default.id}"
  }
}
`, sgName, appName, envName)
}

func testAccBeanstalkEnv_TemplateChange_stack(appName, envName, cfgTplName string) string {
	return fmt.Sprintf(`
provider "aws" {
  region = "us-east-1"
}

resource "aws_elastic_beanstalk_application" "app" {
  name        = "%s"
  description = ""
}

resource "aws_elastic_beanstalk_environment" "environment" {
  name        = "%s"
  application = "${aws_elastic_beanstalk_application.app.name}"

  # Go 1.4
  solution_stack_name = "64bit Amazon Linux 2016.03 v2.1.0 running Go 1.4"
}

resource "aws_elastic_beanstalk_configuration_template" "template" {
  name        = "%s"
  application = "${aws_elastic_beanstalk_application.app.name}"

  # Go 1.5
  solution_stack_name = "64bit Amazon Linux 2016.03 v2.1.3 running Go 1.5"
}
`, appName, envName, cfgTplName)
}

func testAccBeanstalkEnv_TemplateChange_temp(appName, envName, cfgTplName string) string {
	return fmt.Sprintf(`
provider "aws" {
  region = "us-east-1"
}

resource "aws_elastic_beanstalk_application" "app" {
  name        = "%s"
  description = ""
}

resource "aws_elastic_beanstalk_environment" "environment" {
  name        = "%s"
  application = "${aws_elastic_beanstalk_application.app.name}"

  # Go 1.4
  template_name = "${aws_elastic_beanstalk_configuration_template.template.name}"
}

resource "aws_elastic_beanstalk_configuration_template" "template" {
  name        = "%s"
  application = "${aws_elastic_beanstalk_application.app.name}"

  # Go 1.5
  solution_stack_name = "64bit Amazon Linux 2016.03 v2.1.3 running Go 1.5"
}
`, appName, envName, cfgTplName)
}

func testAccBeanstalkEnvApplicationVersionConfig(bucketName, appName, appVersionName, envName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "default" {
  bucket = "%s"
}

resource "aws_s3_bucket_object" "default" {
  bucket = "${aws_s3_bucket.default.id}"
  key = "python-v1.zip"
  source = "test-fixtures/python-v1.zip"
}

resource "aws_elastic_beanstalk_application" "default" {
  name = "%s"
  description = "tf-test-desc"
}

resource "aws_elastic_beanstalk_application_version" "default" {
  application = "${aws_elastic_beanstalk_application.default.name}"
  name = "%s"
  bucket = "${aws_s3_bucket.default.id}"
  key = "${aws_s3_bucket_object.default.id}"
}

resource "aws_elastic_beanstalk_environment" "default" {
  name = "%s"
  application = "${aws_elastic_beanstalk_application.default.name}"
  version_label = "${aws_elastic_beanstalk_application_version.default.name}"
  solution_stack_name = "64bit Amazon Linux running Python"
}
`, bucketName, appName, appVersionName, envName)
}

func testAccBeanstalkEnvSettingJsonValue(appName, queueName, keyPairName, instanceProfileName, roleName, policyName, envName string) string {
	return fmt.Sprintf(`
resource "aws_elastic_beanstalk_application" "app" {
  name = "%s"
  description = "This is a description"
}

resource "aws_sqs_queue" "test" {
  name = "%s"
}

resource "aws_key_pair" "test" {
  key_name   = "%s"
  public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 email@example.com"
}

resource "aws_iam_instance_profile" "app" {
  name  = "%s"
  role = "${aws_iam_role.test.name}"
}

resource "aws_iam_role" "test" {
  name = "%s"
  path = "/"

  assume_role_policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": "sts:AssumeRole",
            "Principal": {
               "Service": "ec2.amazonaws.com"
            },
            "Effect": "Allow",
            "Sid": ""
        }
    ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = "%s"
  role = "${aws_iam_role.test.id}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "dynamodb:*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_elastic_beanstalk_environment" "default" {
  name = "%s"
  application = "${aws_elastic_beanstalk_application.app.name}"
  tier = "Worker"
  solution_stack_name = "64bit Amazon Linux 2016.03 v2.1.0 running Docker 1.9.1"

  setting = {
    namespace = "aws:elasticbeanstalk:command"
    name = "BatchSize"
    value = "30"
  }

  setting = {
    namespace = "aws:elasticbeanstalk:command"
    name = "BatchSizeType"
    value = "Percentage"
  }

  setting = {
    namespace = "aws:elasticbeanstalk:command"
    name = "DeploymentPolicy"
    value = "Rolling"
  }

  setting = {
    namespace = "aws:elasticbeanstalk:sns:topics"
    name = "Notification Endpoint"
    value = "example@example.com"
  }

  setting = {
    namespace = "aws:elasticbeanstalk:sqsd"
    name = "ErrorVisibilityTimeout"
    value = "2"
  }

  setting = {
    namespace = "aws:elasticbeanstalk:sqsd"
    name = "HttpPath"
    value = "/event-message"
  }

  setting = {
    namespace = "aws:elasticbeanstalk:sqsd"
    name = "WorkerQueueURL"
    value = "${aws_sqs_queue.test.id}"
  }

  setting = {
    namespace = "aws:elasticbeanstalk:sqsd"
    name = "VisibilityTimeout"
    value = "300"
  }

  setting = {
    namespace = "aws:elasticbeanstalk:sqsd"
    name = "HttpConnections"
    value = "10"
  }

  setting = {
    namespace = "aws:elasticbeanstalk:sqsd"
    name = "InactivityTimeout"
    value = "299"
  }

  setting = {
    namespace = "aws:elasticbeanstalk:sqsd"
    name = "MimeType"
    value = "application/json"
  }

  setting = {
    namespace = "aws:elasticbeanstalk:environment"
    name = "ServiceRole"
    value = "aws-elasticbeanstalk-service-role"
  }

  setting = {
    namespace = "aws:elasticbeanstalk:environment"
    name = "EnvironmentType"
    value = "LoadBalanced"
  }

  setting = {
    namespace = "aws:elasticbeanstalk:application"
    name = "Application Healthcheck URL"
    value = "/health"
  }

  setting = {
    namespace = "aws:elasticbeanstalk:healthreporting:system"
    name = "SystemType"
    value = "enhanced"
  }

  setting = {
    namespace = "aws:elasticbeanstalk:healthreporting:system"
    name = "HealthCheckSuccessThreshold"
    value = "Ok"
  }

  setting = {
    namespace = "aws:autoscaling:launchconfiguration"
    name = "IamInstanceProfile"
    value = "${aws_iam_instance_profile.app.name}"
  }

  setting = {
    namespace = "aws:autoscaling:launchconfiguration"
    name = "InstanceType"
    value = "t2.micro"
  }

  setting = {
    namespace = "aws:autoscaling:launchconfiguration"
    name = "EC2KeyName"
    value = "${aws_key_pair.test.key_name}"
  }

  setting = {
    namespace = "aws:autoscaling:updatepolicy:rollingupdate"
    name = "RollingUpdateEnabled"
    value = "false"
  }

  setting = {
    namespace = "aws:elasticbeanstalk:healthreporting:system"
    name = "ConfigDocument"
    value = <<EOF
{
	"Version": 1,
	"CloudWatchMetrics": {
		"Instance": {
			"ApplicationRequestsTotal": 60
		},
		"Environment": {
			"ApplicationRequests5xx": 60,
			"ApplicationRequests4xx": 60,
			"ApplicationRequests2xx": 60
		}
	}
}
EOF
  }
}
`, appName, queueName, keyPairName, instanceProfileName, roleName, policyName, envName)
}
