package aws

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func init() {
	resource.AddTestSweepers("aws_s3_bucket", &resource.Sweeper{
		Name: "aws_s3_bucket",
		F:    testSweepS3Buckets,
		Dependencies: []string{
			"aws_s3_access_point",
			"aws_s3_bucket_object",
		},
	})
}

func testSweepS3Buckets(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*AWSClient).s3conn
	input := &s3.ListBucketsInput{}

	output, err := conn.ListBuckets(input)

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping S3 Buckets sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing S3 Buckets: %s", err)
	}

	if len(output.Buckets) == 0 {
		log.Print("[DEBUG] No S3 Buckets to sweep")
		return nil
	}

	defaultNameRegexp := regexp.MustCompile(`^terraform-\d+$`)
	for _, bucket := range output.Buckets {
		name := aws.StringValue(bucket.Name)

		sweepable := false
		prefixes := []string{"mybucket.", "mylogs.", "tf-acc", "tf-object-test", "tf-test", "tf-emr-bootstrap", "terraform-remote-s3-test"}

		for _, prefix := range prefixes {
			if strings.HasPrefix(name, prefix) {
				sweepable = true
				break
			}
		}

		if defaultNameRegexp.MatchString(name) {
			sweepable = true
		}

		if !sweepable {
			log.Printf("[INFO] Skipping S3 Bucket: %s", name)
			continue
		}

		bucketRegion, err := testS3BucketRegion(conn, name)

		if err != nil {
			log.Printf("[ERROR] Error getting S3 Bucket (%s) Location: %s", name, err)
			continue
		}

		if bucketRegion != region {
			log.Printf("[INFO] Skipping S3 Bucket (%s) in different region: %s", name, bucketRegion)
			continue
		}

		input := &s3.DeleteBucketInput{
			Bucket: bucket.Name,
		}

		log.Printf("[INFO] Deleting S3 Bucket: %s", name)
		err = resource.Retry(1*time.Minute, func() *resource.RetryError {
			_, err := conn.DeleteBucket(input)

			if isAWSErr(err, s3.ErrCodeNoSuchBucket, "") {
				return nil
			}

			if isAWSErr(err, "BucketNotEmpty", "") {
				return resource.RetryableError(err)
			}

			if err != nil {
				return resource.NonRetryableError(err)
			}

			return nil
		})

		if err != nil {
			return fmt.Errorf("error deleting S3 Bucket (%s): %s", name, err)
		}
	}

	return nil
}

func testS3BucketRegion(conn *s3.S3, bucket string) (string, error) {
	input := &s3.GetBucketLocationInput{
		Bucket: aws.String(bucket),
	}

	output, err := conn.GetBucketLocation(input)

	if err != nil {
		return "", err
	}

	if output == nil || output.LocationConstraint == nil {
		return "us-east-1", nil
	}

	return aws.StringValue(output.LocationConstraint), nil
}

func testS3BucketObjectLockEnabled(conn *s3.S3, bucket string) (bool, error) {
	input := &s3.GetObjectLockConfigurationInput{
		Bucket: aws.String(bucket),
	}

	output, err := conn.GetObjectLockConfiguration(input)

	if isAWSErr(err, "ObjectLockConfigurationNotFoundError", "") {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return aws.StringValue(output.ObjectLockConfiguration.ObjectLockEnabled) == s3.ObjectLockEnabledEnabled, nil
}

func TestAccAWSS3Bucket_basic(t *testing.T) {
	rInt := acctest.RandInt()
	arnRegexp := regexp.MustCompile(`^arn:aws[\w-]*:s3:::`)
	region := testAccGetRegion()
	hostedZoneID, _ := HostedZoneIDForRegion(region)
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		/*
			IDRefreshName:   resourceName,
			IDRefreshIgnore: []string{"force_destroy"},
		*/
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "hosted_zone_id", hostedZoneID),
					resource.TestCheckResourceAttr(
						resourceName, "region", region),
					resource.TestCheckNoResourceAttr(
						resourceName, "website_endpoint"),
					resource.TestMatchResourceAttr(
						resourceName, "arn", arnRegexp),
					resource.TestCheckResourceAttr(
						resourceName, "bucket", testAccBucketName(rInt)),
					testAccCheckS3BucketDomainName(
						resourceName, "bucket_domain_name", testAccBucketName(rInt)),
					resource.TestCheckResourceAttr(
						resourceName, "bucket_regional_domain_name", testAccBucketRegionalDomainName(rInt, region)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy", "acl"},
			},
		},
	})
}

// Support for common Terraform 0.11 pattern
// Reference: https://github.com/terraform-providers/terraform-provider-aws/issues/7868
func TestAccAWSS3Bucket_Bucket_EmptyString(t *testing.T) {
	resourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigBucketEmptyString,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					resource.TestMatchResourceAttr(resourceName, "bucket", regexp.MustCompile("^terraform-")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy", "acl"},
			},
		},
	})
}

func TestAccAWSS3Bucket_tagsWithNoSystemTags(t *testing.T) {
	resourceName := "aws_s3_bucket.bucket"
	rInt := acctest.RandInt()
	bucketName := fmt.Sprintf("tf-test-bucket-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfig_withTags(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "AAA"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "CCC"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy", "acl"},
			},
			{
				Config: testAccAWSS3BucketConfig_withUpdatedTags(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "4"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "XXX"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key4", "DDD"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key5", "EEE"),
				),
			},
			{

				Config: testAccAWSS3BucketConfig_withNoTags(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			// Verify update from 0 tags.
			{
				Config: testAccAWSS3BucketConfig_withTags(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "AAA"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "CCC"),
				),
			},
		},
	})
}

func TestAccAWSS3Bucket_tagsWithSystemTags(t *testing.T) {
	resourceName := "aws_s3_bucket.bucket"
	rInt := acctest.RandInt()
	bucketName := fmt.Sprintf("tf-test-bucket-%d", rInt)
	var stackId string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckAWSS3BucketDestroy,
			func(s *terraform.State) error {
				// Tear down CF stack.
				conn := testAccProvider.Meta().(*AWSClient).cfconn

				req := &cloudformation.DeleteStackInput{
					StackName: aws.String(stackId),
				}

				log.Printf("[DEBUG] Deleting CloudFormation stack: %#v", req)
				if _, err := conn.DeleteStack(req); err != nil {
					return fmt.Errorf("Error deleting CloudFormation stack: %s", err)
				}

				if err := waitForCloudFormationStackDeletion(conn, stackId, 10*time.Minute); err != nil {
					return fmt.Errorf("Error waiting for CloudFormation stack deletion: %s", err)
				}

				return nil
			},
		),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfig_withNoTags(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					testAccCheckAWSS3DestroyBucket(resourceName),
					testAccCheckAWSS3BucketCreateViaCloudFormation(bucketName, &stackId),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy", "acl"},
			},
			{
				Config: testAccAWSS3BucketConfig_withTags(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "AAA"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "CCC"),
					testAccCheckAWSS3BucketTagKeys(resourceName, "aws:cloudformation:stack-name", "aws:cloudformation:stack-id", "aws:cloudformation:logical-id"),
				),
			},
			{
				Config: testAccAWSS3BucketConfig_withUpdatedTags(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "4"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "XXX"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key4", "DDD"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key5", "EEE"),
					testAccCheckAWSS3BucketTagKeys(resourceName, "aws:cloudformation:stack-name", "aws:cloudformation:stack-id", "aws:cloudformation:logical-id"),
				),
			},
			{

				Config: testAccAWSS3BucketConfig_withNoTags(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					testAccCheckAWSS3BucketTagKeys(resourceName, "aws:cloudformation:stack-name", "aws:cloudformation:stack-id", "aws:cloudformation:logical-id"),
				),
			},
		},
	})
}

func TestAccAWSS3MultiBucket_withTags(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "aws_s3_bucket.bucket1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3MultiBucketConfigWithTags(rInt),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy", "acl"},
			},
		},
	})
}

func TestAccAWSS3Bucket_namePrefix(t *testing.T) {
	resourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfig_namePrefix,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					resource.TestMatchResourceAttr(
						resourceName, "bucket", regexp.MustCompile("^tf-test-")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy", "acl", "bucket_prefix"},
			},
		},
	})
}

func TestAccAWSS3Bucket_generatedName(t *testing.T) {
	resourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfig_generatedName,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy", "acl", "bucket_prefix"},
			},
		},
	})
}

func TestAccAWSS3Bucket_region(t *testing.T) {
	resourceName := "aws_s3_bucket.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigWithRegion(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "region", testAccGetRegion()),
				),
			},
		},
	})
}

func TestAccAWSS3Bucket_acceleration(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck("cloudfront", t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigWithAcceleration(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "acceleration_status", "Enabled"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy", "acl"},
			},
			{
				Config: testAccAWSS3BucketConfigWithoutAcceleration(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "acceleration_status", "Suspended"),
				),
			},
		},
	})
}

func TestAccAWSS3Bucket_RequestPayer(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigRequestPayerBucketOwner(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName,
						"request_payer",
						"BucketOwner"),
					testAccCheckAWSS3RequestPayer(
						resourceName,
						"BucketOwner"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy", "acl"},
			},
			{
				Config: testAccAWSS3BucketConfigRequestPayerRequester(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName,
						"request_payer",
						"Requester"),
					testAccCheckAWSS3RequestPayer(
						resourceName,
						"Requester"),
				),
			},
		},
	})
}

func TestAccAWSS3Bucket_Policy(t *testing.T) {
	rInt := acctest.RandInt()
	partition := testAccGetPartition()
	resourceName := "aws_s3_bucket.bucket"

	checkFn := func(s []*terraform.InstanceState) error {
		// Expect 2: bucket + policy
		if len(s) != 2 {
			return fmt.Errorf("expected 2 states: %#v", s)
		}
		bucketState, policyState := s[0], s[1]

		expectedBucketId := fmt.Sprintf("tf-test-bucket-%d", rInt)

		if bucketState.ID != expectedBucketId {
			return fmt.Errorf("expected bucket of ID %s, %s received",
				expectedBucketId, bucketState.ID)
		}

		if policyState.ID != expectedBucketId {
			return fmt.Errorf("expected policy of ID %s, %s received",
				expectedBucketId, bucketState.ID)
		}

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigWithPolicy(rInt, partition),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					testAccCheckAWSS3BucketPolicy(
						resourceName, testAccAWSS3BucketPolicy(rInt, partition)),
				),
			},
			{
				ResourceName:     resourceName,
				ImportState:      true,
				ImportStateCheck: checkFn,
			},
			{
				Config: testAccAWSS3BucketConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					testAccCheckAWSS3BucketPolicy(
						resourceName, ""),
				),
			},
			{
				Config: testAccAWSS3BucketConfigWithEmptyPolicy(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					testAccCheckAWSS3BucketPolicy(
						resourceName, ""),
				),
			},
		},
	})
}

func TestAccAWSS3Bucket_UpdateAcl(t *testing.T) {
	ri := acctest.RandInt()
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigWithAcl(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "acl", "public-read"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy", "acl", "grant"},
			},
			{
				Config: testAccAWSS3BucketConfigWithAclUpdate(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "acl", "private"),
				),
			},
		},
	})
}

func TestAccAWSS3Bucket_UpdateGrant(t *testing.T) {
	resourceName := "aws_s3_bucket.bucket"
	ri := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigWithGrants(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "grant.#", "1"),
					testAccCheckAWSS3BucketUpdateGrantSingle(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy", "acl"},
			},
			{
				Config: testAccAWSS3BucketConfigWithGrantsUpdate(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "grant.#", "2"),
					testAccCheckAWSS3BucketUpdateGrantMulti(resourceName),
				),
			},
			{
				Config: testAccAWSS3BucketBasic(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "grant.#", "0"),
				),
			},
		},
	})
}

