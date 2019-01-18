package aws

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform/helper/schema"
)

func TestAccAWSS3Bucket_importBasic(t *testing.T) {
	resourceName := "aws_s3_bucket.bucket"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfig(rInt),
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

func TestAccAWSS3Bucket_importWithPolicy(t *testing.T) {
	rInt := acctest.RandInt()
	partition := testAccGetPartition()

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
			},

			{
				ResourceName:     "aws_s3_bucket.bucket",
				ImportState:      true,
				ImportStateCheck: checkFn,
			},
		},
	})
}

func TestAccAWSS3Bucket_basic(t *testing.T) {
	rInt := acctest.RandInt()
	arnRegexp := regexp.MustCompile(`^arn:aws[\w-]*:s3:::`)
	region := testAccGetRegion()
	hostedZoneID, _ := HostedZoneIDForRegion(region)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		/*
			IDRefreshName:   "aws_s3_bucket.bucket",
			IDRefreshIgnore: []string{"force_destroy"},
		*/
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists("aws_s3_bucket.bucket"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "hosted_zone_id", hostedZoneID),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "region", region),
					resource.TestCheckNoResourceAttr(
						"aws_s3_bucket.bucket", "website_endpoint"),
					resource.TestMatchResourceAttr(
						"aws_s3_bucket.bucket", "arn", arnRegexp),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "bucket", testAccBucketName(rInt)),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "bucket_domain_name", testAccBucketDomainName(rInt)),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "bucket_regional_domain_name", testAccBucketRegionalDomainName(rInt, region)),
				),
			},
		},
	})
}

func TestAccAWSS3MultiBucket_withTags(t *testing.T) {
	rInt := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3MultiBucketConfigWithTags(rInt),
			},
		},
	})
}

func TestAccAWSS3Bucket_namePrefix(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfig_namePrefix,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists("aws_s3_bucket.test"),
					resource.TestMatchResourceAttr(
						"aws_s3_bucket.test", "bucket", regexp.MustCompile("^tf-test-")),
				),
			},
		},
	})
}

func TestAccAWSS3Bucket_generatedName(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfig_generatedName,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists("aws_s3_bucket.test"),
				),
			},
		},
	})
}

func TestAccAWSS3Bucket_region(t *testing.T) {
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccMultipleRegionsPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigWithRegion(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists("aws_s3_bucket.bucket"),
					resource.TestCheckResourceAttr("aws_s3_bucket.bucket", "region", "eu-west-1"),
				),
			},
		},
	})
}

func TestAccAWSS3Bucket_acceleration(t *testing.T) {
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccHasServicePreCheck("cloudfront", t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigWithAcceleration(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists("aws_s3_bucket.bucket"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "acceleration_status", "Enabled"),
				),
			},
			{
				Config: testAccAWSS3BucketConfigWithoutAcceleration(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists("aws_s3_bucket.bucket"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "acceleration_status", "Suspended"),
				),
			},
		},
	})
}

func TestAccAWSS3Bucket_RequestPayer(t *testing.T) {
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigRequestPayerBucketOwner(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists("aws_s3_bucket.bucket"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket",
						"request_payer",
						"BucketOwner"),
					testAccCheckAWSS3RequestPayer(
						"aws_s3_bucket.bucket",
						"BucketOwner"),
				),
			},
			{
				Config: testAccAWSS3BucketConfigRequestPayerRequester(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists("aws_s3_bucket.bucket"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket",
						"request_payer",
						"Requester"),
					testAccCheckAWSS3RequestPayer(
						"aws_s3_bucket.bucket",
						"Requester"),
				),
			},
		},
	})
}

func TestAccAWSS3Bucket_Policy(t *testing.T) {
	rInt := acctest.RandInt()
	partition := testAccGetPartition()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigWithPolicy(rInt, partition),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists("aws_s3_bucket.bucket"),
					testAccCheckAWSS3BucketPolicy(
						"aws_s3_bucket.bucket", testAccAWSS3BucketPolicy(rInt, partition)),
				),
			},
			{
				Config: testAccAWSS3BucketConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists("aws_s3_bucket.bucket"),
					testAccCheckAWSS3BucketPolicy(
						"aws_s3_bucket.bucket", ""),
				),
			},
			{
				Config: testAccAWSS3BucketConfigWithEmptyPolicy(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists("aws_s3_bucket.bucket"),
					testAccCheckAWSS3BucketPolicy(
						"aws_s3_bucket.bucket", ""),
				),
			},
		},
	})
}

func TestAccAWSS3Bucket_UpdateAcl(t *testing.T) {
	ri := acctest.RandInt()
	preConfig := fmt.Sprintf(testAccAWSS3BucketConfigWithAcl, ri)
	postConfig := fmt.Sprintf(testAccAWSS3BucketConfigWithAclUpdate, ri)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists("aws_s3_bucket.bucket"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "acl", "public-read"),
				),
			},
			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists("aws_s3_bucket.bucket"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "acl", "private"),
				),
			},
		},
	})
}

func TestAccAWSS3Bucket_Website_Simple(t *testing.T) {
	rInt := acctest.RandInt()
	region := testAccGetRegion()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketWebsiteConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists("aws_s3_bucket.bucket"),
					testAccCheckAWSS3BucketWebsite(
						"aws_s3_bucket.bucket", "index.html", "", "", ""),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "website_endpoint", testAccWebsiteEndpoint(rInt, region)),
				),
			},
			{
				Config: testAccAWSS3BucketWebsiteConfigWithError(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists("aws_s3_bucket.bucket"),
					testAccCheckAWSS3BucketWebsite(
						"aws_s3_bucket.bucket", "index.html", "error.html", "", ""),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "website_endpoint", testAccWebsiteEndpoint(rInt, region)),
				),
			},
			{
				Config: testAccAWSS3BucketConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists("aws_s3_bucket.bucket"),
					testAccCheckAWSS3BucketWebsite(
						"aws_s3_bucket.bucket", "", "", "", ""),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "website_endpoint", ""),
				),
			},
		},
	})
}

