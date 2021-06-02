package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudtrail"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_cloudtrail", &resource.Sweeper{
		Name: "aws_cloudtrail",
		F:    testSweepCloudTrails,
	})
}

func testSweepCloudTrails(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*AWSClient).cloudtrailconn
	var sweeperErrs *multierror.Error

	err = conn.ListTrailsPages(&cloudtrail.ListTrailsInput{}, func(page *cloudtrail.ListTrailsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, trail := range page.Trails {
			name := aws.StringValue(trail.Name)

			if name == "AWSMacieTrail-DO-NOT-EDIT" {
				log.Printf("[INFO] Skipping AWSMacieTrail-DO-NOT-EDIT for Macie Classic, which is not automatically recreated by the service")
				continue
			}

			output, err := conn.DescribeTrails(&cloudtrail.DescribeTrailsInput{
				TrailNameList: aws.StringSlice([]string{name}),
			})
			if err != nil {
				sweeperErr := fmt.Errorf("error describing CloudTrail (%s): %w", name, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}

			if len(output.TrailList) == 0 {
				log.Printf("[INFO] CloudTrail (%s) not found, skipping", name)
				continue
			}

			if aws.BoolValue(output.TrailList[0].IsOrganizationTrail) {
				log.Printf("[INFO] CloudTrail (%s) is an organization trail, skipping", name)
				continue
			}

			log.Printf("[INFO] Deleting CloudTrail: %s", name)
			_, err = conn.DeleteTrail(&cloudtrail.DeleteTrailInput{
				Name: aws.String(name),
			})
			if isAWSErr(err, cloudtrail.ErrCodeTrailNotFoundException, "") {
				continue
			}
			if err != nil {
				sweeperErr := fmt.Errorf("error deleting CloudTrail (%s): %w", name, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})
	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping CloudTrail sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving CloudTrails: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSCloudTrail_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Trail": {
			"basic":                 testAccAWSCloudTrail_basic,
			"cloudwatch":            testAccAWSCloudTrail_cloudwatch,
			"enableLogging":         testAccAWSCloudTrail_enableLogging,
			"globalServiceEvents":   testAccAWSCloudTrail_globalServiceEvents,
			"multiRegion":           testAccAWSCloudTrail_multiRegion,
			"organization":          testAccAWSCloudTrail_organization,
			"logValidation":         testAccAWSCloudTrail_logValidation,
			"kmsKey":                testAccAWSCloudTrail_kmsKey,
			"tags":                  testAccAWSCloudTrail_tags,
			"eventSelector":         testAccAWSCloudTrail_eventSelector,
			"eventSelectorDynamoDB": testAccAWSCloudTrail_eventSelectorDynamoDB,
			"insightSelector":       testAccAWSCloudTrail_insightSelector,
		},
	}

	for group, m := range testCases {
		m := m
		t.Run(group, func(t *testing.T) {
			for name, tc := range m {
				tc := tc
				t.Run(name, func(t *testing.T) {
					tc(t)
				})
			}
		})
	}
}

func testAccAWSCloudTrail_basic(t *testing.T) {
	var trail cloudtrail.Trail
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudtrail.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, cloudtrail.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudTrailDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudTrailConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, "include_global_service_events", "true"),
					resource.TestCheckResourceAttr(resourceName, "is_organization_trail", "false"),
					testAccCheckCloudTrailLogValidationEnabled(resourceName, false, &trail),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCloudTrailModifiedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, "s3_key_prefix", "prefix"),
					resource.TestCheckResourceAttr(resourceName, "include_global_service_events", "false"),
					testAccCheckCloudTrailLogValidationEnabled(resourceName, false, &trail),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
				),
			},
		},
	})
}

func testAccAWSCloudTrail_cloudwatch(t *testing.T) {
	var trail cloudtrail.Trail
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudtrail.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, cloudtrail.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudTrailDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudTrailCloudWatchConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					resource.TestCheckResourceAttrSet(resourceName, "cloud_watch_logs_group_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "cloud_watch_logs_role_arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCloudTrailCloudWatchModifiedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					resource.TestCheckResourceAttrSet(resourceName, "cloud_watch_logs_group_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "cloud_watch_logs_role_arn"),
				),
			},
		},
	})
}

