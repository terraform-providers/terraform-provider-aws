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

	err = conn.ListTrailsPages(&cloudtrail.ListTrailsInput{}, func(page *cloudtrail.ListTrailsOutput, isLast bool) bool {
		if page == nil {
			return !isLast
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

		return !isLast
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
			"basic":                      testAccAWSCloudTrail_basic,
			"cloudwatch":                 testAccAWSCloudTrail_cloudwatch,
			"enableLogging":              testAccAWSCloudTrail_enable_logging,
			"includeGlobalServiceEvents": testAccAWSCloudTrail_include_global_service_events,
			"isMultiRegion":              testAccAWSCloudTrail_is_multi_region,
			"isOrganization":             testAccAWSCloudTrail_is_organization,
			"logValidation":              testAccAWSCloudTrail_logValidation,
			"kmsKey":                     testAccAWSCloudTrail_kmsKey,
			"tags":                       testAccAWSCloudTrail_tags,
			"eventSelector":              testAccAWSCloudTrail_event_selector,
			"insightSelector":            testAccAWSCloudTrail_insight_selector,
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
	cloudTrailRandInt := acctest.RandInt()
	resourceName := "aws_cloudtrail.foobar"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudTrailDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudTrailConfig(cloudTrailRandInt),
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
				Config: testAccAWSCloudTrailConfigModified(cloudTrailRandInt),
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
	randInt := acctest.RandInt()
	resourceName := "aws_cloudtrail.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudTrailDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudTrailConfigCloudWatch(randInt),
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
				Config: testAccAWSCloudTrailConfigCloudWatchModified(randInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					resource.TestCheckResourceAttrSet(resourceName, "cloud_watch_logs_group_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "cloud_watch_logs_role_arn"),
				),
			},
		},
	})
}

