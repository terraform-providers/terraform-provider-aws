package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/synthetics"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/synthetics/finder"
)

func init() {
	resource.AddTestSweepers("aws_synthetics_canary", &resource.Sweeper{
		Name: "aws_synthetics_canary",
		F:    testSweepSyntheticsCanaries,
		Dependencies: []string{
			"aws_lambda_function",
			"aws_lambda_layer",
			"aws_cloudwatch_log_group",
		},
	})
}

func testSweepSyntheticsCanaries(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).syntheticsconn
	input := &synthetics.DescribeCanariesInput{}
	var sweeperErrs *multierror.Error

	for {
		output, err := conn.DescribeCanaries(input)
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Synthetics Canary sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error retrieving Synthetics Canaries: %w", err)
		}

		for _, canary := range output.Canaries {
			name := aws.StringValue(canary.Name)
			log.Printf("[INFO] Deleting Synthetics Canary: %s", name)

			r := resourceAwsSyntheticsCanary()
			d := r.Data(nil)
			d.SetId(name)
			err := r.Delete(d, client)

			if err != nil {
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}
		input.NextToken = output.NextToken
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSSyntheticsCanary_basic(t *testing.T) {
	var conf1, conf2 synthetics.Canary
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(8))
	resourceName := "aws_synthetics_canary.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSyntheticsCanaryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSyntheticsCanaryBasicConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsSyntheticsCanaryExists(resourceName, &conf1),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", synthetics.ServiceName, regexp.MustCompile(`canary:.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_version", "syn-1.0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.memory_in_mb", "1000"),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.timeout_in_seconds", "840"),
					resource.TestCheckResourceAttr(resourceName, "failure_retention_period", "31"),
					resource.TestCheckResourceAttr(resourceName, "success_retention_period", "31"),
					resource.TestCheckResourceAttr(resourceName, "handler", "exports.handler"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.duration_in_seconds", "0"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.expression", "rate(0 hour)"),
					testAccMatchResourceAttrRegionalARN(resourceName, "engine_arn", "lambda", regexp.MustCompile(fmt.Sprintf(`function:cwsyn-%s.+`, rName))),
					testAccMatchResourceAttrRegionalARN(resourceName, "source_location_arn", "lambda", regexp.MustCompile(fmt.Sprintf(`layer:cwsyn-%s.+`, rName))),
					resource.TestCheckResourceAttrPair(resourceName, "execution_role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "artifact_s3_location", fmt.Sprintf("%s/", rName)),
					resource.TestCheckResourceAttr(resourceName, "timeline.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "timeline.0.created"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"zip_file", "start_canary"},
			},
			{
				Config: testAccAWSSyntheticsCanaryZipUpdatedConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsSyntheticsCanaryExists(resourceName, &conf2),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", synthetics.ServiceName, regexp.MustCompile(`canary:.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_version", "syn-1.0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.memory_in_mb", "1000"),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.timeout_in_seconds", "840"),
					resource.TestCheckResourceAttr(resourceName, "failure_retention_period", "31"),
					resource.TestCheckResourceAttr(resourceName, "success_retention_period", "31"),
					resource.TestCheckResourceAttr(resourceName, "handler", "exports.handler"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.duration_in_seconds", "0"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.expression", "rate(0 hour)"),
					testAccMatchResourceAttrRegionalARN(resourceName, "engine_arn", "lambda", regexp.MustCompile(fmt.Sprintf(`function:cwsyn-%s.+`, rName))),
					testAccMatchResourceAttrRegionalARN(resourceName, "source_location_arn", "lambda", regexp.MustCompile(fmt.Sprintf(`layer:cwsyn-%s.+`, rName))),
					resource.TestCheckResourceAttrPair(resourceName, "execution_role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "artifact_s3_location", fmt.Sprintf("%s/", rName)),
					resource.TestCheckResourceAttr(resourceName, "timeline.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "timeline.0.created"),
					resource.TestCheckResourceAttrSet(resourceName, "timeline.0.last_modified"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					testAccCheckAwsSyntheticsCanaryIsUpdated(&conf1, &conf2),
				),
			},
		},
	})
}

func TestAccAWSSyntheticsCanary_runtimeVersion(t *testing.T) {
	var conf1 synthetics.Canary
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(8))
	resourceName := "aws_synthetics_canary.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSyntheticsCanaryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSyntheticsCanaryRuntimeVersionConfig(rName, "syn-nodejs-2.1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsSyntheticsCanaryExists(resourceName, &conf1),
					resource.TestCheckResourceAttr(resourceName, "runtime_version", "syn-nodejs-2.1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"zip_file", "start_canary"},
			},
			{
				Config: testAccAWSSyntheticsCanaryRuntimeVersionConfig(rName, "syn-nodejs-2.2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsSyntheticsCanaryExists(resourceName, &conf1),
					resource.TestCheckResourceAttr(resourceName, "runtime_version", "syn-nodejs-2.2"),
				),
			},
		},
	})
}