func TestAccAWSS3Bucket_AclToGrant(t *testing.T) {
	resourceName := "aws_s3_bucket.bucket"
	ri := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigWithAcl(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "acl", "public-read"),
					resource.TestCheckResourceAttr(resourceName, "grant.#", "0"),
				),
			},
			{
				Config: testAccAWSS3BucketConfigWithGrants(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "grant.#", "1"),
					// check removed ACLs
				),
			},
		},
	})
}

func TestAccAWSS3Bucket_GrantToAcl(t *testing.T) {
	resourceName := "aws_s3_bucket.bucket"
	ri := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigWithGrants(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "grant.#", "1"),
				),
			},
			{
				Config: testAccAWSS3BucketConfigWithAcl(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "acl", "public-read"),
					resource.TestCheckResourceAttr(resourceName, "grant.#", "0"),
					// check removed grants
				),
			},
		},
	})
}

func TestAccAWSS3Bucket_Website_Simple(t *testing.T) {
	rInt := acctest.RandInt()
	region := testAccGetRegion()
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketWebsiteConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					testAccCheckAWSS3BucketWebsite(
						resourceName, "index.html", "", "", ""),
					testAccCheckS3BucketWebsiteEndpoint(
						resourceName, "website_endpoint", testAccBucketName(rInt), region),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy", "acl", "grant"},
			},
			{
				Config: testAccAWSS3BucketWebsiteConfigWithError(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					testAccCheckAWSS3BucketWebsite(
						resourceName, "index.html", "error.html", "", ""),
					testAccCheckS3BucketWebsiteEndpoint(
						resourceName, "website_endpoint", testAccBucketName(rInt), region),
				),
			},
			{
				Config: testAccAWSS3BucketConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					testAccCheckAWSS3BucketWebsite(
						resourceName, "", "", "", ""),
					resource.TestCheckResourceAttr(
						resourceName, "website_endpoint", ""),
				),
			},
		},
	})
}

func TestAccAWSS3Bucket_WebsiteRedirect(t *testing.T) {
	rInt := acctest.RandInt()
	region := testAccGetRegion()
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketWebsiteConfigWithRedirect(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					testAccCheckAWSS3BucketWebsite(
						resourceName, "", "", "", "hashicorp.com?my=query"),
					testAccCheckS3BucketWebsiteEndpoint(
						resourceName, "website_endpoint", testAccBucketName(rInt), region),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy", "acl", "grant"},
			},
			{
				Config: testAccAWSS3BucketWebsiteConfigWithHttpsRedirect(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					testAccCheckAWSS3BucketWebsite(
						resourceName, "", "", "https", "hashicorp.com?my=query"),
					testAccCheckS3BucketWebsiteEndpoint(
						resourceName, "website_endpoint", testAccBucketName(rInt), region),
				),
			},
			{
				Config: testAccAWSS3BucketConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					testAccCheckAWSS3BucketWebsite(
						resourceName, "", "", "", ""),
					resource.TestCheckResourceAttr(
						resourceName, "website_endpoint", ""),
				),
			},
		},
	})
}

func TestAccAWSS3Bucket_WebsiteRoutingRules(t *testing.T) {
	rInt := acctest.RandInt()
	region := testAccGetRegion()
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketWebsiteConfigWithRoutingRules(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					testAccCheckAWSS3BucketWebsite(
						resourceName, "index.html", "error.html", "", ""),
					testAccCheckAWSS3BucketWebsiteRoutingRules(
						resourceName,
						[]*s3.RoutingRule{
							{
								Condition: &s3.Condition{
									KeyPrefixEquals: aws.String("docs/"),
								},
								Redirect: &s3.Redirect{
									ReplaceKeyPrefixWith: aws.String("documents/"),
								},
							},
						},
					),
					testAccCheckS3BucketWebsiteEndpoint(
						resourceName, "website_endpoint", testAccBucketName(rInt), region),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy", "acl", "grant"},
			},
			{
				Config: testAccAWSS3BucketConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					testAccCheckAWSS3BucketWebsite(
						resourceName, "", "", "", ""),
					testAccCheckAWSS3BucketWebsiteRoutingRules(resourceName, nil),
					resource.TestCheckResourceAttr(
						resourceName, "website_endpoint", ""),
				),
			},
		},
	})
}

func TestAccAWSS3Bucket_enableDefaultEncryption_whenTypical(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "aws_s3_bucket.arbitrary"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketEnableDefaultEncryption(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.0.rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.0.rule.0.apply_server_side_encryption_by_default.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.0.rule.0.apply_server_side_encryption_by_default.0.sse_algorithm", "aws:kms"),
					resource.TestMatchResourceAttr(resourceName, "server_side_encryption_configuration.0.rule.0.apply_server_side_encryption_by_default.0.kms_master_key_id", regexp.MustCompile("^arn")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy", "acl"},
			},
		},
	})
}

func TestAccAWSS3Bucket_enableDefaultEncryption_whenAES256IsUsed(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "aws_s3_bucket.arbitrary"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketEnableDefaultEncryptionWithAES256(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.0.rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.0.rule.0.apply_server_side_encryption_by_default.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.0.rule.0.apply_server_side_encryption_by_default.0.sse_algorithm", "AES256"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.0.rule.0.apply_server_side_encryption_by_default.0.kms_master_key_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy", "acl"},
			},
		},
	})
}

func TestAccAWSS3Bucket_disableDefaultEncryption_whenDefaultEncryptionIsEnabled(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "aws_s3_bucket.arbitrary"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketEnableDefaultEncryptionWithDefaultKey(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy", "acl"},
			},
			{
				Config: testAccAWSS3BucketDisableDefaultEncryption(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.#", "0"),
				),
			},
		},
	})
}

// Test TestAccAWSS3Bucket_shouldFailNotFound is designed to fail with a "plan
// not empty" error in Terraform, to check against regresssions.
// See https://github.com/hashicorp/terraform/pull/2925
func TestAccAWSS3Bucket_shouldFailNotFound(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketDestroyedConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					testAccCheckAWSS3DestroyBucket(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSS3Bucket_Versioning(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					testAccCheckAWSS3BucketVersioning(
						resourceName, ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy", "acl"},
			},
			{
				Config: testAccAWSS3BucketConfigWithVersioning(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					testAccCheckAWSS3BucketVersioning(
						resourceName, s3.BucketVersioningStatusEnabled),
				),
			},
			{
				Config: testAccAWSS3BucketConfigWithDisableVersioning(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					testAccCheckAWSS3BucketVersioning(
						resourceName, s3.BucketVersioningStatusSuspended),
				),
			},
		},
	})
}

func TestAccAWSS3Bucket_Cors_Update(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "aws_s3_bucket.bucket"

	updateBucketCors := func(n string) resource.TestCheckFunc {
		return func(s *terraform.State) error {
			rs, ok := s.RootModule().Resources[n]
			if !ok {
				return fmt.Errorf("Not found: %s", n)
			}

			conn := testAccProvider.Meta().(*AWSClient).s3conn
			_, err := conn.PutBucketCors(&s3.PutBucketCorsInput{
				Bucket: aws.String(rs.Primary.ID),
				CORSConfiguration: &s3.CORSConfiguration{
					CORSRules: []*s3.CORSRule{
						{
							AllowedHeaders: []*string{aws.String("*")},
							AllowedMethods: []*string{aws.String("GET")},
							AllowedOrigins: []*string{aws.String("https://www.example.com")},
						},
					},
				},
			})
			if err != nil && !isAWSErr(err, "NoSuchCORSConfiguration", "") {
				return err
			}
			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigWithCORS(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					testAccCheckAWSS3BucketCors(
						resourceName,
						[]*s3.CORSRule{
							{
								AllowedHeaders: []*string{aws.String("*")},
								AllowedMethods: []*string{aws.String("PUT"), aws.String("POST")},
								AllowedOrigins: []*string{aws.String("https://www.example.com")},
								ExposeHeaders:  []*string{aws.String("x-amz-server-side-encryption"), aws.String("ETag")},
								MaxAgeSeconds:  aws.Int64(3000),
							},
						},
					),
					updateBucketCors(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy", "acl"},
			},
			{
				Config: testAccAWSS3BucketConfigWithCORS(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					testAccCheckAWSS3BucketCors(
						resourceName,
						[]*s3.CORSRule{
							{
								AllowedHeaders: []*string{aws.String("*")},
								AllowedMethods: []*string{aws.String("PUT"), aws.String("POST")},
								AllowedOrigins: []*string{aws.String("https://www.example.com")},
								ExposeHeaders:  []*string{aws.String("x-amz-server-side-encryption"), aws.String("ETag")},
								MaxAgeSeconds:  aws.Int64(3000),
							},
						},
					),
				),
			},
		},
	})
}

func TestAccAWSS3Bucket_Cors_Delete(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "aws_s3_bucket.bucket"

	deleteBucketCors := func(n string) resource.TestCheckFunc {
		return func(s *terraform.State) error {
			rs, ok := s.RootModule().Resources[n]
			if !ok {
				return fmt.Errorf("Not found: %s", n)
			}

			conn := testAccProvider.Meta().(*AWSClient).s3conn
			_, err := conn.DeleteBucketCors(&s3.DeleteBucketCorsInput{
				Bucket: aws.String(rs.Primary.ID),
			})
			if err != nil && !isAWSErr(err, "NoSuchCORSConfiguration", "") {
				return err
			}
			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigWithCORS(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					deleteBucketCors(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSS3Bucket_Cors_EmptyOrigin(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigWithCORSEmptyOrigin(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					testAccCheckAWSS3BucketCors(
						resourceName,
						[]*s3.CORSRule{
							{
								AllowedHeaders: []*string{aws.String("*")},
								AllowedMethods: []*string{aws.String("PUT"), aws.String("POST")},
								AllowedOrigins: []*string{aws.String("")},
								ExposeHeaders:  []*string{aws.String("x-amz-server-side-encryption"), aws.String("ETag")},
								MaxAgeSeconds:  aws.Int64(3000),
							},
						},
					),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy", "acl"},
			},
		},
	})
}

func TestAccAWSS3Bucket_Logging(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigWithLogging(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					testAccCheckAWSS3BucketLogging(
						resourceName, "aws_s3_bucket.log_bucket", "log/"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy", "acl"},
			},
		},
	})
}