func testAccAWSCloudTrail_enable_logging(t *testing.T) {
	var trail cloudtrail.Trail
	cloudTrailRandInt := acctest.RandInt()
	resourceName := "aws_cloudtrail.foobar"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudTrailDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudTrailConfig(cloudTrailRandInt),
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
				Config: testAccAWSCloudTrailConfigModified(cloudTrailRandInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					testAccCheckCloudTrailLoggingEnabled(resourceName, false),
					testAccCheckCloudTrailLogValidationEnabled(resourceName, false, &trail),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
				),
			},
			{
				Config: testAccAWSCloudTrailConfig(cloudTrailRandInt),
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

func testAccAWSCloudTrail_is_multi_region(t *testing.T) {
	var trail cloudtrail.Trail
	cloudTrailRandInt := acctest.RandInt()
	resourceName := "aws_cloudtrail.foobar"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudTrailDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudTrailConfig(cloudTrailRandInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, "is_multi_region_trail", "false"),
					testAccCheckCloudTrailLogValidationEnabled(resourceName, false, &trail),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
				),
			},
			{
				Config: testAccAWSCloudTrailConfigMultiRegion(cloudTrailRandInt),
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
				Config: testAccAWSCloudTrailConfig(cloudTrailRandInt),
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

func testAccAWSCloudTrail_is_organization(t *testing.T) {
	var trail cloudtrail.Trail
	cloudTrailRandInt := acctest.RandInt()
	resourceName := "aws_cloudtrail.foobar"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudTrailDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudTrailConfigOrganization(cloudTrailRandInt),
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
				Config: testAccAWSCloudTrailConfig(cloudTrailRandInt),
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
	cloudTrailRandInt := acctest.RandInt()
	resourceName := "aws_cloudtrail.foobar"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudTrailDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudTrailConfig_logValidation(cloudTrailRandInt),
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
				Config: testAccAWSCloudTrailConfig_logValidationModified(cloudTrailRandInt),
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
	cloudTrailRandInt := acctest.RandInt()

	resourceName := "aws_cloudtrail.foobar"
	kmsResourceName := "aws_kms_key.foo"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudTrailDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudTrailConfig_kmsKey(cloudTrailRandInt),
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
	cloudTrailRandInt := acctest.RandInt()
	resourceName := "aws_cloudtrail.foobar"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudTrailDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudTrailConfig_tags(cloudTrailRandInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					testAccCheckCloudTrailLoadTags(&trail, &trailTags),
					resource.TestCheckResourceAttr(resourceName, "tags.Foo", "moo"),
					resource.TestCheckResourceAttr(resourceName, "tags.Pooh", "hi"),
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
				Config: testAccAWSCloudTrailConfig_tagsModified(cloudTrailRandInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					testAccCheckCloudTrailLoadTags(&trail, &trailTagsModified),
					resource.TestCheckResourceAttr(resourceName, "tags.Foo", "moo"),
					resource.TestCheckResourceAttr(resourceName, "tags.Moo", "boom"),
					resource.TestCheckResourceAttr(resourceName, "tags.Pooh", "hi"),
					testAccCheckCloudTrailLogValidationEnabled(resourceName, false, &trail),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
				),
			},
			{
				Config: testAccAWSCloudTrailConfig_tagsModifiedAgain(cloudTrailRandInt),
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

func testAccAWSCloudTrail_include_global_service_events(t *testing.T) {
	var trail cloudtrail.Trail
	cloudTrailRandInt := acctest.RandInt()
	resourceName := "aws_cloudtrail.foobar"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudTrailDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudTrailConfig_include_global_service_events(cloudTrailRandInt),
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

func testAccAWSCloudTrail_event_selector(t *testing.T) {
	cloudTrailRandInt := acctest.RandInt()
	resourceName := "aws_cloudtrail.foobar"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudTrailDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudTrailConfig_eventSelector(cloudTrailRandInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "event_selector.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.data_resource.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.data_resource.0.type", "AWS::S3::Object"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.data_resource.0.values.#", "2"),
					resource.TestMatchResourceAttr(resourceName, "event_selector.0.data_resource.0.values.0", regexp.MustCompile(`^arn:[^:]+:s3:::.+/foobar$`)),
					resource.TestMatchResourceAttr(resourceName, "event_selector.0.data_resource.0.values.1", regexp.MustCompile(`^arn:[^:]+:s3:::.+/baz$`)),
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
				Config: testAccAWSCloudTrailConfig_eventSelectorReadWriteType(cloudTrailRandInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "event_selector.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.include_management_events", "true"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.read_write_type", "WriteOnly"),
				),
			},
			{
				Config: testAccAWSCloudTrailConfig_eventSelectorModified(cloudTrailRandInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "event_selector.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.data_resource.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.data_resource.0.type", "AWS::S3::Object"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.data_resource.0.values.#", "2"),
					resource.TestMatchResourceAttr(resourceName, "event_selector.0.data_resource.0.values.0", regexp.MustCompile(`^arn:[^:]+:s3:::.+/foobar$`)),
					resource.TestMatchResourceAttr(resourceName, "event_selector.0.data_resource.0.values.1", regexp.MustCompile(`^arn:[^:]+:s3:::.+/baz$`)),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.include_management_events", "true"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.read_write_type", "ReadOnly"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.1.data_resource.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.1.data_resource.0.type", "AWS::S3::Object"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.1.data_resource.0.values.#", "2"),
					resource.TestMatchResourceAttr(resourceName, "event_selector.1.data_resource.0.values.0", regexp.MustCompile(`^arn:[^:]+:s3:::.+/tf1$`)),
					resource.TestMatchResourceAttr(resourceName, "event_selector.1.data_resource.0.values.1", regexp.MustCompile(`^arn:[^:]+:s3:::.+/tf2$`)),
					resource.TestCheckResourceAttr(resourceName, "event_selector.1.data_resource.1.type", "AWS::Lambda::Function"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.1.data_resource.1.values.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "event_selector.1.data_resource.1.values.0", regexp.MustCompile(`^arn:[^:]+:lambda:.+:tf-test-trail-event-select-\d+$`)),
					resource.TestCheckResourceAttr(resourceName, "event_selector.1.include_management_events", "false"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.1.read_write_type", "All"),
				),
			},
			{
				Config: testAccAWSCloudTrailConfig_eventSelectorNone(cloudTrailRandInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "event_selector.#", "0"),
				),
			},
		},
	})
}