func TestAccAWSSyntheticsCanary_startCanary(t *testing.T) {
	var conf1, conf2, conf3 synthetics.Canary
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(8))
	resourceName := "aws_synthetics_canary.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSyntheticsCanaryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSyntheticsCanaryStartCanaryConfig(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsSyntheticsCanaryExists(resourceName, &conf1),
					resource.TestCheckResourceAttr(resourceName, "timeline.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "timeline.0.last_started"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"zip_file", "start_canary"},
			},
			{
				Config: testAccAWSSyntheticsCanaryStartCanaryConfig(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsSyntheticsCanaryExists(resourceName, &conf2),
					resource.TestCheckResourceAttr(resourceName, "timeline.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "timeline.0.last_started"),
					resource.TestCheckResourceAttrSet(resourceName, "timeline.0.last_stopped"),
				),
			},
			{
				Config: testAccAWSSyntheticsCanaryStartCanaryConfig(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsSyntheticsCanaryExists(resourceName, &conf3),
					resource.TestCheckResourceAttr(resourceName, "timeline.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "timeline.0.last_started"),
					resource.TestCheckResourceAttrSet(resourceName, "timeline.0.last_stopped"),
					testAccCheckAwsSyntheticsCanaryIsStartedAfter(&conf2, &conf3),
				),
			},
		},
	})
}

func TestAccAWSSyntheticsCanary_startCanary_codeChanges(t *testing.T) {
	var conf1, conf2 synthetics.Canary
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(8))
	resourceName := "aws_synthetics_canary.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSyntheticsCanaryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSyntheticsCanaryStartCanaryConfig(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsSyntheticsCanaryExists(resourceName, &conf1),
					resource.TestCheckResourceAttr(resourceName, "status", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "timeline.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "timeline.0.last_started"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"zip_file", "start_canary"},
			},
			{
				Config: testAccAWSSyntheticsCanaryStartCanaryZipUpdatedConfig(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsSyntheticsCanaryExists(resourceName, &conf2),
					resource.TestCheckResourceAttr(resourceName, "status", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "timeline.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "timeline.0.last_started"),
					resource.TestCheckResourceAttrSet(resourceName, "timeline.0.last_stopped"),
					testAccCheckAwsSyntheticsCanaryIsStartedAfter(&conf1, &conf2),
				),
			},
		},
	})
}

