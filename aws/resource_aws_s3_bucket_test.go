package aws

import (
	"context"
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
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudformation/waiter"
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
	region, err := s3manager.GetBucketRegionWithClient(context.Background(), conn, bucket, func(r *request.Request) {
		// By default, GetBucketRegion forces virtual host addressing, which
		// is not compatible with many non-AWS implementations. Instead, pass
		// the provider s3_force_path_style configuration, which defaults to
		// false, but allows override.
		r.Config.S3ForcePathStyle = conn.Config.S3ForcePathStyle
	})
	if err != nil {
		return "", err
	}

	return region, nil
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
	bucketName := acctest.RandomWithPrefix("tf-test-bucket")
	region := testAccGetRegion()
	hostedZoneID, _ := HostedZoneIDForRegion(region)
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfig_Basic(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "hosted_zone_id", hostedZoneID),
					resource.TestCheckResourceAttr(resourceName, "region", region),
					resource.TestCheckNoResourceAttr(resourceName, "website_endpoint"),
					testAccCheckResourceAttrGlobalARNNoAccount(resourceName, "arn", "s3", bucketName),
					resource.TestCheckResourceAttr(resourceName, "bucket", bucketName),
					testAccCheckS3BucketDomainName(resourceName, "bucket_domain_name", bucketName),
					resource.TestCheckResourceAttr(resourceName, "bucket_regional_domain_name", testAccBucketRegionalDomainName(bucketName, region)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
		},
	})
}

// Support for common Terraform 0.11 pattern
// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/7868
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
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
		},
	})
}

func TestAccAWSS3Bucket_tagsWithNoSystemTags(t *testing.T) {
	resourceName := "aws_s3_bucket.bucket"
	bucketName := acctest.RandomWithPrefix("tf-test-bucket")

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
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
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
	bucketName := acctest.RandomWithPrefix("tf-test-bucket")

	var stackID string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckAWSS3BucketDestroy,
			func(s *terraform.State) error {
				// Tear down CF stack.
				conn := testAccProvider.Meta().(*AWSClient).cfconn

				requestToken := resource.UniqueId()
				req := &cloudformation.DeleteStackInput{
					StackName:          aws.String(stackID),
					ClientRequestToken: aws.String(requestToken),
				}

				log.Printf("[DEBUG] Deleting CloudFormation stack: %s", req)
				if _, err := conn.DeleteStack(req); err != nil {
					return fmt.Errorf("error deleting CloudFormation stack: %w", err)
				}

				if _, err := waiter.StackDeleted(conn, stackID, requestToken, 10*time.Minute); err != nil {
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
					testAccCheckAWSS3BucketCreateViaCloudFormation(bucketName, &stackID),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
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

func TestAccAWSS3Bucket_ignoreTags(t *testing.T) {
	resourceName := "aws_s3_bucket.bucket"
	bucketName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: composeConfig(
					testAccProviderConfigIgnoreTagsKeyPrefixes1("ignorekey"),
					testAccAWSS3BucketConfig_withNoTags(bucketName)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					testAccCheckAWSS3BucketUpdateTags(resourceName, nil, map[string]string{"ignorekey1": "ignorevalue1"}),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					testAccCheckAWSS3BucketCheckTags(resourceName, map[string]string{
						"ignorekey1": "ignorevalue1",
					}),
				),
			},
			{
				Config: composeConfig(
					testAccProviderConfigIgnoreTagsKeyPrefixes1("ignorekey"),
					testAccAWSS3BucketConfig_withTags(bucketName)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "AAA"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "CCC"),
					testAccCheckAWSS3BucketCheckTags(resourceName, map[string]string{
						"ignorekey1": "ignorevalue1",
						"Key1":       "AAA",
						"Key2":       "BBB",
						"Key3":       "CCC",
					}),
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
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
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
					resource.TestMatchResourceAttr(resourceName, "bucket", regexp.MustCompile("^tf-test-")),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl", "bucket_prefix"},
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
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl", "bucket_prefix"},
			},
		},
	})
}

func TestAccAWSS3Bucket_acceleration(t *testing.T) {
	bucketName := acctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigWithAcceleration(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "acceleration_status", "Enabled"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
			{
				Config: testAccAWSS3BucketConfigWithoutAcceleration(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "acceleration_status", "Suspended"),
				),
			},
		},
	})
}

func TestAccAWSS3Bucket_RequestPayer(t *testing.T) {
	bucketName := acctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigRequestPayerBucketOwner(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "request_payer", "BucketOwner"),
					testAccCheckAWSS3RequestPayer(resourceName, "BucketOwner"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
			{
				Config: testAccAWSS3BucketConfigRequestPayerRequester(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "request_payer", "Requester"),
					testAccCheckAWSS3RequestPayer(resourceName, "Requester"),
				),
			},
		},
	})
}

func TestAccAWSS3Bucket_Policy(t *testing.T) {
	bucketName := acctest.RandomWithPrefix("tf-test-bucket")
	partition := testAccGetPartition()
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigWithPolicy(bucketName, partition),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					testAccCheckAWSS3BucketPolicy(resourceName, testAccAWSS3BucketPolicy(bucketName, partition)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"acl",
					"force_destroy",
					"grant",
					// NOTE: Prior to Terraform AWS Provider 3.0, this attribute did not import correctly either.
					//       The Read function does not require GetBucketPolicy, if the argument is not configured.
					//       Rather than introduce that breaking change as well with 3.0, instead we leave the
					//       current Read behavior and note this will be deprecated in a later 3.x release along
					//       with other inline policy attributes across the provider.
					"policy",
				},
			},
			{
				Config: testAccAWSS3BucketConfig_Basic(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					testAccCheckAWSS3BucketPolicy(resourceName, ""),
				),
			},
			{
				Config: testAccAWSS3BucketConfigWithEmptyPolicy(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					testAccCheckAWSS3BucketPolicy(resourceName, ""),
				),
			},
		},
	})
}

func TestAccAWSS3Bucket_UpdateAcl(t *testing.T) {
	bucketName := acctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigWithAcl(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "acl", "public-read"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl", "grant"},
			},
			{
				Config: testAccAWSS3BucketConfigWithAclUpdate(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "acl", "private"),
				),
			},
		},
	})
}

func TestAccAWSS3Bucket_UpdateGrant(t *testing.T) {
	bucketName := acctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigWithGrants(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "grant.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "grant.*", map[string]string{
						"permissions.#": "2",
						"type":          "CanonicalUser",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "grant.*.permissions.*", "FULL_CONTROL"),
					resource.TestCheckTypeSetElemAttr(resourceName, "grant.*.permissions.*", "WRITE"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
			{
				Config: testAccAWSS3BucketConfigWithGrantsUpdate(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "grant.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "grant.*", map[string]string{
						"permissions.#": "1",
						"type":          "CanonicalUser",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "grant.*.permissions.*", "READ"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "grant.*", map[string]string{
						"permissions.#": "1",
						"type":          "Group",
						"uri":           "http://acs.amazonaws.com/groups/s3/LogDelivery",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "grant.*.permissions.*", "READ_ACP"),
				),
			},
			{
				Config: testAccAWSS3BucketConfig_Basic(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "grant.#", "0"),
				),
			},
		},
	})
}