func testAccAWSCloudTrail_insight_selector(t *testing.T) {
	resourceName := "aws_cloudtrail.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudTrailDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudTrailConfig_insightSelector(rName),
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

func testAccAWSCloudTrailConfig(cloudTrailRandInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudtrail" "foobar" {
  name           = "tf-trail-foobar-%[1]d"
  s3_bucket_name = aws_s3_bucket.foo.id
}

data "aws_partition" "current" {}

resource "aws_s3_bucket" "foo" {
  bucket        = "tf-test-trail-%[1]d"
  force_destroy = true

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AWSCloudTrailAclCheck",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:GetBucketAcl",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::tf-test-trail-%[1]d"
    },
    {
      "Sid": "AWSCloudTrailWrite",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:PutObject",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::tf-test-trail-%[1]d/*",
      "Condition": {
        "StringEquals": {
          "s3:x-amz-acl": "bucket-owner-full-control"
        }
      }
    }
  ]
}
POLICY
}
`, cloudTrailRandInt)
}

func testAccAWSCloudTrailConfigModified(cloudTrailRandInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudtrail" "foobar" {
  name                          = "tf-trail-foobar-%[1]d"
  s3_bucket_name                = aws_s3_bucket.foo.id
  s3_key_prefix                 = "prefix"
  include_global_service_events = false
  enable_logging                = false
}

data "aws_partition" "current" {}

resource "aws_s3_bucket" "foo" {
  bucket        = "tf-test-trail-%[1]d"
  force_destroy = true

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AWSCloudTrailAclCheck",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:GetBucketAcl",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::tf-test-trail-%[1]d"
    },
    {
      "Sid": "AWSCloudTrailWrite",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:PutObject",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::tf-test-trail-%[1]d/*",
      "Condition": {
        "StringEquals": {
          "s3:x-amz-acl": "bucket-owner-full-control"
        }
      }
    }
  ]
}
POLICY
}
`, cloudTrailRandInt)
}

func testAccAWSCloudTrailConfigCloudWatch(randInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  name           = "tf-acc-test-%[1]d"
  s3_bucket_name = aws_s3_bucket.test.id

  cloud_watch_logs_group_arn = "${aws_cloudwatch_log_group.test.arn}:*"
  cloud_watch_logs_role_arn  = aws_iam_role.test.arn
}

data "aws_partition" "current" {}

resource "aws_s3_bucket" "test" {
  bucket        = "tf-test-trail-%[1]d"
  force_destroy = true

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AWSCloudTrailAclCheck",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:GetBucketAcl",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::tf-test-trail-%[1]d"
    },
    {
      "Sid": "AWSCloudTrailWrite",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:PutObject",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::tf-test-trail-%[1]d/*",
      "Condition": {
        "StringEquals": {
          "s3:x-amz-acl": "bucket-owner-full-control"
        }
      }
    }
  ]
}
POLICY
}

resource "aws_cloudwatch_log_group" "test" {
  name = "tf-acc-test-cloudtrail-%[1]d"
}

resource "aws_iam_role" "test" {
  name = "tf-acc-test-cloudtrail-%[1]d"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "cloudtrail.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy" "test" {
  name = "tf-acc-test-cloudtrail-%[1]d"
  role = aws_iam_role.test.id

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AWSCloudTrailCreateLogStream",
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "${aws_cloudwatch_log_group.test.arn}:*"
    }
  ]
}
POLICY
}
`, randInt)
}

func testAccAWSCloudTrailConfigCloudWatchModified(randInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  name           = "tf-acc-test-%[1]d"
  s3_bucket_name = aws_s3_bucket.test.id

  cloud_watch_logs_group_arn = "${aws_cloudwatch_log_group.second.arn}:*"
  cloud_watch_logs_role_arn  = aws_iam_role.test.arn
}

data "aws_partition" "current" {}

resource "aws_s3_bucket" "test" {
  bucket        = "tf-test-trail-%[1]d"
  force_destroy = true

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AWSCloudTrailAclCheck",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:GetBucketAcl",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::tf-test-trail-%[1]d"
    },
    {
      "Sid": "AWSCloudTrailWrite",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:PutObject",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::tf-test-trail-%[1]d/*",
      "Condition": {
        "StringEquals": {
          "s3:x-amz-acl": "bucket-owner-full-control"
        }
      }
    }
  ]
}
POLICY
}

resource "aws_cloudwatch_log_group" "test" {
  name = "tf-acc-test-cloudtrail-%[1]d"
}

resource "aws_cloudwatch_log_group" "second" {
  name = "tf-acc-test-cloudtrail-second-%[1]d"
}

resource "aws_iam_role" "test" {
  name = "tf-acc-test-cloudtrail-%[1]d"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "cloudtrail.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy" "test" {
  name = "tf-acc-test-cloudtrail-%[1]d"
  role = aws_iam_role.test.id

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AWSCloudTrailCreateLogStream",
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "${aws_cloudwatch_log_group.second.arn}:*"
    }
  ]
}
POLICY
}
`, randInt)
}