func TestAccAWSSyntheticsCanary_s3(t *testing.T) {
	var conf synthetics.Canary
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(8))
	resourceName := "aws_synthetics_canary.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSyntheticsCanaryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSyntheticsCanaryBasicS3CodeConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsSyntheticsCanaryExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", synthetics.ServiceName, regexp.MustCompile(`canary:.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_version", "syn-1.0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.memory_in_mb", "1000"),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.timeout_in_seconds", "840"),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.active_tracing", "false"),
					resource.TestCheckResourceAttr(resourceName, "failure_retention_period", "31"),
					resource.TestCheckResourceAttr(resourceName, "success_retention_period", "31"),
					resource.TestCheckResourceAttr(resourceName, "handler", "exports.handler"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.duration_in_seconds", "0"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.expression", "rate(0 hour)"),
					testAccMatchResourceAttrRegionalARN(resourceName, "engine_arn", "lambda", regexp.MustCompile(fmt.Sprintf(`function:cwsyn-%s.+`, rName))),
					testAccMatchResourceAttrRegionalARN(resourceName, "source_location_arn", "lambda", regexp.MustCompile(fmt.Sprintf(`layer:cwsyn-%s.+`, rName))),
					resource.TestCheckResourceAttrPair(resourceName, "execution_role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "artifact_s3_location", fmt.Sprintf("%s/", rName)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"s3_bucket", "s3_key", "s3_version", "start_canary"},
			},
		},
	})
}

func TestAccAWSSyntheticsCanary_runConfig(t *testing.T) {
	var conf synthetics.Canary
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(8))
	resourceName := "aws_synthetics_canary.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSyntheticsCanaryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSyntheticsCanaryRunConfigConfig1(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsSyntheticsCanaryExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.memory_in_mb", "1000"),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.timeout_in_seconds", "60"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"zip_file", "start_canary"},
			},
			{
				Config: testAccAWSSyntheticsCanaryRunConfigConfig2(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsSyntheticsCanaryExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.memory_in_mb", "960"),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.timeout_in_seconds", "120"),
				),
			},
			{
				Config: testAccAWSSyntheticsCanaryRunConfigConfig1(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsSyntheticsCanaryExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.memory_in_mb", "960"),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.timeout_in_seconds", "60"),
				),
			},
		},
	})
}

func TestAccAWSSyntheticsCanary_runConfigTracing(t *testing.T) {
	var conf synthetics.Canary
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(8))
	resourceName := "aws_synthetics_canary.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSyntheticsCanaryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSyntheticsCanaryRunConfigTracingConfig(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsSyntheticsCanaryExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.active_tracing", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"zip_file", "start_canary"},
			},
			{
				Config: testAccAWSSyntheticsCanaryRunConfigTracingConfig(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsSyntheticsCanaryExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.active_tracing", "false"),
				),
			},
			{
				Config: testAccAWSSyntheticsCanaryRunConfigTracingConfig(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsSyntheticsCanaryExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.active_tracing", "true"),
				),
			},
		},
	})
}

func TestAccAWSSyntheticsCanary_vpc(t *testing.T) {
	var conf synthetics.Canary
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(8))
	resourceName := "aws_synthetics_canary.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSyntheticsCanaryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSyntheticsCanaryVPCConfig1(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsSyntheticsCanaryExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnet_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.vpc_id", "aws_vpc.test", "id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"zip_file", "start_canary"},
			},
			{
				Config: testAccAWSSyntheticsCanaryVPCConfig2(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsSyntheticsCanaryExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "2"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.vpc_id", "aws_vpc.test", "id"),
				),
			},
			{
				Config: testAccAWSSyntheticsCanaryVPCConfig3(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsSyntheticsCanaryExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnet_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.vpc_id", "aws_vpc.test", "id"),
					testAccCheckAwsSyntheticsCanaryDeleteImplicitResources(resourceName),
				),
			},
		},
	})
}

func TestAccAWSSyntheticsCanary_tags(t *testing.T) {
	var conf synthetics.Canary
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(8))
	resourceName := "aws_synthetics_canary.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSyntheticsCanaryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSyntheticsCanaryConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsSyntheticsCanaryExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"zip_file", "start_canary"},
			},
			{
				Config: testAccAWSSyntheticsCanaryConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsSyntheticsCanaryExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSSyntheticsCanaryConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsSyntheticsCanaryExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSSyntheticsCanary_disappears(t *testing.T) {
	var conf synthetics.Canary
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(8))
	resourceName := "aws_synthetics_canary.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSyntheticsCanaryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSyntheticsCanaryBasicConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsSyntheticsCanaryExists(resourceName, &conf),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSyntheticsCanary(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAwsSyntheticsCanaryDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).syntheticsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_synthetics_canary" {
			continue
		}

		name := rs.Primary.ID
		_, err := finder.CanaryByName(conn, name)
		if err != nil {
			if isAWSErr(err, synthetics.ErrCodeResourceNotFoundException, "") {
				return nil
			}
			return err
		}
	}

	return nil
}

func testAccCheckAwsSyntheticsCanaryExists(n string, canary *synthetics.Canary) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("synthetics Canary not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("synthetics Canary name not set")
		}

		name := rs.Primary.ID
		conn := testAccProvider.Meta().(*AWSClient).syntheticsconn

		out, err := finder.CanaryByName(conn, name)
		if err != nil {
			return fmt.Errorf("syntherics Canary %s not found in AWS", name)
		}

		*canary = *out.Canary

		return nil
	}
}

func testAccCheckAwsSyntheticsCanaryDeleteImplicitResources(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("synthetics Canary not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("synthetics Canary name not set")
		}

		lambdaConn := testAccProvider.Meta().(*AWSClient).lambdaconn

		layerArn := rs.Primary.Attributes["source_location_arn"]
		layerArnObj, err := arn.Parse(layerArn)
		if err != nil {
			return fmt.Errorf("synthetics Canary Lambda Layer %s incorrect arn format: %w", layerArn, err)
		}

		layerName := strings.Split(layerArnObj.Resource, ":")

		deleteLayerVersionInput := &lambda.DeleteLayerVersionInput{
			LayerName:     aws.String(layerName[1]),
			VersionNumber: aws.Int64(1),
		}

		_, err = lambdaConn.DeleteLayerVersion(deleteLayerVersionInput)
		if err != nil {
			return fmt.Errorf("synthetics Canary Lambda Layer %s could not be deleted: %w", layerArn, err)
		}

		lambdaArn := rs.Primary.Attributes["engine_arn"]
		lambdaArnObj, err := arn.Parse(layerArn)
		if err != nil {
			return fmt.Errorf("synthetics Canary Lambda %s incorrect arn format: %w", lambdaArn, err)
		}
		lambdaArnParts := strings.Split(lambdaArnObj.Resource, ":")

		deleteLambdaInput := &lambda.DeleteFunctionInput{
			FunctionName: aws.String(lambdaArnParts[1]),
		}

		_, err = lambdaConn.DeleteFunction(deleteLambdaInput)
		if err != nil {
			return fmt.Errorf("synthetics Canary Lambda %s could not be deleted: %w", lambdaArn, err)
		}

		return nil
	}
}

func testAccCheckAwsSyntheticsCanaryIsUpdated(first, second *synthetics.Canary) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !second.Timeline.LastModified.After(*first.Timeline.LastModified) {
			return fmt.Errorf("synthetics Canary not updated")

		}

		return nil
	}
}

func testAccCheckAwsSyntheticsCanaryIsStartedAfter(first, second *synthetics.Canary) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !second.Timeline.LastStarted.After(*first.Timeline.LastStarted) {
			return fmt.Errorf("synthetics Canary not updated")

		}

		return nil
	}
}

func testAccAWSSyntheticsCanaryConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  acl           = "private"
  force_destroy = true

  versioning {
    enabled = true
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF

  tags = {
    Name = %[1]q
  }
}

data "aws_partition" "current" {}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
        "Effect": "Allow",
        "Action": [
            "logs:CreateLogGroup",
            "logs:CreateLogStream",
            "logs:PutLogEvents"
        ],
        "Resource": "arn:${data.aws_partition.current.partition}:logs:*:*:*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "s3:PutObject"
      ],
      "Resource": [
        "${aws_s3_bucket.test.arn}/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetBucketLocation",
        "s3:ListAllMyBuckets"
      ],
      "Resource": [
        "*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "cloudwatch:PutMetricData"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF
}
`, rName)
}