func TestAccAWSS3Bucket_AclToGrant(t *testing.T) {
	bucketName := acctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigWithAcl(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "acl", "public-read"),
					resource.TestCheckResourceAttr(resourceName, "grant.#", "0"),
				),
			},
			{
				Config: testAccAWSS3BucketConfigWithGrants(bucketName),
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
	bucketName := acctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigWithGrants(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "grant.#", "1"),
				),
			},
			{
				Config: testAccAWSS3BucketConfigWithAcl(bucketName),
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
	bucketName := acctest.RandomWithPrefix("tf-test-bucket")
	region := testAccGetRegion()
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketWebsiteConfig(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					testAccCheckAWSS3BucketWebsite(resourceName, "index.html", "", "", ""),
					testAccCheckS3BucketWebsiteEndpoint(resourceName, "website_endpoint", bucketName, region),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl", "grant"},
			},
			{
				Config: testAccAWSS3BucketWebsiteConfigWithError(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					testAccCheckAWSS3BucketWebsite(resourceName, "index.html", "error.html", "", ""),
					testAccCheckS3BucketWebsiteEndpoint(resourceName, "website_endpoint", bucketName, region),
				),
			},
			{
				Config: testAccAWSS3BucketConfig_Basic(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					testAccCheckAWSS3BucketWebsite(resourceName, "", "", "", ""),
					resource.TestCheckResourceAttr(resourceName, "website_endpoint", ""),
				),
			},
		},
	})
}

func TestAccAWSS3Bucket_WebsiteRedirect(t *testing.T) {
	bucketName := acctest.RandomWithPrefix("tf-test-bucket")
	region := testAccGetRegion()
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketWebsiteConfigWithRedirect(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					testAccCheckAWSS3BucketWebsite(resourceName, "", "", "", "hashicorp.com?my=query"),
					testAccCheckS3BucketWebsiteEndpoint(resourceName, "website_endpoint", bucketName, region),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl", "grant"},
			},
			{
				Config: testAccAWSS3BucketWebsiteConfigWithHttpsRedirect(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					testAccCheckAWSS3BucketWebsite(resourceName, "", "", "https", "hashicorp.com?my=query"),
					testAccCheckS3BucketWebsiteEndpoint(resourceName, "website_endpoint", bucketName, region),
				),
			},
			{
				Config: testAccAWSS3BucketConfig_Basic(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					testAccCheckAWSS3BucketWebsite(resourceName, "", "", "", ""),
					resource.TestCheckResourceAttr(resourceName, "website_endpoint", ""),
				),
			},
		},
	})
}

func TestAccAWSS3Bucket_WebsiteRoutingRules(t *testing.T) {
	bucketName := acctest.RandomWithPrefix("tf-test-bucket")
	region := testAccGetRegion()
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketWebsiteConfigWithRoutingRules(bucketName),
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
					testAccCheckS3BucketWebsiteEndpoint(resourceName, "website_endpoint", bucketName, region),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl", "grant"},
			},
			{
				Config: testAccAWSS3BucketConfig_Basic(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					testAccCheckAWSS3BucketWebsite(resourceName, "", "", "", ""),
					testAccCheckAWSS3BucketWebsiteRoutingRules(resourceName, nil),
					resource.TestCheckResourceAttr(resourceName, "website_endpoint", ""),
				),
			},
		},
	})
}

func TestAccAWSS3Bucket_enableDefaultEncryption_whenTypical(t *testing.T) {
	bucketName := acctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.arbitrary"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketEnableDefaultEncryption(bucketName),
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
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
		},
	})
}

func TestAccAWSS3Bucket_enableDefaultEncryption_whenAES256IsUsed(t *testing.T) {
	bucketName := acctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.arbitrary"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketEnableDefaultEncryptionWithAES256(bucketName),
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
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
		},
	})
}

func TestAccAWSS3Bucket_disableDefaultEncryption_whenDefaultEncryptionIsEnabled(t *testing.T) {
	bucketName := acctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.arbitrary"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketEnableDefaultEncryptionWithDefaultKey(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
			{
				Config: testAccAWSS3BucketDisableDefaultEncryption(bucketName),
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
	bucketName := acctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketDestroyedConfig(bucketName),
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
	bucketName := acctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfig_Basic(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					testAccCheckAWSS3BucketVersioning(resourceName, ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
			{
				Config: testAccAWSS3BucketConfigWithVersioning(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					testAccCheckAWSS3BucketVersioning(resourceName, s3.BucketVersioningStatusEnabled),
				),
			},
			{
				Config: testAccAWSS3BucketConfigWithDisableVersioning(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					testAccCheckAWSS3BucketVersioning(resourceName, s3.BucketVersioningStatusSuspended),
				),
			},
		},
	})
}

func TestAccAWSS3Bucket_Cors_Update(t *testing.T) {
	bucketName := acctest.RandomWithPrefix("tf-test-bucket")
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
				Config: testAccAWSS3BucketConfigWithCORS(bucketName),
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
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
			{
				Config: testAccAWSS3BucketConfigWithCORS(bucketName),
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
	bucketName := acctest.RandomWithPrefix("tf-test-bucket")
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
				Config: testAccAWSS3BucketConfigWithCORS(bucketName),
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
	bucketName := acctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigWithCORSEmptyOrigin(bucketName),
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
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
		},
	})
}

func TestAccAWSS3Bucket_Logging(t *testing.T) {
	bucketName := acctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigWithLogging(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					testAccCheckAWSS3BucketLogging(resourceName, "aws_s3_bucket.log_bucket", "log/"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
		},
	})
}

func TestAccAWSS3Bucket_LifecycleBasic(t *testing.T) {
	bucketName := acctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigWithLifecycle(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.id", "id1"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.prefix", "path1/"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.expiration.0.days", "365"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.expiration.0.date", ""),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.expiration.0.expired_object_delete_marker", "false"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lifecycle_rule.0.transition.*", map[string]string{
						"date":          "",
						"days":          "30",
						"storage_class": "STANDARD_IA",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lifecycle_rule.0.transition.*", map[string]string{
						"date":          "",
						"days":          "60",
						"storage_class": "INTELLIGENT_TIERING",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lifecycle_rule.0.transition.*", map[string]string{
						"date":          "",
						"days":          "90",
						"storage_class": "ONEZONE_IA",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lifecycle_rule.0.transition.*", map[string]string{
						"date":          "",
						"days":          "120",
						"storage_class": "GLACIER",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lifecycle_rule.0.transition.*", map[string]string{
						"date":          "",
						"days":          "210",
						"storage_class": "DEEP_ARCHIVE",
					}),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.1.id", "id2"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.1.prefix", "path2/"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.1.expiration.0.date", "2016-01-12"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.1.expiration.0.days", "0"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.1.expiration.0.expired_object_delete_marker", "false"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.2.id", "id3"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.2.prefix", "path3/"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lifecycle_rule.2.transition.*", map[string]string{
						"days": "0",
					}),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.3.id", "id4"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.3.prefix", "path4/"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.3.tags.tagKey", "tagValue"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.3.tags.terraform", "hashicorp"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.4.id", "id5"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.4.tags.tagKey", "tagValue"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.4.tags.terraform", "hashicorp"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lifecycle_rule.4.transition.*", map[string]string{
						"days":          "0",
						"storage_class": "GLACIER",
					}),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.5.id", "id6"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.5.tags.tagKey", "tagValue"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lifecycle_rule.5.transition.*", map[string]string{
						"days":          "0",
						"storage_class": "GLACIER",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
			{
				Config: testAccAWSS3BucketConfigWithVersioningLifecycle(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.id", "id1"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.prefix", "path1/"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.noncurrent_version_expiration.0.days", "365"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lifecycle_rule.0.noncurrent_version_transition.*", map[string]string{
						"days":          "30",
						"storage_class": "STANDARD_IA",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lifecycle_rule.0.noncurrent_version_transition.*", map[string]string{
						"days":          "60",
						"storage_class": "GLACIER",
					}),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.1.id", "id2"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.1.prefix", "path2/"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.1.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.1.noncurrent_version_expiration.0.days", "365"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.2.id", "id3"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.2.prefix", "path3/"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lifecycle_rule.2.noncurrent_version_transition.*", map[string]string{
						"days":          "0",
						"storage_class": "GLACIER",
					}),
				),
			},
			{
				Config: testAccAWSS3BucketConfig_Basic(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
				),
			},
		},
	})
}

