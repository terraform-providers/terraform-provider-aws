package aws

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/codedeploy"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSCodeDeployDeploymentGroup_basic(t *testing.T) {
	var group codedeploy.DeploymentGroupInfo

	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeDeployDeploymentGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeDeployDeploymentGroup(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "app_name", "foo_app_"+rName),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "deployment_group_name", "foo_"+rName),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "deployment_config_name", "CodeDeployDefault.OneAtATime"),
					resource.TestMatchResourceAttr(
						"aws_codedeploy_deployment_group.foo", "service_role_arn",
						regexp.MustCompile("arn:aws:iam::[0-9]{12}:role/foo_role_.*")),

					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "ec2_tag_set.#", "0"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "ec2_tag_filter.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "ec2_tag_filter.2916377465.key", "filterkey"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "ec2_tag_filter.2916377465.type", "KEY_AND_VALUE"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "ec2_tag_filter.2916377465.value", "filtervalue"),

					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "alarm_configuration.#", "0"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "auto_rollback_configuration.#", "0"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "trigger_configuration.#", "0"),
				),
			},
			{
				Config: testAccAWSCodeDeployDeploymentGroupModified(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "app_name", "foo_app_"+rName),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "deployment_group_name", "bar_"+rName),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "deployment_config_name", "CodeDeployDefault.OneAtATime"),
					resource.TestMatchResourceAttr(
						"aws_codedeploy_deployment_group.foo", "service_role_arn",
						regexp.MustCompile("arn:aws:iam::[0-9]{12}:role/bar_role_.*")),

					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "ec2_tag_set.#", "0"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "ec2_tag_filter.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "ec2_tag_filter.2369538975.key", "filterkey"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "ec2_tag_filter.2369538975.type", "KEY_AND_VALUE"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "ec2_tag_filter.2369538975.value", "anotherfiltervalue"),

					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "alarm_configuration.#", "0"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "auto_rollback_configuration.#", "0"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "trigger_configuration.#", "0"),
				),
			},
			{
				ResourceName:      "aws_codedeploy_deployment_group.foo",
				ImportState:       true,
				ImportStateIdFunc: testAccAWSCodeDeployDeploymentGroupImportStateIdFunc("aws_codedeploy_deployment_group.foo"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSCodeDeployDeploymentGroup_basic_tagSet(t *testing.T) {
	var group codedeploy.DeploymentGroupInfo

	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeDeployDeploymentGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeDeployDeploymentGroup(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "app_name", "foo_app_"+rName),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "deployment_group_name", "foo_"+rName),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "deployment_config_name", "CodeDeployDefault.OneAtATime"),
					resource.TestMatchResourceAttr(
						"aws_codedeploy_deployment_group.foo", "service_role_arn",
						regexp.MustCompile("arn:aws:iam::[0-9]{12}:role/foo_role_.*")),

					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "ec2_tag_set.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "ec2_tag_set.2916377593.ec2_tag_filter.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "ec2_tag_set.2916377593.ec2_tag_filter.2916377465.key", "filterkey"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "ec2_tag_set.2916377593.ec2_tag_filter.2916377465.type", "KEY_AND_VALUE"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "ec2_tag_set.2916377593.ec2_tag_filter.2916377465.value", "filtervalue"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "ec2_tag_filter.#", "0"),

					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "alarm_configuration.#", "0"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "auto_rollback_configuration.#", "0"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "trigger_configuration.#", "0"),
				),
			},
			{
				Config: testAccAWSCodeDeployDeploymentGroupModified(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "app_name", "foo_app_"+rName),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "deployment_group_name", "bar_"+rName),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "deployment_config_name", "CodeDeployDefault.OneAtATime"),
					resource.TestMatchResourceAttr(
						"aws_codedeploy_deployment_group.foo", "service_role_arn",
						regexp.MustCompile("arn:aws:iam::[0-9]{12}:role/bar_role_.*")),

					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "ec2_tag_set.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "ec2_tag_set.2369538847.ec2_tag_filter.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "ec2_tag_set.2369538847.ec2_tag_filter.2369538975.key", "filterkey"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "ec2_tag_set.2369538847.ec2_tag_filter.2369538975.type", "KEY_AND_VALUE"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "ec2_tag_set.2369538847.ec2_tag_filter.2369538975.value", "anotherfiltervalue"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "ec2_tag_filter.#", "0"),

					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "alarm_configuration.#", "0"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "auto_rollback_configuration.#", "0"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "trigger_configuration.#", "0"),
				),
			},
			{
				ResourceName:      "aws_codedeploy_deployment_group.foo",
				ImportState:       true,
				ImportStateIdFunc: testAccAWSCodeDeployDeploymentGroupImportStateIdFunc("aws_codedeploy_deployment_group.foo"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSCodeDeployDeploymentGroup_onPremiseTag(t *testing.T) {
	var group codedeploy.DeploymentGroupInfo

	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeDeployDeploymentGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeDeployDeploymentGroupOnPremiseTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "app_name", "foo_app_"+rName),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "deployment_group_name", "foo_"+rName),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "deployment_config_name", "CodeDeployDefault.OneAtATime"),

					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "on_premises_instance_tag_filter.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "on_premises_instance_tag_filter.2916377465.key", "filterkey"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "on_premises_instance_tag_filter.2916377465.type", "KEY_AND_VALUE"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo", "on_premises_instance_tag_filter.2916377465.value", "filtervalue"),
				),
			},
			{
				ResourceName:      "aws_codedeploy_deployment_group.foo",
				ImportState:       true,
				ImportStateIdFunc: testAccAWSCodeDeployDeploymentGroupImportStateIdFunc("aws_codedeploy_deployment_group.foo"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSCodeDeployDeploymentGroup_disappears(t *testing.T) {
	var group codedeploy.DeploymentGroupInfo
	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeDeployDeploymentGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeDeployDeploymentGroup(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo", &group),
					testAccAWSCodeDeployDeploymentGroupDisappears(&group),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSCodeDeployDeploymentGroup_triggerConfiguration_basic(t *testing.T) {
	var group codedeploy.DeploymentGroupInfo

	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeDeployDeploymentGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeDeployDeploymentGroup_triggerConfiguration_create(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "app_name", "foo-app-"+rName),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "deployment_group_name", "foo-group-"+rName),
					testAccCheckTriggerEvents(&group, "foo-trigger", []string{
						"DeploymentFailure",
					}),
				),
			},
			{
				Config: testAccAWSCodeDeployDeploymentGroup_triggerConfiguration_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "app_name", "foo-app-"+rName),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "deployment_group_name", "foo-group-"+rName),
					testAccCheckTriggerEvents(&group, "foo-trigger", []string{
						"DeploymentFailure",
						"DeploymentSuccess",
					}),
				),
			},
			{
				ResourceName:      "aws_codedeploy_deployment_group.foo_group",
				ImportState:       true,
				ImportStateIdFunc: testAccAWSCodeDeployDeploymentGroupImportStateIdFunc("aws_codedeploy_deployment_group.foo_group"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSCodeDeployDeploymentGroup_triggerConfiguration_multiple(t *testing.T) {
	var group codedeploy.DeploymentGroupInfo

	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeDeployDeploymentGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeDeployDeploymentGroup_triggerConfiguration_createMultiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "app_name", "foo-app-"+rName),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "deployment_group_name", "foo-group-"+rName),
					testAccCheckTriggerEvents(&group, "foo-trigger", []string{
						"DeploymentFailure",
					}),
					testAccCheckTriggerEvents(&group, "bar-trigger", []string{
						"InstanceFailure",
					}),
					testAccCheckTriggerTargetArn(&group, "bar-trigger",
						regexp.MustCompile("^arn:aws:sns:[^:]+:[0-9]{12}:bar-topic-"+rName+"$")),
				),
			},
			{
				Config: testAccAWSCodeDeployDeploymentGroup_triggerConfiguration_updateMultiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "app_name", "foo-app-"+rName),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "deployment_group_name", "foo-group-"+rName),
					testAccCheckTriggerEvents(&group, "foo-trigger", []string{
						"DeploymentFailure",
						"DeploymentStart",
						"DeploymentStop",
						"DeploymentSuccess",
					}),
					testAccCheckTriggerEvents(&group, "bar-trigger", []string{
						"InstanceFailure",
					}),
					testAccCheckTriggerTargetArn(&group, "bar-trigger",
						regexp.MustCompile("^arn:aws:sns:[^:]+:[0-9]{12}:baz-topic-"+rName+"$")),
				),
			},
			{
				ResourceName:      "aws_codedeploy_deployment_group.foo_group",
				ImportState:       true,
				ImportStateIdFunc: testAccAWSCodeDeployDeploymentGroupImportStateIdFunc("aws_codedeploy_deployment_group.foo_group"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSCodeDeployDeploymentGroup_autoRollbackConfiguration_create(t *testing.T) {
	var group codedeploy.DeploymentGroupInfo

	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeDeployDeploymentGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: test_config_auto_rollback_configuration_delete(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "auto_rollback_configuration.#", "0"),
				),
			},
			{
				Config: test_config_auto_rollback_configuration_create(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "auto_rollback_configuration.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "auto_rollback_configuration.0.enabled", "true"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "auto_rollback_configuration.0.events.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "auto_rollback_configuration.0.events.135881253", "DEPLOYMENT_FAILURE"),
				),
			},
			{
				ResourceName:      "aws_codedeploy_deployment_group.foo_group",
				ImportState:       true,
				ImportStateIdFunc: testAccAWSCodeDeployDeploymentGroupImportStateIdFunc("aws_codedeploy_deployment_group.foo_group"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSCodeDeployDeploymentGroup_autoRollbackConfiguration_update(t *testing.T) {
	var group codedeploy.DeploymentGroupInfo

	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeDeployDeploymentGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: test_config_auto_rollback_configuration_create(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "auto_rollback_configuration.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "auto_rollback_configuration.0.enabled", "true"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "auto_rollback_configuration.0.events.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "auto_rollback_configuration.0.events.135881253", "DEPLOYMENT_FAILURE"),
				),
			},
			{
				Config: test_config_auto_rollback_configuration_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "auto_rollback_configuration.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "auto_rollback_configuration.0.enabled", "true"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "auto_rollback_configuration.0.events.#", "2"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "auto_rollback_configuration.0.events.104943466", "DEPLOYMENT_STOP_ON_ALARM"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "auto_rollback_configuration.0.events.135881253", "DEPLOYMENT_FAILURE"),
				),
			},
			{
				ResourceName:      "aws_codedeploy_deployment_group.foo_group",
				ImportState:       true,
				ImportStateIdFunc: testAccAWSCodeDeployDeploymentGroupImportStateIdFunc("aws_codedeploy_deployment_group.foo_group"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSCodeDeployDeploymentGroup_autoRollbackConfiguration_delete(t *testing.T) {
	var group codedeploy.DeploymentGroupInfo

	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeDeployDeploymentGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: test_config_auto_rollback_configuration_create(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "auto_rollback_configuration.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "auto_rollback_configuration.0.enabled", "true"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "auto_rollback_configuration.0.events.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "auto_rollback_configuration.0.events.135881253", "DEPLOYMENT_FAILURE"),
				),
			},
			{
				Config: test_config_auto_rollback_configuration_delete(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "auto_rollback_configuration.#", "0"),
				),
			},
			{
				ResourceName:      "aws_codedeploy_deployment_group.foo_group",
				ImportState:       true,
				ImportStateIdFunc: testAccAWSCodeDeployDeploymentGroupImportStateIdFunc("aws_codedeploy_deployment_group.foo_group"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSCodeDeployDeploymentGroup_autoRollbackConfiguration_disable(t *testing.T) {
	var group codedeploy.DeploymentGroupInfo

	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeDeployDeploymentGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: test_config_auto_rollback_configuration_create(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "auto_rollback_configuration.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "auto_rollback_configuration.0.enabled", "true"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "auto_rollback_configuration.0.events.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "auto_rollback_configuration.0.events.135881253", "DEPLOYMENT_FAILURE"),
				),
			},
			{
				Config: test_config_auto_rollback_configuration_disable(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "auto_rollback_configuration.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "auto_rollback_configuration.0.enabled", "false"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "auto_rollback_configuration.0.events.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "auto_rollback_configuration.0.events.135881253", "DEPLOYMENT_FAILURE"),
				),
			},
			{
				ResourceName:      "aws_codedeploy_deployment_group.foo_group",
				ImportState:       true,
				ImportStateIdFunc: testAccAWSCodeDeployDeploymentGroupImportStateIdFunc("aws_codedeploy_deployment_group.foo_group"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSCodeDeployDeploymentGroup_alarmConfiguration_create(t *testing.T) {
	var group codedeploy.DeploymentGroupInfo

	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeDeployDeploymentGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: test_config_alarm_configuration_delete(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "alarm_configuration.#", "0"),
				),
			},
			{
				Config: test_config_alarm_configuration_create(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "alarm_configuration.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "alarm_configuration.0.enabled", "true"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "alarm_configuration.0.alarms.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "alarm_configuration.0.alarms.2356372769", "foo"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "alarm_configuration.0.ignore_poll_alarm_failure", "false"),
				),
			},
			{
				ResourceName:      "aws_codedeploy_deployment_group.foo_group",
				ImportState:       true,
				ImportStateIdFunc: testAccAWSCodeDeployDeploymentGroupImportStateIdFunc("aws_codedeploy_deployment_group.foo_group"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSCodeDeployDeploymentGroup_alarmConfiguration_update(t *testing.T) {
	var group codedeploy.DeploymentGroupInfo

	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeDeployDeploymentGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: test_config_alarm_configuration_create(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "alarm_configuration.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "alarm_configuration.0.enabled", "true"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "alarm_configuration.0.alarms.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "alarm_configuration.0.alarms.2356372769", "foo"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "alarm_configuration.0.ignore_poll_alarm_failure", "false"),
				),
			},
			{
				Config: test_config_alarm_configuration_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "alarm_configuration.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "alarm_configuration.0.enabled", "true"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "alarm_configuration.0.alarms.#", "2"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "alarm_configuration.0.alarms.1996459178", "bar"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "alarm_configuration.0.alarms.2356372769", "foo"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "alarm_configuration.0.ignore_poll_alarm_failure", "true"),
				),
			},
			{
				ResourceName:      "aws_codedeploy_deployment_group.foo_group",
				ImportState:       true,
				ImportStateIdFunc: testAccAWSCodeDeployDeploymentGroupImportStateIdFunc("aws_codedeploy_deployment_group.foo_group"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSCodeDeployDeploymentGroup_alarmConfiguration_delete(t *testing.T) {
	var group codedeploy.DeploymentGroupInfo

	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeDeployDeploymentGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: test_config_alarm_configuration_create(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "alarm_configuration.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "alarm_configuration.0.enabled", "true"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "alarm_configuration.0.alarms.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "alarm_configuration.0.alarms.2356372769", "foo"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "alarm_configuration.0.ignore_poll_alarm_failure", "false"),
				),
			},
			{
				Config: test_config_alarm_configuration_delete(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "alarm_configuration.#", "0"),
				),
			},
			{
				ResourceName:      "aws_codedeploy_deployment_group.foo_group",
				ImportState:       true,
				ImportStateIdFunc: testAccAWSCodeDeployDeploymentGroupImportStateIdFunc("aws_codedeploy_deployment_group.foo_group"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSCodeDeployDeploymentGroup_alarmConfiguration_disable(t *testing.T) {
	var group codedeploy.DeploymentGroupInfo

	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeDeployDeploymentGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: test_config_alarm_configuration_create(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "alarm_configuration.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "alarm_configuration.0.enabled", "true"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "alarm_configuration.0.alarms.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "alarm_configuration.0.alarms.2356372769", "foo"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "alarm_configuration.0.ignore_poll_alarm_failure", "false"),
				),
			},
			{
				Config: test_config_alarm_configuration_disable(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "alarm_configuration.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "alarm_configuration.0.enabled", "false"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "alarm_configuration.0.alarms.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "alarm_configuration.0.alarms.2356372769", "foo"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "alarm_configuration.0.ignore_poll_alarm_failure", "false"),
				),
			},
			{
				ResourceName:      "aws_codedeploy_deployment_group.foo_group",
				ImportState:       true,
				ImportStateIdFunc: testAccAWSCodeDeployDeploymentGroupImportStateIdFunc("aws_codedeploy_deployment_group.foo_group"),
				ImportStateVerify: true,
			},
		},
	})
}