func testAccAWSCloudTrailConfigMultiRegion(cloudTrailRandInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudtrail" "foobar" {
  name                  = "tf-trail-foobar-%[1]d"
  s3_bucket_name        = aws_s3_bucket.foo.id
  is_multi_region_trail = true
}

data "aws_partition" "current" {}

resource "aws_s3_bucket" "foo" {
  bucket        = "tf-test-trail-%[1]d"
  force_destroy = true

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AWSCloudTrailAclCheck",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:GetBucketAcl",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::tf-test-trail-%[1]d"
    },
    {
      "Sid": "AWSCloudTrailWrite",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:PutObject",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::tf-test-trail-%[1]d/*",
      "Condition": {
        "StringEquals": {
          "s3:x-amz-acl": "bucket-owner-full-control"
        }
      }
    }
  ]
}
POLICY
}
`, cloudTrailRandInt)
}

func testAccAWSCloudTrailConfigOrganization(cloudTrailRandInt int) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_organizations_organization" "test" {
  aws_service_access_principals = ["cloudtrail.${data.aws_partition.current.dns_suffix}"]
}

resource "aws_cloudtrail" "foobar" {
  is_organization_trail = true
  name                  = "tf-trail-foobar-%[1]d"
  s3_bucket_name        = aws_s3_bucket.foo.id
}

resource "aws_s3_bucket" "foo" {
  bucket        = "tf-test-trail-%[1]d"
  force_destroy = true

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AWSCloudTrailAclCheck",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:GetBucketAcl",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::tf-test-trail-%[1]d"
    },
    {
      "Sid": "AWSCloudTrailWrite",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:PutObject",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::tf-test-trail-%[1]d/*",
      "Condition": {
        "StringEquals": {
          "s3:x-amz-acl": "bucket-owner-full-control"
        }
      }
    }
  ]
}
POLICY
}
`, cloudTrailRandInt)
}

func testAccAWSCloudTrailConfig_logValidation(cloudTrailRandInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudtrail" "foobar" {
  name                          = "tf-acc-trail-log-validation-test-%[1]d"
  s3_bucket_name                = aws_s3_bucket.foo.id
  is_multi_region_trail         = true
  include_global_service_events = true
  enable_log_file_validation    = true
}

data "aws_partition" "current" {}

resource "aws_s3_bucket" "foo" {
  bucket        = "tf-test-trail-%[1]d"
  force_destroy = true

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AWSCloudTrailAclCheck",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:GetBucketAcl",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::tf-test-trail-%[1]d"
    },
    {
      "Sid": "AWSCloudTrailWrite",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:PutObject",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::tf-test-trail-%[1]d/*",
      "Condition": {
        "StringEquals": {
          "s3:x-amz-acl": "bucket-owner-full-control"
        }
      }
    }
  ]
}
POLICY
}
`, cloudTrailRandInt)
}

func testAccAWSCloudTrailConfig_logValidationModified(cloudTrailRandInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudtrail" "foobar" {
  name                          = "tf-acc-trail-log-validation-test-%[1]d"
  s3_bucket_name                = aws_s3_bucket.foo.id
  include_global_service_events = true
}

data "aws_partition" "current" {}

resource "aws_s3_bucket" "foo" {
  bucket        = "tf-test-trail-%[1]d"
  force_destroy = true

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AWSCloudTrailAclCheck",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:GetBucketAcl",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::tf-test-trail-%[1]d"
    },
    {
      "Sid": "AWSCloudTrailWrite",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:PutObject",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::tf-test-trail-%[1]d/*",
      "Condition": {
        "StringEquals": {
          "s3:x-amz-acl": "bucket-owner-full-control"
        }
      }
    }
  ]
}
POLICY
}
`, cloudTrailRandInt)
}

func testAccAWSCloudTrailConfig_kmsKey(cloudTrailRandInt int) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "foo" {
  description = "Terraform acc test %[1]d"

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}

data "aws_partition" "current" {}

resource "aws_cloudtrail" "foobar" {
  name                          = "tf-acc-trail-log-validation-test-%[1]d"
  s3_bucket_name                = aws_s3_bucket.foo.id
  include_global_service_events = true
  kms_key_id                    = aws_kms_key.foo.arn
}