func TestAccAWSS3Bucket_LifecycleExpireMarkerOnly(t *testing.T) {
	bucketName := acctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigWithLifecycleExpireMarker(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.id", "id1"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.prefix", "path1/"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.expiration.0.days", "0"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.expiration.0.date", ""),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.expiration.0.expired_object_delete_marker", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
			{
				Config: testAccAWSS3BucketConfig_Basic(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/11420
func TestAccAWSS3Bucket_LifecycleRule_Expiration_EmptyConfigurationBlock(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigLifecycleRuleExpirationEmptyConfigurationBlock(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/15138
func TestAccAWSS3Bucket_LifecycleRule_AbortIncompleteMultipartUploadDays_NoExpiration(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigLifecycleRuleAbortIncompleteMultipartUploadDays(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
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
			testAccMultipleRegionPreCheck(t, 2)
		},
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
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
				Config:                  testAccAWSS3BucketConfigReplication(rInt),
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
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

func TestAccAWSS3Bucket_Replication_MultipleDestinations_EmptyFilter(t *testing.T) {
	rInt := acctest.RandInt()
	alternateRegion := testAccGetAlternateRegion()
	region := testAccGetRegion()
	resourceName := "aws_s3_bucket.bucket"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccMultipleRegionPreCheck(t, 2)
		},
		ErrorCheck:        testAccErrorCheckSkipS3(t),
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckWithProviders(testAccCheckAWSS3BucketDestroyWithProvider, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigReplicationWithMultipleDestinationsEmptyFilter(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExistsWithProvider(resourceName, testAccAwsRegionProviderFunc(region, &providers)),
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.destination", testAccAwsRegionProviderFunc(alternateRegion, &providers)),
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.destination2", testAccAwsRegionProviderFunc(alternateRegion, &providers)),
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.destination3", testAccAwsRegionProviderFunc(alternateRegion, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replication_configuration.0.rules.*", map[string]string{
						"id":                          "rule1",
						"priority":                    "1",
						"status":                      "Enabled",
						"filter.#":                    "1",
						"filter.0.prefix":             "",
						"destination.#":               "1",
						"destination.0.storage_class": "STANDARD",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replication_configuration.0.rules.*", map[string]string{
						"id":                          "rule2",
						"priority":                    "2",
						"status":                      "Enabled",
						"filter.#":                    "1",
						"filter.0.prefix":             "",
						"destination.#":               "1",
						"destination.0.storage_class": "STANDARD_IA",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replication_configuration.0.rules.*", map[string]string{
						"id":                          "rule3",
						"priority":                    "3",
						"status":                      "Disabled",
						"filter.#":                    "1",
						"filter.0.prefix":             "",
						"destination.#":               "1",
						"destination.0.storage_class": "ONEZONE_IA",
					}),
				),
			},
			{
				Config:                  testAccAWSS3BucketConfigReplicationWithMultipleDestinationsEmptyFilter(rInt),
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
		},
	})
}

func TestAccAWSS3Bucket_Replication_MultipleDestinations_NonEmptyFilter(t *testing.T) {
	rInt := acctest.RandInt()
	alternateRegion := testAccGetAlternateRegion()
	region := testAccGetRegion()
	resourceName := "aws_s3_bucket.bucket"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccMultipleRegionPreCheck(t, 2)
		},
		ErrorCheck:        testAccErrorCheckSkipS3(t),
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckWithProviders(testAccCheckAWSS3BucketDestroyWithProvider, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigReplicationWithMultipleDestinationsNonEmptyFilter(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExistsWithProvider(resourceName, testAccAwsRegionProviderFunc(region, &providers)),
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.destination", testAccAwsRegionProviderFunc(alternateRegion, &providers)),
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.destination2", testAccAwsRegionProviderFunc(alternateRegion, &providers)),
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.destination3", testAccAwsRegionProviderFunc(alternateRegion, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replication_configuration.0.rules.*", map[string]string{
						"id":                          "rule1",
						"priority":                    "1",
						"status":                      "Enabled",
						"filter.#":                    "1",
						"filter.0.prefix":             "prefix1",
						"destination.#":               "1",
						"destination.0.storage_class": "STANDARD",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replication_configuration.0.rules.*", map[string]string{
						"id":                          "rule2",
						"priority":                    "2",
						"status":                      "Enabled",
						"filter.#":                    "1",
						"filter.0.tags.%":             "1",
						"filter.0.tags.Key2":          "Value2",
						"destination.#":               "1",
						"destination.0.storage_class": "STANDARD_IA",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replication_configuration.0.rules.*", map[string]string{
						"id":                          "rule3",
						"priority":                    "3",
						"status":                      "Disabled",
						"filter.#":                    "1",
						"filter.0.prefix":             "prefix3",
						"filter.0.tags.%":             "1",
						"filter.0.tags.Key3":          "Value3",
						"destination.#":               "1",
						"destination.0.storage_class": "ONEZONE_IA",
					}),
				),
			},
			{
				Config:                  testAccAWSS3BucketConfigReplicationWithMultipleDestinationsNonEmptyFilter(rInt),
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
		},
	})
}