// When no configuration is provided, a deploymentStyle object with default values is computed
func TestAccAWSCodeDeployDeploymentGroup_deploymentStyle_default(t *testing.T) {
	var group codedeploy.DeploymentGroupInfo

	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeDeployDeploymentGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: test_config_deployment_style_default(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "deployment_style.#", "1"),
					resource.TestCheckResourceAttrSet(
						"aws_codedeploy_deployment_group.foo_group", "deployment_style.0.deployment_option"),
					resource.TestCheckResourceAttrSet(
						"aws_codedeploy_deployment_group.foo_group", "deployment_style.0.deployment_type"),
				),
			},
			{
				ResourceName:      "aws_codedeploy_deployment_group.foo_group",
				ImportState:       true,
				ImportStateIdFunc: testAccAWSCodeDeployDeploymentGroupImportStateIdFunc("aws_codedeploy_deployment_group.foo_group"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSCodeDeployDeploymentGroup_deploymentStyle_create(t *testing.T) {
	var group codedeploy.DeploymentGroupInfo

	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeDeployDeploymentGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: test_config_deployment_style_default(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "deployment_style.#", "1"),
					resource.TestCheckResourceAttrSet(
						"aws_codedeploy_deployment_group.foo_group", "deployment_style.0.deployment_option"),
					resource.TestCheckResourceAttrSet(
						"aws_codedeploy_deployment_group.foo_group", "deployment_style.0.deployment_type"),
				),
			},
			{
				Config: test_config_deployment_style_create(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "deployment_style.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "deployment_style.0.deployment_option", "WITH_TRAFFIC_CONTROL"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "deployment_style.0.deployment_type", "BLUE_GREEN"),

					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "load_balancer_info.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "load_balancer_info.0.elb_info.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "load_balancer_info.0.elb_info.2441772102.name", "foo-elb"),
				),
			},
			{
				ResourceName:      "aws_codedeploy_deployment_group.foo_group",
				ImportState:       true,
				ImportStateIdFunc: testAccAWSCodeDeployDeploymentGroupImportStateIdFunc("aws_codedeploy_deployment_group.foo_group"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSCodeDeployDeploymentGroup_deploymentStyle_update(t *testing.T) {
	var group codedeploy.DeploymentGroupInfo

	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeDeployDeploymentGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: test_config_deployment_style_create(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "deployment_style.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "deployment_style.0.deployment_option", "WITH_TRAFFIC_CONTROL"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "deployment_style.0.deployment_type", "BLUE_GREEN"),
				),
			},
			{
				Config: test_config_deployment_style_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "deployment_style.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "deployment_style.0.deployment_option", "WITHOUT_TRAFFIC_CONTROL"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "deployment_style.0.deployment_type", "IN_PLACE"),
				),
			},
			{
				ResourceName:      "aws_codedeploy_deployment_group.foo_group",
				ImportState:       true,
				ImportStateIdFunc: testAccAWSCodeDeployDeploymentGroupImportStateIdFunc("aws_codedeploy_deployment_group.foo_group"),
				ImportStateVerify: true,
			},
		},
	})
}

// Removing deployment_style from configuration does not trigger an update
// to the default state, but the previous state is instead retained...
func TestAccAWSCodeDeployDeploymentGroup_deploymentStyle_delete(t *testing.T) {
	var group codedeploy.DeploymentGroupInfo

	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeDeployDeploymentGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: test_config_deployment_style_create(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "deployment_style.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "deployment_style.0.deployment_option", "WITH_TRAFFIC_CONTROL"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "deployment_style.0.deployment_type", "BLUE_GREEN"),
				),
			},
			{
				Config: test_config_deployment_style_default(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "deployment_style.#", "1"),
					resource.TestCheckResourceAttrSet(
						"aws_codedeploy_deployment_group.foo_group", "deployment_style.0.deployment_option"),
					resource.TestCheckResourceAttrSet(
						"aws_codedeploy_deployment_group.foo_group", "deployment_style.0.deployment_type"),
				),
			},
			{
				ResourceName:      "aws_codedeploy_deployment_group.foo_group",
				ImportState:       true,
				ImportStateIdFunc: testAccAWSCodeDeployDeploymentGroupImportStateIdFunc("aws_codedeploy_deployment_group.foo_group"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSCodeDeployDeploymentGroup_loadBalancerInfo_create(t *testing.T) {
	var group codedeploy.DeploymentGroupInfo

	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeDeployDeploymentGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: test_config_load_balancer_info_default(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "load_balancer_info.#", "0"),
				),
			},
			{
				Config: test_config_load_balancer_info_create(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "load_balancer_info.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "load_balancer_info.0.elb_info.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "load_balancer_info.0.elb_info.2441772102.name", "foo-elb"),
				),
			},
			{
				ResourceName:      "aws_codedeploy_deployment_group.foo_group",
				ImportState:       true,
				ImportStateIdFunc: testAccAWSCodeDeployDeploymentGroupImportStateIdFunc("aws_codedeploy_deployment_group.foo_group"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSCodeDeployDeploymentGroup_loadBalancerInfo_update(t *testing.T) {
	var group codedeploy.DeploymentGroupInfo

	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeDeployDeploymentGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: test_config_load_balancer_info_create(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "load_balancer_info.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "load_balancer_info.0.elb_info.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "load_balancer_info.0.elb_info.2441772102.name", "foo-elb"),
				),
			},
			{
				Config: test_config_load_balancer_info_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "load_balancer_info.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "load_balancer_info.0.elb_info.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "load_balancer_info.0.elb_info.4206303396.name", "bar-elb"),
				),
			},
			{
				ResourceName:      "aws_codedeploy_deployment_group.foo_group",
				ImportState:       true,
				ImportStateIdFunc: testAccAWSCodeDeployDeploymentGroupImportStateIdFunc("aws_codedeploy_deployment_group.foo_group"),
				ImportStateVerify: true,
			},
		},
	})
}