func TestAccAWSS3Bucket_LifecycleBasic(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigWithLifecycle(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.0.id", "id1"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.0.prefix", "path1/"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.0.expiration.2613713285.days", "365"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.0.expiration.2613713285.date", ""),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.0.expiration.2613713285.expired_object_delete_marker", "false"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.0.transition.2000431762.date", ""),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.0.transition.2000431762.days", "30"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.0.transition.2000431762.storage_class", "STANDARD_IA"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.0.transition.3601168188.date", ""),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.0.transition.3601168188.days", "60"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.0.transition.3601168188.storage_class", "INTELLIGENT_TIERING"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.0.transition.3854926587.date", ""),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.0.transition.3854926587.days", "90"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.0.transition.3854926587.storage_class", "ONEZONE_IA"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.0.transition.962205413.date", ""),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.0.transition.962205413.days", "120"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.0.transition.962205413.storage_class", "GLACIER"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.0.transition.1571523406.date", ""),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.0.transition.1571523406.days", "210"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.0.transition.1571523406.storage_class", "DEEP_ARCHIVE"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.1.id", "id2"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.1.prefix", "path2/"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.1.expiration.2855832418.date", "2016-01-12"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.1.expiration.2855832418.days", "0"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.1.expiration.2855832418.expired_object_delete_marker", "false"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.2.id", "id3"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.2.prefix", "path3/"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.2.transition.460947558.days", "0"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.3.id", "id4"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.3.prefix", "path4/"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.3.tags.tagKey", "tagValue"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.3.tags.terraform", "hashicorp"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.4.id", "id5"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.4.tags.tagKey", "tagValue"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.4.tags.terraform", "hashicorp"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.4.transition.460947558.days", "0"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.4.transition.460947558.storage_class", "GLACIER"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.5.id", "id6"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.5.tags.tagKey", "tagValue"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.5.transition.460947558.days", "0"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.5.transition.460947558.storage_class", "GLACIER"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy", "acl"},
			},
			{
				Config: testAccAWSS3BucketConfigWithVersioningLifecycle(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.0.id", "id1"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.0.prefix", "path1/"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.0.enabled", "true"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.0.noncurrent_version_expiration.80908210.days", "365"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.0.noncurrent_version_transition.1377917700.days", "30"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.0.noncurrent_version_transition.1377917700.storage_class", "STANDARD_IA"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.0.noncurrent_version_transition.2528035817.days", "60"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.0.noncurrent_version_transition.2528035817.storage_class", "GLACIER"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.1.id", "id2"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.1.prefix", "path2/"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.1.enabled", "false"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.1.noncurrent_version_expiration.80908210.days", "365"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.2.id", "id3"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.2.prefix", "path3/"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.2.noncurrent_version_transition.3732708140.days", "0"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.2.noncurrent_version_transition.3732708140.storage_class", "GLACIER"),
				),
			},
			{
				Config: testAccAWSS3BucketConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
				),
			},
		},
	})
}

func TestAccAWSS3Bucket_LifecycleExpireMarkerOnly(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigWithLifecycleExpireMarker(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.0.id", "id1"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.0.prefix", "path1/"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.0.expiration.3591068768.days", "0"),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.0.expiration.3591068768.date", ""),
					resource.TestCheckResourceAttr(
						resourceName, "lifecycle_rule.0.expiration.3591068768.expired_object_delete_marker", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy", "acl"},
			},
			{
				Config: testAccAWSS3BucketConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
				),
			},
		},
	})
}

func TestAccAWSS3Bucket_Replication(t *testing.T) {
	rInt := acctest.RandInt()
	alternateRegion := testAccGetAlternateRegion()
	region := testAccGetRegion()
	partition := testAccGetPartition()
	iamRoleResourceName := "aws_iam_role.role"
	resourceName := "aws_s3_bucket.bucket"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccMultipleRegionsPreCheck(t)
			testAccAlternateRegionPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckWithProviders(testAccCheckAWSS3BucketDestroyWithProvider, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigReplication(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExistsWithProvider(resourceName, testAccAwsRegionProviderFunc(region, &providers)),
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.destination", testAccAwsRegionProviderFunc(alternateRegion, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "0"),
				),
			},
			{
				Config:            testAccAWSS3BucketConfigReplication(rInt),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy", "acl"},
			},
			{
				Config: testAccAWSS3BucketConfigReplicationWithConfiguration(rInt, "STANDARD"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExistsWithProvider(resourceName, testAccAwsRegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", "1"),
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.destination", testAccAwsRegionProviderFunc(alternateRegion, &providers)),
					testAccCheckAWSS3BucketReplicationRules(
						resourceName,
						testAccAwsRegionProviderFunc(region, &providers),
						[]*s3.ReplicationRule{
							{
								ID: aws.String("foobar"),
								Destination: &s3.Destination{
									Bucket:       aws.String(fmt.Sprintf("arn:%s:s3:::tf-test-bucket-destination-%d", partition, rInt)),
									StorageClass: aws.String(s3.StorageClassStandard),
								},
								Prefix: aws.String("foo"),
								Status: aws.String(s3.ReplicationRuleStatusEnabled),
							},
						},
					),
				),
			},
			{
				Config: testAccAWSS3BucketConfigReplicationWithConfiguration(rInt, "GLACIER"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExistsWithProvider(resourceName, testAccAwsRegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", "1"),
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.destination", testAccAwsRegionProviderFunc(alternateRegion, &providers)),
					testAccCheckAWSS3BucketReplicationRules(
						resourceName,
						testAccAwsRegionProviderFunc(region, &providers),
						[]*s3.ReplicationRule{
							{
								ID: aws.String("foobar"),
								Destination: &s3.Destination{
									Bucket:       aws.String(fmt.Sprintf("arn:%s:s3:::tf-test-bucket-destination-%d", partition, rInt)),
									StorageClass: aws.String(s3.StorageClassGlacier),
								},
								Prefix: aws.String("foo"),
								Status: aws.String(s3.ReplicationRuleStatusEnabled),
							},
						},
					),
				),
			},
			{
				Config: testAccAWSS3BucketConfigReplicationWithSseKmsEncryptedObjects(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExistsWithProvider(resourceName, testAccAwsRegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", "1"),
					testAccCheckAWSS3BucketReplicationRules(
						resourceName,
						testAccAwsRegionProviderFunc(region, &providers),
						[]*s3.ReplicationRule{
							{
								ID: aws.String("foobar"),
								Destination: &s3.Destination{
									Bucket:       aws.String(fmt.Sprintf("arn:%s:s3:::tf-test-bucket-destination-%d", partition, rInt)),
									StorageClass: aws.String(s3.ObjectStorageClassStandard),
									EncryptionConfiguration: &s3.EncryptionConfiguration{
										ReplicaKmsKeyID: aws.String("${aws_kms_key.replica.arn}"),
									},
								},
								Prefix: aws.String("foo"),
								Status: aws.String(s3.ReplicationRuleStatusEnabled),
								SourceSelectionCriteria: &s3.SourceSelectionCriteria{
									SseKmsEncryptedObjects: &s3.SseKmsEncryptedObjects{
										Status: aws.String(s3.SseKmsEncryptedObjectsStatusEnabled),
									},
								},
							},
						},
					),
				),
			},
		},
	})
}

func TestAccAWSS3Bucket_ReplicationConfiguration_Rule_Destination_AccessControlTranslation(t *testing.T) {
	rInt := acctest.RandInt()
	region := testAccGetRegion()
	partition := testAccGetPartition()
	iamRoleResourceName := "aws_iam_role.role"
	resourceName := "aws_s3_bucket.bucket"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccMultipleRegionsPreCheck(t)
			testAccAlternateRegionPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckWithProviders(testAccCheckAWSS3BucketDestroyWithProvider, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigReplicationWithAccessControlTranslation(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExistsWithProvider(resourceName, testAccAwsRegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", "1"),
					testAccCheckAWSS3BucketReplicationRules(
						resourceName,
						testAccAwsRegionProviderFunc(region, &providers),
						[]*s3.ReplicationRule{
							{
								ID: aws.String("foobar"),
								Destination: &s3.Destination{
									Account:      aws.String("${data.aws_caller_identity.current.account_id}"),
									Bucket:       aws.String(fmt.Sprintf("arn:%s:s3:::tf-test-bucket-destination-%d", partition, rInt)),
									StorageClass: aws.String(s3.ObjectStorageClassStandard),
									AccessControlTranslation: &s3.AccessControlTranslation{
										Owner: aws.String("Destination"),
									},
								},
								Prefix: aws.String("foo"),
								Status: aws.String(s3.ReplicationRuleStatusEnabled),
							},
						},
					),
				),
			},
			{
				Config:            testAccAWSS3BucketConfigReplicationWithAccessControlTranslation(rInt),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy", "acl", "versioning"},
			},
			{
				Config: testAccAWSS3BucketConfigReplicationWithSseKmsEncryptedObjectsAndAccessControlTranslation(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExistsWithProvider(resourceName, testAccAwsRegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", "1"),
					testAccCheckAWSS3BucketReplicationRules(
						resourceName,
						testAccAwsRegionProviderFunc(region, &providers),
						[]*s3.ReplicationRule{
							{
								ID: aws.String("foobar"),
								Destination: &s3.Destination{
									Account:      aws.String("${data.aws_caller_identity.current.account_id}"),
									Bucket:       aws.String(fmt.Sprintf("arn:%s:s3:::tf-test-bucket-destination-%d", partition, rInt)),
									StorageClass: aws.String(s3.ObjectStorageClassStandard),
									EncryptionConfiguration: &s3.EncryptionConfiguration{
										ReplicaKmsKeyID: aws.String("${aws_kms_key.replica.arn}"),
									},
									AccessControlTranslation: &s3.AccessControlTranslation{
										Owner: aws.String("Destination"),
									},
								},
								Prefix: aws.String("foo"),
								Status: aws.String(s3.ReplicationRuleStatusEnabled),
								SourceSelectionCriteria: &s3.SourceSelectionCriteria{
									SseKmsEncryptedObjects: &s3.SseKmsEncryptedObjects{
										Status: aws.String(s3.SseKmsEncryptedObjectsStatusEnabled),
									},
								},
							},
						},
					),
				),
			},
		},
	})
}

// StorageClass issue: https://github.com/hashicorp/terraform/issues/10909
func TestAccAWSS3Bucket_ReplicationWithoutStorageClass(t *testing.T) {
	rInt := acctest.RandInt()
	alternateRegion := testAccGetAlternateRegion()
	region := testAccGetRegion()
	resourceName := "aws_s3_bucket.bucket"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccMultipleRegionsPreCheck(t)
			testAccAlternateRegionPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckWithProviders(testAccCheckAWSS3BucketDestroyWithProvider, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigReplicationWithoutStorageClass(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExistsWithProvider(resourceName, testAccAwsRegionProviderFunc(region, &providers)),
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.destination", testAccAwsRegionProviderFunc(alternateRegion, &providers)),
				),
			},
			{
				Config:            testAccAWSS3BucketConfigReplicationWithoutStorageClass(rInt),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy", "acl"},
			},
		},
	})
}

func TestAccAWSS3Bucket_ReplicationExpectVersioningValidationError(t *testing.T) {
	rInt := acctest.RandInt()

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccMultipleRegionsPreCheck(t)
			testAccAlternateRegionPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckWithProviders(testAccCheckAWSS3BucketDestroyWithProvider, &providers),
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSS3BucketConfigReplicationNoVersioning(rInt),
				ExpectError: regexp.MustCompile(`versioning must be enabled to allow S3 bucket replication`),
			},
		},
	})
}

// Prefix issue: https://github.com/terraform-providers/terraform-provider-aws/issues/6340
func TestAccAWSS3Bucket_ReplicationWithoutPrefix(t *testing.T) {
	rInt := acctest.RandInt()
	alternateRegion := testAccGetAlternateRegion()
	region := testAccGetRegion()
	resourceName := "aws_s3_bucket.bucket"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccMultipleRegionsPreCheck(t)
			testAccAlternateRegionPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckWithProviders(testAccCheckAWSS3BucketDestroyWithProvider, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigReplicationWithoutPrefix(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExistsWithProvider(resourceName, testAccAwsRegionProviderFunc(region, &providers)),
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.destination", testAccAwsRegionProviderFunc(alternateRegion, &providers)),
				),
			},
			{
				Config:            testAccAWSS3BucketConfigReplicationWithoutPrefix(rInt),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy", "acl"},
			},
		},
	})
}