func TestAccAWSS3Bucket_Replication_MultipleDestinations_TwoDestination(t *testing.T) {
	// This tests 2 destinations since GovCloud and possibly other non-standard partitions allow a max of 2
	rInt := acctest.RandInt()
	alternateRegion := testAccGetAlternateRegion()
	region := testAccGetRegion()
	resourceName := "aws_s3_bucket.bucket"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccMultipleRegionPreCheck(t, 2)
		},
		ErrorCheck:        testAccErrorCheckSkipS3(t),
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckWithProviders(testAccCheckAWSS3BucketDestroyWithProvider, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigReplicationWithMultipleDestinationsTwoDestination(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExistsWithProvider(resourceName, testAccAwsRegionProviderFunc(region, &providers)),
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.destination", testAccAwsRegionProviderFunc(alternateRegion, &providers)),
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.destination2", testAccAwsRegionProviderFunc(alternateRegion, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replication_configuration.0.rules.*", map[string]string{
						"id":                          "rule1",
						"priority":                    "1",
						"status":                      "Enabled",
						"filter.#":                    "1",
						"filter.0.prefix":             "prefix1",
						"destination.#":               "1",
						"destination.0.storage_class": "STANDARD",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replication_configuration.0.rules.*", map[string]string{
						"id":                          "rule2",
						"priority":                    "2",
						"status":                      "Enabled",
						"filter.#":                    "1",
						"filter.0.tags.%":             "1",
						"filter.0.tags.Key2":          "Value2",
						"destination.#":               "1",
						"destination.0.storage_class": "STANDARD_IA",
					}),
				),
			},
			{
				Config:                  testAccAWSS3BucketConfigReplicationWithMultipleDestinationsTwoDestination(rInt),
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
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
			testAccMultipleRegionPreCheck(t, 2)
		},
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
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
				Config:                  testAccAWSS3BucketConfigReplicationWithAccessControlTranslation(rInt),
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl", "versioning"},
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

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/12480
func TestAccAWSS3Bucket_ReplicationConfiguration_Rule_Destination_AddAccessControlTranslation(t *testing.T) {
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
			testAccMultipleRegionPreCheck(t, 2)
		},
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckWithProviders(testAccCheckAWSS3BucketDestroyWithProvider, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigReplicationConfigurationRulesDestination(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExistsWithProvider(resourceName, testAccAwsRegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", "1"),
					testAccCheckAWSS3BucketReplicationRules(
						resourceName,
						[]*s3.ReplicationRule{
							{
								ID: aws.String("foobar"),
								Destination: &s3.Destination{
									Account:      aws.String("${data.aws_caller_identity.current.account_id}"),
									Bucket:       aws.String(fmt.Sprintf("arn:%s:s3:::tf-test-bucket-destination-%d", partition, rInt)),
									StorageClass: aws.String(s3.ObjectStorageClassStandard),
								},
								Prefix: aws.String("foo"),
								Status: aws.String(s3.ReplicationRuleStatusEnabled),
							},
						},
					),
				),
			},
			{
				Config:                  testAccAWSS3BucketConfigReplicationWithAccessControlTranslation(rInt),
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl", "versioning"},
			},
			{
				Config: testAccAWSS3BucketConfigReplicationWithAccessControlTranslation(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExistsWithProvider(resourceName, testAccAwsRegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", "1"),
					testAccCheckAWSS3BucketReplicationRules(
						resourceName,
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
			testAccMultipleRegionPreCheck(t, 2)
		},
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
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
				Config:                  testAccAWSS3BucketConfigReplicationWithoutStorageClass(rInt),
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
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
			testAccMultipleRegionPreCheck(t, 2)
		},
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckWithProviders(testAccCheckAWSS3BucketDestroyWithProvider, &providers),
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSS3BucketConfigReplicationNoVersioning(rInt),
				ExpectError: regexp.MustCompile(`versioning must be enabled to allow S3 bucket replication`),
			},
		},
	})
}

// Prefix issue: https://github.com/hashicorp/terraform-provider-aws/issues/6340
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
			testAccMultipleRegionPreCheck(t, 2)
		},
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
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
				Config:                  testAccAWSS3BucketConfigReplicationWithoutPrefix(rInt),
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
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
			testAccMultipleRegionPreCheck(t, 2)
		},
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
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
				Config:                  testAccAWSS3BucketConfigReplicationWithV2ConfigurationNoTags(rInt),
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
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