// Without "Computed: true" on load_balancer_info, removing the resource
// from configuration causes an error, because the remote resource still exists.
func TestAccAWSCodeDeployDeploymentGroup_loadBalancerInfo_delete(t *testing.T) {
	var group codedeploy.DeploymentGroupInfo

	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeDeployDeploymentGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: test_config_load_balancer_info_create(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "load_balancer_info.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "load_balancer_info.0.elb_info.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "load_balancer_info.0.elb_info.2441772102.name", "foo-elb"),
				),
			},
			{
				Config: test_config_load_balancer_info_delete(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "load_balancer_info.#", "1"),
				),
			},
			{
				ResourceName:      "aws_codedeploy_deployment_group.foo_group",
				ImportState:       true,
				ImportStateIdFunc: testAccAWSCodeDeployDeploymentGroupImportStateIdFunc("aws_codedeploy_deployment_group.foo_group"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSCodeDeployDeploymentGroup_loadBalancerInfo_targetGroupInfo_create(t *testing.T) {
	var group codedeploy.DeploymentGroupInfo

	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeDeployDeploymentGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: test_config_load_balancer_info_default(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "load_balancer_info.#", "0"),
				),
			},
			{
				Config: test_config_load_balancer_info_target_group_info_create(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "load_balancer_info.#", "1"),

					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "load_balancer_info.0.target_group_info.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "load_balancer_info.0.target_group_info.4178177480.name", "foo-tg"),
				),
			},
			{
				ResourceName:      "aws_codedeploy_deployment_group.foo_group",
				ImportState:       true,
				ImportStateIdFunc: testAccAWSCodeDeployDeploymentGroupImportStateIdFunc("aws_codedeploy_deployment_group.foo_group"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSCodeDeployDeploymentGroup_loadBalancerInfo_targetGroupInfo_update(t *testing.T) {
	var group codedeploy.DeploymentGroupInfo

	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeDeployDeploymentGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: test_config_load_balancer_info_target_group_info_create(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "load_balancer_info.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "load_balancer_info.0.target_group_info.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "load_balancer_info.0.target_group_info.4178177480.name", "foo-tg"),
				),
			},
			{
				Config: test_config_load_balancer_info_target_group_info_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "load_balancer_info.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "load_balancer_info.0.target_group_info.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "load_balancer_info.0.target_group_info.2940009368.name", "bar-tg"),
				),
			},
			{
				ResourceName:      "aws_codedeploy_deployment_group.foo_group",
				ImportState:       true,
				ImportStateIdFunc: testAccAWSCodeDeployDeploymentGroupImportStateIdFunc("aws_codedeploy_deployment_group.foo_group"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSCodeDeployDeploymentGroup_loadBalancerInfo_targetGroupInfo_delete(t *testing.T) {
	var group codedeploy.DeploymentGroupInfo

	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeDeployDeploymentGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: test_config_load_balancer_info_target_group_info_create(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "load_balancer_info.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "load_balancer_info.0.target_group_info.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "load_balancer_info.0.target_group_info.4178177480.name", "foo-tg"),
				),
			},
			{
				Config: test_config_load_balancer_info_target_group_info_delete(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "load_balancer_info.#", "1"),
				),
			},
			{
				ResourceName:      "aws_codedeploy_deployment_group.foo_group",
				ImportState:       true,
				ImportStateIdFunc: testAccAWSCodeDeployDeploymentGroupImportStateIdFunc("aws_codedeploy_deployment_group.foo_group"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSCodeDeployDeploymentGroup_in_place_deployment_with_traffic_control_create(t *testing.T) {
	var group codedeploy.DeploymentGroupInfo

	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeDeployDeploymentGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: test_config_deployment_style_default(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "deployment_style.#", "1"),
					resource.TestCheckResourceAttrSet(
						"aws_codedeploy_deployment_group.foo_group", "deployment_style.0.deployment_option"),
					resource.TestCheckResourceAttrSet(
						"aws_codedeploy_deployment_group.foo_group", "deployment_style.0.deployment_type"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "load_balancer_info.#", "0"),
				),
			},
			{
				Config: test_config_in_place_deployment_with_traffic_control_create(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "deployment_style.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "deployment_style.0.deployment_option", "WITH_TRAFFIC_CONTROL"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "deployment_style.0.deployment_type", "IN_PLACE"),

					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "load_balancer_info.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "load_balancer_info.0.elb_info.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "load_balancer_info.0.elb_info.2441772102.name", "foo-elb"),
				),
			},
			{
				ResourceName:      "aws_codedeploy_deployment_group.foo_group",
				ImportState:       true,
				ImportStateIdFunc: testAccAWSCodeDeployDeploymentGroupImportStateIdFunc("aws_codedeploy_deployment_group.foo_group"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSCodeDeployDeploymentGroup_in_place_deployment_with_traffic_control_update(t *testing.T) {
	var group codedeploy.DeploymentGroupInfo

	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeDeployDeploymentGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: test_config_in_place_deployment_with_traffic_control_create(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "deployment_style.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "deployment_style.0.deployment_option", "WITH_TRAFFIC_CONTROL"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "deployment_style.0.deployment_type", "IN_PLACE"),

					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "load_balancer_info.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "load_balancer_info.0.elb_info.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "load_balancer_info.0.elb_info.2441772102.name", "foo-elb"),
				),
			},
			{
				Config: test_config_in_place_deployment_with_traffic_control_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "deployment_style.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "deployment_style.0.deployment_option", "WITH_TRAFFIC_CONTROL"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "deployment_style.0.deployment_type", "BLUE_GREEN"),

					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "load_balancer_info.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "load_balancer_info.0.elb_info.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "load_balancer_info.0.elb_info.2441772102.name", "foo-elb"),

					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.deployment_ready_option.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.deployment_ready_option.0.action_on_timeout", "CONTINUE_DEPLOYMENT"),

					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.green_fleet_provisioning_option.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.green_fleet_provisioning_option.0.action", "DISCOVER_EXISTING"),

					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.terminate_blue_instances_on_deployment_success.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.terminate_blue_instances_on_deployment_success.0.action", "KEEP_ALIVE"),
				),
			},
			{
				ResourceName:      "aws_codedeploy_deployment_group.foo_group",
				ImportState:       true,
				ImportStateIdFunc: testAccAWSCodeDeployDeploymentGroupImportStateIdFunc("aws_codedeploy_deployment_group.foo_group"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSCodeDeployDeploymentGroup_blueGreenDeploymentConfiguration_create(t *testing.T) {
	var group codedeploy.DeploymentGroupInfo

	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeDeployDeploymentGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: test_config_blue_green_deployment_config_default(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.#", "0"),
				),
			},
			{
				Config: test_config_blue_green_deployment_config_create_with_asg(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.#", "1"),

					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.deployment_ready_option.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.deployment_ready_option.0.action_on_timeout", "STOP_DEPLOYMENT"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.deployment_ready_option.0.wait_time_in_minutes", "60"),

					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.green_fleet_provisioning_option.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.green_fleet_provisioning_option.0.action", "COPY_AUTO_SCALING_GROUP"),

					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.terminate_blue_instances_on_deployment_success.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.terminate_blue_instances_on_deployment_success.0.action", "TERMINATE"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.terminate_blue_instances_on_deployment_success.0.termination_wait_time_in_minutes", "120"),
				),
			},
			{
				ResourceName:      "aws_codedeploy_deployment_group.foo_group",
				ImportState:       true,
				ImportStateIdFunc: testAccAWSCodeDeployDeploymentGroupImportStateIdFunc("aws_codedeploy_deployment_group.foo_group"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSCodeDeployDeploymentGroup_blueGreenDeploymentConfiguration_update_with_asg(t *testing.T) {
	var group codedeploy.DeploymentGroupInfo

	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeDeployDeploymentGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: test_config_blue_green_deployment_config_create_with_asg(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.#", "1"),

					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.deployment_ready_option.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.deployment_ready_option.0.action_on_timeout", "STOP_DEPLOYMENT"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.deployment_ready_option.0.wait_time_in_minutes", "60"),

					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.green_fleet_provisioning_option.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.green_fleet_provisioning_option.0.action", "COPY_AUTO_SCALING_GROUP"),

					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.terminate_blue_instances_on_deployment_success.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.terminate_blue_instances_on_deployment_success.0.action", "TERMINATE"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.terminate_blue_instances_on_deployment_success.0.termination_wait_time_in_minutes", "120"),
				),
			},
			{
				Config: test_config_blue_green_deployment_config_update_with_asg(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.#", "1"),

					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.deployment_ready_option.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.deployment_ready_option.0.action_on_timeout", "STOP_DEPLOYMENT"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.deployment_ready_option.0.wait_time_in_minutes", "60"),

					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.green_fleet_provisioning_option.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.green_fleet_provisioning_option.0.action", "COPY_AUTO_SCALING_GROUP"),

					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.terminate_blue_instances_on_deployment_success.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.terminate_blue_instances_on_deployment_success.0.action", "KEEP_ALIVE"),
				),
			},
		},
	})
}

func TestAccAWSCodeDeployDeploymentGroup_blueGreenDeploymentConfiguration_update(t *testing.T) {
	var group codedeploy.DeploymentGroupInfo

	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeDeployDeploymentGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: test_config_blue_green_deployment_config_create_no_asg(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.#", "1"),

					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.deployment_ready_option.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.deployment_ready_option.0.action_on_timeout", "STOP_DEPLOYMENT"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.deployment_ready_option.0.wait_time_in_minutes", "60"),

					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.green_fleet_provisioning_option.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.green_fleet_provisioning_option.0.action", "DISCOVER_EXISTING"),

					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.terminate_blue_instances_on_deployment_success.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.terminate_blue_instances_on_deployment_success.0.action", "TERMINATE"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.terminate_blue_instances_on_deployment_success.0.termination_wait_time_in_minutes", "120"),
				),
			},
			{
				Config: test_config_blue_green_deployment_config_update_no_asg(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.#", "1"),

					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.deployment_ready_option.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.deployment_ready_option.0.action_on_timeout", "CONTINUE_DEPLOYMENT"),

					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.green_fleet_provisioning_option.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.green_fleet_provisioning_option.0.action", "DISCOVER_EXISTING"),

					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.terminate_blue_instances_on_deployment_success.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.terminate_blue_instances_on_deployment_success.0.action", "KEEP_ALIVE"),
				),
			},
		},
	})
}