func testAccAWSCloudTrail_enableLogging(t *testing.T) {
	var trail cloudtrail.Trail
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudtrail.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, cloudtrail.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudTrailDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudTrailConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					// AWS will create the trail with logging turned off.
					// Test that "enable_logging" default works.
					testAccCheckCloudTrailLoggingEnabled(resourceName, true),
					testAccCheckCloudTrailLogValidationEnabled(resourceName, false, &trail),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCloudTrailModifiedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					testAccCheckCloudTrailLoggingEnabled(resourceName, false),
					testAccCheckCloudTrailLogValidationEnabled(resourceName, false, &trail),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
				),
			},
			{
				Config: testAccAWSCloudTrailConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					testAccCheckCloudTrailLoggingEnabled(resourceName, true),
					testAccCheckCloudTrailLogValidationEnabled(resourceName, false, &trail),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
				),
			},
		},
	})
}

func testAccAWSCloudTrail_multiRegion(t *testing.T) {
	var trail cloudtrail.Trail
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudtrail.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, cloudtrail.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudTrailDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudTrailConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, "is_multi_region_trail", "false"),
					testAccCheckCloudTrailLogValidationEnabled(resourceName, false, &trail),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
				),
			},
			{
				Config: testAccAWSCloudTrailMultiRegionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, "is_multi_region_trail", "true"),
					testAccCheckCloudTrailLogValidationEnabled(resourceName, false, &trail),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCloudTrailConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, "is_multi_region_trail", "false"),
					testAccCheckCloudTrailLogValidationEnabled(resourceName, false, &trail),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
				),
			},
		},
	})
}

func testAccAWSCloudTrail_organization(t *testing.T) {
	var trail cloudtrail.Trail
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudtrail.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, cloudtrail.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudTrailDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudTrailOrganizationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, "is_organization_trail", "true"),
					testAccCheckCloudTrailLogValidationEnabled(resourceName, false, &trail),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCloudTrailConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, "is_organization_trail", "false"),
					testAccCheckCloudTrailLogValidationEnabled(resourceName, false, &trail),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
				),
			},
		},
	})
}

func testAccAWSCloudTrail_logValidation(t *testing.T) {
	var trail cloudtrail.Trail
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudtrail.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, cloudtrail.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudTrailDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudTrailLogValidationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, "s3_key_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "include_global_service_events", "true"),
					testAccCheckCloudTrailLogValidationEnabled(resourceName, true, &trail),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCloudTrailLogValidationModifiedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, "s3_key_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "include_global_service_events", "true"),
					testAccCheckCloudTrailLogValidationEnabled(resourceName, false, &trail),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
				),
			},
		},
	})
}

func testAccAWSCloudTrail_kmsKey(t *testing.T) {
	var trail cloudtrail.Trail
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resourceName := "aws_cloudtrail.test"
	kmsResourceName := "aws_kms_key.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, cloudtrail.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudTrailDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudTrailKMSKeyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, "s3_key_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "include_global_service_events", "true"),
					testAccCheckCloudTrailLogValidationEnabled(resourceName, false, &trail),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", kmsResourceName, "arn"),
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

func testAccAWSCloudTrail_tags(t *testing.T) {
	var trail cloudtrail.Trail
	var trailTags []*cloudtrail.Tag
	var trailTagsModified []*cloudtrail.Tag
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudtrail.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, cloudtrail.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudTrailDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudTrailTagsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					testAccCheckCloudTrailLoadTags(&trail, &trailTags),
					resource.TestCheckResourceAttr(resourceName, "tags.Yak", "milk"),
					resource.TestCheckResourceAttr(resourceName, "tags.Fox", "tail"),
					testAccCheckCloudTrailLogValidationEnabled(resourceName, false, &trail),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCloudTrailTagsModifiedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					testAccCheckCloudTrailLoadTags(&trail, &trailTagsModified),
					resource.TestCheckResourceAttr(resourceName, "tags.Yak", "milk"),
					resource.TestCheckResourceAttr(resourceName, "tags.Emu", "toes"),
					resource.TestCheckResourceAttr(resourceName, "tags.Fox", "tail"),
					testAccCheckCloudTrailLogValidationEnabled(resourceName, false, &trail),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
				),
			},
			{
				Config: testAccAWSCloudTrailTagsModifiedAgainConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					testAccCheckCloudTrailLoadTags(&trail, &trailTagsModified),
					testAccCheckCloudTrailLogValidationEnabled(resourceName, false, &trail),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
				),
			},
		},
	})
}