resource "aws_s3_bucket" "foo" {
  bucket        = "tf-test-trail-%[1]d"
  force_destroy = true

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AWSCloudTrailAclCheck",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:GetBucketAcl",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::tf-test-trail-%[1]d"
    },
    {
      "Sid": "AWSCloudTrailWrite",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:PutObject",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::tf-test-trail-%[1]d/*",
      "Condition": {
        "StringEquals": {
          "s3:x-amz-acl": "bucket-owner-full-control"
        }
      }
    }
  ]
}
POLICY
}
`, cloudTrailRandInt)
}

var testAccAWSCloudTrailConfig_tags_tpl = `
resource "aws_cloudtrail" "foobar" {
  name           = "tf-acc-trail-log-validation-test-%[1]d"
  s3_bucket_name = aws_s3_bucket.foo.id
  %[2]s
}

data "aws_partition" "current" {}

resource "aws_s3_bucket" "foo" {
  bucket        = "tf-test-trail-%[1]d"
  force_destroy = true

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AWSCloudTrailAclCheck",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:GetBucketAcl",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::tf-test-trail-%[1]d"
    },
    {
      "Sid": "AWSCloudTrailWrite",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:PutObject",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::tf-test-trail-%[1]d/*",
      "Condition": {
        "StringEquals": {
          "s3:x-amz-acl": "bucket-owner-full-control"
        }
      }
    }
  ]
}
POLICY
}
`

func testAccAWSCloudTrailConfig_include_global_service_events(cloudTrailRandInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudtrail" "foobar" {
  name                          = "tf-trail-foobar-%[1]d"
  s3_bucket_name                = aws_s3_bucket.foo.id
  include_global_service_events = false
}

data "aws_partition" "current" {}

resource "aws_s3_bucket" "foo" {
  bucket        = "tf-test-trail-%[1]d"
  force_destroy = true

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AWSCloudTrailAclCheck",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:GetBucketAcl",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::tf-test-trail-%[1]d"
    },
    {
      "Sid": "AWSCloudTrailWrite",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:PutObject",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::tf-test-trail-%[1]d/*",
      "Condition": {
        "StringEquals": {
          "s3:x-amz-acl": "bucket-owner-full-control"
        }
      }
    }
  ]
}
POLICY
}
`, cloudTrailRandInt)
}

func testAccAWSCloudTrailConfig_tags(cloudTrailRandInt int) string {
	tagsString := `
tags = {
  Foo  = "moo"
  Pooh = "hi"
}
`

	return fmt.Sprintf(testAccAWSCloudTrailConfig_tags_tpl, cloudTrailRandInt, tagsString)
}

func testAccAWSCloudTrailConfig_tagsModified(cloudTrailRandInt int) string {
	tagsString := `
tags = {
  Foo  = "moo"
  Pooh = "hi"
  Moo  = "boom"
}
`

	return fmt.Sprintf(testAccAWSCloudTrailConfig_tags_tpl, cloudTrailRandInt, tagsString)
}

func testAccAWSCloudTrailConfig_tagsModifiedAgain(cloudTrailRandInt int) string {
	return fmt.Sprintf(testAccAWSCloudTrailConfig_tags_tpl, cloudTrailRandInt, "")
}

func testAccAWSCloudTrailConfig_eventSelector(cloudTrailRandInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudtrail" "foobar" {
  name           = "tf-trail-foobar-%[1]d"
  s3_bucket_name = aws_s3_bucket.foo.id

  event_selector {
    read_write_type           = "ReadOnly"
    include_management_events = false

    data_resource {
      type = "AWS::S3::Object"

      values = [
        "${aws_s3_bucket.bar.arn}/foobar",
        "${aws_s3_bucket.bar.arn}/baz",
      ]
    }
  }
}

data "aws_partition" "current" {}

resource "aws_s3_bucket" "foo" {
  bucket        = "tf-test-trail-%[1]d"
  force_destroy = true

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AWSCloudTrailAclCheck",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:GetBucketAcl",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::tf-test-trail-%[1]d"
    },
    {
      "Sid": "AWSCloudTrailWrite",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:PutObject",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::tf-test-trail-%[1]d/*",
      "Condition": {
        "StringEquals": {
          "s3:x-amz-acl": "bucket-owner-full-control"
        }
      }
    }
  ]
}
POLICY
}