// Without "Computed: true" on blue_green_deployment_config, removing the resource
// from configuration causes an error, because the remote resource still exists.
func TestAccAWSCodeDeployDeploymentGroup_blueGreenDeploymentConfiguration_delete(t *testing.T) {
	var group codedeploy.DeploymentGroupInfo

	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeDeployDeploymentGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: test_config_blue_green_deployment_config_create_no_asg(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.#", "1"),

					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.deployment_ready_option.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.deployment_ready_option.0.action_on_timeout", "STOP_DEPLOYMENT"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.deployment_ready_option.0.wait_time_in_minutes", "60"),

					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.green_fleet_provisioning_option.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.green_fleet_provisioning_option.0.action", "DISCOVER_EXISTING"),

					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.terminate_blue_instances_on_deployment_success.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.terminate_blue_instances_on_deployment_success.0.action", "TERMINATE"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.terminate_blue_instances_on_deployment_success.0.termination_wait_time_in_minutes", "120"),
				),
			},
			{
				Config: test_config_blue_green_deployment_config_default(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.#", "1"),
				),
			},
			{
				ResourceName:      "aws_codedeploy_deployment_group.foo_group",
				ImportState:       true,
				ImportStateIdFunc: testAccAWSCodeDeployDeploymentGroupImportStateIdFunc("aws_codedeploy_deployment_group.foo_group"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSCodeDeployDeploymentGroup_blueGreenDeployment_complete(t *testing.T) {
	var group codedeploy.DeploymentGroupInfo

	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeDeployDeploymentGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: test_config_blue_green_deployment_complete(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),

					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "deployment_style.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "deployment_style.0.deployment_option", "WITH_TRAFFIC_CONTROL"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "deployment_style.0.deployment_type", "BLUE_GREEN"),

					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "load_balancer_info.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "load_balancer_info.0.elb_info.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "load_balancer_info.0.elb_info.2441772102.name", "foo-elb"),

					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.#", "1"),

					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.deployment_ready_option.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.deployment_ready_option.0.action_on_timeout", "STOP_DEPLOYMENT"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.deployment_ready_option.0.wait_time_in_minutes", "60"),

					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.green_fleet_provisioning_option.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.green_fleet_provisioning_option.0.action", "DISCOVER_EXISTING"),

					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.terminate_blue_instances_on_deployment_success.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.terminate_blue_instances_on_deployment_success.0.action", "KEEP_ALIVE"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.terminate_blue_instances_on_deployment_success.0.termination_wait_time_in_minutes", "0"),
				),
			},
			{
				Config: test_config_blue_green_deployment_complete_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists("aws_codedeploy_deployment_group.foo_group", &group),

					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "deployment_style.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "deployment_style.0.deployment_option", "WITH_TRAFFIC_CONTROL"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "deployment_style.0.deployment_type", "BLUE_GREEN"),

					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "load_balancer_info.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "load_balancer_info.0.elb_info.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "load_balancer_info.0.elb_info.2441772102.name", "foo-elb"),

					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.#", "1"),

					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.deployment_ready_option.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.deployment_ready_option.0.action_on_timeout", "CONTINUE_DEPLOYMENT"),

					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.green_fleet_provisioning_option.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.green_fleet_provisioning_option.0.action", "DISCOVER_EXISTING"),

					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.terminate_blue_instances_on_deployment_success.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.terminate_blue_instances_on_deployment_success.0.action", "KEEP_ALIVE"),
					resource.TestCheckResourceAttr(
						"aws_codedeploy_deployment_group.foo_group", "blue_green_deployment_config.0.terminate_blue_instances_on_deployment_success.0.termination_wait_time_in_minutes", "0"),
				),
			},
			{
				ResourceName:      "aws_codedeploy_deployment_group.foo_group",
				ImportState:       true,
				ImportStateIdFunc: testAccAWSCodeDeployDeploymentGroupImportStateIdFunc("aws_codedeploy_deployment_group.foo_group"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSCodeDeployDeploymentGroup_ECS_BlueGreen(t *testing.T) {
	var group codedeploy.DeploymentGroupInfo
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))
	ecsClusterResourceName := "aws_ecs_cluster.test"
	ecsServiceResourceName := "aws_ecs_service.test"
	lbTargetGroupBlueResourceName := "aws_lb_target_group.blue"
	lbTargetGroupGreenResourceName := "aws_lb_target_group.green"
	resourceName := "aws_codedeploy_deployment_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeDeployDeploymentGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeDeployDeploymentGroupConfigEcsBlueGreen(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployDeploymentGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "ecs_service.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "ecs_service.0.cluster_name", ecsClusterResourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "ecs_service.0.service_name", ecsServiceResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_info.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_info.0.target_group_pair_info.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_info.0.target_group_pair_info.0.prod_traffic_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_info.0.target_group_pair_info.0.prod_traffic_route.0.listener_arns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_info.0.target_group_pair_info.0.target_group.#", "2"),
					resource.TestCheckResourceAttrPair(resourceName, "load_balancer_info.0.target_group_pair_info.0.target_group.0.name", lbTargetGroupBlueResourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "load_balancer_info.0.target_group_pair_info.0.target_group.1.name", lbTargetGroupGreenResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_info.0.target_group_pair_info.0.test_traffic_route.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSCodeDeployDeploymentGroupImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAWSCodeDeployDeploymentGroup_buildTriggerConfigs(t *testing.T) {
	input := []interface{}{
		map[string]interface{}{
			"trigger_events": schema.NewSet(schema.HashString, []interface{}{
				"DeploymentFailure",
			}),
			"trigger_name":       "foo-trigger",
			"trigger_target_arn": "arn:aws:sns:us-west-2:123456789012:foo-topic",
		},
	}

	expected := []*codedeploy.TriggerConfig{
		{
			TriggerEvents: []*string{
				aws.String("DeploymentFailure"),
			},
			TriggerName:      aws.String("foo-trigger"),
			TriggerTargetArn: aws.String("arn:aws:sns:us-west-2:123456789012:foo-topic"),
		},
	}

	actual := buildTriggerConfigs(input)

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("buildTriggerConfigs output is not correct.\nGot:\n%#v\nExpected:\n%#v\n",
			actual, expected)
	}
}

func TestAWSCodeDeployDeploymentGroup_triggerConfigsToMap(t *testing.T) {
	input := []*codedeploy.TriggerConfig{
		{
			TriggerEvents: []*string{
				aws.String("DeploymentFailure"),
				aws.String("InstanceFailure"),
			},
			TriggerName:      aws.String("bar-trigger"),
			TriggerTargetArn: aws.String("arn:aws:sns:us-west-2:123456789012:bar-topic"),
		},
	}

	expected := map[string]interface{}{
		"trigger_events": schema.NewSet(schema.HashString, []interface{}{
			"DeploymentFailure",
			"InstanceFailure",
		}),
		"trigger_name":       "bar-trigger",
		"trigger_target_arn": "arn:aws:sns:us-west-2:123456789012:bar-topic",
	}

	actual := triggerConfigsToMap(input)[0]

	fatal := false

	if actual["trigger_name"] != expected["trigger_name"] {
		fatal = true
	}

	if actual["trigger_target_arn"] != expected["trigger_target_arn"] {
		fatal = true
	}

	actualEvents := actual["trigger_events"].(*schema.Set)
	expectedEvents := expected["trigger_events"].(*schema.Set)
	if !actualEvents.Equal(expectedEvents) {
		fatal = true
	}

	if fatal {
		t.Fatalf("triggerConfigsToMap output is not correct.\nGot:\n%#v\nExpected:\n%#v\n",
			actual, expected)
	}
}

func TestAWSCodeDeployDeploymentGroup_buildAutoRollbackConfig(t *testing.T) {
	input := []interface{}{
		map[string]interface{}{
			"events": schema.NewSet(schema.HashString, []interface{}{
				"DEPLOYMENT_FAILURE",
			}),
			"enabled": true,
		},
	}

	expected := &codedeploy.AutoRollbackConfiguration{
		Events: []*string{
			aws.String("DEPLOYMENT_FAILURE"),
		},
		Enabled: aws.Bool(true),
	}

	actual := buildAutoRollbackConfig(input)

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("buildAutoRollbackConfig output is not correct.\nGot:\n%#v\nExpected:\n%#v\n",
			actual, expected)
	}
}

func TestAWSCodeDeployDeploymentGroup_autoRollbackConfigToMap(t *testing.T) {
	input := &codedeploy.AutoRollbackConfiguration{
		Events: []*string{
			aws.String("DEPLOYMENT_FAILURE"),
			aws.String("DEPLOYMENT_STOP_ON_ALARM"),
		},
		Enabled: aws.Bool(false),
	}

	expected := map[string]interface{}{
		"events": schema.NewSet(schema.HashString, []interface{}{
			"DEPLOYMENT_FAILURE",
			"DEPLOYMENT_STOP_ON_ALARM",
		}),
		"enabled": false,
	}

	actual := autoRollbackConfigToMap(input)[0]

	fatal := false

	if actual["enabled"] != expected["enabled"] {
		fatal = true
	}

	actualEvents := actual["events"].(*schema.Set)
	expectedEvents := expected["events"].(*schema.Set)
	if !actualEvents.Equal(expectedEvents) {
		fatal = true
	}

	if fatal {
		t.Fatalf("autoRollbackConfigToMap output is not correct.\nGot:\n%#v\nExpected:\n%#v\n",
			actual, expected)
	}
}

func TestAWSCodeDeployDeploymentGroup_expandDeploymentStyle(t *testing.T) {
	input := []interface{}{
		map[string]interface{}{
			"deployment_option": "WITH_TRAFFIC_CONTROL",
			"deployment_type":   "BLUE_GREEN",
		},
	}

	expected := &codedeploy.DeploymentStyle{
		DeploymentOption: aws.String("WITH_TRAFFIC_CONTROL"),
		DeploymentType:   aws.String("BLUE_GREEN"),
	}

	actual := expandDeploymentStyle(input)

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expandDeploymentStyle output is not correct.\nGot:\n%#v\nExpected:\n%#v\n",
			actual, expected)
	}
}

func TestAWSCodeDeployDeploymentGroup_flattenDeploymentStyle(t *testing.T) {
	expected := map[string]interface{}{
		"deployment_option": "WITHOUT_TRAFFIC_CONTROL",
		"deployment_type":   "IN_PLACE",
	}

	input := &codedeploy.DeploymentStyle{
		DeploymentOption: aws.String("WITHOUT_TRAFFIC_CONTROL"),
		DeploymentType:   aws.String("IN_PLACE"),
	}

	actual := flattenDeploymentStyle(input)[0]

	fatal := false

	if actual["deployment_option"] != expected["deployment_option"] {
		fatal = true
	}

	if actual["deployment_type"] != expected["deployment_type"] {
		fatal = true
	}

	if fatal {
		t.Fatalf("flattenDeploymentStyle output is not correct.\nGot:\n%#v\nExpected:\n%#v\n",
			actual, expected)
	}
}

func TestAWSCodeDeployDeploymentGroup_expandLoadBalancerInfo(t *testing.T) {
	input := []interface{}{
		map[string]interface{}{
			"elb_info": schema.NewSet(loadBalancerInfoHash, []interface{}{
				map[string]interface{}{
					"name": "foo-elb",
				},
				map[string]interface{}{
					"name": "bar-elb",
				},
			}),
		},
	}

	expected := &codedeploy.LoadBalancerInfo{
		ElbInfoList: []*codedeploy.ELBInfo{
			{
				Name: aws.String("foo-elb"),
			},
			{
				Name: aws.String("bar-elb"),
			},
		},
	}

	actual := expandLoadBalancerInfo(input)

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expandLoadBalancerInfo output is not correct.\nGot:\n%#v\nExpected:\n%#v\n",
			actual, expected)
	}
}