func testAccAWSCloudTrail_globalServiceEvents(t *testing.T) {
	var trail cloudtrail.Trail
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudtrail.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, cloudtrail.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudTrailDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudTrailGlobalServiceEventsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, "include_global_service_events", "false"),
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

func testAccAWSCloudTrail_eventSelector(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudtrail.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, cloudtrail.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudTrailDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudTrailEventSelectorConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "event_selector.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.data_resource.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.data_resource.0.type", "AWS::S3::Object"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.data_resource.0.values.#", "2"),
					testAccCheckResourceAttrGlobalARNNoAccount(resourceName, "event_selector.0.data_resource.0.values.0", "s3", fmt.Sprintf("%s-2/isen", rName)),
					testAccCheckResourceAttrGlobalARNNoAccount(resourceName, "event_selector.0.data_resource.0.values.1", "s3", fmt.Sprintf("%s-2/ko", rName)),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.include_management_events", "false"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.read_write_type", "ReadOnly"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCloudTrailEventSelectorReadWriteTypeConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "event_selector.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.include_management_events", "true"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.read_write_type", "WriteOnly"),
				),
			},
			{
				Config: testAccAWSCloudTrailEventSelectorModifiedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "event_selector.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.data_resource.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.data_resource.0.type", "AWS::S3::Object"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.data_resource.0.values.#", "2"),
					testAccCheckResourceAttrGlobalARNNoAccount(resourceName, "event_selector.0.data_resource.0.values.0", "s3", fmt.Sprintf("%s-2/isen", rName)),
					testAccCheckResourceAttrGlobalARNNoAccount(resourceName, "event_selector.0.data_resource.0.values.1", "s3", fmt.Sprintf("%s-2/ko", rName)),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.include_management_events", "true"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.read_write_type", "ReadOnly"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.1.data_resource.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.1.data_resource.0.type", "AWS::S3::Object"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.1.data_resource.0.values.#", "2"),
					testAccCheckResourceAttrGlobalARNNoAccount(resourceName, "event_selector.1.data_resource.0.values.0", "s3", fmt.Sprintf("%s-2/tf1", rName)),
					testAccCheckResourceAttrGlobalARNNoAccount(resourceName, "event_selector.1.data_resource.0.values.1", "s3", fmt.Sprintf("%s-2/tf2", rName)),
					resource.TestCheckResourceAttr(resourceName, "event_selector.1.data_resource.1.type", "AWS::Lambda::Function"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.1.data_resource.1.values.#", "1"),
					testAccCheckResourceAttrRegionalARN(resourceName, "event_selector.1.data_resource.1.values.0", "lambda", fmt.Sprintf("function:%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "event_selector.1.include_management_events", "false"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.1.read_write_type", "All"),
				),
			},
			{
				Config: testAccAWSCloudTrailEventSelectorNoneConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "event_selector.#", "0"),
				),
			},
		},
	})
}

func testAccAWSCloudTrail_eventSelectorDynamoDB(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudtrail.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, cloudtrail.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudTrailDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudTrailEventSelectorDynamoDBConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "event_selector.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.data_resource.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.data_resource.0.type", "AWS::DynamoDB::Table"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.data_resource.0.values.#", "1"),
					testAccMatchResourceAttrRegionalARN(resourceName, "event_selector.0.data_resource.0.values.0", "dynamodb", regexp.MustCompile(`table/tf-acc-test-.+`)),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.include_management_events", "true"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.read_write_type", "All"),
				),
			},
		},
	})
}

func testAccAWSCloudTrail_insightSelector(t *testing.T) {
	resourceName := "aws_cloudtrail.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, cloudtrail.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudTrailDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudTrailInsightSelectorConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "insight_selector.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "insight_selector.0.insight_type", "ApiCallRateInsight"),
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

func testAccCheckCloudTrailExists(n string, trail *cloudtrail.Trail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).cloudtrailconn
		params := cloudtrail.DescribeTrailsInput{
			TrailNameList: []*string{aws.String(rs.Primary.ID)},
		}
		resp, err := conn.DescribeTrails(&params)
		if err != nil {
			return err
		}
		if len(resp.TrailList) == 0 {
			return fmt.Errorf("Trail not found")
		}
		*trail = *resp.TrailList[0]

		return nil
	}
}