func testAccAWSSyntheticsCanaryRunConfigConfig1(rName string) string {
	return composeConfig(testAccAWSSyntheticsCanaryConfigBase(rName), fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn
  handler              = "exports.handler"
  zip_file             = "test-fixtures/lambdatest.zip"
  runtime_version      = "syn-1.0"

  schedule {
    expression = "rate(0 minute)"
  }

  run_config {
    timeout_in_seconds = 60
  }
}
`, rName))
}

func testAccAWSSyntheticsCanaryRunConfigConfig2(rName string) string {
	return composeConfig(testAccAWSSyntheticsCanaryConfigBase(rName), fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn
  handler              = "exports.handler"
  zip_file             = "test-fixtures/lambdatest.zip"
  runtime_version      = "syn-1.0"

  schedule {
    expression = "rate(0 minute)"
  }

  run_config {
    timeout_in_seconds = 120
    memory_in_mb       = 960
  }
}
`, rName))
}

func testAccAWSSyntheticsCanaryRunConfigTracingConfig(rName string, tracing bool) string {
	return composeConfig(testAccAWSSyntheticsCanaryConfigBase(rName), fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn
  handler              = "exports.handler"
  zip_file             = "test-fixtures/lambdatest.zip"
  runtime_version      = "syn-nodejs-2.0"

  schedule {
    expression = "rate(0 minute)"
  }

  run_config {
    active_tracing     = %[2]t
    timeout_in_seconds = 60
  }
}
`, rName, tracing))
}

func testAccAWSSyntheticsCanaryBasicConfig(rName string) string {
	return composeConfig(testAccAWSSyntheticsCanaryConfigBase(rName), fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn
  handler              = "exports.handler"
  zip_file             = "test-fixtures/lambdatest.zip"
  runtime_version      = "syn-1.0"

  schedule {
    expression = "rate(0 minute)"
  }
}
`, rName))
}

func testAccAWSSyntheticsCanaryRuntimeVersionConfig(rName, version string) string {
	return composeConfig(testAccAWSSyntheticsCanaryConfigBase(rName), fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn
  handler              = "exports.handler"
  zip_file             = "test-fixtures/lambdatest.zip"
  runtime_version      = %[2]q

  schedule {
    expression = "rate(0 minute)"
  }
}
`, rName, version))
}

func testAccAWSSyntheticsCanaryZipUpdatedConfig(rName string) string {
	return composeConfig(testAccAWSSyntheticsCanaryConfigBase(rName), fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn
  handler              = "exports.handler"
  zip_file             = "test-fixtures/lambdatest_modified.zip"
  runtime_version      = "syn-1.0"

  schedule {
    expression = "rate(0 minute)"
  }
}
`, rName))
}