func TestAWSCodeDeployDeploymentGroup_flattenLoadBalancerInfo(t *testing.T) {
	input := &codedeploy.LoadBalancerInfo{
		TargetGroupInfoList: []*codedeploy.TargetGroupInfo{
			{
				Name: aws.String("abc-tg"),
			},
			{
				Name: aws.String("xyz-tg"),
			},
		},
	}

	expected := map[string]interface{}{
		"target_group_info": schema.NewSet(loadBalancerInfoHash, []interface{}{
			map[string]interface{}{
				"name": "abc-tg",
			},
			map[string]interface{}{
				"name": "xyz-tg",
			},
		}),
	}

	actual := flattenLoadBalancerInfo(input)[0].(map[string]interface{})

	fatal := false

	a := actual["target_group_info"].(*schema.Set)
	e := expected["target_group_info"].(*schema.Set)
	if !a.Equal(e) {
		fatal = true
	}

	if fatal {
		t.Fatalf("flattenLoadBalancerInfo output is not correct.\nGot:\n%#v\nExpected:\n%#v\n",
			actual, expected)
	}
}

func TestAWSCodeDeployDeploymentGroup_expandBlueGreenDeploymentConfig(t *testing.T) {
	input := []interface{}{
		map[string]interface{}{
			"deployment_ready_option": []interface{}{
				map[string]interface{}{
					"action_on_timeout":    "CONTINUE_DEPLOYMENT",
					"wait_time_in_minutes": 60,
				},
			},

			"green_fleet_provisioning_option": []interface{}{
				map[string]interface{}{
					"action": "COPY_AUTO_SCALING_GROUP",
				},
			},

			"terminate_blue_instances_on_deployment_success": []interface{}{
				map[string]interface{}{
					"action":                           "TERMINATE",
					"termination_wait_time_in_minutes": 90,
				},
			},
		},
	}

	expected := &codedeploy.BlueGreenDeploymentConfiguration{
		DeploymentReadyOption: &codedeploy.DeploymentReadyOption{
			ActionOnTimeout:   aws.String("CONTINUE_DEPLOYMENT"),
			WaitTimeInMinutes: aws.Int64(60),
		},

		GreenFleetProvisioningOption: &codedeploy.GreenFleetProvisioningOption{
			Action: aws.String("COPY_AUTO_SCALING_GROUP"),
		},

		TerminateBlueInstancesOnDeploymentSuccess: &codedeploy.BlueInstanceTerminationOption{
			Action:                       aws.String("TERMINATE"),
			TerminationWaitTimeInMinutes: aws.Int64(90),
		},
	}

	actual := expandBlueGreenDeploymentConfig(input)

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expandBlueGreenDeploymentConfig output is not correct.\nGot:\n%#v\nExpected:\n%#v\n",
			actual, expected)
	}
}

func TestAWSCodeDeployDeploymentGroup_flattenBlueGreenDeploymentConfig(t *testing.T) {
	input := &codedeploy.BlueGreenDeploymentConfiguration{
		DeploymentReadyOption: &codedeploy.DeploymentReadyOption{
			ActionOnTimeout:   aws.String("STOP_DEPLOYMENT"),
			WaitTimeInMinutes: aws.Int64(120),
		},

		GreenFleetProvisioningOption: &codedeploy.GreenFleetProvisioningOption{
			Action: aws.String("DISCOVER_EXISTING"),
		},

		TerminateBlueInstancesOnDeploymentSuccess: &codedeploy.BlueInstanceTerminationOption{
			Action:                       aws.String("KEEP_ALIVE"),
			TerminationWaitTimeInMinutes: aws.Int64(90),
		},
	}

	expected := map[string]interface{}{
		"deployment_ready_option": []map[string]interface{}{
			{
				"action_on_timeout":    "STOP_DEPLOYMENT",
				"wait_time_in_minutes": 120,
			},
		},

		"green_fleet_provisioning_option": []map[string]interface{}{
			{
				"action": "DISCOVER_EXISTING",
			},
		},

		"terminate_blue_instances_on_deployment_success": []map[string]interface{}{
			{
				"action":                           "KEEP_ALIVE",
				"termination_wait_time_in_minutes": 90,
			},
		},
	}

	actual := flattenBlueGreenDeploymentConfig(input)[0]

	fatal := false

	a := actual["deployment_ready_option"].([]map[string]interface{})[0]
	if a["action_on_timeout"].(string) != "STOP_DEPLOYMENT" {
		fatal = true
	}

	if a["wait_time_in_minutes"].(int64) != 120 {
		fatal = true
	}

	b := actual["green_fleet_provisioning_option"].([]map[string]interface{})[0]
	if b["action"].(string) != "DISCOVER_EXISTING" {
		fatal = true
	}

	c := actual["terminate_blue_instances_on_deployment_success"].([]map[string]interface{})[0]
	if c["action"].(string) != "KEEP_ALIVE" {
		fatal = true
	}

	if c["termination_wait_time_in_minutes"].(int64) != 90 {
		fatal = true
	}

	if fatal {
		t.Fatalf("flattenBlueGreenDeploymentConfig output is not correct.\nGot:\n%#v\nExpected:\n%#v\n",
			actual, expected)
	}
}

func TestAWSCodeDeployDeploymentGroup_buildAlarmConfig(t *testing.T) {
	input := []interface{}{
		map[string]interface{}{
			"alarms": schema.NewSet(schema.HashString, []interface{}{
				"foo-alarm",
			}),
			"enabled":                   true,
			"ignore_poll_alarm_failure": false,
		},
	}

	expected := &codedeploy.AlarmConfiguration{
		Alarms: []*codedeploy.Alarm{
			{
				Name: aws.String("foo-alarm"),
			},
		},
		Enabled:                aws.Bool(true),
		IgnorePollAlarmFailure: aws.Bool(false),
	}

	actual := buildAlarmConfig(input)

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("buildAlarmConfig output is not correct.\nGot:\n%#v\nExpected:\n%#v\n",
			actual, expected)
	}
}

func TestAWSCodeDeployDeploymentGroup_alarmConfigToMap(t *testing.T) {
	input := &codedeploy.AlarmConfiguration{
		Alarms: []*codedeploy.Alarm{
			{
				Name: aws.String("bar-alarm"),
			},
			{
				Name: aws.String("foo-alarm"),
			},
		},
		Enabled:                aws.Bool(false),
		IgnorePollAlarmFailure: aws.Bool(true),
	}

	expected := map[string]interface{}{
		"alarms": schema.NewSet(schema.HashString, []interface{}{
			"bar-alarm",
			"foo-alarm",
		}),
		"enabled":                   false,
		"ignore_poll_alarm_failure": true,
	}

	actual := alarmConfigToMap(input)[0]

	fatal := false

	if actual["enabled"] != expected["enabled"] {
		fatal = true
	}

	if actual["ignore_poll_alarm_failure"] != expected["ignore_poll_alarm_failure"] {
		fatal = true
	}

	actualAlarms := actual["alarms"].(*schema.Set)
	expectedAlarms := expected["alarms"].(*schema.Set)
	if !actualAlarms.Equal(expectedAlarms) {
		fatal = true
	}

	if fatal {
		t.Fatalf("alarmConfigToMap output is not correct.\nGot:\n%#v\nExpected:\n%#v\n",
			actual, expected)
	}
}

func testAccCheckTriggerEvents(group *codedeploy.DeploymentGroupInfo, triggerName string, expectedEvents []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		for _, actual := range group.TriggerConfigurations {
			if *actual.TriggerName == triggerName {

				numberOfEvents := len(actual.TriggerEvents)
				if numberOfEvents != len(expectedEvents) {
					return fmt.Errorf("Trigger events do not match. Expected: %d. Got: %d.",
						len(expectedEvents), numberOfEvents)
				}

				actualEvents := make([]string, 0, numberOfEvents)
				for _, event := range actual.TriggerEvents {
					actualEvents = append(actualEvents, *event)
				}
				sort.Strings(actualEvents)

				if !reflect.DeepEqual(actualEvents, expectedEvents) {
					return fmt.Errorf("Trigger events do not match.\nExpected: %v\nGot: %v\n",
						expectedEvents, actualEvents)
				}
				break
			}
		}
		return nil
	}
}

func testAccCheckTriggerTargetArn(group *codedeploy.DeploymentGroupInfo, triggerName string, r *regexp.Regexp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, actual := range group.TriggerConfigurations {
			if *actual.TriggerName == triggerName {
				if !r.MatchString(*actual.TriggerTargetArn) {
					return fmt.Errorf("Trigger target arn does not match regular expression.\nRegex: %v\nTriggerTargetArn: %v\n",
						r, *actual.TriggerTargetArn)
				}
				break
			}
		}
		return nil
	}
}

func testAccCheckAWSCodeDeployDeploymentGroupDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).codedeployconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_codedeploy_deployment_group" {
			continue
		}

		resp, err := conn.GetDeploymentGroup(&codedeploy.GetDeploymentGroupInput{
			ApplicationName:     aws.String(rs.Primary.Attributes["app_name"]),
			DeploymentGroupName: aws.String(rs.Primary.Attributes["deployment_group_name"]),
		})

		if ae, ok := err.(awserr.Error); ok && ae.Code() == "ApplicationDoesNotExistException" {
			continue
		}

		if err == nil {
			if resp.DeploymentGroupInfo.DeploymentGroupName != nil {
				return fmt.Errorf("CodeDeploy deployment group still exists:\n%#v", *resp.DeploymentGroupInfo.DeploymentGroupName)
			}
		}

		return err
	}

	return nil
}

func testAccAWSCodeDeployDeploymentGroupDisappears(group *codedeploy.DeploymentGroupInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).codedeployconn
		opts := &codedeploy.DeleteDeploymentGroupInput{
			ApplicationName:     group.ApplicationName,
			DeploymentGroupName: group.DeploymentGroupName,
		}
		if _, err := conn.DeleteDeploymentGroup(opts); err != nil {
			return err
		}
		return resource.Retry(40*time.Minute, func() *resource.RetryError {
			opts := &codedeploy.GetDeploymentGroupInput{
				ApplicationName:     group.ApplicationName,
				DeploymentGroupName: group.DeploymentGroupName,
			}
			_, err := conn.GetDeploymentGroup(opts)
			if err != nil {
				codedeploy, ok := err.(awserr.Error)
				if ok && codedeploy.Code() == "DeploymentGroupDoesNotExistException" {
					return nil
				}
				return resource.NonRetryableError(
					fmt.Errorf("Error retrieving CodeDeploy Deployment Group: %s", err))
			}
			return resource.RetryableError(fmt.Errorf(
				"Waiting for CodeDeploy Deployment Group: %v", group.DeploymentGroupName))
		})
	}
}

func testAccCheckAWSCodeDeployDeploymentGroupExists(name string, group *codedeploy.DeploymentGroupInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).codedeployconn

		resp, err := conn.GetDeploymentGroup(&codedeploy.GetDeploymentGroupInput{
			ApplicationName:     aws.String(rs.Primary.Attributes["app_name"]),
			DeploymentGroupName: aws.String(rs.Primary.Attributes["deployment_group_name"]),
		})

		if err != nil {
			return err
		}

		*group = *resp.DeploymentGroupInfo

		return nil
	}
}