func testAccCheckCloudTrailLoggingEnabled(n string, desired bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).cloudtrailconn
		params := cloudtrail.GetTrailStatusInput{
			Name: aws.String(rs.Primary.ID),
		}
		resp, err := conn.GetTrailStatus(&params)

		if err != nil {
			return err
		}
		if *resp.IsLogging != desired {
			return fmt.Errorf("Expected logging status %t, given %t", desired, *resp.IsLogging)
		}

		return nil
	}
}

func testAccCheckCloudTrailLogValidationEnabled(n string, desired bool, trail *cloudtrail.Trail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if trail.LogFileValidationEnabled == nil {
			return fmt.Errorf("No LogFileValidationEnabled attribute present in trail: %s", trail)
		}

		if *trail.LogFileValidationEnabled != desired {
			return fmt.Errorf("Expected log validation status %t, given %t", desired,
				*trail.LogFileValidationEnabled)
		}

		// local state comparison
		enabled, ok := rs.Primary.Attributes["enable_log_file_validation"]
		if !ok {
			return fmt.Errorf("No enable_log_file_validation attribute defined for %s, expected %t",
				n, desired)
		}
		desiredInString := fmt.Sprintf("%t", desired)
		if enabled != desiredInString {
			return fmt.Errorf("Expected log validation status %s, saved %s", desiredInString, enabled)
		}

		return nil
	}
}

func testAccCheckAWSCloudTrailDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cloudtrailconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudtrail" {
			continue
		}

		params := cloudtrail.DescribeTrailsInput{
			TrailNameList: []*string{aws.String(rs.Primary.ID)},
		}

		resp, err := conn.DescribeTrails(&params)

		if err == nil {
			if len(resp.TrailList) != 0 &&
				*resp.TrailList[0].Name == rs.Primary.ID {
				return fmt.Errorf("CloudTrail still exists: %s", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccCheckCloudTrailLoadTags(trail *cloudtrail.Trail, tags *[]*cloudtrail.Tag) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).cloudtrailconn
		input := cloudtrail.ListTagsInput{
			ResourceIdList: []*string{trail.TrailARN},
		}
		out, err := conn.ListTags(&input)
		if err != nil {
			return err
		}
		log.Printf("[DEBUG] Received CloudTrail tags during test: %s", out)
		if len(out.ResourceTagList) > 0 {
			*tags = out.ResourceTagList[0].TagsList
		}
		log.Printf("[DEBUG] Loading CloudTrail tags into a var: %s", *tags)
		return nil
	}
}

func testAccAWSCloudTrailBaseConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid       = "AWSCloudTrailAclCheck"
        Effect    = "Allow"
        Principal = "*"
        Action    = "s3:GetBucketAcl"
        Resource  = "arn:${data.aws_partition.current.partition}:s3:::%[1]s"
      },
      {
        Sid       = "AWSCloudTrailWrite"
        Effect    = "Allow"
        Principal = "*"
        Action    = "s3:PutObject"
        Resource  = "arn:${data.aws_partition.current.partition}:s3:::%[1]s/*"
        Condition = {
          StringEquals = {
            "s3:x-amz-acl" = "bucket-owner-full-control"
          }
        }
      }
    ]
  })
}
`, rName)
}

func testAccAWSCloudTrailConfig(rName string) string {
	return composeConfig(testAccAWSCloudTrailBaseConfig(rName), fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  name           = %[1]q
  s3_bucket_name = aws_s3_bucket.test.id
}
`, rName))
}

func testAccAWSCloudTrailModifiedConfig(rName string) string {
	return composeConfig(testAccAWSCloudTrailBaseConfig(rName), fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  name                          = %[1]q
  s3_bucket_name                = aws_s3_bucket.test.id
  s3_key_prefix                 = "prefix"
  include_global_service_events = false
  enable_logging                = false
}
`, rName))
}

func testAccAWSCloudTrailCloudWatchConfig(rName string) string {
	return composeConfig(testAccAWSCloudTrailBaseConfig(rName), fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  name           = %[1]q
  s3_bucket_name = aws_s3_bucket.test.id

  cloud_watch_logs_group_arn = "${aws_cloudwatch_log_group.test.arn}:*"
  cloud_watch_logs_role_arn  = aws_iam_role.test.arn
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"

    Statement = [{
      Sid    = ""
      Effect = "Allow"
      Action = "sts:AssumeRole"

      Principal = {
        Service = "cloudtrail.${data.aws_partition.current.dns_suffix}"
      }
    }]
  })
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"

    Statement = [{
      Sid      = "AWSCloudTrailCreateLogStream"
      Effect   = "Allow"
      Resource = "${aws_cloudwatch_log_group.test.arn}:*"

      Action = [
        "logs:CreateLogStream",
        "logs:PutLogEvents",
      ]
    }]
  })
}
`, rName))
}