func TestAccAWSS3Bucket_ReplicationSchemaV2(t *testing.T) {
	rInt := acctest.RandInt()
	alternateRegion := testAccGetAlternateRegion()
	region := testAccGetRegion()
	partition := testAccGetPartition()
	iamRoleResourceName := "aws_iam_role.role"
	resourceName := "aws_s3_bucket.bucket"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccMultipleRegionsPreCheck(t)
			testAccAlternateRegionPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckWithProviders(testAccCheckAWSS3BucketDestroyWithProvider, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigReplicationWithV2ConfigurationNoTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExistsWithProvider(resourceName, testAccAwsRegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", "1"),
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.destination", testAccAwsRegionProviderFunc(alternateRegion, &providers)),
					testAccCheckAWSS3BucketReplicationRules(
						resourceName,
						testAccAwsRegionProviderFunc(region, &providers),
						[]*s3.ReplicationRule{
							{
								ID: aws.String("foobar"),
								Destination: &s3.Destination{
									Bucket:       aws.String(fmt.Sprintf("arn:%s:s3:::tf-test-bucket-destination-%d", partition, rInt)),
									StorageClass: aws.String(s3.ObjectStorageClassStandard),
								},
								Status: aws.String(s3.ReplicationRuleStatusEnabled),
								Filter: &s3.ReplicationRuleFilter{
									Prefix: aws.String("foo"),
								},
								Priority: aws.Int64(0),
								DeleteMarkerReplication: &s3.DeleteMarkerReplication{
									Status: aws.String(s3.DeleteMarkerReplicationStatusDisabled),
								},
							},
						},
					),
				),
			},
			{
				Config:            testAccAWSS3BucketConfigReplicationWithV2ConfigurationNoTags(rInt),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy", "acl"},
			},
			{
				Config: testAccAWSS3BucketConfigReplicationWithV2ConfigurationOnlyOneTag(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExistsWithProvider(resourceName, testAccAwsRegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", "1"),
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.destination", testAccAwsRegionProviderFunc(alternateRegion, &providers)),
					testAccCheckAWSS3BucketReplicationRules(
						resourceName,
						testAccAwsRegionProviderFunc(region, &providers),
						[]*s3.ReplicationRule{
							{
								ID: aws.String("foobar"),
								Destination: &s3.Destination{
									Bucket:       aws.String(fmt.Sprintf("arn:%s:s3:::tf-test-bucket-destination-%d", partition, rInt)),
									StorageClass: aws.String(s3.ObjectStorageClassStandard),
								},
								Status: aws.String(s3.ReplicationRuleStatusEnabled),
								Filter: &s3.ReplicationRuleFilter{
									And: &s3.ReplicationRuleAndOperator{
										Prefix: aws.String(""),
										Tags: []*s3.Tag{
											{
												Key:   aws.String("ReplicateMe"),
												Value: aws.String("Yes"),
											},
										},
									},
								},
								Priority: aws.Int64(42),
								DeleteMarkerReplication: &s3.DeleteMarkerReplication{
									Status: aws.String(s3.DeleteMarkerReplicationStatusDisabled),
								},
							},
						},
					),
				),
			},
			{
				Config: testAccAWSS3BucketConfigReplicationWithV2ConfigurationPrefixAndTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExistsWithProvider(resourceName, testAccAwsRegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", "1"),
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.destination", testAccAwsRegionProviderFunc(alternateRegion, &providers)),
					testAccCheckAWSS3BucketReplicationRules(
						resourceName,
						testAccAwsRegionProviderFunc(region, &providers),
						[]*s3.ReplicationRule{
							{
								ID: aws.String("foobar"),
								Destination: &s3.Destination{
									Bucket:       aws.String(fmt.Sprintf("arn:%s:s3:::tf-test-bucket-destination-%d", partition, rInt)),
									StorageClass: aws.String(s3.ObjectStorageClassStandard),
								},
								Status: aws.String(s3.ReplicationRuleStatusEnabled),
								Filter: &s3.ReplicationRuleFilter{
									And: &s3.ReplicationRuleAndOperator{
										Prefix: aws.String("foo"),
										Tags: []*s3.Tag{
											{
												Key:   aws.String("ReplicateMe"),
												Value: aws.String("Yes"),
											},
											{
												Key:   aws.String("AnotherTag"),
												Value: aws.String("OK"),
											},
										},
									},
								},
								Priority: aws.Int64(41),
								DeleteMarkerReplication: &s3.DeleteMarkerReplication{
									Status: aws.String(s3.DeleteMarkerReplicationStatusDisabled),
								},
							},
						},
					),
				),
			},
			{
				Config: testAccAWSS3BucketConfigReplicationWithV2ConfigurationMultipleTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExistsWithProvider(resourceName, testAccAwsRegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", "1"),
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.destination", testAccAwsRegionProviderFunc(alternateRegion, &providers)),
					testAccCheckAWSS3BucketReplicationRules(
						resourceName,
						testAccAwsRegionProviderFunc(region, &providers),
						[]*s3.ReplicationRule{
							{
								ID: aws.String("foobar"),
								Destination: &s3.Destination{
									Bucket:       aws.String(fmt.Sprintf("arn:%s:s3:::tf-test-bucket-destination-%d", partition, rInt)),
									StorageClass: aws.String(s3.ObjectStorageClassStandard),
								},
								Status: aws.String(s3.ReplicationRuleStatusEnabled),
								Filter: &s3.ReplicationRuleFilter{
									And: &s3.ReplicationRuleAndOperator{
										Prefix: aws.String(""),
										Tags: []*s3.Tag{
											{
												Key:   aws.String("ReplicateMe"),
												Value: aws.String("Yes"),
											},
											{
												Key:   aws.String("AnotherTag"),
												Value: aws.String("OK"),
											},
											{
												Key:   aws.String("Foo"),
												Value: aws.String("Bar"),
											},
										},
									},
								},
								Priority: aws.Int64(0),
								DeleteMarkerReplication: &s3.DeleteMarkerReplication{
									Status: aws.String(s3.DeleteMarkerReplicationStatusDisabled),
								},
							},
						},
					),
				),
			},
		},
	})
}

func TestAccAWSS3Bucket_objectLock(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "aws_s3_bucket.arbitrary"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketObjectLockEnabledNoDefaultRetention(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "object_lock_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_configuration.0.object_lock_enabled", "Enabled"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_configuration.0.rule.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy", "acl"},
			},
			{
				Config: testAccAWSS3BucketObjectLockEnabledWithDefaultRetention(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "object_lock_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_configuration.0.object_lock_enabled", "Enabled"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_configuration.0.rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_configuration.0.rule.0.default_retention.0.mode", "COMPLIANCE"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_configuration.0.rule.0.default_retention.0.days", "3"),
				),
			},
		},
	})
}

func TestAccAWSS3Bucket_forceDestroy(t *testing.T) {
	resourceName := "aws_s3_bucket.bucket"
	rInt := acctest.RandInt()
	bucketName := fmt.Sprintf("tf-test-bucket-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfig_forceDestroy(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					testAccCheckAWSS3BucketAddObjects(resourceName, "data.txt", "prefix/more_data.txt"),
				),
			},
		},
	})
}

// By default, the AWS Go SDK cleans up URIs by removing extra slashes
// when the service API requests use the URI as part of making a request.
// While the aws_s3_bucket_object resource automatically cleans the key
// to not contain these extra slashes, out-of-band handling and other AWS
// services may create keys with extra slashes (empty "directory" prefixes).
func TestAccAWSS3Bucket_forceDestroyWithEmptyPrefixes(t *testing.T) {
	resourceName := "aws_s3_bucket.bucket"
	rInt := acctest.RandInt()
	bucketName := fmt.Sprintf("tf-test-bucket-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfig_forceDestroy(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					testAccCheckAWSS3BucketAddObjects(resourceName, "data.txt", "/extraleadingslash.txt"),
				),
			},
		},
	})
}

func TestAccAWSS3Bucket_forceDestroyWithObjectLockEnabled(t *testing.T) {
	resourceName := "aws_s3_bucket.bucket"
	rInt := acctest.RandInt()
	bucketName := fmt.Sprintf("tf-test-bucket-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfig_forceDestroyWithObjectLockEnabled(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					testAccCheckAWSS3BucketAddObjectsWithLegalHold(resourceName, "data.txt", "prefix/more_data.txt"),
				),
			},
		},
	})
}

func TestAWSS3BucketName(t *testing.T) {
	validDnsNames := []string{
		"foobar",
		"foo.bar",
		"foo.bar.baz",
		"1234",
		"foo-bar",
		strings.Repeat("x", 63),
	}

	for _, v := range validDnsNames {
		if err := validateS3BucketName(v, "us-west-2"); err != nil {
			t.Fatalf("%q should be a valid S3 bucket name", v)
		}
	}

	invalidDnsNames := []string{
		"foo..bar",
		"Foo.Bar",
		"192.168.0.1",
		"127.0.0.1",
		".foo",
		"bar.",
		"foo_bar",
		strings.Repeat("x", 64),
	}

	for _, v := range invalidDnsNames {
		if err := validateS3BucketName(v, "us-west-2"); err == nil {
			t.Fatalf("%q should not be a valid S3 bucket name", v)
		}
	}

	validEastNames := []string{
		"foobar",
		"foo_bar",
		"127.0.0.1",
		"foo..bar",
		"foo_bar_baz",
		"foo.bar.baz",
		"Foo.Bar",
		strings.Repeat("x", 255),
	}

	for _, v := range validEastNames {
		if err := validateS3BucketName(v, "us-east-1"); err != nil {
			t.Fatalf("%q should be a valid S3 bucket name", v)
		}
	}

	invalidEastNames := []string{
		"foo;bar",
		strings.Repeat("x", 256),
	}

	for _, v := range invalidEastNames {
		if err := validateS3BucketName(v, "us-east-1"); err == nil {
			t.Fatalf("%q should not be a valid S3 bucket name", v)
		}
	}
}

func TestBucketRegionalDomainName(t *testing.T) {
	const bucket = "bucket-name"

	var testCases = []struct {
		ExpectedErrCount int
		ExpectedOutput   string
		Region           string
	}{
		{
			Region:           "",
			ExpectedErrCount: 0,
			ExpectedOutput:   bucket + ".s3.amazonaws.com",
		},
		{
			Region:           "custom",
			ExpectedErrCount: 0,
			ExpectedOutput:   bucket + ".s3.custom.amazonaws.com",
		},
		{
			Region:           "us-east-1",
			ExpectedErrCount: 0,
			ExpectedOutput:   bucket + ".s3.amazonaws.com",
		},
		{
			Region:           "us-west-2",
			ExpectedErrCount: 0,
			ExpectedOutput:   bucket + ".s3.us-west-2.amazonaws.com",
		},
		{
			Region:           "us-gov-west-1",
			ExpectedErrCount: 0,
			ExpectedOutput:   bucket + ".s3.us-gov-west-1.amazonaws.com",
		},
		{
			Region:           "cn-north-1",
			ExpectedErrCount: 0,
			ExpectedOutput:   bucket + ".s3.cn-north-1.amazonaws.com.cn",
		},
	}

	for _, tc := range testCases {
		output, err := BucketRegionalDomainName(bucket, tc.Region)
		if tc.ExpectedErrCount == 0 && err != nil {
			t.Fatalf("expected %q not to trigger an error, received: %s", tc.Region, err)
		}
		if tc.ExpectedErrCount > 0 && err == nil {
			t.Fatalf("expected %q to trigger an error", tc.Region)
		}
		if output != tc.ExpectedOutput {
			t.Fatalf("expected %q, received %q", tc.ExpectedOutput, output)
		}
	}
}