func TestAWSCodeDeployDeploymentGroup_handleCodeDeployApiError(t *testing.T) {
	notAnAwsError := errors.New("Not an awserr")
	invalidRoleException := awserr.New("InvalidRoleException", "Invalid role exception", nil)
	invalidTriggerConfigExceptionNoMatch := awserr.New("InvalidTriggerConfigException", "Some other error message", nil)
	invalidTriggerConfigExceptionMatch := awserr.New("InvalidTriggerConfigException", "Topic ARN foo is not valid", nil)
	fakeAwsError := awserr.New("FakeAwsException", "Not a real AWS error", nil)

	testCases := []struct {
		Input    error
		Expected *resource.RetryError
	}{
		{
			Input:    nil,
			Expected: nil,
		},
		{
			Input:    notAnAwsError,
			Expected: resource.NonRetryableError(notAnAwsError),
		},
		{
			Input:    invalidRoleException,
			Expected: resource.RetryableError(invalidRoleException),
		},
		{
			Input:    invalidTriggerConfigExceptionNoMatch,
			Expected: resource.NonRetryableError(invalidTriggerConfigExceptionNoMatch),
		},
		{
			Input:    invalidTriggerConfigExceptionMatch,
			Expected: resource.RetryableError(invalidTriggerConfigExceptionMatch),
		},
		{
			Input:    fakeAwsError,
			Expected: resource.NonRetryableError(fakeAwsError),
		},
	}

	for _, tc := range testCases {
		actual := handleCodeDeployApiError(tc.Input, "test")
		if !reflect.DeepEqual(actual, tc.Expected) {
			t.Fatalf(`handleCodeDeployApiError output is not correct. Expected: %v, Actual: %v.`,
				tc.Expected, actual)
		}
	}
}

func testAccAWSCodeDeployDeploymentGroupImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s:%s", rs.Primary.Attributes["app_name"], rs.Primary.Attributes["deployment_group_name"]), nil
	}
}

func testAccAWSCodeDeployDeploymentGroup(rName string, tagGroup bool) string {
	var tagGroupOrFilter string
	if tagGroup {
		tagGroupOrFilter = `ec2_tag_set {
    ec2_tag_filter {
      key = "filterkey"
      type = "KEY_AND_VALUE"
      value = "filtervalue"
    }
  }
`
	} else {
		tagGroupOrFilter = `ec2_tag_filter {
    key = "filterkey"
    type = "KEY_AND_VALUE"
    value = "filtervalue"
  }
`
	}

	return fmt.Sprintf(`
resource "aws_codedeploy_app" "foo_app" {
  name = "foo_app_%s"
}

resource "aws_iam_role_policy" "foo_policy" {
  name = "foo_policy_%s"
  role = "${aws_iam_role.foo_role.id}"
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "autoscaling:CompleteLifecycleAction",
        "autoscaling:DeleteLifecycleHook",
        "autoscaling:DescribeAutoScalingGroups",
        "autoscaling:DescribeLifecycleHooks",
        "autoscaling:PutLifecycleHook",
        "autoscaling:RecordLifecycleActionHeartbeat",
        "ec2:DescribeInstances",
        "ec2:DescribeInstanceStatus",
        "tag:GetTags",
        "tag:GetResources"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_role" "foo_role" {
  name = "foo_role_%s"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "codedeploy.amazonaws.com"
        ]
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_codedeploy_deployment_group" "foo" {
  app_name = "${aws_codedeploy_app.foo_app.name}"
  deployment_group_name = "foo_%s"
  service_role_arn = "${aws_iam_role.foo_role.arn}"
  %s
}`, rName, rName, rName, rName, tagGroupOrFilter)
}

func testAccAWSCodeDeployDeploymentGroupModified(rName string, tagGroup bool) string {
	var tagGroupOrFilter string
	if tagGroup {
		tagGroupOrFilter = `ec2_tag_set {
    ec2_tag_filter {
      key = "filterkey"
      type = "KEY_AND_VALUE"
      value = "anotherfiltervalue"
    }
  }
`
	} else {
		tagGroupOrFilter = `ec2_tag_filter {
    key = "filterkey"
    type = "KEY_AND_VALUE"
    value = "anotherfiltervalue"
  }
`
	}

	return fmt.Sprintf(`
resource "aws_codedeploy_app" "foo_app" {
  name = "foo_app_%s"
}

resource "aws_iam_role_policy" "foo_policy" {
  name = "foo_policy_%s"
  role = "${aws_iam_role.bar_role.id}"
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "autoscaling:CompleteLifecycleAction",
        "autoscaling:DeleteLifecycleHook",
        "autoscaling:DescribeAutoScalingGroups",
        "autoscaling:DescribeLifecycleHooks",
        "autoscaling:PutLifecycleHook",
        "autoscaling:RecordLifecycleActionHeartbeat",
        "ec2:DescribeInstances",
        "ec2:DescribeInstanceStatus",
        "tag:GetTags",
        "tag:GetResources"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_role" "bar_role" {
  name = "bar_role_%s"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "codedeploy.amazonaws.com"
        ]
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_codedeploy_deployment_group" "foo" {
  app_name = "${aws_codedeploy_app.foo_app.name}"
  deployment_group_name = "bar_%s"
  service_role_arn = "${aws_iam_role.bar_role.arn}"
  %s
}`, rName, rName, rName, rName, tagGroupOrFilter)
}

func testAccAWSCodeDeployDeploymentGroupOnPremiseTags(rName string) string {
	return fmt.Sprintf(`
resource "aws_codedeploy_app" "foo_app" {
  name = "foo_app_%s"
}

resource "aws_iam_role_policy" "foo_policy" {
  name = "foo_policy_%s"
  role = "${aws_iam_role.foo_role.id}"
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "autoscaling:CompleteLifecycleAction",
        "autoscaling:DeleteLifecycleHook",
        "autoscaling:DescribeAutoScalingGroups",
        "autoscaling:DescribeLifecycleHooks",
        "autoscaling:PutLifecycleHook",
        "autoscaling:RecordLifecycleActionHeartbeat",
        "ec2:DescribeInstances",
        "ec2:DescribeInstanceStatus",
        "tag:GetTags",
        "tag:GetResources"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_role" "foo_role" {
  name = "foo_role_%s"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "codedeploy.amazonaws.com"
        ]
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_codedeploy_deployment_group" "foo" {
  app_name = "${aws_codedeploy_app.foo_app.name}"
  deployment_group_name = "foo_%s"
  service_role_arn = "${aws_iam_role.foo_role.arn}"
  on_premises_instance_tag_filter {
    key = "filterkey"
    type = "KEY_AND_VALUE"
    value = "filtervalue"
  }
}`, rName, rName, rName, rName)
}

func baseCodeDeployConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_codedeploy_app" "foo_app" {
  name = "foo-app-%s"
}