func TestAccAWSS3Bucket_SameRegionReplicationSchemaV2(t *testing.T) {
	resourceName := "aws_s3_bucket.bucket"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	destinationResourceName := "aws_s3_bucket.destination"
	rNameDestination := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigSameRegionReplicationWithV2ConfigurationNoTags(rName, rNameDestination),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					testAccCheckResourceAttrGlobalARN(resourceName, "replication_configuration.0.role", "iam", fmt.Sprintf("role/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", "1"),
					testAccCheckAWSS3BucketExists(destinationResourceName),
					testAccCheckAWSS3BucketReplicationRules(
						resourceName,
						[]*s3.ReplicationRule{
							{
								ID: aws.String("testid"),
								Destination: &s3.Destination{
									Bucket:       aws.String(fmt.Sprintf("arn:%s:s3:::%s", testAccGetPartition(), rNameDestination)),
									StorageClass: aws.String(s3.ObjectStorageClassStandard),
								},
								Status: aws.String(s3.ReplicationRuleStatusEnabled),
								Filter: &s3.ReplicationRuleFilter{
									Prefix: aws.String("testprefix"),
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
				Config:            testAccAWSS3BucketConfigSameRegionReplicationWithV2ConfigurationNoTags(rName, rNameDestination),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy", "acl"},
			},
		},
	})
}

func TestAccAWSS3Bucket_objectLock(t *testing.T) {
	bucketName := acctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.arbitrary"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketObjectLockEnabledNoDefaultRetention(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "object_lock_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_configuration.0.object_lock_enabled", "Enabled"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_configuration.0.rule.#", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
			{
				Config: testAccAWSS3BucketObjectLockEnabledWithDefaultRetention(bucketName),
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
	bucketName := acctest.RandomWithPrefix("tf-test-bucket")

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
	bucketName := acctest.RandomWithPrefix("tf-test-bucket")

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
	bucketName := acctest.RandomWithPrefix("tf-test-bucket")

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
		if err := validateS3BucketName(v, endpoints.UsWest2RegionID); err != nil {
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
		if err := validateS3BucketName(v, endpoints.UsWest2RegionID); err == nil {
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
		if err := validateS3BucketName(v, endpoints.UsEast1RegionID); err != nil {
			t.Fatalf("%q should be a valid S3 bucket name", v)
		}
	}

	invalidEastNames := []string{
		"foo;bar",
		strings.Repeat("x", 256),
	}

	for _, v := range invalidEastNames {
		if err := validateS3BucketName(v, endpoints.UsEast1RegionID); err == nil {
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
			Region:           endpoints.UsEast1RegionID,
			ExpectedErrCount: 0,
			ExpectedOutput:   bucket + ".s3.amazonaws.com",
		},
		{
			Region:           endpoints.UsWest2RegionID,
			ExpectedErrCount: 0,
			ExpectedOutput:   bucket + fmt.Sprintf(".s3.%s.%s", endpoints.UsWest2RegionID, testAccGetPartitionDNSSuffix()),
		},
		{
			Region:           endpoints.UsGovWest1RegionID,
			ExpectedErrCount: 0,
			ExpectedOutput:   bucket + fmt.Sprintf(".s3.%s.%s", endpoints.UsGovWest1RegionID, testAccGetPartitionDNSSuffix()),
		},
		{
			Region:           endpoints.CnNorth1RegionID,
			ExpectedErrCount: 0,
			ExpectedOutput:   bucket + fmt.Sprintf(".s3.%s.amazonaws.com.cn", endpoints.CnNorth1RegionID),
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
				region:    endpoints.UsEast1RegionID,
			},
			LocationConstraint: "",
			Expected:           fmt.Sprintf("bucket-name.s3-website-%s.%s", endpoints.UsEast1RegionID, testAccGetPartitionDNSSuffix()),
		},
		{
			AWSClient: &AWSClient{
				dnsSuffix: "amazonaws.com",
				region:    endpoints.UsWest2RegionID,
			},
			LocationConstraint: endpoints.UsWest2RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website-%s.%s", endpoints.UsWest2RegionID, testAccGetPartitionDNSSuffix()),
		},
		{
			AWSClient: &AWSClient{
				dnsSuffix: "amazonaws.com",
				region:    endpoints.UsWest1RegionID,
			},
			LocationConstraint: endpoints.UsWest1RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website-%s.%s", endpoints.UsWest1RegionID, testAccGetPartitionDNSSuffix()),
		},
		{
			AWSClient: &AWSClient{
				dnsSuffix: "amazonaws.com",
				region:    endpoints.EuWest1RegionID,
			},
			LocationConstraint: endpoints.EuWest1RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website-%s.%s", endpoints.EuWest1RegionID, testAccGetPartitionDNSSuffix()),
		},
		{
			AWSClient: &AWSClient{
				dnsSuffix: "amazonaws.com",
				region:    endpoints.EuWest3RegionID,
			},
			LocationConstraint: endpoints.EuWest3RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website.%s.%s", endpoints.EuWest3RegionID, testAccGetPartitionDNSSuffix()),
		},
		{
			AWSClient: &AWSClient{
				dnsSuffix: "amazonaws.com",
				region:    endpoints.EuCentral1RegionID,
			},
			LocationConstraint: endpoints.EuCentral1RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website.%s.%s", endpoints.EuCentral1RegionID, testAccGetPartitionDNSSuffix()),
		},
		{
			AWSClient: &AWSClient{
				dnsSuffix: "amazonaws.com",
				region:    endpoints.ApSouth1RegionID,
			},
			LocationConstraint: endpoints.ApSouth1RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website.%s.%s", endpoints.ApSouth1RegionID, testAccGetPartitionDNSSuffix()),
		},
		{
			AWSClient: &AWSClient{
				dnsSuffix: "amazonaws.com",
				region:    endpoints.ApSoutheast1RegionID,
			},
			LocationConstraint: endpoints.ApSoutheast1RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website-%s.%s", endpoints.ApSoutheast1RegionID, testAccGetPartitionDNSSuffix()),
		},
		{
			AWSClient: &AWSClient{
				dnsSuffix: "amazonaws.com",
				region:    endpoints.ApNortheast1RegionID,
			},
			LocationConstraint: endpoints.ApNortheast1RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website-%s.%s", endpoints.ApNortheast1RegionID, testAccGetPartitionDNSSuffix()),
		},
		{
			AWSClient: &AWSClient{
				dnsSuffix: "amazonaws.com",
				region:    endpoints.ApSoutheast2RegionID,
			},
			LocationConstraint: endpoints.ApSoutheast2RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website-%s.%s", endpoints.ApSoutheast2RegionID, testAccGetPartitionDNSSuffix()),
		},
		{
			AWSClient: &AWSClient{
				dnsSuffix: "amazonaws.com",
				region:    endpoints.ApNortheast2RegionID,
			},
			LocationConstraint: endpoints.ApNortheast2RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website.%s.%s", endpoints.ApNortheast2RegionID, testAccGetPartitionDNSSuffix()),
		},
		{
			AWSClient: &AWSClient{
				dnsSuffix: "amazonaws.com",
				region:    endpoints.SaEast1RegionID,
			},
			LocationConstraint: endpoints.SaEast1RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website-%s.%s", endpoints.SaEast1RegionID, testAccGetPartitionDNSSuffix()),
		},
		{
			AWSClient: &AWSClient{
				dnsSuffix: "amazonaws.com",
				region:    endpoints.UsGovEast1RegionID,
			},
			LocationConstraint: endpoints.UsGovEast1RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website.%s.%s", endpoints.UsGovEast1RegionID, testAccGetPartitionDNSSuffix()),
		},
		{
			AWSClient: &AWSClient{
				dnsSuffix: "amazonaws.com",
				region:    endpoints.UsGovWest1RegionID,
			},
			LocationConstraint: endpoints.UsGovWest1RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website-%s.%s", endpoints.UsGovWest1RegionID, testAccGetPartitionDNSSuffix()),
		},
		{
			AWSClient: &AWSClient{
				dnsSuffix: "c2s.ic.gov",
				region:    endpoints.UsIsoEast1RegionID,
			},
			LocationConstraint: endpoints.UsIsoEast1RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website.%s.c2s.ic.gov", endpoints.UsIsoEast1RegionID),
		},
		{
			AWSClient: &AWSClient{
				dnsSuffix: "sc2s.sgov.gov",
				region:    endpoints.UsIsobEast1RegionID,
			},
			LocationConstraint: endpoints.UsIsobEast1RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website.%s.sc2s.sgov.gov", endpoints.UsIsobEast1RegionID),
		},
		{
			AWSClient: &AWSClient{
				dnsSuffix: "amazonaws.com.cn",
				region:    endpoints.CnNorthwest1RegionID,
			},
			LocationConstraint: endpoints.CnNorthwest1RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website.%s.amazonaws.com.cn", endpoints.CnNorthwest1RegionID),
		},
		{
			AWSClient: &AWSClient{
				dnsSuffix: "amazonaws.com.cn",
				region:    endpoints.CnNorth1RegionID,
			},
			LocationConstraint: endpoints.CnNorth1RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website.%s.amazonaws.com.cn", endpoints.CnNorth1RegionID),
		},
	}

	for _, testCase := range testCases {
		got := WebsiteEndpoint(testCase.AWSClient, "bucket-name", testCase.LocationConstraint)
		if got.Endpoint != testCase.Expected {
			t.Errorf("WebsiteEndpointUrl(\"bucket-name\", %q) => %q, want %q", testCase.LocationConstraint, got.Endpoint, testCase.Expected)
		}
	}
}

// testAccErrorCheckSkipS3 skips tests that have error messages indicating unsupported features
func testAccErrorCheckSkipS3(t *testing.T) resource.ErrorCheckFunc {
	return testAccErrorCheckSkipMessagesContaining(t,
		"Number of distinct destination bucket ARNs cannot exceed",
	)
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
func testAccCheckAWSS3BucketCreateViaCloudFormation(n string, stackID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).cfconn
		stackName := acctest.RandomWithPrefix("tf-acc-test-s3tags")
		templateBody := fmt.Sprintf(`{
  "Resources": {
    "TfTestBucket": {
      "Type": "AWS::S3::Bucket",
      "Properties": {
        "BucketName": "%s"
      }
    }
  }
}`, n)

		requestToken := resource.UniqueId()
		req := &cloudformation.CreateStackInput{
			StackName:          aws.String(stackName),
			TemplateBody:       aws.String(templateBody),
			ClientRequestToken: aws.String(requestToken),
		}

		log.Printf("[DEBUG] Creating CloudFormation stack: %s", req)
		resp, err := conn.CreateStack(req)
		if err != nil {
			return fmt.Errorf("error creating CloudFormation stack: %w", err)
		}

		stack, err := waiter.StackCreated(conn, aws.StringValue(resp.StackId), requestToken, 10*time.Minute)
		if err != nil {
			return fmt.Errorf("Error waiting for CloudFormation stack creation: %w", err)
		}
		status := aws.StringValue(stack.StackStatus)
		if status != cloudformation.StackStatusCreateComplete {
			return fmt.Errorf("Invalid CloudFormation stack creation status: %s", status)
		}

		*stackID = aws.StringValue(resp.StackId)
		return nil
	}
}

func testAccCheckAWSS3BucketTagKeys(n string, keys ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		conn := testAccProvider.Meta().(*AWSClient).s3conn

		got, err := keyvaluetags.S3BucketListTags(conn, rs.Primary.Attributes["bucket"])
		if err != nil {
			return err
		}

		for _, want := range keys {
			ok := false
			for _, key := range got.Keys() {
				if want == key {
					ok = true
					break
				}
			}
			if !ok {
				return fmt.Errorf("Key %s not found in bucket's tag set", want)
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

func testAccCheckAWSS3BucketReplicationRules(n string, rules []*s3.ReplicationRule) resource.TestCheckFunc {
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

		conn := testAccProvider.Meta().(*AWSClient).s3conn
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

func testAccCheckS3BucketDomainName(resourceName string, attributeName string, bucketName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		expectedValue := testAccProvider.Meta().(*AWSClient).PartitionHostname(fmt.Sprintf("%s.s3", bucketName))

		return resource.TestCheckResourceAttr(resourceName, attributeName, expectedValue)(s)
	}
}

func testAccBucketRegionalDomainName(bucket, region string) string {
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

func testAccCheckAWSS3BucketUpdateTags(n string, oldTags, newTags map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		conn := testAccProvider.Meta().(*AWSClient).s3conn

		return keyvaluetags.S3BucketUpdateTags(conn, rs.Primary.Attributes["bucket"], oldTags, newTags)
	}
}

func testAccCheckAWSS3BucketCheckTags(n string, expectedTags map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		conn := testAccProvider.Meta().(*AWSClient).s3conn

		got, err := keyvaluetags.S3BucketListTags(conn, rs.Primary.Attributes["bucket"])
		if err != nil {
			return err
		}

		want := keyvaluetags.New(expectedTags)
		if !reflect.DeepEqual(want, got) {
			return fmt.Errorf("Incorrect tags, want: %v got: %v", want, got)
		}

		return nil
	}
}

func testAccAWSS3BucketPolicy(bucketName, partition string) string {
	return fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "s3:GetObject",
      "Resource": "arn:%[1]s:s3:::%[2]s/*"
    }
  ]
}`, partition, bucketName)
}

func testAccAWSS3BucketConfig_Basic(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
}
`, bucketName)
}