func TestWebsiteEndpoint(t *testing.T) {
	// https://docs.aws.amazon.com/AmazonS3/latest/dev/WebsiteEndpoints.html
	testCases := []struct {
		AWSClient          *AWSClient
		LocationConstraint string
		Expected           string
	}{
		{
			AWSClient: &AWSClient{
				dnsSuffix: "amazonaws.com",
				region:    "us-east-1",
			},
			LocationConstraint: "",
			Expected:           "bucket-name.s3-website-us-east-1.amazonaws.com",
		},
		{
			AWSClient: &AWSClient{
				dnsSuffix: "amazonaws.com",
				region:    "us-west-2",
			},
			LocationConstraint: "us-west-2",
			Expected:           "bucket-name.s3-website-us-west-2.amazonaws.com",
		},
		{
			AWSClient: &AWSClient{
				dnsSuffix: "amazonaws.com",
				region:    "us-west-1",
			},
			LocationConstraint: "us-west-1",
			Expected:           "bucket-name.s3-website-us-west-1.amazonaws.com",
		},
		{
			AWSClient: &AWSClient{
				dnsSuffix: "amazonaws.com",
				region:    "eu-west-1",
			},
			LocationConstraint: "eu-west-1",
			Expected:           "bucket-name.s3-website-eu-west-1.amazonaws.com",
		},
		{
			AWSClient: &AWSClient{
				dnsSuffix: "amazonaws.com",
				region:    "eu-west-3",
			},
			LocationConstraint: "eu-west-3",
			Expected:           "bucket-name.s3-website.eu-west-3.amazonaws.com"},
		{
			AWSClient: &AWSClient{
				dnsSuffix: "amazonaws.com",
				region:    "eu-central-1",
			},
			LocationConstraint: "eu-central-1",
			Expected:           "bucket-name.s3-website.eu-central-1.amazonaws.com",
		},
		{
			AWSClient: &AWSClient{
				dnsSuffix: "amazonaws.com",
				region:    "ap-south-1",
			},
			LocationConstraint: "ap-south-1",
			Expected:           "bucket-name.s3-website.ap-south-1.amazonaws.com",
		},
		{
			AWSClient: &AWSClient{
				dnsSuffix: "amazonaws.com",
				region:    "ap-southeast-1",
			},
			LocationConstraint: "ap-southeast-1",
			Expected:           "bucket-name.s3-website-ap-southeast-1.amazonaws.com",
		},
		{
			AWSClient: &AWSClient{
				dnsSuffix: "amazonaws.com",
				region:    "ap-northeast-1",
			},
			LocationConstraint: "ap-northeast-1",
			Expected:           "bucket-name.s3-website-ap-northeast-1.amazonaws.com",
		},
		{
			AWSClient: &AWSClient{
				dnsSuffix: "amazonaws.com",
				region:    "ap-southeast-2",
			},
			LocationConstraint: "ap-southeast-2",
			Expected:           "bucket-name.s3-website-ap-southeast-2.amazonaws.com",
		},
		{
			AWSClient: &AWSClient{
				dnsSuffix: "amazonaws.com",
				region:    "ap-northeast-2",
			},
			LocationConstraint: "ap-northeast-2",
			Expected:           "bucket-name.s3-website.ap-northeast-2.amazonaws.com",
		},
		{
			AWSClient: &AWSClient{
				dnsSuffix: "amazonaws.com",
				region:    "sa-east-1",
			},
			LocationConstraint: "sa-east-1",
			Expected:           "bucket-name.s3-website-sa-east-1.amazonaws.com",
		},
		{
			AWSClient: &AWSClient{
				dnsSuffix: "amazonaws.com",
				region:    "us-gov-east-1",
			},
			LocationConstraint: "us-gov-east-1",
			Expected:           "bucket-name.s3-website.us-gov-east-1.amazonaws.com",
		},
		{
			AWSClient: &AWSClient{
				dnsSuffix: "amazonaws.com",
				region:    "us-gov-west-1",
			},
			LocationConstraint: "us-gov-west-1",
			Expected:           "bucket-name.s3-website-us-gov-west-1.amazonaws.com",
		},
		{
			AWSClient: &AWSClient{
				dnsSuffix: "c2s.ic.gov",
				region:    "us-iso-east-1",
			},
			LocationConstraint: "us-iso-east-1",
			Expected:           "bucket-name.s3-website.us-iso-east-1.c2s.ic.gov",
		},
		{
			AWSClient: &AWSClient{
				dnsSuffix: "sc2s.sgov.gov",
				region:    "us-isob-east-1",
			},
			LocationConstraint: "us-isob-east-1",
			Expected:           "bucket-name.s3-website.us-isob-east-1.sc2s.sgov.gov",
		},
		{
			AWSClient: &AWSClient{
				dnsSuffix: "amazonaws.com.cn",
				region:    "cn-northwest-1",
			},
			LocationConstraint: "cn-northwest-1",
			Expected:           "bucket-name.s3-website.cn-northwest-1.amazonaws.com.cn",
		},
		{
			AWSClient: &AWSClient{
				dnsSuffix: "amazonaws.com.cn",
				region:    "cn-north-1",
			},
			LocationConstraint: "cn-north-1",
			Expected:           "bucket-name.s3-website.cn-north-1.amazonaws.com.cn",
		},
	}

	for _, testCase := range testCases {
		got := WebsiteEndpoint(testCase.AWSClient, "bucket-name", testCase.LocationConstraint)
		if got.Endpoint != testCase.Expected {
			t.Errorf("WebsiteEndpointUrl(\"bucket-name\", %q) => %q, want %q", testCase.LocationConstraint, got.Endpoint, testCase.Expected)
		}
	}
}

func testAccCheckAWSS3BucketDestroy(s *terraform.State) error {
	return testAccCheckAWSS3BucketDestroyWithProvider(s, testAccProvider)
}

func testAccCheckAWSS3BucketDestroyWithProvider(s *terraform.State, provider *schema.Provider) error {
	conn := provider.Meta().(*AWSClient).s3conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_s3_bucket" {
			continue
		}

		input := &s3.HeadBucketInput{
			Bucket: aws.String(rs.Primary.ID),
		}

		// Retry for S3 eventual consistency
		err := resource.Retry(1*time.Minute, func() *resource.RetryError {
			_, err := conn.HeadBucket(input)

			if isAWSErr(err, s3.ErrCodeNoSuchBucket, "") || isAWSErr(err, "NotFound", "") {
				return nil
			}

			if err != nil {
				return resource.NonRetryableError(err)
			}

			return resource.RetryableError(fmt.Errorf("AWS S3 Bucket still exists: %s", rs.Primary.ID))
		})

		if isResourceTimeoutError(err) {
			_, err = conn.HeadBucket(input)
		}

		if err != nil {
			return err
		}
	}
	return nil
}

func testAccCheckAWSS3BucketExists(n string) resource.TestCheckFunc {
	return testAccCheckAWSS3BucketExistsWithProvider(n, func() *schema.Provider { return testAccProvider })
}

func testAccCheckAWSS3BucketExistsWithProvider(n string, providerF func() *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		provider := providerF()

		conn := provider.Meta().(*AWSClient).s3conn
		_, err := conn.HeadBucket(&s3.HeadBucketInput{
			Bucket: aws.String(rs.Primary.ID),
		})

		if err != nil {
			if isAWSErr(err, s3.ErrCodeNoSuchBucket, "") {
				return fmt.Errorf("S3 bucket not found")
			}
			return err
		}
		return nil

	}
}

func testAccCheckAWSS3DestroyBucket(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No S3 Bucket ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).s3conn
		_, err := conn.DeleteBucket(&s3.DeleteBucketInput{
			Bucket: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return fmt.Errorf("Error destroying Bucket (%s) in testAccCheckAWSS3DestroyBucket: %s", rs.Primary.ID, err)
		}
		return nil
	}
}

func testAccCheckAWSS3BucketAddObjects(n string, keys ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		conn := testAccProvider.Meta().(*AWSClient).s3connUriCleaningDisabled

		for _, key := range keys {
			_, err := conn.PutObject(&s3.PutObjectInput{
				Bucket: aws.String(rs.Primary.ID),
				Key:    aws.String(key),
			})

			if err != nil {
				return fmt.Errorf("PutObject error: %s", err)
			}
		}

		return nil
	}
}

func testAccCheckAWSS3BucketAddObjectsWithLegalHold(n string, keys ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		conn := testAccProvider.Meta().(*AWSClient).s3conn

		for _, key := range keys {
			_, err := conn.PutObject(&s3.PutObjectInput{
				Bucket:                    aws.String(rs.Primary.ID),
				Key:                       aws.String(key),
				ObjectLockLegalHoldStatus: aws.String(s3.ObjectLockLegalHoldStatusOn),
			})

			if err != nil {
				return fmt.Errorf("PutObject error: %s", err)
			}
		}

		return nil
	}
}

// Create an S3 bucket via a CF stack so that it has system tags.
func testAccCheckAWSS3BucketCreateViaCloudFormation(n string, stackId *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).cfconn
		stackName := fmt.Sprintf("tf-acc-test-s3tags-%d", acctest.RandInt())
		templateBody := fmt.Sprintf(`
        {
            "Resources": {
                "TfTestBucket": {
                    "Type": "AWS::S3::Bucket",
                    "Properties": {
                        "BucketName": "%s",
                    }
                }
            }
        }
        `, n)

		req := &cloudformation.CreateStackInput{
			StackName:    aws.String(stackName),
			TemplateBody: aws.String(templateBody),
		}

		log.Printf("[DEBUG] Creating CloudFormation stack: %#v", req)
		resp, err := conn.CreateStack(req)
		if err != nil {
			return fmt.Errorf("Error creating CloudFormation stack: %s", err)
		}

		status, err := waitForCloudFormationStackCreation(conn, aws.StringValue(resp.StackId), 10*time.Minute)
		if err != nil {
			return fmt.Errorf("Error waiting for CloudFormation stack creation: %s", err)
		}
		if status != cloudformation.StackStatusCreateComplete {
			return fmt.Errorf("Invalid CloudFormation stack creation status: %s", status)
		}

		*stackId = aws.StringValue(resp.StackId)
		return nil
	}
}

func testAccCheckAWSS3BucketTagKeys(n string, keys ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		conn := testAccProvider.Meta().(*AWSClient).s3conn

		resp, err := conn.GetBucketTagging(&s3.GetBucketTaggingInput{
			Bucket: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		for _, key := range keys {
			ok := false
			for _, tag := range resp.TagSet {
				if key == aws.StringValue(tag.Key) {
					ok = true
					break
				}
			}
			if !ok {
				return fmt.Errorf("Key %s not found in bucket's tag set", key)
			}
		}

		return nil
	}
}

func testAccCheckAWSS3BucketPolicy(n string, policy string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		conn := testAccProvider.Meta().(*AWSClient).s3conn

		out, err := conn.GetBucketPolicy(&s3.GetBucketPolicyInput{
			Bucket: aws.String(rs.Primary.ID),
		})

		if policy == "" {
			if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "NoSuchBucketPolicy" {
				// expected
				return nil
			}
			if err == nil {
				return fmt.Errorf("Expected no policy, got: %#v", *out.Policy)
			} else {
				return fmt.Errorf("GetBucketPolicy error: %v, expected %s", err, policy)
			}
		}
		if err != nil {
			return fmt.Errorf("GetBucketPolicy error: %v, expected %s", err, policy)
		}

		if v := out.Policy; v == nil {
			if policy != "" {
				return fmt.Errorf("bad policy, found nil, expected: %s", policy)
			}
		} else {
			expected := make(map[string]interface{})
			if err := json.Unmarshal([]byte(policy), &expected); err != nil {
				return err
			}
			actual := make(map[string]interface{})
			if err := json.Unmarshal([]byte(*v), &actual); err != nil {
				return err
			}

			if !reflect.DeepEqual(expected, actual) {
				return fmt.Errorf("bad policy, expected: %#v, got %#v", expected, actual)
			}
		}

		return nil
	}
}