resource "aws_iam_role_policy" "foo_policy" {
  name = "foo-policy-%s"
  role = "${aws_iam_role.foo_role.id}"
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "autoscaling:CompleteLifecycleAction",
        "autoscaling:DeleteLifecycleHook",
        "autoscaling:DescribeAutoScalingGroups",
        "autoscaling:DescribeLifecycleHooks",
        "autoscaling:PutLifecycleHook",
        "autoscaling:RecordLifecycleActionHeartbeat",
        "codedeploy:*",
        "ec2:DescribeInstances",
        "ec2:DescribeInstanceStatus",
        "tag:GetTags",
        "tag:GetResources",
        "sns:Publish"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_role" "foo_role" {
  name = "foo-role-%s"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "codedeploy.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_sns_topic" "foo_topic" {
  name = "foo-topic-%s"
}`, rName, rName, rName, rName)
}

func testAccAWSCodeDeployDeploymentGroup_triggerConfiguration_create(rName string) string {
	return fmt.Sprintf(`

  %s

resource "aws_codedeploy_deployment_group" "foo_group" {
  app_name = "${aws_codedeploy_app.foo_app.name}"
  deployment_group_name = "foo-group-%s"
  service_role_arn = "${aws_iam_role.foo_role.arn}"

  trigger_configuration {
    trigger_events = ["DeploymentFailure"]
    trigger_name = "foo-trigger"
    trigger_target_arn = "${aws_sns_topic.foo_topic.arn}"
  }
}`, baseCodeDeployConfig(rName), rName)
}

func testAccAWSCodeDeployDeploymentGroup_triggerConfiguration_update(rName string) string {
	return fmt.Sprintf(`

  %s

resource "aws_codedeploy_deployment_group" "foo_group" {
  app_name = "${aws_codedeploy_app.foo_app.name}"
  deployment_group_name = "foo-group-%s"
  service_role_arn = "${aws_iam_role.foo_role.arn}"

  trigger_configuration {
    trigger_events = ["DeploymentSuccess", "DeploymentFailure"]
    trigger_name = "foo-trigger"
    trigger_target_arn = "${aws_sns_topic.foo_topic.arn}"
  }
}`, baseCodeDeployConfig(rName), rName)
}

func testAccAWSCodeDeployDeploymentGroup_triggerConfiguration_createMultiple(rName string) string {
	return fmt.Sprintf(`

  %s

resource "aws_sns_topic" "bar_topic" {
  name = "bar-topic-%s"
}

resource "aws_codedeploy_deployment_group" "foo_group" {
  app_name = "${aws_codedeploy_app.foo_app.name}"
  deployment_group_name = "foo-group-%s"
  service_role_arn = "${aws_iam_role.foo_role.arn}"

  trigger_configuration {
    trigger_events = ["DeploymentFailure"]
    trigger_name = "foo-trigger"
    trigger_target_arn = "${aws_sns_topic.foo_topic.arn}"
  }

  trigger_configuration {
    trigger_events = ["InstanceFailure"]
    trigger_name = "bar-trigger"
    trigger_target_arn = "${aws_sns_topic.bar_topic.arn}"
  }
}`, baseCodeDeployConfig(rName), rName, rName)
}

func testAccAWSCodeDeployDeploymentGroup_triggerConfiguration_updateMultiple(rName string) string {
	return fmt.Sprintf(`

  %s

resource "aws_sns_topic" "bar_topic" {
  name = "bar-topic-%s"
}

resource "aws_sns_topic" "baz_topic" {
  name = "baz-topic-%s"
}

resource "aws_codedeploy_deployment_group" "foo_group" {
  app_name = "${aws_codedeploy_app.foo_app.name}"
  deployment_group_name = "foo-group-%s"
  service_role_arn = "${aws_iam_role.foo_role.arn}"

  trigger_configuration {
    trigger_events = ["DeploymentStart", "DeploymentSuccess", "DeploymentFailure", "DeploymentStop"]
    trigger_name = "foo-trigger"
    trigger_target_arn = "${aws_sns_topic.foo_topic.arn}"
  }

  trigger_configuration {
    trigger_events = ["InstanceFailure"]
    trigger_name = "bar-trigger"
    trigger_target_arn = "${aws_sns_topic.baz_topic.arn}"
  }
}`, baseCodeDeployConfig(rName), rName, rName, rName)
}

func test_config_auto_rollback_configuration_create(rName string) string {
	return fmt.Sprintf(`

  %s

resource "aws_codedeploy_deployment_group" "foo_group" {
  app_name = "${aws_codedeploy_app.foo_app.name}"
  deployment_group_name = "foo-group-%s"
  service_role_arn = "${aws_iam_role.foo_role.arn}"

  auto_rollback_configuration {
    enabled = true
    events = ["DEPLOYMENT_FAILURE"]
  }
}`, baseCodeDeployConfig(rName), rName)
}

func test_config_auto_rollback_configuration_update(rName string) string {
	return fmt.Sprintf(`

  %s

resource "aws_codedeploy_deployment_group" "foo_group" {
  app_name = "${aws_codedeploy_app.foo_app.name}"
  deployment_group_name = "foo-group-%s"
  service_role_arn = "${aws_iam_role.foo_role.arn}"

  auto_rollback_configuration {
    enabled = true
    events = ["DEPLOYMENT_FAILURE", "DEPLOYMENT_STOP_ON_ALARM"]
  }
}`, baseCodeDeployConfig(rName), rName)
}

func test_config_auto_rollback_configuration_delete(rName string) string {
	return fmt.Sprintf(`

  %s

resource "aws_codedeploy_deployment_group" "foo_group" {
  app_name = "${aws_codedeploy_app.foo_app.name}"
  deployment_group_name = "foo-group-%s"
  service_role_arn = "${aws_iam_role.foo_role.arn}"
}`, baseCodeDeployConfig(rName), rName)
}

func test_config_auto_rollback_configuration_disable(rName string) string {
	return fmt.Sprintf(`

  %s

resource "aws_codedeploy_deployment_group" "foo_group" {
  app_name = "${aws_codedeploy_app.foo_app.name}"
  deployment_group_name = "foo-group-%s"
  service_role_arn = "${aws_iam_role.foo_role.arn}"

  auto_rollback_configuration {
    enabled = false
    events = ["DEPLOYMENT_FAILURE"]
  }
}`, baseCodeDeployConfig(rName), rName)
}

func test_config_alarm_configuration_create(rName string) string {
	return fmt.Sprintf(`

  %s

resource "aws_codedeploy_deployment_group" "foo_group" {
  app_name = "${aws_codedeploy_app.foo_app.name}"
  deployment_group_name = "foo-group-%s"
  service_role_arn = "${aws_iam_role.foo_role.arn}"

  alarm_configuration {
    alarms = ["foo"]
    enabled = true
  }
}`, baseCodeDeployConfig(rName), rName)
}

func test_config_alarm_configuration_update(rName string) string {
	return fmt.Sprintf(`

  %s

resource "aws_codedeploy_deployment_group" "foo_group" {
  app_name = "${aws_codedeploy_app.foo_app.name}"
  deployment_group_name = "foo-group-%s"
  service_role_arn = "${aws_iam_role.foo_role.arn}"

  alarm_configuration {
    alarms = ["foo", "bar"]
    enabled = true
    ignore_poll_alarm_failure = true
  }
}`, baseCodeDeployConfig(rName), rName)
}

func test_config_alarm_configuration_delete(rName string) string {
	return fmt.Sprintf(`

  %s

resource "aws_codedeploy_deployment_group" "foo_group" {
  app_name = "${aws_codedeploy_app.foo_app.name}"
  deployment_group_name = "foo-group-%s"
  service_role_arn = "${aws_iam_role.foo_role.arn}"
}`, baseCodeDeployConfig(rName), rName)
}

func test_config_alarm_configuration_disable(rName string) string {
	return fmt.Sprintf(`

  %s

resource "aws_codedeploy_deployment_group" "foo_group" {
  app_name = "${aws_codedeploy_app.foo_app.name}"
  deployment_group_name = "foo-group-%s"
  service_role_arn = "${aws_iam_role.foo_role.arn}"

  alarm_configuration {
    alarms = ["foo"]
    enabled = false
  }
}`, baseCodeDeployConfig(rName), rName)
}

func test_config_deployment_style_default(rName string) string {
	return fmt.Sprintf(`

  %s

resource "aws_codedeploy_deployment_group" "foo_group" {
  app_name = "${aws_codedeploy_app.foo_app.name}"
  deployment_group_name = "foo-group-%s"
  service_role_arn = "${aws_iam_role.foo_role.arn}"
}`, baseCodeDeployConfig(rName), rName)
}

func test_config_deployment_style_create(rName string) string {
	return fmt.Sprintf(`

  %s

resource "aws_codedeploy_deployment_group" "foo_group" {
  app_name = "${aws_codedeploy_app.foo_app.name}"
  deployment_group_name = "foo-group-%s"
  service_role_arn = "${aws_iam_role.foo_role.arn}"

  deployment_style {
    deployment_option = "WITH_TRAFFIC_CONTROL"
    deployment_type = "BLUE_GREEN"
  }

  load_balancer_info {
    elb_info {
      name = "foo-elb"
    }
  }
}`, baseCodeDeployConfig(rName), rName)
}

func test_config_deployment_style_update(rName string) string {
	return fmt.Sprintf(`

  %s

resource "aws_codedeploy_deployment_group" "foo_group" {
  app_name = "${aws_codedeploy_app.foo_app.name}"
  deployment_group_name = "foo-group-%s"
  service_role_arn = "${aws_iam_role.foo_role.arn}"

  deployment_style {
    deployment_option = "WITHOUT_TRAFFIC_CONTROL"
    deployment_type = "IN_PLACE"
  }
}`, baseCodeDeployConfig(rName), rName)
}

func test_config_load_balancer_info_default(rName string) string {
	return fmt.Sprintf(`

  %s

resource "aws_codedeploy_deployment_group" "foo_group" {
  app_name = "${aws_codedeploy_app.foo_app.name}"
  deployment_group_name = "foo-group-%s"
  service_role_arn = "${aws_iam_role.foo_role.arn}"
}`, baseCodeDeployConfig(rName), rName)
}

func test_config_load_balancer_info_create(rName string) string {
	return fmt.Sprintf(`

  %s

resource "aws_codedeploy_deployment_group" "foo_group" {
  app_name = "${aws_codedeploy_app.foo_app.name}"
  deployment_group_name = "foo-group-%s"
  service_role_arn = "${aws_iam_role.foo_role.arn}"

  load_balancer_info {
    elb_info {
      name = "foo-elb"
    }
  }
}`, baseCodeDeployConfig(rName), rName)
}

func test_config_load_balancer_info_update(rName string) string {
	return fmt.Sprintf(`

  %s

resource "aws_codedeploy_deployment_group" "foo_group" {
  app_name = "${aws_codedeploy_app.foo_app.name}"
  deployment_group_name = "foo-group-%s"
  service_role_arn = "${aws_iam_role.foo_role.arn}"

  load_balancer_info {
    elb_info {
      name = "bar-elb"
    }
  }
}`, baseCodeDeployConfig(rName), rName)
}

func test_config_load_balancer_info_delete(rName string) string {
	return fmt.Sprintf(`

  %s

resource "aws_codedeploy_deployment_group" "foo_group" {
  app_name = "${aws_codedeploy_app.foo_app.name}"
  deployment_group_name = "foo-group-%s"
  service_role_arn = "${aws_iam_role.foo_role.arn}"
}`, baseCodeDeployConfig(rName), rName)
}

func test_config_load_balancer_info_target_group_info_create(rName string) string {
	return fmt.Sprintf(`

  %s

resource "aws_codedeploy_deployment_group" "foo_group" {
  app_name = "${aws_codedeploy_app.foo_app.name}"
  deployment_group_name = "foo-group-%s"
  service_role_arn = "${aws_iam_role.foo_role.arn}"

  load_balancer_info {
    target_group_info {
      name = "foo-tg"
    }
  }
}`, baseCodeDeployConfig(rName), rName)
}

func test_config_load_balancer_info_target_group_info_update(rName string) string {
	return fmt.Sprintf(`

  %s

resource "aws_codedeploy_deployment_group" "foo_group" {
  app_name = "${aws_codedeploy_app.foo_app.name}"
  deployment_group_name = "foo-group-%s"
  service_role_arn = "${aws_iam_role.foo_role.arn}"

  load_balancer_info {
    target_group_info {
      name = "bar-tg"
    }
  }
}`, baseCodeDeployConfig(rName), rName)
}

func test_config_load_balancer_info_target_group_info_delete(rName string) string {
	return fmt.Sprintf(`

  %s

resource "aws_codedeploy_deployment_group" "foo_group" {
  app_name = "${aws_codedeploy_app.foo_app.name}"
  deployment_group_name = "foo-group-%s"
  service_role_arn = "${aws_iam_role.foo_role.arn}"
}`, baseCodeDeployConfig(rName), rName)
}

func test_config_in_place_deployment_with_traffic_control_create(rName string) string {
	return fmt.Sprintf(`

  %s

resource "aws_codedeploy_deployment_group" "foo_group" {
  app_name = "${aws_codedeploy_app.foo_app.name}"
  deployment_group_name = "foo-group-%s"
  service_role_arn = "${aws_iam_role.foo_role.arn}"

  deployment_style {
    deployment_option = "WITH_TRAFFIC_CONTROL"
    deployment_type = "IN_PLACE"
  }

  load_balancer_info {
    elb_info {
      name = "foo-elb"
    }
  }
}`, baseCodeDeployConfig(rName), rName)
}

func test_config_in_place_deployment_with_traffic_control_update(rName string) string {
	return fmt.Sprintf(`

  %s

resource "aws_codedeploy_deployment_group" "foo_group" {
  app_name = "${aws_codedeploy_app.foo_app.name}"
  deployment_group_name = "foo-group-%s"
  service_role_arn = "${aws_iam_role.foo_role.arn}"

  deployment_style {
    deployment_option = "WITH_TRAFFIC_CONTROL"
    deployment_type = "BLUE_GREEN"
  }

  load_balancer_info {
    elb_info {
      name = "foo-elb"
    }
  }

  blue_green_deployment_config {
    deployment_ready_option {
      action_on_timeout = "CONTINUE_DEPLOYMENT"
    }

    green_fleet_provisioning_option {
      action = "DISCOVER_EXISTING"
    }

    terminate_blue_instances_on_deployment_success {
      action = "KEEP_ALIVE"
    }
  }
}`, baseCodeDeployConfig(rName), rName)
}

func test_config_blue_green_deployment_config_default(rName string) string {
	return fmt.Sprintf(`

  %s

resource "aws_codedeploy_deployment_group" "foo_group" {
  app_name = "${aws_codedeploy_app.foo_app.name}"
  deployment_group_name = "foo-group-%s"
  service_role_arn = "${aws_iam_role.foo_role.arn}"
}`, baseCodeDeployConfig(rName), rName)
}

func test_config_blue_green_deployment_config_create_with_asg(rName string) string {
	return fmt.Sprintf(`

  %s

data "aws_ami" "amzn-ami-minimal-hvm-ebs" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name = "name"
    values = ["amzn-ami-minimal-hvm-*"]
  }
  filter {
    name = "root-device-type"
    values = ["ebs"]
  }
}

data "aws_availability_zones" "available" {}

data "aws_subnet" "test" {
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  default_for_az    = "true"
}