func testAccAWSS3BucketConfig_withNoTags(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket        = %[1]q
  acl           = "private"
  force_destroy = false
}
`, bucketName)
}

func testAccAWSS3BucketConfig_withTags(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket        = %[1]q
  acl           = "private"
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
  bucket        = %[1]q
  acl           = "private"
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

func testAccAWSS3BucketWebsiteConfig(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
  acl    = "public-read"

  website {
    index_document = "index.html"
  }
}
`, bucketName)
}

func testAccAWSS3BucketWebsiteConfigWithError(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
  acl    = "public-read"

  website {
    index_document = "index.html"
    error_document = "error.html"
  }
}
`, bucketName)
}

func testAccAWSS3BucketWebsiteConfigWithRedirect(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
  acl    = "public-read"

  website {
    redirect_all_requests_to = "hashicorp.com?my=query"
  }
}
`, bucketName)
}

func testAccAWSS3BucketWebsiteConfigWithHttpsRedirect(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
  acl    = "public-read"

  website {
    redirect_all_requests_to = "https://hashicorp.com?my=query"
  }
}
`, bucketName)
}

func testAccAWSS3BucketWebsiteConfigWithRoutingRules(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
  acl    = "public-read"

  website {
    index_document = "index.html"
    error_document = "error.html"

    routing_rules = <<EOF
[
  {
    "Condition": {
      "KeyPrefixEquals": "docs/"
    },
    "Redirect": {
      "ReplaceKeyPrefixWith": "documents/"
    }
  }
]
EOF

  }
}
`, bucketName)
}

func testAccAWSS3BucketConfigWithAcceleration(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket              = %[1]q
  acceleration_status = "Enabled"
}
`, bucketName)
}

func testAccAWSS3BucketConfigWithoutAcceleration(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket              = %[1]q
  acceleration_status = "Suspended"
}
`, bucketName)
}

func testAccAWSS3BucketConfigRequestPayerBucketOwner(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket        = %[1]q
  request_payer = "BucketOwner"
}
`, bucketName)
}

func testAccAWSS3BucketConfigRequestPayerRequester(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket        = %[1]q
  request_payer = "Requester"
}
`, bucketName)
}

func testAccAWSS3BucketConfigWithPolicy(bucketName, partition string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
  acl    = "public-read"
  policy = %[2]s
}
`, bucketName, strconv.Quote(testAccAWSS3BucketPolicy(bucketName, partition)))
}

func testAccAWSS3BucketDestroyedConfig(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
  acl    = "public-read"
}
`, bucketName)
}

func testAccAWSS3BucketEnableDefaultEncryption(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "arbitrary" {
  description             = "KMS Key for Bucket %[1]s"
  deletion_window_in_days = 10
}

resource "aws_s3_bucket" "arbitrary" {
  bucket = %[1]q

  server_side_encryption_configuration {
    rule {
      apply_server_side_encryption_by_default {
        kms_master_key_id = aws_kms_key.arbitrary.arn
        sse_algorithm     = "aws:kms"
      }
    }
  }
}
`, bucketName)
}

func testAccAWSS3BucketEnableDefaultEncryptionWithAES256(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "arbitrary" {
  bucket = %[1]q

  server_side_encryption_configuration {
    rule {
      apply_server_side_encryption_by_default {
        sse_algorithm = "AES256"
      }
    }
  }
}
`, bucketName)
}

func testAccAWSS3BucketEnableDefaultEncryptionWithDefaultKey(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "arbitrary" {
  bucket = %[1]q

  server_side_encryption_configuration {
    rule {
      apply_server_side_encryption_by_default {
        sse_algorithm = "aws:kms"
      }
    }
  }
}
`, bucketName)
}

func testAccAWSS3BucketDisableDefaultEncryption(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "arbitrary" {
  bucket = %[1]q
}
`, bucketName)
}

func testAccAWSS3BucketConfigWithEmptyPolicy(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
  acl    = "public-read"
  policy = ""
}
`, bucketName)
}

func testAccAWSS3BucketConfigWithVersioning(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q

  versioning {
    enabled = true
  }
}
`, bucketName)
}

func testAccAWSS3BucketConfigWithDisableVersioning(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q

  versioning {
    enabled = false
  }
}
`, bucketName)
}

func testAccAWSS3BucketConfigWithCORS(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["PUT", "POST"]
    allowed_origins = ["https://www.example.com"]
    expose_headers  = ["x-amz-server-side-encryption", "ETag"]
    max_age_seconds = 3000
  }
}
`, bucketName)
}