func testAccAWSCloudTrailCloudWatchModifiedConfig(rName string) string {
	return composeConfig(testAccAWSCloudTrailBaseConfig(rName), fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  name           = %[1]q
  s3_bucket_name = aws_s3_bucket.test.id

  cloud_watch_logs_group_arn = "${aws_cloudwatch_log_group.test2.arn}:*"
  cloud_watch_logs_role_arn  = aws_iam_role.test.arn
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_group" "test2" {
  name = "%[1]s-2"
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"

    Statement = [{
      Sid    = ""
      Effect = "Allow"
      Action = "sts:AssumeRole"

      Principal = {
        Service = "cloudtrail.${data.aws_partition.current.dns_suffix}"
      }
    }]
  })
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"

    Statement = [{
      Sid      = "AWSCloudTrailCreateLogStream"
      Effect   = "Allow"
      Resource = "${aws_cloudwatch_log_group.test2.arn}:*"

      Action = [
        "logs:CreateLogStream",
        "logs:PutLogEvents",
      ]
    }]
  })
}
`, rName))
}

func testAccAWSCloudTrailMultiRegionConfig(rName string) string {
	return composeConfig(testAccAWSCloudTrailBaseConfig(rName), fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  name                  = %[1]q
  s3_bucket_name        = aws_s3_bucket.test.id
  is_multi_region_trail = true
}
`, rName))
}

func testAccAWSCloudTrailOrganizationConfig(rName string) string {
	return composeConfig(testAccAWSCloudTrailBaseConfig(rName), fmt.Sprintf(`
resource "aws_organizations_organization" "test" {
  aws_service_access_principals = ["cloudtrail.${data.aws_partition.current.dns_suffix}"]
}

resource "aws_cloudtrail" "test" {
  is_organization_trail = true
  name                  = %[1]q
  s3_bucket_name        = aws_s3_bucket.test.id
}
`, rName))
}

func testAccAWSCloudTrailLogValidationConfig(rName string) string {
	return composeConfig(testAccAWSCloudTrailBaseConfig(rName), fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  name                          = %[1]q
  s3_bucket_name                = aws_s3_bucket.test.id
  is_multi_region_trail         = true
  include_global_service_events = true
  enable_log_file_validation    = true
}
`, rName))
}

func testAccAWSCloudTrailLogValidationModifiedConfig(rName string) string {
	return composeConfig(testAccAWSCloudTrailBaseConfig(rName), fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  name                          = %[1]q
  s3_bucket_name                = aws_s3_bucket.test.id
  include_global_service_events = true
}
`, rName))
}

func testAccAWSCloudTrailKMSKeyConfig(rName string) string {
	return composeConfig(testAccAWSCloudTrailBaseConfig(rName), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description = %[1]q

  policy = jsonencode({
    Version = "2012-10-17"
    Id      = %[1]q

    Statement = [{
      Sid      = "Enable IAM User Permissions"
      Effect   = "Allow"
      Action   = "kms:*"
      Resource = "*"

      Principal = {
        AWS = "*"
      }
    }]
  })
}

resource "aws_cloudtrail" "test" {
  name                          = %[1]q
  s3_bucket_name                = aws_s3_bucket.test.id
  include_global_service_events = true
  kms_key_id                    = aws_kms_key.test.arn
}
`, rName))
}

func testAccAWSCloudTrailGlobalServiceEventsConfig(rName string) string {
	return composeConfig(testAccAWSCloudTrailBaseConfig(rName), fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  name                          = %[1]q
  s3_bucket_name                = aws_s3_bucket.test.id
  include_global_service_events = false
}
`, rName))
}