resource "aws_launch_configuration" "foo_lc" {
  image_id = "${data.aws_ami.amzn-ami-minimal-hvm-ebs.id}"
  instance_type = "t2.micro"
  "name_prefix" = "foo-lc-"

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_autoscaling_group" "foo_asg" {
  name = "foo-asg-%s"
  max_size = 2
  min_size = 0
  desired_capacity = 1

  vpc_zone_identifier = ["${data.aws_subnet.test.id}"]

  launch_configuration = "${aws_launch_configuration.foo_lc.name}"

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_codedeploy_deployment_group" "foo_group" {
  app_name = "${aws_codedeploy_app.foo_app.name}"
  deployment_group_name = "foo-group-%s"
  service_role_arn = "${aws_iam_role.foo_role.arn}"

  autoscaling_groups = ["${aws_autoscaling_group.foo_asg.name}"]

  blue_green_deployment_config {
    deployment_ready_option {
      action_on_timeout = "STOP_DEPLOYMENT"
      wait_time_in_minutes = 60
    }

    green_fleet_provisioning_option {
      action = "COPY_AUTO_SCALING_GROUP"
    }

    terminate_blue_instances_on_deployment_success {
      action = "TERMINATE"
      termination_wait_time_in_minutes = 120
    }
  }
}`, baseCodeDeployConfig(rName), rName, rName)
}

func test_config_blue_green_deployment_config_update_with_asg(rName string) string {
	return fmt.Sprintf(`

  %s

data "aws_ami" "amzn-ami-minimal-hvm-ebs" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name = "name"
    values = ["amzn-ami-minimal-hvm-*"]
  }
  filter {
    name = "root-device-type"
    values = ["ebs"]
  }
}

data "aws_availability_zones" "available" {}

data "aws_subnet" "test" {
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  default_for_az    = "true"
}

resource "aws_launch_configuration" "foo_lc" {
  image_id = "${data.aws_ami.amzn-ami-minimal-hvm-ebs.id}"
  instance_type = "t2.micro"
  "name_prefix" = "foo-lc-"

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_autoscaling_group" "foo_asg" {
  name = "foo-asg-%s"
  max_size = 2
  min_size = 0
  desired_capacity = 1

  vpc_zone_identifier = ["${data.aws_subnet.test.id}"]

  launch_configuration = "${aws_launch_configuration.foo_lc.name}"

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_codedeploy_deployment_group" "foo_group" {
  app_name = "${aws_codedeploy_app.foo_app.name}"
  deployment_group_name = "foo-group-%s"
  service_role_arn = "${aws_iam_role.foo_role.arn}"

  autoscaling_groups = ["${aws_autoscaling_group.foo_asg.name}"]

  blue_green_deployment_config {
    deployment_ready_option {
      action_on_timeout = "STOP_DEPLOYMENT"
      wait_time_in_minutes = 60
    }

    green_fleet_provisioning_option {
      action = "COPY_AUTO_SCALING_GROUP"
    }

    terminate_blue_instances_on_deployment_success {
      action = "KEEP_ALIVE"
    }
  }
}`, baseCodeDeployConfig(rName), rName, rName)
}

func test_config_blue_green_deployment_config_create_no_asg(rName string) string {
	return fmt.Sprintf(`

  %s

resource "aws_codedeploy_deployment_group" "foo_group" {
  app_name = "${aws_codedeploy_app.foo_app.name}"
  deployment_group_name = "foo-group-%s"
  service_role_arn = "${aws_iam_role.foo_role.arn}"

  blue_green_deployment_config {
    deployment_ready_option {
      action_on_timeout = "STOP_DEPLOYMENT"
      wait_time_in_minutes = 60
    }

    green_fleet_provisioning_option {
      action = "DISCOVER_EXISTING"
    }

    terminate_blue_instances_on_deployment_success {
      action = "TERMINATE"
      termination_wait_time_in_minutes = 120
    }
  }
}`, baseCodeDeployConfig(rName), rName)
}

func test_config_blue_green_deployment_config_update_no_asg(rName string) string {
	return fmt.Sprintf(`

  %s

resource "aws_codedeploy_deployment_group" "foo_group" {
  app_name = "${aws_codedeploy_app.foo_app.name}"
  deployment_group_name = "foo-group-%s"
  service_role_arn = "${aws_iam_role.foo_role.arn}"

  blue_green_deployment_config {
    deployment_ready_option {
      action_on_timeout = "CONTINUE_DEPLOYMENT"
    }

    green_fleet_provisioning_option {
      action = "DISCOVER_EXISTING"
    }

    terminate_blue_instances_on_deployment_success {
      action = "KEEP_ALIVE"
    }
  }
}`, baseCodeDeployConfig(rName), rName)
}

func test_config_blue_green_deployment_complete(rName string) string {
	return fmt.Sprintf(`

  %s

resource "aws_codedeploy_deployment_group" "foo_group" {
  app_name = "${aws_codedeploy_app.foo_app.name}"
  deployment_group_name = "foo-group-%s"
  service_role_arn = "${aws_iam_role.foo_role.arn}"

  deployment_style {
    deployment_option = "WITH_TRAFFIC_CONTROL"
    deployment_type = "BLUE_GREEN"
  }

  load_balancer_info {
    elb_info {
      name = "foo-elb"
    }
  }

  blue_green_deployment_config {
    deployment_ready_option {
      action_on_timeout = "STOP_DEPLOYMENT"
      wait_time_in_minutes = 60
    }

    green_fleet_provisioning_option {
      action = "DISCOVER_EXISTING"
    }

    terminate_blue_instances_on_deployment_success {
      action = "KEEP_ALIVE"
    }
  }
}`, baseCodeDeployConfig(rName), rName)
}

func test_config_blue_green_deployment_complete_updated(rName string) string {
	return fmt.Sprintf(`

  %s

resource "aws_codedeploy_deployment_group" "foo_group" {
  app_name = "${aws_codedeploy_app.foo_app.name}"
  deployment_group_name = "foo-group-%s"
  service_role_arn = "${aws_iam_role.foo_role.arn}"

  deployment_style {
    deployment_option = "WITH_TRAFFIC_CONTROL"
    deployment_type = "BLUE_GREEN"
  }

  load_balancer_info {
    elb_info {
      name = "foo-elb"
    }
  }

  blue_green_deployment_config {
    deployment_ready_option {
      action_on_timeout = "CONTINUE_DEPLOYMENT"
    }

    green_fleet_provisioning_option {
      action = "DISCOVER_EXISTING"
    }

    terminate_blue_instances_on_deployment_success {
      action = "KEEP_ALIVE"
    }
  }
}`, baseCodeDeployConfig(rName), rName)
}

func testAccAWSCodeDeployDeploymentGroupConfigEcsBase(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-codedeploy-deployment-group-ecs"
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = "${data.aws_availability_zones.available.names[count.index]}"
  cidr_block        = "${cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)}"
  vpc_id            = "${aws_vpc.test.id}"

  tags = {
    Name = "tf-acc-test-codedeploy-deployment-group-ecs"
  }
}

resource "aws_security_group" "test" {
  name   = %q
  vpc_id = "${aws_vpc.test.id}"

  ingress {
    cidr_blocks = ["${aws_vpc.test.cidr_block}"]
    from_port   = 80
    protocol    = "6"
    to_port     = 8000
  }
}

resource "aws_lb_target_group" "blue" {
  name        = "${aws_lb.test.name}-blue"
  port        = 80
  protocol    = "HTTP"
  target_type = "ip"
  vpc_id      = "${aws_vpc.test.id}"
}

resource "aws_lb_target_group" "green" {
  name        = "${aws_lb.test.name}-green"
  port        = 80
  protocol    = "HTTP"
  target_type = "ip"
  vpc_id      = "${aws_vpc.test.id}"
}

resource "aws_lb" "test" {
  internal = true
  name     = %q
  subnets  = ["${aws_subnet.test.*.id}"]
}

resource "aws_lb_listener" "test" {
  load_balancer_arn = "${aws_lb.test.arn}"
  port              = "80"
  protocol          = "HTTP"

  default_action {
    target_group_arn = "${aws_lb_target_group.blue.arn}"
    type             = "forward"
  }
}

resource "aws_ecs_cluster" "test" {
  name = %q
}

resource "aws_ecs_task_definition" "test" {
  cpu                      = "256"
  family                   = %q
  memory                   = "512"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]

  container_definitions = <<DEFINITION
[
  {
    "cpu": 256,
    "essential": true,
    "image": "mongo:latest",
    "memory": 512,
    "name": "test",
    "networkMode": "awsvpc",
    "portMappings": [
      {
        "containerPort": 80,
        "hostPort": 80
      }
    ]
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  cluster         = "${aws_ecs_cluster.test.id}"
  desired_count   = 1
  launch_type     = "FARGATE"
  name            = %q
  task_definition = "${aws_ecs_task_definition.test.arn}"

  deployment_controller {
    type = "CODE_DEPLOY"
  }

  load_balancer {
    container_name   = "test"
    container_port   = "80"
    target_group_arn = "${aws_lb_target_group.blue.id}"
  }

  network_configuration {
    assign_public_ip = true
    security_groups  = ["${aws_security_group.test.id}"]
    subnets          = ["${aws_subnet.test.*.id}"]
  }
}

resource "aws_codedeploy_app" "test" {
  compute_platform = "ECS"
  name             = %q
}

resource "aws_iam_role" "test" {
  name = %q

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": ["codedeploy.amazonaws.com"]
      }
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy" "test" {
  role = "${aws_iam_role.test.name}"

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "cloudwatch:DescribeAlarms",
        "ecs:CreateTaskSet",
        "ecs:DeleteTaskSet",
        "ecs:DescribeServices",
        "ecs:UpdateServicePrimaryTaskSet",
        "elasticloadbalancing:DescribeListeners",
        "elasticloadbalancing:DescribeRules",
        "elasticloadbalancing:DescribeTargetGroups",
        "elasticloadbalancing:ModifyListener",
        "elasticloadbalancing:ModifyRule",
        "lambda:InvokeFunction",
        "s3:GetObject",
        "s3:GetObjectMetadata",
        "s3:GetObjectVersion",
        "sns:Publish"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
POLICY
}
`, rName, rName, rName, rName, rName, rName, rName)
}

func testAccAWSCodeDeployDeploymentGroupConfigEcsBlueGreen(rName string) string {
	return testAccAWSCodeDeployDeploymentGroupConfigEcsBase(rName) + fmt.Sprintf(`
resource "aws_codedeploy_deployment_group" "test" {
  app_name               = "${aws_codedeploy_app.test.name}"
  deployment_config_name = "CodeDeployDefault.ECSAllAtOnce"
  deployment_group_name  = %q
  service_role_arn       = "${aws_iam_role.test.arn}"

  auto_rollback_configuration {
    enabled = true
    events  = ["DEPLOYMENT_FAILURE"]
  }

  blue_green_deployment_config {
    deployment_ready_option {
      action_on_timeout = "CONTINUE_DEPLOYMENT"
    }

    terminate_blue_instances_on_deployment_success {
      action                           = "TERMINATE"
      termination_wait_time_in_minutes = 5
    }
  }

  deployment_style {
    deployment_option = "WITH_TRAFFIC_CONTROL"
    deployment_type   = "BLUE_GREEN"
  }

  ecs_service {
    cluster_name = "${aws_ecs_cluster.test.name}"
    service_name = "${aws_ecs_service.test.name}"
  }

  load_balancer_info {
    target_group_pair_info {
      prod_traffic_route {
        listener_arns = ["${aws_lb_listener.test.arn}"]
      }

      target_group {
        name = "${aws_lb_target_group.blue.name}"
      }

      target_group {
        name = "${aws_lb_target_group.green.name}"
      }
    }
  }
}
`, rName)
}