func testAccAWSSyntheticsCanaryStartCanaryConfig(rName string, state bool) string {
	return composeConfig(testAccAWSSyntheticsCanaryConfigBase(rName), fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn
  handler              = "exports.handler"
  zip_file             = "test-fixtures/lambdatest.zip"
  start_canary         = %[2]t
  runtime_version      = "syn-1.0"

  schedule {
    expression = "rate(0 minute)"
  }
}
`, rName, state))
}

func testAccAWSSyntheticsCanaryStartCanaryZipUpdatedConfig(rName string, state bool) string {
	return composeConfig(testAccAWSSyntheticsCanaryConfigBase(rName), fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn
  handler              = "exports.handler"
  zip_file             = "test-fixtures/lambdatest_modified.zip"
  start_canary         = %[2]t
  runtime_version      = "syn-1.0"

  schedule {
    expression = "rate(0 minute)"
  }
}
`, rName, state))
}

func testAccAWSSyntheticsCanaryBasicS3CodeConfig(rName string) string {
	return composeConfig(testAccAWSSyntheticsCanaryConfigBase(rName), fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn
  handler              = "exports.handler"
  s3_bucket            = aws_s3_bucket_object.test.bucket
  s3_key               = aws_s3_bucket_object.test.key
  s3_version           = aws_s3_bucket_object.test.version_id
  runtime_version      = "syn-1.0"

  schedule {
    expression = "rate(0 minute)"
  }
}

resource "aws_s3_bucket_object" "test" {
  bucket = aws_s3_bucket.test.bucket
  key    = %[1]q
  source = "test-fixtures/lambdatest.zip"
  etag   = filemd5("test-fixtures/lambdatest.zip")
}

`, rName))
}

func testAccAWSSyntheticsCanaryVPCConfigBase(rName string) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test1" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 2, 0)
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 2, 1)
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test1" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test2" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"
  role       = aws_iam_role.test.name
}
`, rName))
}

func testAccAWSSyntheticsCanaryVPCConfig1(rName string) string {
	return composeConfig(
		testAccAWSSyntheticsCanaryConfigBase(rName),
		testAccAWSSyntheticsCanaryVPCConfigBase(rName),
		fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn
  handler              = "exports.handler"
  zip_file             = "test-fixtures/lambdatest.zip"
  runtime_version      = "syn-1.0"

  schedule {
    expression = "rate(0 minute)"
  }

  vpc_config {
    subnet_ids         = [aws_subnet.test1.id]
    security_group_ids = [aws_security_group.test1.id]
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccAWSSyntheticsCanaryVPCConfig2(rName string) string {
	return composeConfig(
		testAccAWSSyntheticsCanaryConfigBase(rName),
		testAccAWSSyntheticsCanaryVPCConfigBase(rName),
		fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn
  handler              = "exports.handler"
  zip_file             = "test-fixtures/lambdatest.zip"
  runtime_version      = "syn-1.0"

  schedule {
    expression = "rate(0 minute)"
  }

  vpc_config {
    subnet_ids         = [aws_subnet.test1.id, aws_subnet.test2.id]
    security_group_ids = [aws_security_group.test1.id, aws_security_group.test2.id]
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccAWSSyntheticsCanaryVPCConfig3(rName string) string {
	return composeConfig(
		testAccAWSSyntheticsCanaryConfigBase(rName),
		testAccAWSSyntheticsCanaryVPCConfigBase(rName),
		fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn
  handler              = "exports.handler"
  zip_file             = "test-fixtures/lambdatest.zip"
  runtime_version      = "syn-1.0"

  schedule {
    expression = "rate(0 minute)"
  }

  vpc_config {
    subnet_ids         = [aws_subnet.test2.id]
    security_group_ids = [aws_security_group.test2.id]
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccAWSSyntheticsCanaryConfigTags1(rName, tagKey1, tagValue1 string) string {
	return composeConfig(testAccAWSSyntheticsCanaryConfigBase(rName), fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn
  handler              = "exports.handler"
  zip_file             = "test-fixtures/lambdatest.zip"
  runtime_version      = "syn-1.0"

  schedule {
    expression = "rate(0 minute)"
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccAWSSyntheticsCanaryConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return composeConfig(testAccAWSSyntheticsCanaryConfigBase(rName), fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn
  handler              = "exports.handler"
  zip_file             = "test-fixtures/lambdatest.zip"
  runtime_version      = "syn-1.0"

  schedule {
    expression = "rate(0 minute)"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