func testAccAWSCloudTrailTagsConfig(rName string) string {
	return composeConfig(testAccAWSCloudTrailBaseConfig(rName), fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  name           = %[1]q
  s3_bucket_name = aws_s3_bucket.test.id

  tags = {
    Yak = "milk"
    Fox = "tail"
  }
}
`, rName))
}

func testAccAWSCloudTrailTagsModifiedConfig(rName string) string {
	return composeConfig(testAccAWSCloudTrailBaseConfig(rName), fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  name           = %[1]q
  s3_bucket_name = aws_s3_bucket.test.id

  tags = {
    Yak = "milk"
    Fox = "tail"
    Emu = "toes"
  }
}
`, rName))
}

func testAccAWSCloudTrailTagsModifiedAgainConfig(rName string) string {
	return composeConfig(testAccAWSCloudTrailBaseConfig(rName), fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  name           = %[1]q
  s3_bucket_name = aws_s3_bucket.test.id
}
`, rName))
}

func testAccAWSCloudTrailEventSelectorConfig(rName string) string {
	return composeConfig(testAccAWSCloudTrailBaseConfig(rName), fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  name           = %[1]q
  s3_bucket_name = aws_s3_bucket.test.id

  event_selector {
    read_write_type           = "ReadOnly"
    include_management_events = false

    data_resource {
      type = "AWS::S3::Object"

      values = [
        "${aws_s3_bucket.test2.arn}/isen",
        "${aws_s3_bucket.test2.arn}/ko",
      ]
    }
  }
}

resource "aws_s3_bucket" "test2" {
  bucket        = "%[1]s-2"
  force_destroy = true
}
`, rName))
}

func testAccAWSCloudTrailEventSelectorReadWriteTypeConfig(rName string) string {
	return composeConfig(testAccAWSCloudTrailBaseConfig(rName), fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  name           = %[1]q
  s3_bucket_name = aws_s3_bucket.test.id

  event_selector {
    read_write_type           = "WriteOnly"
    include_management_events = true
  }
}
`, rName))
}

func testAccAWSCloudTrailEventSelectorModifiedConfig(rName string) string {
	return composeConfig(testAccAWSCloudTrailBaseConfig(rName), fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  name           = %[1]q
  s3_bucket_name = aws_s3_bucket.test.id

  event_selector {
    read_write_type           = "ReadOnly"
    include_management_events = true

    data_resource {
      type = "AWS::S3::Object"

      values = [
        "${aws_s3_bucket.test2.arn}/isen",
        "${aws_s3_bucket.test2.arn}/ko",
      ]
    }
  }

  event_selector {
    read_write_type           = "All"
    include_management_events = false

    data_resource {
      type = "AWS::S3::Object"

      values = [
        "${aws_s3_bucket.test2.arn}/tf1",
        "${aws_s3_bucket.test2.arn}/tf2",
      ]
    }

    data_resource {
      type = "AWS::Lambda::Function"

      values = [
        aws_lambda_function.test.arn,
      ]
    }
  }
}

resource "aws_s3_bucket" "test2" {
  bucket        = "%[1]s-2"
  force_destroy = true
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"

    Statement = [{
      Sid    = ""
      Effect = "Allow"
      Action = "sts:AssumeRole"

      Principal = {
        Service = "lambda.${data.aws_partition.current.dns_suffix}"
      }
    }]
  })
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.test.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"
}
`, rName))
}

func testAccAWSCloudTrailEventSelectorNoneConfig(rName string) string {
	return composeConfig(testAccAWSCloudTrailBaseConfig(rName), fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  name           = %[1]q
  s3_bucket_name = aws_s3_bucket.test.id
}
`, rName))
}

func testAccAWSCloudTrailEventSelectorDynamoDBConfig(rName string) string {
	return composeConfig(testAccAWSCloudTrailBaseConfig(rName), fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  name           = %[1]q
  s3_bucket_name = aws_s3_bucket.test.id

  event_selector {
    read_write_type           = "All"
    include_management_events = true

    data_resource {
      type = "AWS::DynamoDB::Table"

      values = [
        aws_dynamodb_table.test.arn,
      ]
    }
  }
}

resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = 1
  write_capacity = 1
  hash_key       = %[1]q

  attribute {
    name = %[1]q
    type = "S"
  }
}
`, rName))
}

func testAccAWSCloudTrailInsightSelectorConfig(rName string) string {
	return composeConfig(testAccAWSCloudTrailBaseConfig(rName), fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  name           = %[1]q
  s3_bucket_name = aws_s3_bucket.test.id


  insight_selector {
    insight_type = "ApiCallRateInsight"
  }
}
`, rName))
}