func TestAccAWSS3Bucket_WebsiteRedirect(t *testing.T) {
	rInt := acctest.RandInt()
	region := testAccGetRegion()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketWebsiteConfigWithRedirect(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists("aws_s3_bucket.bucket"),
					testAccCheckAWSS3BucketWebsite(
						"aws_s3_bucket.bucket", "", "", "", "hashicorp.com?my=query"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "website_endpoint", testAccWebsiteEndpoint(rInt, region)),
				),
			},
			{
				Config: testAccAWSS3BucketWebsiteConfigWithHttpsRedirect(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists("aws_s3_bucket.bucket"),
					testAccCheckAWSS3BucketWebsite(
						"aws_s3_bucket.bucket", "", "", "https", "hashicorp.com?my=query"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "website_endpoint", testAccWebsiteEndpoint(rInt, region)),
				),
			},
			{
				Config: testAccAWSS3BucketConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists("aws_s3_bucket.bucket"),
					testAccCheckAWSS3BucketWebsite(
						"aws_s3_bucket.bucket", "", "", "", ""),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "website_endpoint", ""),
				),
			},
		},
	})
}

func TestAccAWSS3Bucket_WebsiteRoutingRules(t *testing.T) {
	rInt := acctest.RandInt()
	region := testAccGetRegion()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketWebsiteConfigWithRoutingRules(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists("aws_s3_bucket.bucket"),
					testAccCheckAWSS3BucketWebsite(
						"aws_s3_bucket.bucket", "index.html", "error.html", "", ""),
					testAccCheckAWSS3BucketWebsiteRoutingRules(
						"aws_s3_bucket.bucket",
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
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "website_endpoint", testAccWebsiteEndpoint(rInt, region)),
				),
			},
			{
				Config: testAccAWSS3BucketConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists("aws_s3_bucket.bucket"),
					testAccCheckAWSS3BucketWebsite(
						"aws_s3_bucket.bucket", "", "", "", ""),
					testAccCheckAWSS3BucketWebsiteRoutingRules("aws_s3_bucket.bucket", nil),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "website_endpoint", ""),
				),
			},
		},
	})
}

func TestAccAWSS3Bucket_enableDefaultEncryption_whenTypical(t *testing.T) {
	rInt := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketEnableDefaultEncryption(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists("aws_s3_bucket.arbitrary"),
					resource.TestCheckResourceAttr("aws_s3_bucket.arbitrary", "server_side_encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr("aws_s3_bucket.arbitrary", "server_side_encryption_configuration.0.rule.#", "1"),
					resource.TestCheckResourceAttr("aws_s3_bucket.arbitrary", "server_side_encryption_configuration.0.rule.0.apply_server_side_encryption_by_default.#", "1"),
					resource.TestCheckResourceAttr("aws_s3_bucket.arbitrary", "server_side_encryption_configuration.0.rule.0.apply_server_side_encryption_by_default.0.sse_algorithm", "aws:kms"),
					resource.TestMatchResourceAttr("aws_s3_bucket.arbitrary", "server_side_encryption_configuration.0.rule.0.apply_server_side_encryption_by_default.0.kms_master_key_id", regexp.MustCompile("^arn")),
				),
			},
		},
	})
}

func TestAccAWSS3Bucket_enableDefaultEncryption_whenAES256IsUsed(t *testing.T) {
	rInt := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketEnableDefaultEncryptionWithAES256(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists("aws_s3_bucket.arbitrary"),
					resource.TestCheckResourceAttr("aws_s3_bucket.arbitrary", "server_side_encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr("aws_s3_bucket.arbitrary", "server_side_encryption_configuration.0.rule.#", "1"),
					resource.TestCheckResourceAttr("aws_s3_bucket.arbitrary", "server_side_encryption_configuration.0.rule.0.apply_server_side_encryption_by_default.#", "1"),
					resource.TestCheckResourceAttr("aws_s3_bucket.arbitrary", "server_side_encryption_configuration.0.rule.0.apply_server_side_encryption_by_default.0.sse_algorithm", "AES256"),
					resource.TestCheckResourceAttr("aws_s3_bucket.arbitrary", "server_side_encryption_configuration.0.rule.0.apply_server_side_encryption_by_default.0.kms_master_key_id", ""),
				),
			},
		},
	})
}

func TestAccAWSS3Bucket_disableDefaultEncryption_whenDefaultEncryptionIsEnabled(t *testing.T) {
	rInt := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketEnableDefaultEncryptionWithDefaultKey(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists("aws_s3_bucket.arbitrary"),
				),
			},
			{
				Config: testAccAWSS3BucketDisableDefaultEncryption(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists("aws_s3_bucket.arbitrary"),
					resource.TestCheckResourceAttr("aws_s3_bucket.arbitrary", "server_side_encryption_configuration.#", "0"),
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
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketDestroyedConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists("aws_s3_bucket.bucket"),
					testAccCheckAWSS3DestroyBucket("aws_s3_bucket.bucket"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSS3Bucket_Versioning(t *testing.T) {
	rInt := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists("aws_s3_bucket.bucket"),
					testAccCheckAWSS3BucketVersioning(
						"aws_s3_bucket.bucket", ""),
				),
			},
			{
				Config: testAccAWSS3BucketConfigWithVersioning(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists("aws_s3_bucket.bucket"),
					testAccCheckAWSS3BucketVersioning(
						"aws_s3_bucket.bucket", s3.BucketVersioningStatusEnabled),
				),
			},
			{
				Config: testAccAWSS3BucketConfigWithDisableVersioning(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists("aws_s3_bucket.bucket"),
					testAccCheckAWSS3BucketVersioning(
						"aws_s3_bucket.bucket", s3.BucketVersioningStatusSuspended),
				),
			},
		},
	})
}