resource "aws_s3_bucket" "bar" {
  bucket        = "tf-test-trail-event-select-%[1]d"
  force_destroy = true
}
`, cloudTrailRandInt)
}

func testAccAWSCloudTrailConfig_eventSelectorReadWriteType(cloudTrailRandInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudtrail" "foobar" {
  name           = "tf-trail-foobar-%[1]d"
  s3_bucket_name = aws_s3_bucket.foo.id

  event_selector {
    read_write_type           = "WriteOnly"
    include_management_events = true
  }
}

data "aws_partition" "current" {}

resource "aws_s3_bucket" "foo" {
  bucket        = "tf-test-trail-%[1]d"
  force_destroy = true

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AWSCloudTrailAclCheck",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:GetBucketAcl",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::tf-test-trail-%[1]d"
    },
    {
      "Sid": "AWSCloudTrailWrite",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:PutObject",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::tf-test-trail-%[1]d/*",
      "Condition": {
        "StringEquals": {
          "s3:x-amz-acl": "bucket-owner-full-control"
        }
      }
    }
  ]
}
POLICY
}
`, cloudTrailRandInt)
}

func testAccAWSCloudTrailConfig_eventSelectorModified(cloudTrailRandInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudtrail" "foobar" {
  name           = "tf-trail-foobar-%[1]d"
  s3_bucket_name = aws_s3_bucket.foo.id

  event_selector {
    read_write_type           = "ReadOnly"
    include_management_events = true

    data_resource {
      type = "AWS::S3::Object"

      values = [
        "${aws_s3_bucket.bar.arn}/foobar",
        "${aws_s3_bucket.bar.arn}/baz",
      ]
    }
  }

  event_selector {
    read_write_type           = "All"
    include_management_events = false

    data_resource {
      type = "AWS::S3::Object"

      values = [
        "${aws_s3_bucket.bar.arn}/tf1",
        "${aws_s3_bucket.bar.arn}/tf2",
      ]
    }

    data_resource {
      type = "AWS::Lambda::Function"

      values = [
        aws_lambda_function.lambda_function_test.arn,
      ]
    }
  }
}

data "aws_partition" "current" {}

resource "aws_s3_bucket" "foo" {
  bucket        = "tf-test-trail-%[1]d"
  force_destroy = true

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AWSCloudTrailAclCheck",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:GetBucketAcl",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::tf-test-trail-%[1]d"
    },
    {
      "Sid": "AWSCloudTrailWrite",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:PutObject",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::tf-test-trail-%[1]d/*",
      "Condition": {
        "StringEquals": {
          "s3:x-amz-acl": "bucket-owner-full-control"
        }
      }
    }
  ]
}
POLICY
}

resource "aws_s3_bucket" "bar" {
  bucket        = "tf-test-trail-event-select-%[1]d"
  force_destroy = true
}

resource "aws_iam_role" "iam_for_lambda" {
  name = "tf-test-trail-event-select-%[1]d"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_lambda_function" "lambda_function_test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "tf-test-trail-event-select-%[1]d"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"
}
`, cloudTrailRandInt)
}

func testAccAWSCloudTrailConfig_eventSelectorNone(cloudTrailRandInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudtrail" "foobar" {
  name           = "tf-trail-foobar-%[1]d"
  s3_bucket_name = aws_s3_bucket.foo.id
}

data "aws_partition" "current" {}

resource "aws_s3_bucket" "foo" {
  bucket        = "tf-test-trail-%[1]d"
  force_destroy = true

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AWSCloudTrailAclCheck",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:GetBucketAcl",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::tf-test-trail-%[1]d"
    },
    {
      "Sid": "AWSCloudTrailWrite",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:PutObject",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::tf-test-trail-%[1]d/*",
      "Condition": {
        "StringEquals": {
          "s3:x-amz-acl": "bucket-owner-full-control"
        }
      }
    }
  ]
}
POLICY
}
`, cloudTrailRandInt)
}

func testAccAWSCloudTrailConfig_insightSelector(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  name           = %[1]q
  s3_bucket_name = aws_s3_bucket.test.id


  insight_selector {
    insight_type = "ApiCallRateInsight"
  }
}

data "aws_partition" "current" {}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true

  policy = <<POLICY
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Sid": "AWSCloudTrailAclCheck",
			"Effect": "Allow",
			"Principal": "*",
			"Action": "s3:GetBucketAcl",
			"Resource": "arn:${data.aws_partition.current.partition}:s3:::%[1]s"
		},
		{
			"Sid": "AWSCloudTrailWrite",
			"Effect": "Allow",
			"Principal": "*",
			"Action": "s3:PutObject",
			"Resource": "arn:${data.aws_partition.current.partition}:s3:::%[1]s/*",
			"Condition": {
				"StringEquals": {
					"s3:x-amz-acl": "bucket-owner-full-control"
				}
			}
		}
	]
}
POLICY
}
`, rName)
}