func testAccCheckAWSS3BucketWebsite(n string, indexDoc string, errorDoc string, redirectProtocol string, redirectTo string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		conn := testAccProvider.Meta().(*AWSClient).s3conn

		out, err := conn.GetBucketWebsite(&s3.GetBucketWebsiteInput{
			Bucket: aws.String(rs.Primary.ID),
		})

		if err != nil {
			if indexDoc == "" {
				// If we want to assert that the website is not there, than
				// this error is expected
				return nil
			} else {
				return fmt.Errorf("S3BucketWebsite error: %v", err)
			}
		}

		if v := out.IndexDocument; v == nil {
			if indexDoc != "" {
				return fmt.Errorf("bad index doc, found nil, expected: %s", indexDoc)
			}
		} else {
			if *v.Suffix != indexDoc {
				return fmt.Errorf("bad index doc, expected: %s, got %#v", indexDoc, out.IndexDocument)
			}
		}

		if v := out.ErrorDocument; v == nil {
			if errorDoc != "" {
				return fmt.Errorf("bad error doc, found nil, expected: %s", errorDoc)
			}
		} else {
			if *v.Key != errorDoc {
				return fmt.Errorf("bad error doc, expected: %s, got %#v", errorDoc, out.ErrorDocument)
			}
		}

		if v := out.RedirectAllRequestsTo; v == nil {
			if redirectTo != "" {
				return fmt.Errorf("bad redirect to, found nil, expected: %s", redirectTo)
			}
		} else {
			if *v.HostName != redirectTo {
				return fmt.Errorf("bad redirect to, expected: %s, got %#v", redirectTo, out.RedirectAllRequestsTo)
			}
			if redirectProtocol != "" && v.Protocol != nil && *v.Protocol != redirectProtocol {
				return fmt.Errorf("bad redirect protocol to, expected: %s, got %#v", redirectProtocol, out.RedirectAllRequestsTo)
			}
		}

		return nil
	}
}

func testAccCheckAWSS3BucketWebsiteRoutingRules(n string, routingRules []*s3.RoutingRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		conn := testAccProvider.Meta().(*AWSClient).s3conn

		out, err := conn.GetBucketWebsite(&s3.GetBucketWebsiteInput{
			Bucket: aws.String(rs.Primary.ID),
		})

		if err != nil {
			if routingRules == nil {
				return nil
			}
			return fmt.Errorf("GetBucketWebsite error: %v", err)
		}

		if !reflect.DeepEqual(out.RoutingRules, routingRules) {
			return fmt.Errorf("bad routing rule, expected: %v, got %v", routingRules, out.RoutingRules)
		}

		return nil
	}
}

func testAccCheckAWSS3BucketVersioning(n string, versioningStatus string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		conn := testAccProvider.Meta().(*AWSClient).s3conn

		out, err := conn.GetBucketVersioning(&s3.GetBucketVersioningInput{
			Bucket: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return fmt.Errorf("GetBucketVersioning error: %v", err)
		}

		if v := out.Status; v == nil {
			if versioningStatus != "" {
				return fmt.Errorf("bad error versioning status, found nil, expected: %s", versioningStatus)
			}
		} else {
			if *v != versioningStatus {
				return fmt.Errorf("bad error versioning status, expected: %s, got %s", versioningStatus, *v)
			}
		}

		return nil
	}
}

func testAccCheckAWSS3BucketCors(n string, corsRules []*s3.CORSRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		conn := testAccProvider.Meta().(*AWSClient).s3conn

		out, err := conn.GetBucketCors(&s3.GetBucketCorsInput{
			Bucket: aws.String(rs.Primary.ID),
		})

		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() != "NoSuchCORSConfiguration" {
				return fmt.Errorf("GetBucketCors error: %v", err)
			}
		}

		if !reflect.DeepEqual(out.CORSRules, corsRules) {
			return fmt.Errorf("bad error cors rule, expected: %v, got %v", corsRules, out.CORSRules)
		}

		return nil
	}
}

func testAccCheckAWSS3RequestPayer(n, expectedPayer string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		conn := testAccProvider.Meta().(*AWSClient).s3conn

		out, err := conn.GetBucketRequestPayment(&s3.GetBucketRequestPaymentInput{
			Bucket: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return fmt.Errorf("GetBucketRequestPayment error: %v", err)
		}

		if *out.Payer != expectedPayer {
			return fmt.Errorf("bad error request payer type, expected: %v, got %v",
				expectedPayer, out.Payer)
		}

		return nil
	}
}

func testAccCheckAWSS3BucketLogging(n, b, p string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		conn := testAccProvider.Meta().(*AWSClient).s3conn

		out, err := conn.GetBucketLogging(&s3.GetBucketLoggingInput{
			Bucket: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return fmt.Errorf("GetBucketLogging error: %v", err)
		}

		if out.LoggingEnabled == nil {
			return fmt.Errorf("logging not enabled for bucket: %s", rs.Primary.ID)
		}

		tb := s.RootModule().Resources[b]

		if v := out.LoggingEnabled.TargetBucket; v == nil {
			if tb.Primary.ID != "" {
				return fmt.Errorf("bad target bucket, found nil, expected: %s", tb.Primary.ID)
			}
		} else {
			if *v != tb.Primary.ID {
				return fmt.Errorf("bad target bucket, expected: %s, got %s", tb.Primary.ID, *v)
			}
		}

		if v := out.LoggingEnabled.TargetPrefix; v == nil {
			if p != "" {
				return fmt.Errorf("bad target prefix, found nil, expected: %s", p)
			}
		} else {
			if *v != p {
				return fmt.Errorf("bad target prefix, expected: %s, got %s", p, *v)
			}
		}

		return nil
	}
}

func testAccCheckAWSS3BucketReplicationRules(n string, providerF func() *schema.Provider, rules []*s3.ReplicationRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		for _, rule := range rules {
			if dest := rule.Destination; dest != nil {
				if account := dest.Account; account != nil && strings.HasPrefix(aws.StringValue(dest.Account), "${") {
					resourceReference := strings.Replace(aws.StringValue(dest.Account), "${", "", 1)
					resourceReference = strings.Replace(resourceReference, "}", "", 1)
					resourceReferenceParts := strings.Split(resourceReference, ".")
					resourceAttribute := resourceReferenceParts[len(resourceReferenceParts)-1]
					resourceName := strings.Join(resourceReferenceParts[:len(resourceReferenceParts)-1], ".")
					value := s.RootModule().Resources[resourceName].Primary.Attributes[resourceAttribute]
					dest.Account = aws.String(value)
				}
				if ec := dest.EncryptionConfiguration; ec != nil {
					if ec.ReplicaKmsKeyID != nil {
						key_arn := s.RootModule().Resources["aws_kms_key.replica"].Primary.Attributes["arn"]
						ec.ReplicaKmsKeyID = aws.String(strings.Replace(*ec.ReplicaKmsKeyID, "${aws_kms_key.replica.arn}", key_arn, -1))
					}
				}
			}
			// Sort filter tags by key.
			if filter := rule.Filter; filter != nil {
				if and := filter.And; and != nil {
					if tags := and.Tags; tags != nil {
						sort.Slice(tags, func(i, j int) bool { return *tags[i].Key < *tags[j].Key })
					}
				}
			}
		}

		provider := providerF()

		conn := provider.Meta().(*AWSClient).s3conn
		out, err := conn.GetBucketReplication(&s3.GetBucketReplicationInput{
			Bucket: aws.String(rs.Primary.ID),
		})
		if err != nil {
			if isAWSErr(err, s3.ErrCodeNoSuchBucket, "") {
				return fmt.Errorf("S3 bucket not found")
			}
			if rules == nil {
				return nil
			}
			return fmt.Errorf("GetReplicationConfiguration error: %v", err)
		}

		for _, rule := range out.ReplicationConfiguration.Rules {
			// Sort filter tags by key.
			if filter := rule.Filter; filter != nil {
				if and := filter.And; and != nil {
					if tags := and.Tags; tags != nil {
						sort.Slice(tags, func(i, j int) bool { return *tags[i].Key < *tags[j].Key })
					}
				}
			}
		}
		if !reflect.DeepEqual(out.ReplicationConfiguration.Rules, rules) {
			return fmt.Errorf("bad replication rules, expected: %v, got %v", rules, out.ReplicationConfiguration.Rules)
		}

		return nil
	}
}

func testAccCheckAWSS3BucketUpdateGrantSingle(resourceName string) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		id := s.RootModule().Resources["data.aws_canonical_user_id.current"].Primary.ID
		gh := fmt.Sprintf("grant.%v", grantHash(map[string]interface{}{
			"id":   id,
			"type": "CanonicalUser",
			"uri":  "",
			"permissions": schema.NewSet(
				schema.HashString,
				[]interface{}{"FULL_CONTROL", "WRITE"},
			),
		}))
		for _, t := range []resource.TestCheckFunc{
			resource.TestCheckResourceAttr(resourceName, gh+".permissions.#", "2"),
			resource.TestCheckResourceAttr(resourceName, gh+".permissions.3535167073", "FULL_CONTROL"),
			resource.TestCheckResourceAttr(resourceName, gh+".permissions.2319431919", "WRITE"),
			resource.TestCheckResourceAttr(resourceName, gh+".type", "CanonicalUser"),
		} {
			if err := t(s); err != nil {
				return err
			}
		}
		return nil
	}
}

func testAccCheckAWSS3BucketUpdateGrantMulti(resourceName string) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		id := s.RootModule().Resources["data.aws_canonical_user_id.current"].Primary.ID
		gh1 := fmt.Sprintf("grant.%v", grantHash(map[string]interface{}{
			"id":   id,
			"type": "CanonicalUser",
			"uri":  "",
			"permissions": schema.NewSet(
				schema.HashString,
				[]interface{}{"READ"},
			),
		}))
		gh2 := fmt.Sprintf("grant.%v", grantHash(map[string]interface{}{
			"id":   "",
			"type": "Group",
			"uri":  "http://acs.amazonaws.com/groups/s3/LogDelivery",
			"permissions": schema.NewSet(
				schema.HashString,
				[]interface{}{"READ_ACP"},
			),
		}))
		for _, t := range []resource.TestCheckFunc{
			resource.TestCheckResourceAttr(resourceName, gh1+".permissions.#", "1"),
			resource.TestCheckResourceAttr(resourceName, gh1+".permissions.2931993811", "READ"),
			resource.TestCheckResourceAttr(resourceName, gh1+".type", "CanonicalUser"),
			resource.TestCheckResourceAttr(resourceName, gh2+".permissions.#", "1"),
			resource.TestCheckResourceAttr(resourceName, gh2+".permissions.1600971645", "READ_ACP"),
			resource.TestCheckResourceAttr(resourceName, gh2+".type", "Group"),
			resource.TestCheckResourceAttr(resourceName, gh2+".uri", "http://acs.amazonaws.com/groups/s3/LogDelivery"),
		} {
			if err := t(s); err != nil {
				return err
			}
		}
		return nil
	}
}

// These need a bit of randomness as the name can only be used once globally
// within AWS
func testAccBucketName(randInt int) string {
	return fmt.Sprintf("tf-test-bucket-%d", randInt)
}

func testAccCheckS3BucketDomainName(resourceName string, attributeName string, bucketName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		expectedValue := testAccProvider.Meta().(*AWSClient).PartitionHostname(fmt.Sprintf("%s.s3", bucketName))

		return resource.TestCheckResourceAttr(resourceName, attributeName, expectedValue)(s)
	}
}