func TestAccAWSS3Bucket_Cors_Update(t *testing.T) {
	rInt := acctest.RandInt()

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
					testAccCheckAWSS3BucketExists("aws_s3_bucket.bucket"),
					testAccCheckAWSS3BucketCors(
						"aws_s3_bucket.bucket",
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
					updateBucketCors("aws_s3_bucket.bucket"),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccAWSS3BucketConfigWithCORS(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists("aws_s3_bucket.bucket"),
					testAccCheckAWSS3BucketCors(
						"aws_s3_bucket.bucket",
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
					testAccCheckAWSS3BucketExists("aws_s3_bucket.bucket"),
					deleteBucketCors("aws_s3_bucket.bucket"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSS3Bucket_Cors_EmptyOrigin(t *testing.T) {
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigWithCORSEmptyOrigin(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists("aws_s3_bucket.bucket"),
					testAccCheckAWSS3BucketCors(
						"aws_s3_bucket.bucket",
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
		},
	})
}

func TestAccAWSS3Bucket_Logging(t *testing.T) {
	rInt := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigWithLogging(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists("aws_s3_bucket.bucket"),
					testAccCheckAWSS3BucketLogging(
						"aws_s3_bucket.bucket", "aws_s3_bucket.log_bucket", "log/"),
				),
			},
		},
	})
}

func TestAccAWSS3Bucket_Lifecycle(t *testing.T) {
	rInt := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigWithLifecycle(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists("aws_s3_bucket.bucket"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "lifecycle_rule.0.id", "id1"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "lifecycle_rule.0.prefix", "path1/"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "lifecycle_rule.0.expiration.2613713285.days", "365"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "lifecycle_rule.0.expiration.2613713285.date", ""),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "lifecycle_rule.0.expiration.2613713285.expired_object_delete_marker", "false"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "lifecycle_rule.0.transition.2000431762.date", ""),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "lifecycle_rule.0.transition.2000431762.days", "30"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "lifecycle_rule.0.transition.2000431762.storage_class", "STANDARD_IA"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "lifecycle_rule.0.transition.3601168188.date", ""),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "lifecycle_rule.0.transition.3601168188.days", "60"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "lifecycle_rule.0.transition.3601168188.storage_class", "INTELLIGENT_TIERING"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "lifecycle_rule.0.transition.3854926587.date", ""),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "lifecycle_rule.0.transition.3854926587.days", "90"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "lifecycle_rule.0.transition.3854926587.storage_class", "ONEZONE_IA"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "lifecycle_rule.0.transition.962205413.date", ""),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "lifecycle_rule.0.transition.962205413.days", "120"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "lifecycle_rule.0.transition.962205413.storage_class", "GLACIER"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "lifecycle_rule.1.id", "id2"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "lifecycle_rule.1.prefix", "path2/"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "lifecycle_rule.1.expiration.2855832418.date", "2016-01-12"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "lifecycle_rule.1.expiration.2855832418.days", "0"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "lifecycle_rule.1.expiration.2855832418.expired_object_delete_marker", "false"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "lifecycle_rule.2.id", "id3"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "lifecycle_rule.2.prefix", "path3/"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "lifecycle_rule.2.transition.460947558.days", "0"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "lifecycle_rule.3.id", "id4"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "lifecycle_rule.3.prefix", "path4/"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "lifecycle_rule.3.tags.tagKey", "tagValue"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "lifecycle_rule.3.tags.terraform", "hashicorp"),
				),
			},
			{
				Config: testAccAWSS3BucketConfigWithVersioningLifecycle(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists("aws_s3_bucket.bucket"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "lifecycle_rule.0.id", "id1"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "lifecycle_rule.0.prefix", "path1/"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "lifecycle_rule.0.enabled", "true"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "lifecycle_rule.0.noncurrent_version_expiration.80908210.days", "365"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "lifecycle_rule.0.noncurrent_version_transition.1377917700.days", "30"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "lifecycle_rule.0.noncurrent_version_transition.1377917700.storage_class", "STANDARD_IA"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "lifecycle_rule.0.noncurrent_version_transition.2528035817.days", "60"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "lifecycle_rule.0.noncurrent_version_transition.2528035817.storage_class", "GLACIER"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "lifecycle_rule.1.id", "id2"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "lifecycle_rule.1.prefix", "path2/"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "lifecycle_rule.1.enabled", "false"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "lifecycle_rule.1.noncurrent_version_expiration.80908210.days", "365"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "lifecycle_rule.2.id", "id3"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "lifecycle_rule.2.prefix", "path3/"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "lifecycle_rule.2.noncurrent_version_transition.3732708140.days", "0"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "lifecycle_rule.2.noncurrent_version_transition.3732708140.storage_class", "GLACIER"),
				),
			},
			{
				Config: testAccAWSS3BucketConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists("aws_s3_bucket.bucket"),
				),
			},
		},
	})
}