func testAccAWSS3BucketConfigWithCORSEmptyOrigin(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["PUT", "POST"]
    allowed_origins = [""]
    expose_headers  = ["x-amz-server-side-encryption", "ETag"]
    max_age_seconds = 3000
  }
}
`, bucketName)
}

func testAccAWSS3BucketConfigWithAcl(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
  acl    = "public-read"
}
`, bucketName)
}

func testAccAWSS3BucketConfigWithAclUpdate(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
  acl    = "private"
}
`, bucketName)
}

func testAccAWSS3BucketConfigWithGrants(bucketName string) string {
	return fmt.Sprintf(`
data "aws_canonical_user_id" "current" {}

resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q

  grant {
    id          = data.aws_canonical_user_id.current.id
    type        = "CanonicalUser"
    permissions = ["FULL_CONTROL", "WRITE"]
  }
}
`, bucketName)
}

func testAccAWSS3BucketConfigWithGrantsUpdate(bucketName string) string {
	return fmt.Sprintf(`
data "aws_canonical_user_id" "current" {}

resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q

  grant {
    id          = data.aws_canonical_user_id.current.id
    type        = "CanonicalUser"
    permissions = ["READ"]
  }

  grant {
    type        = "Group"
    permissions = ["READ_ACP"]
    uri         = "http://acs.amazonaws.com/groups/s3/LogDelivery"
  }
}
`, bucketName)
}

func testAccAWSS3BucketConfigWithLogging(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "log_bucket" {
  bucket = "%[1]s-log"
  acl    = "log-delivery-write"
}

resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
  acl    = "private"

  logging {
    target_bucket = aws_s3_bucket.log_bucket.id
    target_prefix = "log/"
  }
}
`, bucketName)
}

func testAccAWSS3BucketConfigWithLifecycle(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
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
    id      = "id5"
    enabled = true

    tags = {
      "tagKey"    = "tagValue"
      "terraform" = "hashicorp"
    }

    transition {
      days          = 0
      storage_class = "GLACIER"
    }
  }

  lifecycle_rule {
    id      = "id6"
    enabled = true

    tags = {
      "tagKey" = "tagValue"
    }

    transition {
      days          = 0
      storage_class = "GLACIER"
    }
  }
}
`, bucketName)
}

func testAccAWSS3BucketConfigWithLifecycleExpireMarker(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
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
`, bucketName)
}

func testAccAWSS3BucketConfigWithVersioningLifecycle(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
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
`, bucketName)
}

func testAccAWSS3BucketConfigLifecycleRuleExpirationEmptyConfigurationBlock(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q

  lifecycle_rule {
    enabled = true
    id      = "id1"

    expiration {}
  }
}
`, rName)
}

func testAccAWSS3BucketConfigLifecycleRuleAbortIncompleteMultipartUploadDays(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q

  lifecycle_rule {
    abort_incomplete_multipart_upload_days = 7
    enabled                                = true
    id                                     = "id1"
  }
}
`, rName)
}

func testAccAWSS3BucketConfigReplicationBasic(randInt int) string {
	return testAccAlternateRegionProviderConfig() + fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "role" {
  name = "tf-iam-role-replication-%[1]d"

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
  provider = "awsalternate"
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
  bucket = "tf-test-bucket-%[1]d"
  acl    = "private"

  versioning {
    enabled = true
  }
}
`, randInt)
}

func testAccAWSS3BucketConfigReplicationWithConfiguration(randInt int, storageClass string) string {
	return testAccAWSS3BucketConfigReplicationBasic(randInt) + fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-%[1]d"
  acl    = "private"

  versioning {
    enabled = true
  }

  replication_configuration {
    role = aws_iam_role.role.arn

    rules {
      id     = "foobar"
      prefix = "foo"
      status = "Enabled"

      destination {
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "%[2]s"
      }
    }
  }
}
`, randInt, storageClass)
}

func testAccAWSS3BucketConfigReplicationWithMultipleDestinationsEmptyFilter(randInt int) string {
	return composeConfig(
		testAccAWSS3BucketConfigReplicationBasic(randInt),
		fmt.Sprintf(`
resource "aws_s3_bucket" "destination2" {
  provider = "awsalternate"
  bucket   = "tf-test-bucket-destination2-%[1]d"

  versioning {
    enabled = true
  }
}

resource "aws_s3_bucket" "destination3" {
  provider = "awsalternate"
  bucket   = "tf-test-bucket-destination3-%[1]d"

  versioning {
    enabled = true
  }
}

resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-%[1]d"
  acl    = "private"

  versioning {
    enabled = true
  }

  replication_configuration {
    role = aws_iam_role.role.arn

    rules {
      id       = "rule1"
      priority = 1
      status   = "Enabled"

      filter {}

      destination {
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"
      }
    }

    rules {
      id       = "rule2"
      priority = 2
      status   = "Enabled"

      filter {}

      destination {
        bucket        = aws_s3_bucket.destination2.arn
        storage_class = "STANDARD_IA"
      }
    }

    rules {
      id       = "rule3"
      priority = 3
      status   = "Disabled"

      filter {}

      destination {
        bucket        = aws_s3_bucket.destination3.arn
        storage_class = "ONEZONE_IA"
      }
    }
  }
}
`, randInt))
}

func testAccAWSS3BucketConfigReplicationWithMultipleDestinationsNonEmptyFilter(randInt int) string {
	return composeConfig(
		testAccAWSS3BucketConfigReplicationBasic(randInt),
		fmt.Sprintf(`
resource "aws_s3_bucket" "destination2" {
  provider = "awsalternate"
  bucket   = "tf-test-bucket-destination2-%[1]d"

  versioning {
    enabled = true
  }
}

resource "aws_s3_bucket" "destination3" {
  provider = "awsalternate"
  bucket   = "tf-test-bucket-destination3-%[1]d"

  versioning {
    enabled = true
  }
}

resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-%[1]d"
  acl    = "private"

  versioning {
    enabled = true
  }

  replication_configuration {
    role = aws_iam_role.role.arn

    rules {
      id       = "rule1"
      priority = 1
      status   = "Enabled"

      filter {
        prefix = "prefix1"
      }

      destination {
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"
      }
    }

    rules {
      id       = "rule2"
      priority = 2
      status   = "Enabled"

      filter {
        tags = {
          Key2 = "Value2"
        }
      }

      destination {
        bucket        = aws_s3_bucket.destination2.arn
        storage_class = "STANDARD_IA"
      }
    }

    rules {
      id       = "rule3"
      priority = 3
      status   = "Disabled"

      filter {
        prefix = "prefix3"

        tags = {
          Key3 = "Value3"
        }
      }

      destination {
        bucket        = aws_s3_bucket.destination3.arn
        storage_class = "ONEZONE_IA"
      }
    }
  }
}
`, randInt))
}

func testAccAWSS3BucketConfigReplicationWithMultipleDestinationsTwoDestination(randInt int) string {
	return composeConfig(
		testAccAWSS3BucketConfigReplicationBasic(randInt),
		fmt.Sprintf(`
resource "aws_s3_bucket" "destination2" {
  provider = "awsalternate"
  bucket   = "tf-test-bucket-destination2-%[1]d"

  versioning {
    enabled = true
  }
}

resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-%[1]d"
  acl    = "private"

  versioning {
    enabled = true
  }

  replication_configuration {
    role = aws_iam_role.role.arn

    rules {
      id       = "rule1"
      priority = 1
      status   = "Enabled"

      filter {
        prefix = "prefix1"
      }

      destination {
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"
      }
    }

    rules {
      id       = "rule2"
      priority = 2
      status   = "Enabled"

      filter {
        tags = {
          Key2 = "Value2"
        }
      }

      destination {
        bucket        = aws_s3_bucket.destination2.arn
        storage_class = "STANDARD_IA"
      }
    }
  }
}
`, randInt))
}

func testAccAWSS3BucketConfigReplicationWithSseKmsEncryptedObjects(randInt int) string {
	return testAccAWSS3BucketConfigReplicationBasic(randInt) + fmt.Sprintf(`
resource "aws_kms_key" "replica" {
  provider                = "awsalternate"
  description             = "TF Acceptance Test S3 repl KMS key"
  deletion_window_in_days = 7
}

resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-%[1]d"
  acl    = "private"

  versioning {
    enabled = true
  }

  replication_configuration {
    role = aws_iam_role.role.arn

    rules {
      id     = "foobar"
      prefix = "foo"
      status = "Enabled"

      destination {
        bucket             = aws_s3_bucket.destination.arn
        storage_class      = "STANDARD"
        replica_kms_key_id = aws_kms_key.replica.arn
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
  bucket = "tf-test-bucket-%[1]d"
  acl    = "private"

  versioning {
    enabled = true
  }

  replication_configuration {
    role = aws_iam_role.role.arn

    rules {
      id     = "foobar"
      prefix = "foo"
      status = "Enabled"

      destination {
        account_id    = data.aws_caller_identity.current.account_id
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"

        access_control_translation {
          owner = "Destination"
        }
      }
    }
  }
}
`, randInt)
}

func testAccAWSS3BucketConfigReplicationConfigurationRulesDestination(randInt int) string {
	return testAccAWSS3BucketConfigReplicationBasic(randInt) + fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_s3_bucket" "bucket" {
  acl    = "private"
  bucket = "tf-test-bucket-%[1]d"

  replication_configuration {
    role = aws_iam_role.role.arn

    rules {
      id     = "foobar"
      prefix = "foo"
      status = "Enabled"

      destination {
        account_id    = data.aws_caller_identity.current.account_id
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"
      }
    }
  }

  versioning {
    enabled = true
  }
}
`, randInt)
}

func testAccAWSS3BucketConfigReplicationWithSseKmsEncryptedObjectsAndAccessControlTranslation(randInt int) string {
	return testAccAWSS3BucketConfigReplicationBasic(randInt) + fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_kms_key" "replica" {
  provider                = "awsalternate"
  description             = "TF Acceptance Test S3 repl KMS key"
  deletion_window_in_days = 7
}

resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-%[1]d"
  acl    = "private"

  versioning {
    enabled = true
  }

  replication_configuration {
    role = aws_iam_role.role.arn

    rules {
      id     = "foobar"
      prefix = "foo"
      status = "Enabled"

      destination {
        account_id         = data.aws_caller_identity.current.account_id
        bucket             = aws_s3_bucket.destination.arn
        storage_class      = "STANDARD"
        replica_kms_key_id = aws_kms_key.replica.arn

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
  bucket = "tf-test-bucket-%[1]d"
  acl    = "private"

  versioning {
    enabled = true
  }

  replication_configuration {
    role = aws_iam_role.role.arn

    rules {
      id     = "foobar"
      prefix = "foo"
      status = "Enabled"

      destination {
        bucket = aws_s3_bucket.destination.arn
      }
    }
  }
}
`, randInt)
}

func testAccAWSS3BucketConfigReplicationWithoutPrefix(randInt int) string {
	return testAccAWSS3BucketConfigReplicationBasic(randInt) + fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-%[1]d"
  acl    = "private"

  versioning {
    enabled = true
  }

  replication_configuration {
    role = aws_iam_role.role.arn

    rules {
      id     = "foobar"
      status = "Enabled"

      destination {
        bucket        = aws_s3_bucket.destination.arn
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
  bucket = "tf-test-bucket-%[1]d"
  acl    = "private"

  replication_configuration {
    role = aws_iam_role.role.arn

    rules {
      id     = "foobar"
      prefix = "foo"
      status = "Enabled"

      destination {
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"
      }
    }
  }
}
`, randInt)
}

func testAccAWSS3BucketConfigSameRegionReplicationWithV2ConfigurationNoTags(rName, rNameDestination string) string {
	return composeConfig(testAccAWSS3BucketReplicationConfig_iamPolicy(rName), fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
  acl    = "private"

  versioning {
    enabled = true
  }

  replication_configuration {
    role = aws_iam_role.test.arn

    rules {
      id     = "testid"
      status = "Enabled"

      filter {
        prefix = "testprefix"
      }

      destination {
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"
      }
    }
  }
}

resource "aws_s3_bucket" "destination" {
  bucket = %[2]q

  versioning {
    enabled = true
  }
}
`, rName, rNameDestination))
}

func testAccAWSS3BucketConfigReplicationWithV2ConfigurationNoTags(randInt int) string {
	return testAccAWSS3BucketConfigReplicationBasic(randInt) + fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-%[1]d"
  acl    = "private"

  versioning {
    enabled = true
  }

  replication_configuration {
    role = aws_iam_role.role.arn

    rules {
      id     = "foobar"
      status = "Enabled"

      filter {
        prefix = "foo"
      }

      destination {
        bucket        = aws_s3_bucket.destination.arn
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
  bucket = "tf-test-bucket-%[1]d"
  acl    = "private"

  versioning {
    enabled = true
  }

  replication_configuration {
    role = aws_iam_role.role.arn

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
        bucket        = aws_s3_bucket.destination.arn
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
  bucket = "tf-test-bucket-%[1]d"
  acl    = "private"

  versioning {
    enabled = true
  }

  replication_configuration {
    role = aws_iam_role.role.arn

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
        bucket        = aws_s3_bucket.destination.arn
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
  bucket = "tf-test-bucket-%[1]d"
  acl    = "private"

  versioning {
    enabled = true
  }

  replication_configuration {
    role = aws_iam_role.role.arn

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
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"
      }
    }
  }
}
`, randInt)
}

func testAccAWSS3BucketObjectLockEnabledNoDefaultRetention(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "arbitrary" {
  bucket = %[1]q

  object_lock_configuration {
    object_lock_enabled = "Enabled"
  }
}
`, bucketName)
}

func testAccAWSS3BucketObjectLockEnabledWithDefaultRetention(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "arbitrary" {
  bucket = %[1]q

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
`, bucketName)
}

func testAccAWSS3BucketConfig_forceDestroy(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket        = "%s"
  acl           = "private"
  force_destroy = true
}
`, bucketName)
}

func testAccAWSS3BucketConfig_forceDestroyWithObjectLockEnabled(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket        = "%s"
  acl           = "private"
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

func testAccAWSS3BucketReplicationConfig_iamPolicy(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "s3.amazonaws.com"
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