func testAccBucketRegionalDomainName(randInt int, region string) string {
	bucket := fmt.Sprintf("tf-test-bucket-%d", randInt)
	regionalEndpoint, err := BucketRegionalDomainName(bucket, region)
	if err != nil {
		return fmt.Sprintf("Regional endpoint not found for bucket %s", bucket)
	}
	return regionalEndpoint
}

func testAccCheckS3BucketWebsiteEndpoint(resourceName string, attributeName string, bucketName string, region string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		website := WebsiteEndpoint(testAccProvider.Meta().(*AWSClient), bucketName, region)
		expectedValue := website.Endpoint

		return resource.TestCheckResourceAttr(resourceName, attributeName, expectedValue)(s)
	}
}

func testAccAWSS3BucketPolicy(randInt int, partition string) string {
	return fmt.Sprintf(`{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Sid": "",
			"Effect": "Allow",
			"Principal": {"AWS": "*"},
			"Action": "s3:GetObject",
			"Resource": "arn:%s:s3:::tf-test-bucket-%d/*"
		}
	]
}
`, partition, randInt)
}

func testAccAWSS3BucketConfig(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-%d"
}
`, randInt)
}

func testAccAWSS3BucketBasic(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
	bucket = "tf-test-bucket-%d"
}
`, randInt)
}

func testAccAWSS3BucketConfig_withNoTags(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
  acl = "private"
  force_destroy = false
}
`, bucketName)
}

func testAccAWSS3BucketConfig_withTags(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
  acl = "private"
  force_destroy = false
  tags = {
    Key1 = "AAA"
    Key2 = "BBB"
    Key3 = "CCC"
  }
}
`, bucketName)
}

func testAccAWSS3BucketConfig_withUpdatedTags(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
  acl = "private"
  force_destroy = false
  tags = {
    Key2 = "BBB"
    Key3 = "XXX"
    Key4 = "DDD"
    Key5 = "EEE"
  }
}
`, bucketName)
}

func testAccAWSS3MultiBucketConfigWithTags(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket1" {
  bucket        = "tf-test-bucket-1-%[1]d"
  acl           = "private"
  force_destroy = true

  tags = {
    Name        = "tf-test-bucket-1-%[1]d"
    Environment = "%[1]d"
  }
}

resource "aws_s3_bucket" "bucket2" {
  bucket        = "tf-test-bucket-2-%[1]d"
  acl           = "private"
  force_destroy = true

  tags = {
    Name        = "tf-test-bucket-2-%[1]d"
    Environment = "%[1]d"
  }
}

resource "aws_s3_bucket" "bucket3" {
  bucket        = "tf-test-bucket-3-%[1]d"
  acl           = "private"
  force_destroy = true

  tags = {
    Name        = "tf-test-bucket-3-%[1]d"
    Environment = "%[1]d"
  }
}

resource "aws_s3_bucket" "bucket4" {
  bucket        = "tf-test-bucket-4-%[1]d"
  acl           = "private"
  force_destroy = true

  tags = {
    Name        = "tf-test-bucket-4-%[1]d"
    Environment = "%[1]d"
  }
}

resource "aws_s3_bucket" "bucket5" {
  bucket        = "tf-test-bucket-5-%[1]d"
  acl           = "private"
  force_destroy = true

  tags = {
    Name        = "tf-test-bucket-5-%[1]d"
    Environment = "%[1]d"
  }
}

resource "aws_s3_bucket" "bucket6" {
  bucket        = "tf-test-bucket-6-%[1]d"
  acl           = "private"
  force_destroy = true

  tags = {
    Name        = "tf-test-bucket-6-%[1]d"
    Environment = "%[1]d"
  }
}
`, randInt)
}

func testAccAWSS3BucketConfigWithRegion(bucketName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
  region = data.aws_region.current.name
}
`, bucketName)
}

func testAccAWSS3BucketWebsiteConfig(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-%d"
  acl    = "public-read"

  website {
    index_document = "index.html"
  }
}
`, randInt)
}

func testAccAWSS3BucketWebsiteConfigWithError(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-%d"
  acl    = "public-read"

  website {
    index_document = "index.html"
    error_document = "error.html"
  }
}
`, randInt)
}

func testAccAWSS3BucketWebsiteConfigWithRedirect(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-%d"
  acl    = "public-read"

  website {
    redirect_all_requests_to = "hashicorp.com?my=query"
  }
}
`, randInt)
}

func testAccAWSS3BucketWebsiteConfigWithHttpsRedirect(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-%d"
  acl    = "public-read"

  website {
    redirect_all_requests_to = "https://hashicorp.com?my=query"
  }
}
`, randInt)
}

func testAccAWSS3BucketWebsiteConfigWithRoutingRules(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-%d"
  acl    = "public-read"

  website {
    index_document = "index.html"
    error_document = "error.html"

    routing_rules = <<EOF
[{
	"Condition": {
		"KeyPrefixEquals": "docs/"
	},
	"Redirect": {
		"ReplaceKeyPrefixWith": "documents/"
	}
}]
EOF
  }
}
`, randInt)
}

func testAccAWSS3BucketConfigWithAcceleration(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket              = "tf-test-bucket-%d"
  acceleration_status = "Enabled"
}
`, randInt)
}

func testAccAWSS3BucketConfigWithoutAcceleration(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket              = "tf-test-bucket-%d"
  acceleration_status = "Suspended"
}
`, randInt)
}

func testAccAWSS3BucketConfigRequestPayerBucketOwner(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket        = "tf-test-bucket-%d"
  request_payer = "BucketOwner"
}
`, randInt)
}

func testAccAWSS3BucketConfigRequestPayerRequester(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket        = "tf-test-bucket-%d"
  request_payer = "Requester"
}
`, randInt)
}

func testAccAWSS3BucketConfigWithPolicy(randInt int, partition string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-%d"
  acl    = "public-read"
  policy = %s
}
`, randInt, strconv.Quote(testAccAWSS3BucketPolicy(randInt, partition)))
}

func testAccAWSS3BucketDestroyedConfig(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-%d"
  acl    = "public-read"
}
`, randInt)
}

func testAccAWSS3BucketEnableDefaultEncryption(randInt int) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "arbitrary" {
  description             = "KMS Key for Bucket Testing %d"
  deletion_window_in_days = 10
}

resource "aws_s3_bucket" "arbitrary" {
  bucket = "tf-test-bucket-%d"

  server_side_encryption_configuration {
    rule {
      apply_server_side_encryption_by_default {
        kms_master_key_id = "${aws_kms_key.arbitrary.arn}"
        sse_algorithm     = "aws:kms"
      }
    }
  }
}
`, randInt, randInt)
}

func testAccAWSS3BucketEnableDefaultEncryptionWithAES256(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "arbitrary" {
  bucket = "tf-test-bucket-%d"

  server_side_encryption_configuration {
    rule {
      apply_server_side_encryption_by_default {
        sse_algorithm = "AES256"
      }
    }
  }
}
`, randInt)
}

func testAccAWSS3BucketEnableDefaultEncryptionWithDefaultKey(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "arbitrary" {
  bucket = "tf-test-bucket-%d"

  server_side_encryption_configuration {
    rule {
      apply_server_side_encryption_by_default {
        sse_algorithm = "aws:kms"
      }
    }
  }
}
`, randInt)
}

func testAccAWSS3BucketDisableDefaultEncryption(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "arbitrary" {
  bucket = "tf-test-bucket-%d"
}
`, randInt)
}

func testAccAWSS3BucketConfigWithEmptyPolicy(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-%d"
  acl    = "public-read"
  policy = ""
}
`, randInt)
}

func testAccAWSS3BucketConfigWithVersioning(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-%d"

  versioning {
    enabled = true
  }
}
`, randInt)
}

func testAccAWSS3BucketConfigWithDisableVersioning(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-%d"

  versioning {
    enabled = false
  }
}
`, randInt)
}

func testAccAWSS3BucketConfigWithCORS(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-%d"

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["PUT", "POST"]
    allowed_origins = ["https://www.example.com"]
    expose_headers  = ["x-amz-server-side-encryption", "ETag"]
    max_age_seconds = 3000
  }
}
`, randInt)
}

func testAccAWSS3BucketConfigWithCORSEmptyOrigin(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-%d"

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["PUT", "POST"]
    allowed_origins = [""]
    expose_headers  = ["x-amz-server-side-encryption", "ETag"]
    max_age_seconds = 3000
  }
}
`, randInt)
}

func testAccAWSS3BucketConfigWithAcl(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
	bucket = "tf-test-bucket-%d"
	acl = "public-read"
}
`, randInt)
}

func testAccAWSS3BucketConfigWithAclUpdate(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
	bucket = "tf-test-bucket-%d"
	acl = "private"
}
`, randInt)
}

func testAccAWSS3BucketConfigWithGrants(randInt int) string {
	return fmt.Sprintf(`
data "aws_canonical_user_id" "current" {}
resource "aws_s3_bucket" "bucket" {
	bucket = "tf-test-bucket-%d"
	grant {
        id = "${data.aws_canonical_user_id.current.id}"
        type = "CanonicalUser"
        permissions = ["FULL_CONTROL", "WRITE"]
    }
}
`, randInt)
}

func testAccAWSS3BucketConfigWithGrantsUpdate(randInt int) string {
	return fmt.Sprintf(`
data "aws_canonical_user_id" "current" {}
resource "aws_s3_bucket" "bucket" {
	bucket = "tf-test-bucket-%d"
	grant {
        id = "${data.aws_canonical_user_id.current.id}"
        type = "CanonicalUser"
        permissions = ["READ"]
    }
    grant {
        type = "Group"
        permissions = ["READ_ACP"]
        uri = "http://acs.amazonaws.com/groups/s3/LogDelivery"
    }
}
`, randInt)
}

func testAccAWSS3BucketConfigWithLogging(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "log_bucket" {
  bucket = "tf-test-log-bucket-%d"
  acl    = "log-delivery-write"
}

resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-%d"
  acl    = "private"

  logging {
    target_bucket = "${aws_s3_bucket.log_bucket.id}"
    target_prefix = "log/"
  }
}
`, randInt, randInt)
}

func testAccAWSS3BucketConfigWithLifecycle(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-%d"
  acl    = "private"

  lifecycle_rule {
    id      = "id1"
    prefix  = "path1/"
    enabled = true

    expiration {
      days = 365
    }

    transition {
      days          = 30
      storage_class = "STANDARD_IA"
    }

    transition {
      days          = 60
      storage_class = "INTELLIGENT_TIERING"
    }

    transition {
      days          = 90
      storage_class = "ONEZONE_IA"
    }

    transition {
      days          = 120
      storage_class = "GLACIER"
    }

    transition {
      days          = 210
      storage_class = "DEEP_ARCHIVE"
    }
  }

  lifecycle_rule {
    id      = "id2"
    prefix  = "path2/"
    enabled = true

    expiration {
      date = "2016-01-12"
    }
  }

  lifecycle_rule {
    id      = "id3"
    prefix  = "path3/"
    enabled = true

    transition {
      days          = 0
      storage_class = "GLACIER"
    }
  }

  lifecycle_rule {
    id      = "id4"
    prefix  = "path4/"
    enabled = true

    tags = {
      "tagKey"    = "tagValue"
      "terraform" = "hashicorp"
    }

    expiration {
      date = "2016-01-12"
    }
  }
	lifecycle_rule {
		id = "id5"
		enabled = true

		tags = {
			"tagKey" = "tagValue"
			"terraform" = "hashicorp"
		}

		transition {
			days = 0
			storage_class = "GLACIER"
		}
	}
	lifecycle_rule {
		id = "id6"
		enabled = true

		tags = {
			"tagKey" = "tagValue"
		}

		transition {
			days = 0
			storage_class = "GLACIER"
		}
	}
}
`, randInt)
}

func testAccAWSS3BucketConfigWithLifecycleExpireMarker(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-%d"
  acl    = "private"

  lifecycle_rule {
    id      = "id1"
    prefix  = "path1/"
    enabled = true

    expiration {
      expired_object_delete_marker = "true"
    }
  }
}
`, randInt)
}