func TestAccAWSS3Bucket_LifecycleExpireMarkerOnly(t *testing.T) {
	rInt := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigWithLifecycleExpireMarker(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists("aws_s3_bucket.bucket"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "lifecycle_rule.0.id", "id1"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "lifecycle_rule.0.prefix", "path1/"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "lifecycle_rule.0.expiration.3591068768.days", "0"),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "lifecycle_rule.0.expiration.3591068768.date", ""),
					resource.TestCheckResourceAttr(
						"aws_s3_bucket.bucket", "lifecycle_rule.0.expiration.3591068768.expired_object_delete_marker", "true"),
				),
			},
			{
				Config: testAccAWSS3BucketConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists("aws_s3_bucket.bucket"),
				),
			},
		},
	})
}

func TestAccAWSS3Bucket_Replication(t *testing.T) {
	rInt := acctest.RandInt()
	region := testAccGetRegion()
	partition := testAccGetPartition()

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccMultipleRegionsPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckWithProviders(testAccCheckAWSS3BucketDestroyWithProvider, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigReplication(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.bucket", testAccAwsRegionProviderFunc(region, &providers)),
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.destination", testAccAwsRegionProviderFunc("eu-west-1", &providers)),
					resource.TestCheckResourceAttr("aws_s3_bucket.bucket", "replication_configuration.#", "0"),
				),
			},
			{
				Config: testAccAWSS3BucketConfigReplicationWithConfiguration(rInt, "STANDARD"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.bucket", testAccAwsRegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr("aws_s3_bucket.bucket", "replication_configuration.#", "1"),
					resource.TestMatchResourceAttr("aws_s3_bucket.bucket", "replication_configuration.0.role", regexp.MustCompile(fmt.Sprintf("^arn:aws[\\w-]*:iam::[\\d+]+:role/tf-iam-role-replication-%d", rInt))),
					resource.TestCheckResourceAttr("aws_s3_bucket.bucket", "replication_configuration.0.rules.#", "1"),
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.destination", testAccAwsRegionProviderFunc("eu-west-1", &providers)),
					testAccCheckAWSS3BucketReplicationRules(
						"aws_s3_bucket.bucket",
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
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.bucket", testAccAwsRegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr("aws_s3_bucket.bucket", "replication_configuration.#", "1"),
					resource.TestMatchResourceAttr("aws_s3_bucket.bucket", "replication_configuration.0.role", regexp.MustCompile(fmt.Sprintf("^arn:aws[\\w-]*:iam::[\\d+]+:role/tf-iam-role-replication-%d", rInt))),
					resource.TestCheckResourceAttr("aws_s3_bucket.bucket", "replication_configuration.0.rules.#", "1"),
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.destination", testAccAwsRegionProviderFunc("eu-west-1", &providers)),
					testAccCheckAWSS3BucketReplicationRules(
						"aws_s3_bucket.bucket",
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
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.bucket", testAccAwsRegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr("aws_s3_bucket.bucket", "replication_configuration.#", "1"),
					resource.TestMatchResourceAttr("aws_s3_bucket.bucket", "replication_configuration.0.role", regexp.MustCompile(fmt.Sprintf("^arn:aws[\\w-]*:iam::[\\d+]+:role/tf-iam-role-replication-%d", rInt))),
					resource.TestCheckResourceAttr("aws_s3_bucket.bucket", "replication_configuration.0.rules.#", "1"),
					testAccCheckAWSS3BucketReplicationRules(
						"aws_s3_bucket.bucket",
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

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccMultipleRegionsPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckWithProviders(testAccCheckAWSS3BucketDestroyWithProvider, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigReplicationWithAccessControlTranslation(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.bucket", testAccAwsRegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr("aws_s3_bucket.bucket", "replication_configuration.#", "1"),
					resource.TestMatchResourceAttr("aws_s3_bucket.bucket", "replication_configuration.0.role", regexp.MustCompile(fmt.Sprintf("^arn:%s:iam::[\\d+]+:role/tf-iam-role-replication-%d", partition, rInt))),
					resource.TestCheckResourceAttr("aws_s3_bucket.bucket", "replication_configuration.0.rules.#", "1"),
					testAccCheckAWSS3BucketReplicationRules(
						"aws_s3_bucket.bucket",
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
				Config: testAccAWSS3BucketConfigReplicationWithSseKmsEncryptedObjectsAndAccessControlTranslation(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.bucket", testAccAwsRegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr("aws_s3_bucket.bucket", "replication_configuration.#", "1"),
					resource.TestMatchResourceAttr("aws_s3_bucket.bucket", "replication_configuration.0.role", regexp.MustCompile(fmt.Sprintf("^arn:%s:iam::[\\d+]+:role/tf-iam-role-replication-%d", partition, rInt))),
					resource.TestCheckResourceAttr("aws_s3_bucket.bucket", "replication_configuration.0.rules.#", "1"),
					testAccCheckAWSS3BucketReplicationRules(
						"aws_s3_bucket.bucket",
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
	region := testAccGetRegion()

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccMultipleRegionsPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckWithProviders(testAccCheckAWSS3BucketDestroyWithProvider, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigReplicationWithoutStorageClass(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.bucket", testAccAwsRegionProviderFunc(region, &providers)),
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.destination", testAccAwsRegionProviderFunc("eu-west-1", &providers)),
				),
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
	region := testAccGetRegion()

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccMultipleRegionsPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckWithProviders(testAccCheckAWSS3BucketDestroyWithProvider, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigReplicationWithoutPrefix(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.bucket", testAccAwsRegionProviderFunc(region, &providers)),
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.destination", testAccAwsRegionProviderFunc("eu-west-1", &providers)),
				),
			},
		},
	})
}

func TestAccAWSS3Bucket_ReplicationSchemaV2(t *testing.T) {
	rInt := acctest.RandInt()
	region := testAccGetRegion()
	partition := testAccGetPartition()

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccMultipleRegionsPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckWithProviders(testAccCheckAWSS3BucketDestroyWithProvider, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigReplicationWithV2ConfigurationNoTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.bucket", testAccAwsRegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr("aws_s3_bucket.bucket", "replication_configuration.#", "1"),
					resource.TestMatchResourceAttr("aws_s3_bucket.bucket", "replication_configuration.0.role", regexp.MustCompile(fmt.Sprintf("^arn:aws[\\w-]*:iam::[\\d+]+:role/tf-iam-role-replication-%d", rInt))),
					resource.TestCheckResourceAttr("aws_s3_bucket.bucket", "replication_configuration.0.rules.#", "1"),
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.destination", testAccAwsRegionProviderFunc("eu-west-1", &providers)),
					testAccCheckAWSS3BucketReplicationRules(
						"aws_s3_bucket.bucket",
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
				Config: testAccAWSS3BucketConfigReplicationWithV2ConfigurationOnlyOneTag(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.bucket", testAccAwsRegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr("aws_s3_bucket.bucket", "replication_configuration.#", "1"),
					resource.TestMatchResourceAttr("aws_s3_bucket.bucket", "replication_configuration.0.role", regexp.MustCompile(fmt.Sprintf("^arn:aws[\\w-]*:iam::[\\d+]+:role/tf-iam-role-replication-%d", rInt))),
					resource.TestCheckResourceAttr("aws_s3_bucket.bucket", "replication_configuration.0.rules.#", "1"),
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.destination", testAccAwsRegionProviderFunc("eu-west-1", &providers)),
					testAccCheckAWSS3BucketReplicationRules(
						"aws_s3_bucket.bucket",
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
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.bucket", testAccAwsRegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr("aws_s3_bucket.bucket", "replication_configuration.#", "1"),
					resource.TestMatchResourceAttr("aws_s3_bucket.bucket", "replication_configuration.0.role", regexp.MustCompile(fmt.Sprintf("^arn:aws[\\w-]*:iam::[\\d+]+:role/tf-iam-role-replication-%d", rInt))),
					resource.TestCheckResourceAttr("aws_s3_bucket.bucket", "replication_configuration.0.rules.#", "1"),
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.destination", testAccAwsRegionProviderFunc("eu-west-1", &providers)),
					testAccCheckAWSS3BucketReplicationRules(
						"aws_s3_bucket.bucket",
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
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.bucket", testAccAwsRegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr("aws_s3_bucket.bucket", "replication_configuration.#", "1"),
					resource.TestMatchResourceAttr("aws_s3_bucket.bucket", "replication_configuration.0.role", regexp.MustCompile(fmt.Sprintf("^arn:aws[\\w-]*:iam::[\\d+]+:role/tf-iam-role-replication-%d", rInt))),
					resource.TestCheckResourceAttr("aws_s3_bucket.bucket", "replication_configuration.0.rules.#", "1"),
					testAccCheckAWSS3BucketExistsWithProvider("aws_s3_bucket.destination", testAccAwsRegionProviderFunc("eu-west-1", &providers)),
					testAccCheckAWSS3BucketReplicationRules(
						"aws_s3_bucket.bucket",
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketObjectLockEnabledNoDefaultRetention(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists("aws_s3_bucket.arbitrary"),
					resource.TestCheckResourceAttr("aws_s3_bucket.arbitrary", "object_lock_configuration.#", "1"),
					resource.TestCheckResourceAttr("aws_s3_bucket.arbitrary", "object_lock_configuration.0.object_lock_enabled", "Enabled"),
					resource.TestCheckResourceAttr("aws_s3_bucket.arbitrary", "object_lock_configuration.0.rule.#", "0"),
				),
			},
			{
				Config: testAccAWSS3BucketObjectLockEnabledWithDefaultRetention(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists("aws_s3_bucket.arbitrary"),
					resource.TestCheckResourceAttr("aws_s3_bucket.arbitrary", "object_lock_configuration.#", "1"),
					resource.TestCheckResourceAttr("aws_s3_bucket.arbitrary", "object_lock_configuration.0.object_lock_enabled", "Enabled"),
					resource.TestCheckResourceAttr("aws_s3_bucket.arbitrary", "object_lock_configuration.0.rule.#", "1"),
					resource.TestCheckResourceAttr("aws_s3_bucket.arbitrary", "object_lock_configuration.0.rule.0.default_retention.0.mode", "COMPLIANCE"),
					resource.TestCheckResourceAttr("aws_s3_bucket.arbitrary", "object_lock_configuration.0.rule.0.default_retention.0.days", "3"),
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

func testAccCheckAWSS3BucketDestroy(s *terraform.State) error {
	return testAccCheckAWSS3BucketDestroyWithProvider(s, testAccProvider)
}

func testAccCheckAWSS3BucketDestroyWithProvider(s *terraform.State, provider *schema.Provider) error {
	conn := provider.Meta().(*AWSClient).s3conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_s3_bucket" {
			continue
		}
		_, err := conn.DeleteBucket(&s3.DeleteBucketInput{
			Bucket: aws.String(rs.Primary.ID),
		})
		if err != nil {
			if isAWSErr(err, s3.ErrCodeNoSuchBucket, "") {
				return nil
			}
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

// These need a bit of randomness as the name can only be used once globally
// within AWS
func testAccBucketName(randInt int) string {
	return fmt.Sprintf("tf-test-bucket-%d", randInt)
}

func testAccBucketDomainName(randInt int) string {
	return fmt.Sprintf("tf-test-bucket-%d.s3.amazonaws.com", randInt)
}

func testAccBucketRegionalDomainName(randInt int, region string) string {
	bucket := fmt.Sprintf("tf-test-bucket-%d", randInt)
	regionalEndpoint, err := BucketRegionalDomainName(bucket, region)
	if err != nil {
		return fmt.Sprintf("Regional endpoint not found for bucket %s", bucket)
	}
	return regionalEndpoint
}

func testAccWebsiteEndpoint(randInt int, region string) string {
	return fmt.Sprintf("tf-test-bucket-%d.s3-website-%s.amazonaws.com", randInt, region)
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
}`, partition, randInt)
}

func testAccAWSS3BucketConfig(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
	bucket = "tf-test-bucket-%d"
	acl = "public-read"
}
`, randInt)
}

func testAccAWSS3MultiBucketConfigWithTags(randInt int) string {
	t := template.Must(template.New("t1").
		Parse(`
resource "aws_s3_bucket" "bucket1" {
	bucket = "tf-test-bucket-1-{{.GUID}}"
	acl = "private"
	force_destroy = true
	tags = {
		Name = "tf-test-bucket-1-{{.GUID}}"
		Environment = "{{.GUID}}"
	}
}

resource "aws_s3_bucket" "bucket2" {
	bucket = "tf-test-bucket-2-{{.GUID}}"
	acl = "private"
	force_destroy = true
	tags = {
		Name = "tf-test-bucket-2-{{.GUID}}"
		Environment = "{{.GUID}}"
	}
}

resource "aws_s3_bucket" "bucket3" {
	bucket = "tf-test-bucket-3-{{.GUID}}"
	acl = "private"
	force_destroy = true
	tags = {
		Name = "tf-test-bucket-3-{{.GUID}}"
		Environment = "{{.GUID}}"
	}
}

resource "aws_s3_bucket" "bucket4" {
	bucket = "tf-test-bucket-4-{{.GUID}}"
	acl = "private"
	force_destroy = true
	tags = {
		Name = "tf-test-bucket-4-{{.GUID}}"
		Environment = "{{.GUID}}"
	}
}

resource "aws_s3_bucket" "bucket5" {
	bucket = "tf-test-bucket-5-{{.GUID}}"
	acl = "private"
	force_destroy = true
	tags = {
		Name = "tf-test-bucket-5-{{.GUID}}"
		Environment = "{{.GUID}}"
	}
}

resource "aws_s3_bucket" "bucket6" {
	bucket = "tf-test-bucket-6-{{.GUID}}"
	acl = "private"
	force_destroy = true
	tags = {
		Name = "tf-test-bucket-6-{{.GUID}}"
		Environment = "{{.GUID}}"
	}
}
`))
	var doc bytes.Buffer
	t.Execute(&doc, struct{ GUID int }{GUID: randInt})
	return doc.String()
}

func testAccAWSS3BucketConfigWithRegion(randInt int) string {
	return fmt.Sprintf(`
provider "aws" {
	alias = "west"
	region = "eu-west-1"
}

resource "aws_s3_bucket" "bucket" {
	provider = "aws.west"
	bucket = "tf-test-bucket-%d"
	region = "eu-west-1"
}
`, randInt)
}

func testAccAWSS3BucketWebsiteConfig(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
	bucket = "tf-test-bucket-%d"
	acl = "public-read"

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
	acl = "public-read"

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
	acl = "public-read"

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
	acl = "public-read"

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
	acl = "public-read"

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
	bucket = "tf-test-bucket-%d"
	acl = "public-read"
	acceleration_status = "Enabled"
}
`, randInt)
}

func testAccAWSS3BucketConfigWithoutAcceleration(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
	bucket = "tf-test-bucket-%d"
	acl = "public-read"
	acceleration_status = "Suspended"
}
`, randInt)
}

func testAccAWSS3BucketConfigRequestPayerBucketOwner(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
	bucket = "tf-test-bucket-%d"
	acl = "public-read"
	request_payer = "BucketOwner"
}
`, randInt)
}

func testAccAWSS3BucketConfigRequestPayerRequester(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
	bucket = "tf-test-bucket-%d"
	acl = "public-read"
	request_payer = "Requester"
}
`, randInt)
}

func testAccAWSS3BucketConfigWithPolicy(randInt int, partition string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
	bucket = "tf-test-bucket-%d"
	acl = "public-read"
	policy = %s
}
`, randInt, strconv.Quote(testAccAWSS3BucketPolicy(randInt, partition)))
}

func testAccAWSS3BucketDestroyedConfig(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
	bucket = "tf-test-bucket-%d"
	acl = "public-read"
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
	  	sse_algorithm     = "AES256"
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
	  	sse_algorithm     = "aws:kms"
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
	acl = "public-read"
	policy = ""
}
`, randInt)
}

func testAccAWSS3BucketConfigWithVersioning(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
	bucket = "tf-test-bucket-%d"
	acl = "public-read"
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
	acl = "public-read"
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
	acl = "public-read"
	cors_rule {
			allowed_headers = ["*"]
			allowed_methods = ["PUT","POST"]
			allowed_origins = ["https://www.example.com"]
			expose_headers = ["x-amz-server-side-encryption","ETag"]
			max_age_seconds = 3000
	}
}
`, randInt)
}

func testAccAWSS3BucketConfigWithCORSEmptyOrigin(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
	bucket = "tf-test-bucket-%d"
	acl = "public-read"
	cors_rule {
			allowed_headers = ["*"]
			allowed_methods = ["PUT","POST"]
			allowed_origins = [""]
			expose_headers = ["x-amz-server-side-encryption","ETag"]
			max_age_seconds = 3000
	}
}
`, randInt)
}

var testAccAWSS3BucketConfigWithAcl = `
resource "aws_s3_bucket" "bucket" {
	bucket = "tf-test-bucket-%d"
	acl = "public-read"
}
`

var testAccAWSS3BucketConfigWithAclUpdate = `
resource "aws_s3_bucket" "bucket" {
	bucket = "tf-test-bucket-%d"
	acl = "private"
}
`

func testAccAWSS3BucketConfigWithLogging(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "log_bucket" {
	bucket = "tf-test-log-bucket-%d"
	acl = "log-delivery-write"
}
resource "aws_s3_bucket" "bucket" {
	bucket = "tf-test-bucket-%d"
	acl = "private"
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
	acl = "private"
	lifecycle_rule {
		id = "id1"
		prefix = "path1/"
		enabled = true

		expiration {
			days = 365
		}

		transition {
			days = 30
			storage_class = "STANDARD_IA"
		}

		transition {
			days = 60
			storage_class = "INTELLIGENT_TIERING"
		}

		transition {
			days = 90
			storage_class = "ONEZONE_IA"
		}

		transition {
			days = 120
			storage_class = "GLACIER"
		}
	}
	lifecycle_rule {
		id = "id2"
		prefix = "path2/"
		enabled = true

		expiration {
			date = "2016-01-12"
		}
	}
	lifecycle_rule {
		id = "id3"
		prefix = "path3/"
		enabled = true

		transition {
			days = 0
			storage_class = "GLACIER"
		}
	}
	lifecycle_rule {
		id = "id4"
		prefix = "path4/"
		enabled = true

	tags = {
			"tagKey" = "tagValue"
			"terraform" = "hashicorp"
		}

		expiration {
			date = "2016-01-12"
		}
	}
}
`, randInt)
}

func testAccAWSS3BucketConfigWithLifecycleExpireMarker(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
	bucket = "tf-test-bucket-%d"
	acl = "private"
	lifecycle_rule {
		id = "id1"
		prefix = "path1/"
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
	acl = "private"
	versioning {
	  enabled = false
	}
	lifecycle_rule {
		id = "id1"
		prefix = "path1/"
		enabled = true

		noncurrent_version_expiration {
			days = 365
		}
		noncurrent_version_transition {
			days = 30
			storage_class = "STANDARD_IA"
		}
		noncurrent_version_transition {
			days = 60
			storage_class = "GLACIER"
		}
	}
	lifecycle_rule {
		id = "id2"
		prefix = "path2/"
		enabled = false

		noncurrent_version_expiration {
			days = 365
		}
	}
	lifecycle_rule {
		id = "id3"
		prefix = "path3/"
		enabled = true

		noncurrent_version_transition {
			days = 0
			storage_class = "GLACIER"
		}
	}
}
`, randInt)
}

const testAccAWSS3BucketConfigReplicationBasic = `
provider "aws" {
  alias  = "euwest"
  region = "eu-west-1"
}

provider "aws" {
  alias  = "uswest2"
  region = "us-west-2"
}

resource "aws_iam_role" "role" {
  name               = "tf-iam-role-replication-%d"
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
`

func testAccAWSS3BucketConfigReplication(randInt int) string {
	return fmt.Sprintf(testAccAWSS3BucketConfigReplicationBasic+`
resource "aws_s3_bucket" "bucket" {
    provider = "aws.uswest2"
    bucket   = "tf-test-bucket-%d"
    acl      = "private"

    versioning {
        enabled = true
    }
}

resource "aws_s3_bucket" "destination" {
    provider = "aws.euwest"
    bucket   = "tf-test-bucket-destination-%d"
    region   = "eu-west-1"

    versioning {
        enabled = true
    }
}
`, randInt, randInt, randInt)
}

func testAccAWSS3BucketConfigReplicationWithConfiguration(randInt int, storageClass string) string {
	return fmt.Sprintf(testAccAWSS3BucketConfigReplicationBasic+`
resource "aws_s3_bucket" "bucket" {
    provider = "aws.uswest2"
    bucket   = "tf-test-bucket-%d"
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
                storage_class = "%s"
            }
        }
    }
}

resource "aws_s3_bucket" "destination" {
    provider = "aws.euwest"
    bucket   = "tf-test-bucket-destination-%d"
    region   = "eu-west-1"

    versioning {
        enabled = true
    }
}
`, randInt, randInt, storageClass, randInt)
}

func testAccAWSS3BucketConfigReplicationWithSseKmsEncryptedObjects(randInt int) string {
	return fmt.Sprintf(testAccAWSS3BucketConfigReplicationBasic+`
resource "aws_kms_key" "replica" {
  provider                = "aws.euwest"
  description             = "TF Acceptance Test S3 repl KMS key"
  deletion_window_in_days = 7
}

resource "aws_s3_bucket" "bucket" {
    provider = "aws.uswest2"
    bucket   = "tf-test-bucket-%d"
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

resource "aws_s3_bucket" "destination" {
    provider = "aws.euwest"
    bucket   = "tf-test-bucket-destination-%d"
    region   = "eu-west-1"

    versioning {
        enabled = true
    }
}
`, randInt, randInt, randInt)
}

func testAccAWSS3BucketConfigReplicationWithAccessControlTranslation(randInt int) string {
	return fmt.Sprintf(testAccAWSS3BucketConfigReplicationBasic+`
data "aws_caller_identity" "current" {}

resource "aws_s3_bucket" "bucket" {
    provider = "aws.uswest2"
    bucket   = "tf-test-bucket-%d"
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

resource "aws_s3_bucket" "destination" {
    provider = "aws.euwest"
    bucket   = "tf-test-bucket-destination-%d"
    region   = "eu-west-1"

    versioning {
        enabled = true
    }
}
`, randInt, randInt, randInt)
}

func testAccAWSS3BucketConfigReplicationWithSseKmsEncryptedObjectsAndAccessControlTranslation(randInt int) string {
	return fmt.Sprintf(testAccAWSS3BucketConfigReplicationBasic+`
data "aws_caller_identity" "current" {}

resource "aws_kms_key" "replica" {
  provider                = "aws.euwest"
  description             = "TF Acceptance Test S3 repl KMS key"
  deletion_window_in_days = 7
}

resource "aws_s3_bucket" "bucket" {
    provider = "aws.uswest2"
    bucket   = "tf-test-bucket-%d"
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

resource "aws_s3_bucket" "destination" {
    provider = "aws.euwest"
    bucket   = "tf-test-bucket-destination-%d"
    region   = "eu-west-1"

    versioning {
        enabled = true
    }
}
`, randInt, randInt, randInt)
}

func testAccAWSS3BucketConfigReplicationWithoutStorageClass(randInt int) string {
	return fmt.Sprintf(testAccAWSS3BucketConfigReplicationBasic+`
resource "aws_s3_bucket" "bucket" {
    provider = "aws.uswest2"
    bucket   = "tf-test-bucket-%d"
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

resource "aws_s3_bucket" "destination" {
    provider = "aws.euwest"
    bucket   = "tf-test-bucket-destination-%d"
    region   = "eu-west-1"

    versioning {
        enabled = true
    }
}
`, randInt, randInt, randInt)
}

func testAccAWSS3BucketConfigReplicationWithoutPrefix(randInt int) string {
	return fmt.Sprintf(testAccAWSS3BucketConfigReplicationBasic+`
resource "aws_s3_bucket" "bucket" {
    provider = "aws.uswest2"
    bucket   = "tf-test-bucket-%d"
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

resource "aws_s3_bucket" "destination" {
    provider = "aws.euwest"
    bucket   = "tf-test-bucket-destination-%d"
    region   = "eu-west-1"

    versioning {
        enabled = true
    }
}
`, randInt, randInt, randInt)
}

func testAccAWSS3BucketConfigReplicationNoVersioning(randInt int) string {
	return fmt.Sprintf(testAccAWSS3BucketConfigReplicationBasic+`
resource "aws_s3_bucket" "bucket" {
    provider = "aws.uswest2"
    bucket   = "tf-test-bucket-%d"
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

resource "aws_s3_bucket" "destination" {
    provider = "aws.euwest"
    bucket   = "tf-test-bucket-destination-%d"
    region   = "eu-west-1"

    versioning {
        enabled = true
    }
}
`, randInt, randInt, randInt)
}

func testAccAWSS3BucketConfigReplicationWithV2ConfigurationNoTags(randInt int) string {
	return fmt.Sprintf(testAccAWSS3BucketConfigReplicationBasic+`
resource "aws_s3_bucket" "bucket" {
    provider = "aws.uswest2"
    bucket   = "tf-test-bucket-%d"
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

resource "aws_s3_bucket" "destination" {
    provider = "aws.euwest"
    bucket   = "tf-test-bucket-destination-%d"
    region   = "eu-west-1"

    versioning {
        enabled = true
    }
}
`, randInt, randInt, randInt)
}

func testAccAWSS3BucketConfigReplicationWithV2ConfigurationOnlyOneTag(randInt int) string {
	return fmt.Sprintf(testAccAWSS3BucketConfigReplicationBasic+`
resource "aws_s3_bucket" "bucket" {
    provider = "aws.uswest2"
    bucket   = "tf-test-bucket-%d"
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

resource "aws_s3_bucket" "destination" {
    provider = "aws.euwest"
    bucket   = "tf-test-bucket-destination-%d"
    region   = "eu-west-1"

    versioning {
        enabled = true
    }
}
`, randInt, randInt, randInt)
}

func testAccAWSS3BucketConfigReplicationWithV2ConfigurationPrefixAndTags(randInt int) string {
	return fmt.Sprintf(testAccAWSS3BucketConfigReplicationBasic+`
resource "aws_s3_bucket" "bucket" {
    provider = "aws.uswest2"
    bucket   = "tf-test-bucket-%d"
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

resource "aws_s3_bucket" "destination" {
    provider = "aws.euwest"
    bucket   = "tf-test-bucket-destination-%d"
    region   = "eu-west-1"

    versioning {
        enabled = true
    }
}
`, randInt, randInt, randInt)
}

func testAccAWSS3BucketConfigReplicationWithV2ConfigurationMultipleTags(randInt int) string {
	return fmt.Sprintf(testAccAWSS3BucketConfigReplicationBasic+`
resource "aws_s3_bucket" "bucket" {
    provider = "aws.uswest2"
    bucket   = "tf-test-bucket-%d"
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

resource "aws_s3_bucket" "destination" {
    provider = "aws.euwest"
    bucket   = "tf-test-bucket-destination-%d"
    region   = "eu-west-1"

    versioning {
        enabled = true
    }
}
`, randInt, randInt, randInt)
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