func testAccAWSS3BucketConfigWithVersioningLifecycle(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-%d"
  acl    = "private"

  versioning {
    enabled = false
  }

  lifecycle_rule {
    id      = "id1"
    prefix  = "path1/"
    enabled = true

    noncurrent_version_expiration {
      days = 365
    }

    noncurrent_version_transition {
      days          = 30
      storage_class = "STANDARD_IA"
    }

    noncurrent_version_transition {
      days          = 60
      storage_class = "GLACIER"
    }
  }

  lifecycle_rule {
    id      = "id2"
    prefix  = "path2/"
    enabled = false

    noncurrent_version_expiration {
      days = 365
    }
  }

  lifecycle_rule {
    id      = "id3"
    prefix  = "path3/"
    enabled = true

    noncurrent_version_transition {
      days          = 0
      storage_class = "GLACIER"
    }
  }
}
`, randInt)
}

func testAccAWSS3BucketConfigReplicationBasic(randInt int) string {
	return testAccAlternateRegionProviderConfig() + fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "role" {
  name               = "tf-iam-role-replication-%[1]d"
  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "s3.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_s3_bucket" "destination" {
  provider = "aws.alternate"
  bucket   = "tf-test-bucket-destination-%[1]d"

  versioning {
    enabled = true
  }
}
`, randInt)
}

func testAccAWSS3BucketConfigReplication(randInt int) string {
	return testAccAWSS3BucketConfigReplicationBasic(randInt) + fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
    bucket   = "tf-test-bucket-%[1]d"
    acl      = "private"

    versioning {
        enabled = true
    }
}
`, randInt)
}

func testAccAWSS3BucketConfigReplicationWithConfiguration(randInt int, storageClass string) string {
	return testAccAWSS3BucketConfigReplicationBasic(randInt) + fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
    bucket   = "tf-test-bucket-%[1]d"
    acl      = "private"

    versioning {
        enabled = true
    }

    replication_configuration {
        role = "${aws_iam_role.role.arn}"
        rules {
            id     = "foobar"
            prefix = "foo"
            status = "Enabled"

            destination {
                bucket        = "${aws_s3_bucket.destination.arn}"
                storage_class = "%[2]s"
            }
        }
    }
}
`, randInt, storageClass)
}

func testAccAWSS3BucketConfigReplicationWithSseKmsEncryptedObjects(randInt int) string {
	return testAccAWSS3BucketConfigReplicationBasic(randInt) + fmt.Sprintf(`
resource "aws_kms_key" "replica" {
  provider                = "aws.alternate"
  description             = "TF Acceptance Test S3 repl KMS key"
  deletion_window_in_days = 7
}

resource "aws_s3_bucket" "bucket" {
    bucket   = "tf-test-bucket-%[1]d"
    acl      = "private"

    versioning {
        enabled = true
    }

    replication_configuration {
        role = "${aws_iam_role.role.arn}"
        rules {
            id     = "foobar"
            prefix = "foo"
            status = "Enabled"

            destination {
                bucket        = "${aws_s3_bucket.destination.arn}"
                storage_class = "STANDARD"
                replica_kms_key_id = "${aws_kms_key.replica.arn}"
            }

            source_selection_criteria {
                sse_kms_encrypted_objects {
                  enabled = true
                }
            }
        }
    }
}
`, randInt)
}

func testAccAWSS3BucketConfigReplicationWithAccessControlTranslation(randInt int) string {
	return testAccAWSS3BucketConfigReplicationBasic(randInt) + fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_s3_bucket" "bucket" {
    bucket   = "tf-test-bucket-%[1]d"
    acl      = "private"

    versioning {
        enabled = true
    }

    replication_configuration {
        role = "${aws_iam_role.role.arn}"
        rules {
            id     = "foobar"
            prefix = "foo"
            status = "Enabled"

            destination {
                account_id         = "${data.aws_caller_identity.current.account_id}"
                bucket             = "${aws_s3_bucket.destination.arn}"
                storage_class      = "STANDARD"

                access_control_translation {
                    owner = "Destination"
                }
            }
        }
    }
}
`, randInt)
}

func testAccAWSS3BucketConfigReplicationWithSseKmsEncryptedObjectsAndAccessControlTranslation(randInt int) string {
	return testAccAWSS3BucketConfigReplicationBasic(randInt) + fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_kms_key" "replica" {
  provider                = "aws.alternate"
  description             = "TF Acceptance Test S3 repl KMS key"
  deletion_window_in_days = 7
}

resource "aws_s3_bucket" "bucket" {
    bucket   = "tf-test-bucket-%[1]d"
    acl      = "private"

    versioning {
        enabled = true
    }

    replication_configuration {
        role = "${aws_iam_role.role.arn}"
        rules {
            id     = "foobar"
            prefix = "foo"
            status = "Enabled"

            destination {
                account_id         = "${data.aws_caller_identity.current.account_id}"
                bucket             = "${aws_s3_bucket.destination.arn}"
                storage_class      = "STANDARD"
                replica_kms_key_id = "${aws_kms_key.replica.arn}"

                access_control_translation {
                    owner = "Destination"
                }
            }

            source_selection_criteria {
                sse_kms_encrypted_objects {
                  enabled = true
                }
            }
        }
    }
}
`, randInt)
}

func testAccAWSS3BucketConfigReplicationWithoutStorageClass(randInt int) string {
	return testAccAWSS3BucketConfigReplicationBasic(randInt) + fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
    bucket   = "tf-test-bucket-%[1]d"
    acl      = "private"

    versioning {
        enabled = true
    }

    replication_configuration {
        role = "${aws_iam_role.role.arn}"
        rules {
            id     = "foobar"
            prefix = "foo"
            status = "Enabled"

            destination {
                bucket        = "${aws_s3_bucket.destination.arn}"
            }
        }
    }
}
`, randInt)
}

func testAccAWSS3BucketConfigReplicationWithoutPrefix(randInt int) string {
	return testAccAWSS3BucketConfigReplicationBasic(randInt) + fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
    bucket   = "tf-test-bucket-%[1]d"
    acl      = "private"

    versioning {
        enabled = true
    }

    replication_configuration {
        role = "${aws_iam_role.role.arn}"
        rules {
            id     = "foobar"
            status = "Enabled"

            destination {
                bucket        = "${aws_s3_bucket.destination.arn}"
                storage_class = "STANDARD"
            }
        }
    }
}
`, randInt)
}

func testAccAWSS3BucketConfigReplicationNoVersioning(randInt int) string {
	return testAccAWSS3BucketConfigReplicationBasic(randInt) + fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
    bucket   = "tf-test-bucket-%[1]d"
    acl      = "private"

    replication_configuration {
        role = "${aws_iam_role.role.arn}"
        rules {
            id     = "foobar"
            prefix = "foo"
            status = "Enabled"

            destination {
                bucket        = "${aws_s3_bucket.destination.arn}"
                storage_class = "STANDARD"
            }
        }
    }
}
`, randInt)
}

func testAccAWSS3BucketConfigReplicationWithV2ConfigurationNoTags(randInt int) string {
	return testAccAWSS3BucketConfigReplicationBasic(randInt) + fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
    bucket   = "tf-test-bucket-%[1]d"
    acl      = "private"

    versioning {
        enabled = true
    }

    replication_configuration {
        role = "${aws_iam_role.role.arn}"
        rules {
            id     = "foobar"
            status = "Enabled"

            filter {
                prefix = "foo"
            }

            destination {
                bucket        = "${aws_s3_bucket.destination.arn}"
                storage_class = "STANDARD"
            }
        }
    }
}
`, randInt)
}

func testAccAWSS3BucketConfigReplicationWithV2ConfigurationOnlyOneTag(randInt int) string {
	return testAccAWSS3BucketConfigReplicationBasic(randInt) + fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
    bucket   = "tf-test-bucket-%[1]d"
    acl      = "private"

    versioning {
        enabled = true
    }

    replication_configuration {
        role = "${aws_iam_role.role.arn}"
        rules {
            id     = "foobar"
            status = "Enabled"

            priority = 42

            filter {
  tags = {
                    ReplicateMe = "Yes"
                }
            }

            destination {
                bucket        = "${aws_s3_bucket.destination.arn}"
                storage_class = "STANDARD"
            }
        }
    }
}
`, randInt)
}

func testAccAWSS3BucketConfigReplicationWithV2ConfigurationPrefixAndTags(randInt int) string {
	return testAccAWSS3BucketConfigReplicationBasic(randInt) + fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
    bucket   = "tf-test-bucket-%[1]d"
    acl      = "private"

    versioning {
        enabled = true
    }

    replication_configuration {
        role = "${aws_iam_role.role.arn}"
        rules {
            id     = "foobar"
            status = "Enabled"

            priority = 41

            filter {
                prefix = "foo"

  tags = {
                    AnotherTag  = "OK"
                    ReplicateMe = "Yes"
                }
            }

            destination {
                bucket        = "${aws_s3_bucket.destination.arn}"
                storage_class = "STANDARD"
            }
        }
    }
}
`, randInt)
}

func testAccAWSS3BucketConfigReplicationWithV2ConfigurationMultipleTags(randInt int) string {
	return testAccAWSS3BucketConfigReplicationBasic(randInt) + fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
    bucket   = "tf-test-bucket-%[1]d"
    acl      = "private"

    versioning {
        enabled = true
    }

    replication_configuration {
        role = "${aws_iam_role.role.arn}"
        rules {
            id     = "foobar"
            status = "Enabled"

            filter {
  tags = {
                    AnotherTag  = "OK"
                    Foo         = "Bar"
                    ReplicateMe = "Yes"
                }
            }

            destination {
                bucket        = "${aws_s3_bucket.destination.arn}"
                storage_class = "STANDARD"
            }
        }
    }
}
`, randInt)
}

func testAccAWSS3BucketObjectLockEnabledNoDefaultRetention(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "arbitrary" {
  bucket = "tf-test-bucket-%d"

  object_lock_configuration {
    object_lock_enabled = "Enabled"
  }
}
`, randInt)
}

func testAccAWSS3BucketObjectLockEnabledWithDefaultRetention(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "arbitrary" {
  bucket = "tf-test-bucket-%d"

  object_lock_configuration {
    object_lock_enabled = "Enabled"

    rule {
      default_retention {
        mode = "COMPLIANCE"
        days = 3
      }
    }
  }
}
`, randInt)
}

func testAccAWSS3BucketConfig_forceDestroy(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = "%s"
  acl = "private"
  force_destroy = true
}
`, bucketName)
}

func testAccAWSS3BucketConfig_forceDestroyWithObjectLockEnabled(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = "%s"
  acl = "private"
  force_destroy = true

  versioning {
    enabled = true
  }

  object_lock_configuration {
    object_lock_enabled = "Enabled"
  }
}
`, bucketName)
}

const testAccAWSS3BucketConfigBucketEmptyString = `
resource "aws_s3_bucket" "test" {
  bucket = ""
}
`

const testAccAWSS3BucketConfig_namePrefix = `
resource "aws_s3_bucket" "test" {
	bucket_prefix = "tf-test-"
}
`

const testAccAWSS3BucketConfig_generatedName = `
resource "aws_s3_bucket" "test" {
	bucket_prefix = "tf-test-"
}
`
