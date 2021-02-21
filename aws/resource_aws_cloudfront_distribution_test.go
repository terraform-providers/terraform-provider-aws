package aws

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_cloudfront_distribution", &resource.Sweeper{
		Name: "aws_cloudfront_distribution",
		F:    testSweepCloudFrontDistributions,
	})
}

func testSweepCloudFrontDistributions(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).cloudfrontconn

	distributionSummaries := make([]*cloudfront.DistributionSummary, 0)

	input := &cloudfront.ListDistributionsInput{}
	err = conn.ListDistributionsPages(input, func(page *cloudfront.ListDistributionsOutput, lastPage bool) bool {
		distributionSummaries = append(distributionSummaries, page.DistributionList.Items...)
		return !lastPage
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping CloudFront Distribution sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error listing CloudFront Distributions: %s", err)
	}

	if len(distributionSummaries) == 0 {
		log.Print("[DEBUG] No CloudFront Distributions to sweep")
		return nil
	}

	for _, distributionSummary := range distributionSummaries {
		distributionID := *distributionSummary.Id

		if *distributionSummary.Enabled {
			log.Printf("[WARN] Skipping deletion of enabled CloudFront Distribution: %s", distributionID)
			continue
		}

		output, err := conn.GetDistribution(&cloudfront.GetDistributionInput{
			Id: aws.String(distributionID),
		})
		if err != nil {
			return fmt.Errorf("Error reading CloudFront Distribution %s: %s", distributionID, err)
		}

		_, err = conn.DeleteDistribution(&cloudfront.DeleteDistributionInput{
			Id:      aws.String(distributionID),
			IfMatch: output.ETag,
		})
		if err != nil {
			return fmt.Errorf("Error deleting CloudFront Distribution %s: %s", distributionID, err)
		}
	}

	return nil
}

func TestAccAWSCloudFrontDistribution_disappears(t *testing.T) {
	var distribution cloudfront.Distribution
	resourceName := "aws_cloudfront_distribution.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontDistributionConfigEnabled(false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontDistributionExists(resourceName, &distribution),
					testAccCheckCloudFrontDistributionDisappears(&distribution),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestAccAWSCloudFrontDistribution_S3Origin runs an
// aws_cloudfront_distribution acceptance test with a single S3 origin.
//
// If you are testing manually and can't wait for deletion, set the
// TF_TEST_CLOUDFRONT_RETAIN environment variable.
func TestAccAWSCloudFrontDistribution_S3Origin(t *testing.T) {
	var distribution cloudfront.Distribution
	ri := acctest.RandInt()
	testConfig := fmt.Sprintf(testAccAWSCloudFrontDistributionS3Config, ri, originBucket, logBucket, testAccAWSCloudFrontDistributionRetainConfig())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontDistributionExists("aws_cloudfront_distribution.s3_distribution", &distribution),
					resource.TestCheckResourceAttr(
						"aws_cloudfront_distribution.s3_distribution",
						"hosted_zone_id",
						"Z2FDTNDATAQYW2",
					),
				),
			},
			{
				ResourceName:      "aws_cloudfront_distribution.s3_distribution",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_on_delete",
					"wait_for_deployment",
				},
			},
		},
	})
}

func TestAccAWSCloudFrontDistribution_S3OriginWithTags(t *testing.T) {
	var distribution cloudfront.Distribution
	ri := acctest.RandInt()
	preConfig := fmt.Sprintf(testAccAWSCloudFrontDistributionS3ConfigWithTags, ri, originBucket, logBucket, testAccAWSCloudFrontDistributionRetainConfig())
	postConfig := fmt.Sprintf(testAccAWSCloudFrontDistributionS3ConfigWithTagsUpdated, ri, originBucket, logBucket, testAccAWSCloudFrontDistributionRetainConfig())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontDistributionExists("aws_cloudfront_distribution.s3_distribution", &distribution),
					resource.TestCheckResourceAttr(
						"aws_cloudfront_distribution.s3_distribution", "tags.%", "2"),
					resource.TestCheckResourceAttr(
						"aws_cloudfront_distribution.s3_distribution", "tags.environment", "production"),
					resource.TestCheckResourceAttr(
						"aws_cloudfront_distribution.s3_distribution", "tags.account", "main"),
				),
			},
			{
				ResourceName:      "aws_cloudfront_distribution.s3_distribution",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_on_delete",
					"wait_for_deployment",
				},
			},
			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontDistributionExists("aws_cloudfront_distribution.s3_distribution", &distribution),
					resource.TestCheckResourceAttr(
						"aws_cloudfront_distribution.s3_distribution", "tags.%", "1"),
					resource.TestCheckResourceAttr(
						"aws_cloudfront_distribution.s3_distribution", "tags.environment", "dev"),
				),
			},
		},
	})
}

// TestAccAWSCloudFrontDistribution_customOriginruns an
// aws_cloudfront_distribution acceptance test with a single custom origin.
//
// If you are testing manually and can't wait for deletion, set the
// TF_TEST_CLOUDFRONT_RETAIN environment variable.
func TestAccAWSCloudFrontDistribution_customOrigin(t *testing.T) {
	var distribution cloudfront.Distribution
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontDistributionCustomConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontDistributionExists("aws_cloudfront_distribution.custom_distribution", &distribution),
				),
			},
			{
				ResourceName:      "aws_cloudfront_distribution.custom_distribution",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_on_delete",
					"wait_for_deployment",
				},
			},
		},
	})
}

func TestAccAWSCloudFrontDistribution_originPolicyDefault(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontDistributionOriginRequestPolicyConfigDefault,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("aws_cloudfront_distribution.custom_distribution", "default_cache_behavior.0.origin_request_policy_id", regexp.MustCompile("[A-z0-9]+")),
				),
			},
			{
				ResourceName:      "aws_cloudfront_distribution.custom_distribution",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_on_delete",
					"wait_for_deployment",
				},
			},
		},
	})
}

func TestAccAWSCloudFrontDistribution_originPolicyOrdered(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontDistributionOriginRequestPolicyConfigOrdered,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("aws_cloudfront_distribution.custom_distribution", "ordered_cache_behavior.0.origin_request_policy_id", regexp.MustCompile("[A-z0-9]+")),
				),
			},
			{
				ResourceName:      "aws_cloudfront_distribution.custom_distribution",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_on_delete",
					"wait_for_deployment",
				},
			},
		},
	})
}

// TestAccAWSCloudFrontDistribution_multiOrigin runs an
// aws_cloudfront_distribution acceptance test with multiple origins.
//
// If you are testing manually and can't wait for deletion, set the
// TF_TEST_CLOUDFRONT_RETAIN environment variable.
func TestAccAWSCloudFrontDistribution_multiOrigin(t *testing.T) {
	var distribution cloudfront.Distribution
	resourceName := "aws_cloudfront_distribution.multi_origin_distribution"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontDistributionMultiOriginConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontDistributionExists(resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.0.default_ttl", "50"),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.0.path_pattern", "images1/*.jpg"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_on_delete",
					"wait_for_deployment",
				},
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/188
// TestAccAWSCloudFrontDistribution_orderedCacheBehavior runs an
// aws_cloudfront_distribution acceptance test with multiple and ordered cache behaviors.
//
// If you are testing manually and can't wait for deletion, set the
// TF_TEST_CLOUDFRONT_RETAIN environment variable.
func TestAccAWSCloudFrontDistribution_orderedCacheBehavior(t *testing.T) {
	var distribution cloudfront.Distribution
	resourceName := "aws_cloudfront_distribution.main"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontDistributionOrderedCacheBehavior,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontDistributionExists(resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.0.default_ttl", "50"),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.0.path_pattern", "images1/*.jpg"),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.1.default_ttl", "51"),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.1.path_pattern", "images2/*.jpg"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_on_delete",
					"wait_for_deployment",
				},
			},
		},
	})
}

func TestAccAWSCloudFrontDistribution_orderedCacheBehaviorCachePolicy(t *testing.T) {
	var distribution cloudfront.Distribution
	resourceName := "aws_cloudfront_distribution.main"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontDistributionOrderedCacheBehaviorCachePolicy,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontDistributionExists(resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.0.path_pattern", "images2/*.jpg"),
					resource.TestMatchResourceAttr(resourceName, "ordered_cache_behavior.0.cache_policy_id", regexp.MustCompile(`^[a-z0-9]+`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_on_delete",
					"wait_for_deployment",
				},
			},
		},
	})
}

func TestAccAWSCloudFrontDistribution_Origin_EmptyDomainName(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSCloudFrontDistributionConfig_Origin_EmptyDomainName,
				ExpectError: regexp.MustCompile(`domain_name must not be empty`),
			},
		},
	})
}

func TestAccAWSCloudFrontDistribution_Origin_EmptyOriginID(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSCloudFrontDistributionConfig_Origin_EmptyOriginID,
				ExpectError: regexp.MustCompile(`origin_id must not be empty`),
			},
		},
	})
}

// TestAccAWSCloudFrontDistribution_noOptionalItemsConfig runs an
// aws_cloudfront_distribution acceptance test with no optional items set.
//
// If you are testing manually and can't wait for deletion, set the
// TF_TEST_CLOUDFRONT_RETAIN environment variable.
func TestAccAWSCloudFrontDistribution_noOptionalItemsConfig(t *testing.T) {
	var distribution cloudfront.Distribution
	resourceName := "aws_cloudfront_distribution.no_optional_items"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontDistributionNoOptionalItemsConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontDistributionExists(resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "aliases.#", "0"),
					testAccMatchResourceAttrGlobalARN(resourceName, "arn", "cloudfront", regexp.MustCompile(`distribution/[A-Z0-9]+$`)),
					resource.TestCheckResourceAttr(resourceName, "custom_error_response.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.allowed_methods.#", "7"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.cached_methods.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.compress", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.forwarded_values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.forwarded_values.0.cookies.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.forwarded_values.0.cookies.0.forward", "all"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.forwarded_values.0.cookies.0.whitelisted_names.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.forwarded_values.0.headers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.forwarded_values.0.query_string", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.forwarded_values.0.query_string_cache_keys.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.lambda_function_association.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.min_ttl", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.smooth_streaming", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.target_origin_id", "myCustomOrigin"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.trusted_signers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.viewer_protocol_policy", "allow-all"),
					resource.TestMatchResourceAttr(resourceName, "domain_name", regexp.MustCompile(`^[a-z0-9]+\.cloudfront\.net$`)),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestMatchResourceAttr(resourceName, "etag", regexp.MustCompile(`^[A-Z0-9]+$`)),
					resource.TestCheckResourceAttr(resourceName, "hosted_zone_id", "Z2FDTNDATAQYW2"),
					resource.TestCheckResourceAttrSet(resourceName, "http_version"),
					resource.TestCheckResourceAttr(resourceName, "is_ipv6_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "logging_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "origin.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "origin.*", map[string]string{
						"custom_header.#":                                 "0",
						"custom_origin_config.#":                          "1",
						"custom_origin_config.0.http_port":                "80",
						"custom_origin_config.0.https_port":               "443",
						"custom_origin_config.0.origin_keepalive_timeout": "5",
						"custom_origin_config.0.origin_protocol_policy":   "http-only",
						"custom_origin_config.0.origin_read_timeout":      "30",
						"custom_origin_config.0.origin_ssl_protocols.#":   "2",
						"domain_name": "www.example.com",
					}),
					resource.TestCheckResourceAttr(resourceName, "price_class", "PriceClass_All"),
					resource.TestCheckResourceAttr(resourceName, "restrictions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "restrictions.0.geo_restriction.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "restrictions.0.geo_restriction.0.locations.#", "4"),
					resource.TestCheckResourceAttr(resourceName, "restrictions.0.geo_restriction.0.restriction_type", "whitelist"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "viewer_certificate.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "viewer_certificate.0.cloudfront_default_certificate", "true"),
					resource.TestCheckResourceAttr(resourceName, "wait_for_deployment", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_on_delete",
					"wait_for_deployment",
				},
			},
		},
	})
}

// TestAccAWSCloudFrontDistribution_HTTP11Config runs an
// aws_cloudfront_distribution acceptance test with the HTTP version set to
// 1.1.
//
// If you are testing manually and can't wait for deletion, set the
// TF_TEST_CLOUDFRONT_RETAIN environment variable.
func TestAccAWSCloudFrontDistribution_HTTP11Config(t *testing.T) {
	var distribution cloudfront.Distribution
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontDistributionHTTP11Config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontDistributionExists("aws_cloudfront_distribution.http_1_1", &distribution),
				),
			},
			{
				ResourceName:      "aws_cloudfront_distribution.http_1_1",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_on_delete",
					"wait_for_deployment",
				},
			},
		},
	})
}

func TestAccAWSCloudFrontDistribution_IsIPV6EnabledConfig(t *testing.T) {
	var distribution cloudfront.Distribution
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontDistributionIsIPV6EnabledConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontDistributionExists("aws_cloudfront_distribution.is_ipv6_enabled", &distribution),
					resource.TestCheckResourceAttr(
						"aws_cloudfront_distribution.is_ipv6_enabled", "is_ipv6_enabled", "true"),
				),
			},
			{
				ResourceName:      "aws_cloudfront_distribution.is_ipv6_enabled",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_on_delete",
					"wait_for_deployment",
				},
			},
		},
	})
}

func TestAccAWSCloudFrontDistribution_noCustomErrorResponseConfig(t *testing.T) {
	var distribution cloudfront.Distribution
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontDistributionNoCustomErroResponseInfo,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontDistributionExists("aws_cloudfront_distribution.no_custom_error_responses", &distribution),
				),
			},
			{
				ResourceName:      "aws_cloudfront_distribution.no_custom_error_responses",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_on_delete",
					"wait_for_deployment",
				},
			},
		},
	})
}

func TestAccAWSCloudFrontDistribution_DefaultCacheBehavior_ForwardedValues_Cookies_WhitelistedNames(t *testing.T) {
	var distribution cloudfront.Distribution
	resourceName := "aws_cloudfront_distribution.test"
	retainOnDelete := testAccAWSCloudFrontDistributionRetainOnDeleteFromEnv()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontDistributionConfigDefaultCacheBehaviorForwardedValuesCookiesWhitelistedNamesUnordered3(retainOnDelete),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontDistributionExists(resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.forwarded_values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.forwarded_values.0.cookies.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.forwarded_values.0.cookies.0.whitelisted_names.#", "3"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_on_delete",
					"wait_for_deployment",
				},
			},
			{
				Config: testAccAWSCloudFrontDistributionConfigDefaultCacheBehaviorForwardedValuesCookiesWhitelistedNamesUnordered2(retainOnDelete),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontDistributionExists(resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.forwarded_values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.forwarded_values.0.cookies.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.forwarded_values.0.cookies.0.whitelisted_names.#", "2"),
				),
			},
		},
	})
}

func TestAccAWSCloudFrontDistribution_DefaultCacheBehavior_ForwardedValues_Headers(t *testing.T) {
	var distribution cloudfront.Distribution
	resourceName := "aws_cloudfront_distribution.test"
	retainOnDelete := testAccAWSCloudFrontDistributionRetainOnDeleteFromEnv()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontDistributionConfigDefaultCacheBehaviorForwardedValuesHeadersUnordered3(retainOnDelete),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontDistributionExists(resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.forwarded_values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.forwarded_values.0.headers.#", "3"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_on_delete",
					"wait_for_deployment",
				},
			},
			{
				Config: testAccAWSCloudFrontDistributionConfigDefaultCacheBehaviorForwardedValuesHeadersUnordered2(retainOnDelete),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontDistributionExists(resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.forwarded_values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.forwarded_values.0.headers.#", "2"),
				),
			},
		},
	})
}

func TestAccAWSCloudFrontDistribution_DefaultCacheBehavior_TrustedSigners(t *testing.T) {
	var distribution cloudfront.Distribution
	resourceName := "aws_cloudfront_distribution.test"
	retainOnDelete := testAccAWSCloudFrontDistributionRetainOnDeleteFromEnv()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontDistributionConfigDefaultCacheBehaviorTrustedSignersSelf(retainOnDelete),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontDistributionExists(resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "trusted_signers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "trusted_signers.0.items.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "trusted_signers.0.items.0.aws_account_number", "self"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.0.trusted_signers.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_on_delete",
					"wait_for_deployment",
				},
			},
		},
	})
}

func TestAccAWSCloudFrontDistribution_DefaultCacheBehavior_RealtimeLogConfigArn(t *testing.T) {
	var distribution cloudfront.Distribution
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudfront_distribution.test"
	realtimeLogConfigResourceName := "aws_cloudfront_realtime_log_config.test"
	retainOnDelete := testAccAWSCloudFrontDistributionRetainOnDeleteFromEnv()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontDistributionConfigDefaultCacheBehaviorRealtimeLogConfigArn(rName, retainOnDelete),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontDistributionExists(resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "default_cache_behavior.0.realtime_log_config_arn", realtimeLogConfigResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_on_delete",
					"wait_for_deployment",
				},
			},
		},
	})
}

func TestAccAWSCloudFrontDistribution_OrderedCacheBehavior_RealtimeLogConfigArn(t *testing.T) {
	var distribution cloudfront.Distribution
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudfront_distribution.test"
	realtimeLogConfigResourceName := "aws_cloudfront_realtime_log_config.test"
	retainOnDelete := testAccAWSCloudFrontDistributionRetainOnDeleteFromEnv()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontDistributionConfigOrderedCacheBehaviorRealtimeLogConfigArn(rName, retainOnDelete),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontDistributionExists(resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "ordered_cache_behavior.0.realtime_log_config_arn", realtimeLogConfigResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_on_delete",
					"wait_for_deployment",
				},
			},
		},
	})
}

func TestAccAWSCloudFrontDistribution_Enabled(t *testing.T) {
	var distribution cloudfront.Distribution
	resourceName := "aws_cloudfront_distribution.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontDistributionConfigEnabled(false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontDistributionExists(resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_on_delete",
					"wait_for_deployment",
				},
			},
			{
				Config: testAccAWSCloudFrontDistributionConfigEnabled(true, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontDistributionExists(resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
				),
			},
		},
	})
}

// TestAccAWSCloudFrontDistribution_RetainOnDelete verifies retain_on_delete = true
// This acceptance test performs the following steps:
//  * Trigger a Terraform destroy of the resource, which should only disable the distribution
//  * Check it still exists and is disabled outside Terraform
//  * Destroy for real outside Terraform
func TestAccAWSCloudFrontDistribution_RetainOnDelete(t *testing.T) {
	var distribution cloudfront.Distribution
	resourceName := "aws_cloudfront_distribution.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontDistributionConfigEnabled(true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontDistributionExists(resourceName, &distribution),
				),
			},
			{
				Config:  testAccAWSCloudFrontDistributionConfigEnabled(true, true),
				Destroy: true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontDistributionExistsAPIOnly(&distribution),
					testAccCheckCloudFrontDistributionWaitForDeployment(&distribution),
					testAccCheckCloudFrontDistributionDisabled(&distribution),
					testAccCheckCloudFrontDistributionDisappears(&distribution),
				),
			},
		},
	})
}

func TestAccAWSCloudFrontDistribution_OrderedCacheBehavior_ForwardedValues_Cookies_WhitelistedNames(t *testing.T) {
	var distribution cloudfront.Distribution
	resourceName := "aws_cloudfront_distribution.test"
	retainOnDelete := testAccAWSCloudFrontDistributionRetainOnDeleteFromEnv()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontDistributionConfigOrderedCacheBehaviorForwardedValuesCookiesWhitelistedNamesUnordered3(retainOnDelete),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontDistributionExists(resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.0.forwarded_values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.0.forwarded_values.0.cookies.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.0.forwarded_values.0.cookies.0.whitelisted_names.#", "3"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_on_delete",
					"wait_for_deployment",
				},
			},
			{
				Config: testAccAWSCloudFrontDistributionConfigOrderedCacheBehaviorForwardedValuesCookiesWhitelistedNamesUnordered2(retainOnDelete),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontDistributionExists(resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.0.forwarded_values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.0.forwarded_values.0.cookies.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.0.forwarded_values.0.cookies.0.whitelisted_names.#", "2"),
				),
			},
		},
	})
}

func TestAccAWSCloudFrontDistribution_OrderedCacheBehavior_ForwardedValues_Headers(t *testing.T) {
	var distribution cloudfront.Distribution
	resourceName := "aws_cloudfront_distribution.test"
	retainOnDelete := testAccAWSCloudFrontDistributionRetainOnDeleteFromEnv()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontDistributionConfigOrderedCacheBehaviorForwardedValuesHeadersUnordered3(retainOnDelete),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontDistributionExists(resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.0.forwarded_values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.0.forwarded_values.0.headers.#", "3"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_on_delete",
					"wait_for_deployment",
				},
			},
			{
				Config: testAccAWSCloudFrontDistributionConfigOrderedCacheBehaviorForwardedValuesHeadersUnordered2(retainOnDelete),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontDistributionExists(resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.0.forwarded_values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ordered_cache_behavior.0.forwarded_values.0.headers.#", "2"),
				),
			},
		},
	})
}

func TestAccAWSCloudFrontDistribution_ViewerCertificate_AcmCertificateArn(t *testing.T) {
	var distribution cloudfront.Distribution
	resourceName := "aws_cloudfront_distribution.test"
	retainOnDelete := testAccAWSCloudFrontDistributionRetainOnDeleteFromEnv()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckCloudFrontDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontDistributionConfigViewerCertificateAcmCertificateArn(retainOnDelete),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontDistributionExists(resourceName, &distribution),
				),
			},
			{
				Config:            testAccAWSCloudFrontDistributionConfigViewerCertificateAcmCertificateArn(retainOnDelete),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_on_delete",
					"wait_for_deployment",
				},
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/7773
func TestAccAWSCloudFrontDistribution_ViewerCertificate_AcmCertificateArn_ConflictsWithCloudFrontDefaultCertificate(t *testing.T) {
	var distribution cloudfront.Distribution
	resourceName := "aws_cloudfront_distribution.test"
	retainOnDelete := testAccAWSCloudFrontDistributionRetainOnDeleteFromEnv()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckCloudFrontDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontDistributionConfigViewerCertificateAcmCertificateArnConflictsWithCloudFrontDefaultCertificate(retainOnDelete),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontDistributionExists(resourceName, &distribution),
				),
			},
			{
				Config:            testAccAWSCloudFrontDistributionConfigViewerCertificateAcmCertificateArnConflictsWithCloudFrontDefaultCertificate(retainOnDelete),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_on_delete",
					"wait_for_deployment",
				},
			},
		},
	})
}

func TestAccAWSCloudFrontDistribution_WaitForDeployment(t *testing.T) {
	var distribution cloudfront.Distribution
	resourceName := "aws_cloudfront_distribution.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontDistributionConfigWaitForDeployment(false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontDistributionExists(resourceName, &distribution),
					testAccCheckCloudFrontDistributionStatusInProgress(&distribution),
					testAccCheckCloudFrontDistributionWaitForDeployment(&distribution),
					resource.TestCheckResourceAttr(resourceName, "wait_for_deployment", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_on_delete",
					"wait_for_deployment",
				},
			},
			{
				Config: testAccAWSCloudFrontDistributionConfigWaitForDeployment(true, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontDistributionExists(resourceName, &distribution),
					testAccCheckCloudFrontDistributionStatusInProgress(&distribution),
					resource.TestCheckResourceAttr(resourceName, "wait_for_deployment", "false"),
				),
			},
			{
				Config: testAccAWSCloudFrontDistributionConfigWaitForDeployment(false, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontDistributionExists(resourceName, &distribution),
					testAccCheckCloudFrontDistributionStatusDeployed(&distribution),
					resource.TestCheckResourceAttr(resourceName, "wait_for_deployment", "true"),
				),
			},
		},
	})
}

func testAccCheckCloudFrontDistributionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cloudfrontconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudfront_distribution" {
			continue
		}

		input := &cloudfront.GetDistributionInput{
			Id: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetDistribution(input)

		if isAWSErr(err, cloudfront.ErrCodeNoSuchDistribution, "") {
			continue
		}

		if err != nil {
			return err
		}

		if !testAccAWSCloudFrontDistributionRetainOnDeleteFromEnv() {
			return fmt.Errorf("CloudFront Distribution (%s) still exists", rs.Primary.ID)
		}

		if output != nil && output.Distribution != nil && output.Distribution.DistributionConfig != nil && aws.BoolValue(output.Distribution.DistributionConfig.Enabled) {
			return fmt.Errorf("CloudFront Distribution (%s) not disabled", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckCloudFrontDistributionExists(resourceName string, distribution *cloudfront.Distribution) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource ID not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).cloudfrontconn

		input := &cloudfront.GetDistributionInput{
			Id: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetDistribution(input)

		if err != nil {
			return fmt.Errorf("Error retrieving CloudFront distribution: %s", err)
		}

		*distribution = *output.Distribution

		return nil
	}
}

func testAccCheckCloudFrontDistributionExistsAPIOnly(distribution *cloudfront.Distribution) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).cloudfrontconn

		input := &cloudfront.GetDistributionInput{
			Id: distribution.Id,
		}

		output, err := conn.GetDistribution(input)

		if err != nil {
			return err
		}

		*distribution = *output.Distribution

		return nil
	}
}

func testAccCheckCloudFrontDistributionStatusDeployed(distribution *cloudfront.Distribution) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if distribution == nil {
			return fmt.Errorf("CloudFront Distribution empty")
		}

		if aws.StringValue(distribution.Status) != "Deployed" {
			return fmt.Errorf("CloudFront Distribution (%s) status not Deployed: %s", aws.StringValue(distribution.Id), aws.StringValue(distribution.Status))
		}

		return nil
	}
}

func testAccCheckCloudFrontDistributionStatusInProgress(distribution *cloudfront.Distribution) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if distribution == nil {
			return fmt.Errorf("CloudFront Distribution empty")
		}

		if aws.StringValue(distribution.Status) != "InProgress" {
			return fmt.Errorf("CloudFront Distribution (%s) status not InProgress: %s", aws.StringValue(distribution.Id), aws.StringValue(distribution.Status))
		}

		return nil
	}
}

func testAccCheckCloudFrontDistributionDisabled(distribution *cloudfront.Distribution) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if distribution == nil || distribution.DistributionConfig == nil {
			return fmt.Errorf("CloudFront Distribution configuration empty")
		}

		if aws.BoolValue(distribution.DistributionConfig.Enabled) {
			return fmt.Errorf("CloudFront Distribution (%s) enabled", aws.StringValue(distribution.Id))
		}

		return nil
	}
}

// testAccCheckCloudFrontDistributionDisappears deletes a CloudFront Distribution outside Terraform
// This requires the CloudFront Distribution to previously be disabled and fetches latest ETag automatically.
func testAccCheckCloudFrontDistributionDisappears(distribution *cloudfront.Distribution) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).cloudfrontconn

		getDistributionInput := &cloudfront.GetDistributionInput{
			Id: distribution.Id,
		}

		getDistributionOutput, err := conn.GetDistribution(getDistributionInput)

		if err != nil {
			return err
		}

		deleteDistributionInput := &cloudfront.DeleteDistributionInput{
			Id:      distribution.Id,
			IfMatch: getDistributionOutput.ETag,
		}

		err = resource.Retry(2*time.Minute, func() *resource.RetryError {
			_, err = conn.DeleteDistribution(deleteDistributionInput)

			if isAWSErr(err, cloudfront.ErrCodeDistributionNotDisabled, "") {
				return resource.RetryableError(err)
			}

			if isAWSErr(err, cloudfront.ErrCodeNoSuchDistribution, "") {
				return nil
			}

			if isAWSErr(err, cloudfront.ErrCodePreconditionFailed, "") {
				return resource.RetryableError(err)
			}

			if err != nil {
				return resource.NonRetryableError(err)
			}

			return nil
		})

		if isResourceTimeoutError(err) {
			_, err = conn.DeleteDistribution(deleteDistributionInput)
		}

		return err
	}
}

func testAccCheckCloudFrontDistributionWaitForDeployment(distribution *cloudfront.Distribution) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return resourceAwsCloudFrontDistributionWaitUntilDeployed(aws.StringValue(distribution.Id), testAccProvider.Meta())
	}
}

func testAccAWSCloudFrontDistributionRetainOnDeleteFromEnv() bool {
	_, ok := os.LookupEnv("TF_TEST_CLOUDFRONT_RETAIN")
	return ok
}

func testAccAWSCloudFrontDistributionRetainConfig() string {
	if testAccAWSCloudFrontDistributionRetainOnDeleteFromEnv() {
		return "retain_on_delete = true"
	}
	return ""
}

func TestAccAWSCloudFrontDistribution_OriginGroups(t *testing.T) {
	var distribution cloudfront.Distribution
	resourceName := "aws_cloudfront_distribution.failover_distribution"
	ri := acctest.RandInt()
	testConfig := fmt.Sprintf(testAccAWSCloudFrontDistributionOriginGroupsConfig, ri, originBucket, backupBucket, testAccAWSCloudFrontDistributionRetainConfig())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontDistributionExists(resourceName, &distribution),
					resource.TestCheckResourceAttr(resourceName, "origin_group.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "origin_group.*", map[string]string{
						"origin_id":                          "groupS3",
						"failover_criteria.#":                "1",
						"failover_criteria.0.status_codes.#": "4",
						"member.#":                           "2",
						"member.0.origin_id":                 "primaryS3",
						"member.1.origin_id":                 "failoverS3",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "origin_group.*.failover_criteria.0.status_codes.*", "403"),
					resource.TestCheckTypeSetElemAttr(resourceName, "origin_group.*.failover_criteria.0.status_codes.*", "404"),
					resource.TestCheckTypeSetElemAttr(resourceName, "origin_group.*.failover_criteria.0.status_codes.*", "500"),
					resource.TestCheckTypeSetElemAttr(resourceName, "origin_group.*.failover_criteria.0.status_codes.*", "502"),
				),
			},
		},
	})
}

var originBucket = `
resource "aws_s3_bucket" "s3_bucket_origin" {
  bucket = "mybucket.${var.rand_id}"
  acl    = "public-read"
}
`

var backupBucket = `
resource "aws_s3_bucket" "s3_backup_bucket_origin" {
  bucket = "mybucket-backup.${var.rand_id}"
  acl    = "public-read"
}
`

var logBucket = `
resource "aws_s3_bucket" "s3_bucket_logs" {
  bucket = "mylogs.${var.rand_id}"
  acl    = "public-read"
}
`

var testAccAWSCloudFrontDistributionS3Config = `
variable rand_id {
  default = %d
}

# origin bucket
%s

# log bucket
%s

resource "aws_cloudfront_distribution" "s3_distribution" {
  origin {
    domain_name = "${aws_s3_bucket.s3_bucket_origin.id}.s3.amazonaws.com"
    origin_id   = "myS3Origin"
  }

  enabled             = true
  default_root_object = "index.html"

  logging_config {
    include_cookies = false
    bucket          = "${aws_s3_bucket.s3_bucket_logs.id}.s3.amazonaws.com"
    prefix          = "myprefix"
  }

  default_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "myS3Origin"

    forwarded_values {
      query_string = false

      cookies {
        forward = "none"
      }
    }

    viewer_protocol_policy = "allow-all"
    min_ttl                = 0
    default_ttl            = 3600
    max_ttl                = 86400
  }

  price_class = "PriceClass_200"

  restrictions {
    geo_restriction {
      restriction_type = "whitelist"
      locations        = ["US", "CA", "GB", "DE"]
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }

  %s
}
`

var testAccAWSCloudFrontDistributionS3ConfigWithTags = `
variable rand_id {
  default = %d
}

# origin bucket
%s

# log bucket
%s

resource "aws_cloudfront_distribution" "s3_distribution" {
  origin {
    domain_name = "${aws_s3_bucket.s3_bucket_origin.id}.s3.amazonaws.com"
    origin_id   = "myS3Origin"
  }

  enabled             = true
  default_root_object = "index.html"

  default_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "myS3Origin"

    forwarded_values {
      query_string = false

      cookies {
        forward = "none"
      }
    }

    viewer_protocol_policy = "allow-all"
    min_ttl                = 0
    default_ttl            = 3600
    max_ttl                = 86400
  }

  price_class = "PriceClass_200"

  restrictions {
    geo_restriction {
      restriction_type = "whitelist"
      locations        = ["US", "CA", "GB", "DE"]
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }

  tags = {
    environment = "production"
    account     = "main"
  }

  %s
}
`

var testAccAWSCloudFrontDistributionS3ConfigWithTagsUpdated = `
variable rand_id {
  default = %d
}

# origin bucket
%s

# log bucket
%s

resource "aws_cloudfront_distribution" "s3_distribution" {
  origin {
    domain_name = "${aws_s3_bucket.s3_bucket_origin.id}.s3.amazonaws.com"
    origin_id   = "myS3Origin"
  }

  enabled             = true
  default_root_object = "index.html"

  default_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "myS3Origin"

    forwarded_values {
      query_string = false

      cookies {
        forward = "none"
      }
    }

    viewer_protocol_policy = "allow-all"
    min_ttl                = 0
    default_ttl            = 3600
    max_ttl                = 86400
  }

  price_class = "PriceClass_200"

  restrictions {
    geo_restriction {
      restriction_type = "whitelist"
      locations        = ["US", "CA", "GB", "DE"]
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }

  tags = {
    environment = "dev"
  }

  %s
}
`

var testAccAWSCloudFrontDistributionCustomConfig = fmt.Sprintf(`
variable rand_id {
  default = %d
}

# log bucket
%s

resource "aws_cloudfront_distribution" "custom_distribution" {
  origin {
    domain_name = "www.example.com"
    origin_id   = "myCustomOrigin"

    custom_origin_config {
      http_port                = 80
      https_port               = 443
      origin_protocol_policy   = "http-only"
      origin_ssl_protocols     = ["SSLv3", "TLSv1"]
      origin_read_timeout      = 30
      origin_keepalive_timeout = 5
    }
  }

  enabled             = true
  comment             = "Some comment"
  default_root_object = "index.html"

  logging_config {
    include_cookies = false
    bucket          = "${aws_s3_bucket.s3_bucket_logs.id}.s3.amazonaws.com"
    prefix          = "myprefix"
  }

  default_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "myCustomOrigin"
    smooth_streaming = false

    forwarded_values {
      query_string = false

      cookies {
        forward = "all"
      }
    }

    viewer_protocol_policy = "allow-all"
    min_ttl                = 0
    default_ttl            = 3600
    max_ttl                = 86400
  }

  price_class = "PriceClass_200"

  restrictions {
    geo_restriction {
      restriction_type = "whitelist"
      locations        = ["US", "CA", "GB", "DE"]
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }

  %s
}
`, acctest.RandInt(), logBucket, testAccAWSCloudFrontDistributionRetainConfig())

var testAccAWSCloudFrontDistributionOriginRequestPolicyConfigDefault = fmt.Sprintf(`
variable rand_id {
  default = %[1]d
}

# log bucket
%[2]s

resource "aws_cloudfront_cache_policy" "example" {
  name        = "test-policy%[1]d"
  comment     = "test comment"
  default_ttl = 50
  max_ttl     = 100
  min_ttl     = 1
  parameters_in_cache_key_and_forwarded_to_origin {
    cookies_config {
      cookie_behavior = "whitelist"
      cookies {
        items = ["test"]
      }
    }
    headers_config {
      header_behavior = "whitelist"
      headers {
        items = ["test"]
      }
    }
    query_strings_config {
      query_string_behavior = "whitelist"
      query_strings {
        items = ["test"]
      }
    }
  }
}

resource "aws_cloudfront_origin_request_policy" "test_policy" {
  name    = "test-policy%[1]d"
  comment = "test comment"
  cookies_config {
    cookie_behavior = "whitelist"
    cookies {
      items = ["test"]
    }
  }
  headers_config {
    header_behavior = "whitelist"
    headers {
      items = ["test"]
    }
  }
  query_strings_config {
    query_string_behavior = "whitelist"
    query_strings {
      items = ["test"]
    }
  }
}

resource "aws_cloudfront_distribution" "custom_distribution" {
  origin {
    domain_name = "www.example.com"
    origin_id   = "myCustomOrigin"

    custom_origin_config {
      http_port                = 80
      https_port               = 443
      origin_protocol_policy   = "http-only"
      origin_ssl_protocols     = ["SSLv3", "TLSv1"]
      origin_read_timeout      = 30
      origin_keepalive_timeout = 5
    }
  }

  enabled             = true
  comment             = "Some comment"
  default_root_object = "index.html"

  logging_config {
    include_cookies = false
    bucket          = "${aws_s3_bucket.s3_bucket_logs.id}.s3.amazonaws.com"
    prefix          = "myprefix"
  }

  default_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "myCustomOrigin"
    smooth_streaming = false

    origin_request_policy_id = aws_cloudfront_origin_request_policy.test_policy.id
    cache_policy_id          = aws_cloudfront_cache_policy.example.id

    viewer_protocol_policy = "allow-all"
  }

  price_class = "PriceClass_200"

  restrictions {
    geo_restriction {
      restriction_type = "whitelist"
      locations        = ["US", "CA", "GB", "DE"]
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }
}
`, acctest.RandInt(), logBucket, testAccAWSCloudFrontDistributionRetainConfig())

var testAccAWSCloudFrontDistributionOriginRequestPolicyConfigOrdered = fmt.Sprintf(`
variable rand_id {
  default = %[1]d
}

# log bucket
%[2]s

resource "aws_cloudfront_cache_policy" "example" {
  name        = "test-policy%[1]d"
  comment     = "test comment"
  default_ttl = 50
  max_ttl     = 100
  min_ttl     = 1
  parameters_in_cache_key_and_forwarded_to_origin {
    cookies_config {
      cookie_behavior = "whitelist"
      cookies {
        items = ["test"]
      }
    }
    headers_config {
      header_behavior = "whitelist"
      headers {
        items = ["test"]
      }
    }
    query_strings_config {
      query_string_behavior = "whitelist"
      query_strings {
        items = ["test"]
      }
    }
  }
}

resource "aws_cloudfront_origin_request_policy" "test_policy" {
  name    = "test-policy%[1]d"
  comment = "test comment"
  cookies_config {
    cookie_behavior = "whitelist"
    cookies {
      items = ["test"]
    }
  }
  headers_config {
    header_behavior = "whitelist"
    headers {
      items = ["test"]
    }
  }
  query_strings_config {
    query_string_behavior = "whitelist"
    query_strings {
      items = ["test"]
    }
  }
}

resource "aws_cloudfront_distribution" "custom_distribution" {
  origin {
    domain_name = "www.example.com"
    origin_id   = "myCustomOrigin"

    custom_origin_config {
      http_port                = 80
      https_port               = 443
      origin_protocol_policy   = "http-only"
      origin_ssl_protocols     = ["SSLv3", "TLSv1"]
      origin_read_timeout      = 30
      origin_keepalive_timeout = 5
    }
  }

  enabled             = true
  comment             = "Some comment"
  default_root_object = "index.html"

  logging_config {
    include_cookies = false
    bucket          = "${aws_s3_bucket.s3_bucket_logs.id}.s3.amazonaws.com"
    prefix          = "myprefix"
  }

  default_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "myCustomOrigin"
    smooth_streaming = false

    origin_request_policy_id = aws_cloudfront_origin_request_policy.test_policy.id
    cache_policy_id          = aws_cloudfront_cache_policy.example.id

    viewer_protocol_policy = "allow-all"
  }

  ordered_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "myCustomOrigin"
    smooth_streaming = false
    path_pattern     = "/*"

    origin_request_policy_id = aws_cloudfront_origin_request_policy.test_policy.id
    cache_policy_id          = aws_cloudfront_cache_policy.example.id

    viewer_protocol_policy = "allow-all"
  }

  price_class = "PriceClass_200"

  restrictions {
    geo_restriction {
      restriction_type = "whitelist"
      locations        = ["US", "CA", "GB", "DE"]
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }
}
`, acctest.RandInt(), logBucket, testAccAWSCloudFrontDistributionRetainConfig())

var testAccAWSCloudFrontDistributionMultiOriginConfig = fmt.Sprintf(`
variable rand_id {
  default = %d
}

# origin bucket
%s

# log bucket
%s

resource "aws_cloudfront_distribution" "multi_origin_distribution" {
  origin {
    domain_name = "${aws_s3_bucket.s3_bucket_origin.id}.s3.amazonaws.com"
    origin_id   = "myS3Origin"
  }

  origin {
    domain_name = "www.example.com"
    origin_id   = "myCustomOrigin"

    custom_origin_config {
      http_port                = 80
      https_port               = 443
      origin_protocol_policy   = "http-only"
      origin_ssl_protocols     = ["SSLv3", "TLSv1"]
      origin_keepalive_timeout = 45
    }
  }

  enabled             = true
  comment             = "Some comment"
  default_root_object = "index.html"

  logging_config {
    include_cookies = false
    bucket          = "${aws_s3_bucket.s3_bucket_logs.id}.s3.amazonaws.com"
    prefix          = "myprefix"
  }

  default_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "myS3Origin"
    smooth_streaming = true

    forwarded_values {
      query_string = false

      cookies {
        forward = "all"
      }
    }

    min_ttl                = 100
    default_ttl            = 100
    max_ttl                = 100
    viewer_protocol_policy = "allow-all"
  }

  ordered_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "myS3Origin"

    forwarded_values {
      query_string = true

      cookies {
        forward = "none"
      }
    }

    min_ttl                = 50
    default_ttl            = 50
    max_ttl                = 50
    viewer_protocol_policy = "allow-all"
    path_pattern           = "images1/*.jpg"
  }

  ordered_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "myCustomOrigin"

    forwarded_values {
      query_string = true

      cookies {
        forward = "none"
      }
    }

    min_ttl                = 50
    default_ttl            = 50
    max_ttl                = 50
    viewer_protocol_policy = "allow-all"
    path_pattern           = "images2/*.jpg"
  }

  price_class = "PriceClass_All"

  custom_error_response {
    error_code            = 404
    response_page_path    = "/error-pages/404.html"
    response_code         = 200
    error_caching_min_ttl = 30
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }

  %s
}
`, acctest.RandInt(), originBucket, logBucket, testAccAWSCloudFrontDistributionRetainConfig())

var testAccAWSCloudFrontDistributionNoCustomErroResponseInfo = fmt.Sprintf(`
variable rand_id {
  default = %d
}

resource "aws_cloudfront_distribution" "no_custom_error_responses" {
  origin {
    domain_name = "www.example.com"
    origin_id   = "myCustomOrigin"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "http-only"
      origin_ssl_protocols   = ["SSLv3", "TLSv1"]
    }
  }

  enabled = true
  comment = "Some comment"

  default_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "myCustomOrigin"
    smooth_streaming = false

    forwarded_values {
      query_string = false

      cookies {
        forward = "all"
      }
    }

    viewer_protocol_policy = "allow-all"
    min_ttl                = 0
    default_ttl            = 3600
    max_ttl                = 86400
  }

  custom_error_response {
    error_code            = 404
    error_caching_min_ttl = 30
  }

  restrictions {
    geo_restriction {
      restriction_type = "whitelist"
      locations        = ["US", "CA", "GB", "DE"]
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }

  %s
}
`, acctest.RandInt(), testAccAWSCloudFrontDistributionRetainConfig())

var testAccAWSCloudFrontDistributionNoOptionalItemsConfig = fmt.Sprintf(`
variable rand_id {
  default = %d
}

resource "aws_cloudfront_distribution" "no_optional_items" {
  origin {
    domain_name = "www.example.com"
    origin_id   = "myCustomOrigin"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "http-only"
      origin_ssl_protocols   = ["SSLv3", "TLSv1"]
    }
  }

  enabled = true

  default_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "myCustomOrigin"
    smooth_streaming = false

    forwarded_values {
      query_string = false

      cookies {
        forward = "all"
      }
    }

    viewer_protocol_policy = "allow-all"
  }

  restrictions {
    geo_restriction {
      restriction_type = "whitelist"
      locations        = ["US", "CA", "GB", "DE"]
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }

  %s
}
`, acctest.RandInt(), testAccAWSCloudFrontDistributionRetainConfig())

var testAccAWSCloudFrontDistributionConfig_Origin_EmptyDomainName = fmt.Sprintf(`
resource "aws_cloudfront_distribution" "Origin_EmptyDomainName" {
  origin {
    domain_name = ""
    origin_id   = "myCustomOrigin"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "http-only"
      origin_ssl_protocols   = ["SSLv3", "TLSv1"]
    }
  }

  enabled = true

  default_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "myCustomOrigin"
    smooth_streaming = false

    forwarded_values {
      query_string = false

      cookies {
        forward = "all"
      }
    }

    viewer_protocol_policy = "allow-all"
  }

  restrictions {
    geo_restriction {
      restriction_type = "whitelist"
      locations        = ["US", "CA", "GB", "DE"]
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }

  %s
}
`, testAccAWSCloudFrontDistributionRetainConfig())

var testAccAWSCloudFrontDistributionConfig_Origin_EmptyOriginID = fmt.Sprintf(`
resource "aws_cloudfront_distribution" "Origin_EmptyOriginID" {
  origin {
    domain_name = "www.example.com"
    origin_id   = ""

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "http-only"
      origin_ssl_protocols   = ["SSLv3", "TLSv1"]
    }
  }

  enabled = true

  default_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "myCustomOrigin"
    smooth_streaming = false

    forwarded_values {
      query_string = false

      cookies {
        forward = "all"
      }
    }

    viewer_protocol_policy = "allow-all"
  }

  restrictions {
    geo_restriction {
      restriction_type = "whitelist"
      locations        = ["US", "CA", "GB", "DE"]
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }

  %s
}
`, testAccAWSCloudFrontDistributionRetainConfig())

var testAccAWSCloudFrontDistributionHTTP11Config = fmt.Sprintf(`
variable rand_id {
  default = %d
}

resource "aws_cloudfront_distribution" "http_1_1" {
  origin {
    domain_name = "www.example.com"
    origin_id   = "myCustomOrigin"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "http-only"
      origin_ssl_protocols   = ["SSLv3", "TLSv1"]
    }
  }

  enabled = true
  comment = "Some comment"

  default_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "myCustomOrigin"
    smooth_streaming = false

    forwarded_values {
      query_string = false

      cookies {
        forward = "all"
      }
    }

    viewer_protocol_policy = "allow-all"
    min_ttl                = 0
    default_ttl            = 3600
    max_ttl                = 86400
  }

  http_version = "http1.1"

  restrictions {
    geo_restriction {
      restriction_type = "whitelist"
      locations        = ["US", "CA", "GB", "DE"]
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }

  %s
}
`, acctest.RandInt(), testAccAWSCloudFrontDistributionRetainConfig())

var testAccAWSCloudFrontDistributionIsIPV6EnabledConfig = fmt.Sprintf(`
variable rand_id {
  default = %d
}

resource "aws_cloudfront_distribution" "is_ipv6_enabled" {
  origin {
    domain_name = "www.example.com"
    origin_id   = "myCustomOrigin"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "http-only"
      origin_ssl_protocols   = ["SSLv3", "TLSv1"]
    }
  }

  enabled         = true
  is_ipv6_enabled = true
  comment         = "Some comment"

  default_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "myCustomOrigin"
    smooth_streaming = false

    forwarded_values {
      query_string = false

      cookies {
        forward = "all"
      }
    }

    viewer_protocol_policy = "allow-all"
    min_ttl                = 0
    default_ttl            = 3600
    max_ttl                = 86400
  }

  http_version = "http1.1"

  restrictions {
    geo_restriction {
      restriction_type = "whitelist"
      locations        = ["US", "CA", "GB", "DE"]
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }

  %s
}
`, acctest.RandInt(), testAccAWSCloudFrontDistributionRetainConfig())

var testAccAWSCloudFrontDistributionOrderedCacheBehavior = fmt.Sprintf(`
variable rand_id {
  default = %d
}

resource "aws_cloudfront_distribution" "main" {
  origin {
    domain_name = "www.hashicorp.com"
    origin_id   = "myCustomOrigin"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "http-only"
      origin_ssl_protocols   = ["SSLv3", "TLSv1"]
    }
  }

  enabled = true
  comment = "Some comment"

  default_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "myCustomOrigin"
    smooth_streaming = true

    forwarded_values {
      query_string = false

      cookies {
        forward = "all"
      }
    }

    min_ttl                = 100
    default_ttl            = 100
    max_ttl                = 100
    viewer_protocol_policy = "allow-all"
  }

  ordered_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "myCustomOrigin"

    forwarded_values {
      query_string = true

      cookies {
        forward = "none"
      }
    }

    min_ttl                = 50
    default_ttl            = 50
    max_ttl                = 50
    viewer_protocol_policy = "allow-all"
    path_pattern           = "images1/*.jpg"
  }

  ordered_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "myCustomOrigin"

    forwarded_values {
      query_string = true

      cookies {
        forward = "none"
      }
    }

    min_ttl                = 51
    default_ttl            = 51
    max_ttl                = 51
    viewer_protocol_policy = "allow-all"
    path_pattern           = "images2/*.jpg"
  }

  price_class = "PriceClass_All"

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }

  %s
}
`, acctest.RandInt(), testAccAWSCloudFrontDistributionRetainConfig())

var testAccAWSCloudFrontDistributionOrderedCacheBehaviorCachePolicy = fmt.Sprintf(`
variable rand_id {
  default = %d
}

resource "aws_cloudfront_distribution" "main" {
  origin {
    domain_name = "www.hashicorp.com"
    origin_id   = "myCustomOrigin"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "http-only"
      origin_ssl_protocols   = ["SSLv3", "TLSv1"]
    }
  }

  enabled = true
  comment = "Some comment"

  default_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "myCustomOrigin"
    smooth_streaming = true

    forwarded_values {
      query_string = false

      cookies {
        forward = "all"
      }
    }

    viewer_protocol_policy = "allow-all"
  }

  ordered_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "myCustomOrigin"

    cache_policy_id = aws_cloudfront_cache_policy.cache_policy.id

    viewer_protocol_policy = "allow-all"
    path_pattern           = "images2/*.jpg"
  }

  price_class = "PriceClass_All"

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }

  %s
}

resource "aws_cloudfront_cache_policy" "cache_policy" {
  name        = "test-policy%[1]d"
  comment     = "test comment"
  default_ttl = 50
  max_ttl     = 100
  parameters_in_cache_key_and_forwarded_to_origin {
    cookies_config {
      cookie_behavior = "none"
    }
    headers_config {
      header_behavior = "none"
    }
    query_strings_config {
      query_string_behavior = "none"
    }
  }
}

`, acctest.RandInt(), testAccAWSCloudFrontDistributionRetainConfig())

var testAccAWSCloudFrontDistributionOriginGroupsConfig = `
variable rand_id {
  default = %d
}
# origin bucket
%s

# backup bucket
%s

resource "aws_cloudfront_distribution" "failover_distribution" {
  origin {
    domain_name = aws_s3_bucket.s3_bucket_origin.bucket_regional_domain_name
    origin_id   = "primaryS3"
  }

  origin {
    domain_name = aws_s3_bucket.s3_backup_bucket_origin.bucket_regional_domain_name
    origin_id   = "failoverS3"
  }

  origin_group {
    origin_id = "groupS3"

    failover_criteria {
      status_codes = [403, 404, 500, 502]
    }

    member {
      origin_id = "primaryS3"
    }

    member {
      origin_id = "failoverS3"
    }
  }

  enabled = true

  restrictions {
    geo_restriction {
      restriction_type = "whitelist"
      locations        = ["US", "CA", "GB", "DE"]
    }
  }

  default_cache_behavior {
    allowed_methods  = ["GET", "HEAD"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "groupS3"

    forwarded_values {
      query_string = false

      cookies {
        forward = "none"
      }
    }

    viewer_protocol_policy = "allow-all"
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }
  %s
}
`

func testAccAWSCloudFrontDistributionConfigDefaultCacheBehaviorForwardedValuesCookiesWhitelistedNamesUnordered2(retainOnDelete bool) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_distribution" "test" {
  # Faster acceptance testing
  enabled             = false
  retain_on_delete    = %[1]t
  wait_for_deployment = false

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"

    forwarded_values {
      query_string = false

      cookies {
        forward           = "whitelist"
        whitelisted_names = ["test2", "test1"]
      }
    }
  }

  origin {
    domain_name = "www.example.com"
    origin_id   = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }
}
`, retainOnDelete)
}

func testAccAWSCloudFrontDistributionConfigDefaultCacheBehaviorForwardedValuesCookiesWhitelistedNamesUnordered3(retainOnDelete bool) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_distribution" "test" {
  # Faster acceptance testing
  enabled             = false
  retain_on_delete    = %[1]t
  wait_for_deployment = false

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"

    forwarded_values {
      query_string = false

      cookies {
        forward           = "whitelist"
        whitelisted_names = ["test2", "test3", "test1"]
      }
    }
  }

  origin {
    domain_name = "www.example.com"
    origin_id   = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }
}
`, retainOnDelete)
}

func testAccAWSCloudFrontDistributionConfigDefaultCacheBehaviorForwardedValuesHeadersUnordered2(retainOnDelete bool) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_distribution" "test" {
  # Faster acceptance testing
  enabled             = false
  retain_on_delete    = %[1]t
  wait_for_deployment = false

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"

    forwarded_values {
      headers      = ["Origin", "Access-Control-Request-Headers"]
      query_string = false

      cookies {
        forward = "all"
      }
    }
  }

  origin {
    domain_name = "www.example.com"
    origin_id   = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }
}
`, retainOnDelete)
}

func testAccAWSCloudFrontDistributionConfigDefaultCacheBehaviorForwardedValuesHeadersUnordered3(retainOnDelete bool) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_distribution" "test" {
  # Faster acceptance testing
  enabled             = false
  retain_on_delete    = %[1]t
  wait_for_deployment = false

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"

    forwarded_values {
      headers      = ["Origin", "Access-Control-Request-Headers", "Access-Control-Request-Method"]
      query_string = false

      cookies {
        forward = "all"
      }
    }
  }

  origin {
    domain_name = "www.example.com"
    origin_id   = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }
}
`, retainOnDelete)
}

func testAccAWSCloudFrontDistributionConfigEnabled(enabled, retainOnDelete bool) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_distribution" "test" {
  enabled          = %[1]t
  retain_on_delete = %[2]t

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"

    forwarded_values {
      query_string = false

      cookies {
        forward = "all"
      }
    }
  }

  origin {
    domain_name = "www.example.com"
    origin_id   = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }
}
`, enabled, retainOnDelete)
}

func testAccAWSCloudFrontDistributionConfigOrderedCacheBehaviorForwardedValuesCookiesWhitelistedNamesUnordered2(retainOnDelete bool) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_distribution" "test" {
  # Faster acceptance testing
  enabled             = false
  retain_on_delete    = %[1]t
  wait_for_deployment = false

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"

    forwarded_values {
      query_string = false

      cookies {
        forward = "all"
      }
    }
  }

  ordered_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    path_pattern           = "/test/*"
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"

    forwarded_values {
      query_string = false

      cookies {
        forward           = "whitelist"
        whitelisted_names = ["test2", "test1"]
      }
    }
  }

  origin {
    domain_name = "www.example.com"
    origin_id   = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }
}
`, retainOnDelete)
}

func testAccAWSCloudFrontDistributionConfigOrderedCacheBehaviorForwardedValuesCookiesWhitelistedNamesUnordered3(retainOnDelete bool) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_distribution" "test" {
  # Faster acceptance testing
  enabled             = false
  retain_on_delete    = %[1]t
  wait_for_deployment = false

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"

    forwarded_values {
      query_string = false

      cookies {
        forward = "all"
      }
    }
  }

  ordered_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    path_pattern           = "/test/*"
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"

    forwarded_values {
      query_string = false

      cookies {
        forward           = "whitelist"
        whitelisted_names = ["test2", "test3", "test1"]
      }
    }
  }

  origin {
    domain_name = "www.example.com"
    origin_id   = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }
}
`, retainOnDelete)
}

func testAccAWSCloudFrontDistributionConfigOrderedCacheBehaviorForwardedValuesHeadersUnordered2(retainOnDelete bool) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_distribution" "test" {
  # Faster acceptance testing
  enabled             = false
  retain_on_delete    = %[1]t
  wait_for_deployment = false

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"

    forwarded_values {
      query_string = false

      cookies {
        forward = "all"
      }
    }
  }

  ordered_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    path_pattern           = "/test/*"
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"

    forwarded_values {
      headers      = ["Origin", "Access-Control-Request-Headers"]
      query_string = false

      cookies {
        forward = "all"
      }
    }
  }

  origin {
    domain_name = "www.example.com"
    origin_id   = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }
}
`, retainOnDelete)
}

func testAccAWSCloudFrontDistributionConfigOrderedCacheBehaviorForwardedValuesHeadersUnordered3(retainOnDelete bool) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_distribution" "test" {
  # Faster acceptance testing
  enabled             = false
  retain_on_delete    = %[1]t
  wait_for_deployment = false

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"

    forwarded_values {
      query_string = false

      cookies {
        forward = "all"
      }
    }
  }

  ordered_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    path_pattern           = "/test/*"
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"

    forwarded_values {
      headers      = ["Origin", "Access-Control-Request-Headers", "Access-Control-Request-Method"]
      query_string = false

      cookies {
        forward = "all"
      }
    }
  }

  origin {
    domain_name = "www.example.com"
    origin_id   = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }
}
`, retainOnDelete)
}

func testAccAWSCloudFrontDistributionConfigDefaultCacheBehaviorTrustedSignersSelf(retainOnDelete bool) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_distribution" "test" {
  # Faster acceptance testing
  enabled             = false
  retain_on_delete    = %[1]t
  wait_for_deployment = false

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "test"
    trusted_signers        = ["self"]
    viewer_protocol_policy = "allow-all"

    forwarded_values {
      query_string = false

      cookies {
        forward = "all"
      }
    }
  }

  origin {
    domain_name = "www.example.com"
    origin_id   = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }
}
`, retainOnDelete)
}

// CloudFront Distribution ACM Certificates must be created in us-east-1
func testAccAWSCloudFrontDistributionConfigViewerCertificateAcmCertificateArnBase(commonName string) string {
	key := tlsRsaPrivateKeyPem(2048)
	certificate := tlsRsaX509SelfSignedCertificatePem(key, commonName)

	return testAccCloudfrontRegionProviderConfig() + fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[1]s"
  private_key      = "%[2]s"
}
`, tlsPemEscapeNewlines(certificate), tlsPemEscapeNewlines(key))
}

func testAccAWSCloudFrontDistributionConfigViewerCertificateAcmCertificateArn(retainOnDelete bool) string {
	return testAccAWSCloudFrontDistributionConfigViewerCertificateAcmCertificateArnBase("example.com") + fmt.Sprintf(`
resource "aws_cloudfront_distribution" "test" {
  enabled          = false
  retain_on_delete = %t

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"

    forwarded_values {
      query_string = false

      cookies {
        forward = "all"
      }
    }
  }

  origin {
    domain_name = "www.example.com"
    origin_id   = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    acm_certificate_arn = aws_acm_certificate.test.arn
    ssl_support_method  = "sni-only"
  }
}
`, retainOnDelete)
}

func testAccAWSCloudFrontDistributionConfigViewerCertificateAcmCertificateArnConflictsWithCloudFrontDefaultCertificate(retainOnDelete bool) string {
	return testAccAWSCloudFrontDistributionConfigViewerCertificateAcmCertificateArnBase("example.com") + fmt.Sprintf(`
resource "aws_cloudfront_distribution" "test" {
  enabled          = false
  retain_on_delete = %t

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"

    forwarded_values {
      query_string = false

      cookies {
        forward = "all"
      }
    }
  }

  origin {
    domain_name = "www.example.com"
    origin_id   = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    acm_certificate_arn            = aws_acm_certificate.test.arn
    cloudfront_default_certificate = false
    ssl_support_method             = "sni-only"
  }
}
`, retainOnDelete)
}

func testAccAWSCloudFrontDistributionConfigWaitForDeployment(enabled, waitForDeployment bool) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_distribution" "test" {
  enabled             = %[1]t
  wait_for_deployment = %[2]t

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"

    forwarded_values {
      query_string = false

      cookies {
        forward = "all"
      }
    }
  }

  origin {
    domain_name = "www.example.com"
    origin_id   = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }
}
`, enabled, waitForDeployment)
}

func testAccAWSCloudFrontDistributionConfigCacheBehaviorRealtimeLogConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  name        = %[1]q
  shard_count = 2
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Principal": {
      "Service": "cloudfront.amazonaws.com"
    },
    "Effect": "Allow"
  }]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": [
      "kinesis:DescribeStreamSummary",
      "kinesis:DescribeStream",
      "kinesis:PutRecord",
      "kinesis:PutRecords"
    ],
    "Resource": "${aws_kinesis_stream.test.arn}"
  }]
}
EOF
}

resource "aws_cloudfront_realtime_log_config" "test" {
  name          = %[1]q
  sampling_rate = 50
  fields        = ["timestamp", "c-ip"]

  endpoint {
    stream_type = "Kinesis"

    kinesis_stream_config {
      role_arn   = aws_iam_role.test.arn
      stream_arn = aws_kinesis_stream.test.arn
    }
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName)
}

func testAccAWSCloudFrontDistributionConfigDefaultCacheBehaviorRealtimeLogConfigArn(rName string, retainOnDelete bool) string {
	return composeConfig(
		testAccAWSCloudFrontDistributionConfigCacheBehaviorRealtimeLogConfigBase(rName),
		fmt.Sprintf(`
resource "aws_cloudfront_distribution" "test" {
  # Faster acceptance testing
  enabled             = false
  retain_on_delete    = %[1]t
  wait_for_deployment = false

  default_cache_behavior {
    allowed_methods         = ["GET", "HEAD"]
    cached_methods          = ["GET", "HEAD"]
    target_origin_id        = "test"
    viewer_protocol_policy  = "allow-all"
    realtime_log_config_arn = aws_cloudfront_realtime_log_config.test.arn

    forwarded_values {
      query_string = false

      cookies {
        forward = "none"
      }
    }
  }

  origin {
    domain_name = "www.example.com"
    origin_id   = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }
}
`, retainOnDelete))
}

func testAccAWSCloudFrontDistributionConfigOrderedCacheBehaviorRealtimeLogConfigArn(rName string, retainOnDelete bool) string {
	return composeConfig(
		testAccAWSCloudFrontDistributionConfigCacheBehaviorRealtimeLogConfigBase(rName),
		fmt.Sprintf(`
resource "aws_cloudfront_distribution" "test" {
  # Faster acceptance testing
  enabled             = false
  retain_on_delete    = %[1]t
  wait_for_deployment = false

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"

    forwarded_values {
      query_string = false

      cookies {
        forward = "all"
      }
    }
  }

  ordered_cache_behavior {
    allowed_methods         = ["GET", "HEAD"]
    cached_methods          = ["GET", "HEAD"]
    path_pattern            = "/test/*"
    target_origin_id        = "test"
    viewer_protocol_policy  = "allow-all"
    realtime_log_config_arn = aws_cloudfront_realtime_log_config.test.arn

    forwarded_values {
      query_string = false

      cookies {
        forward = "none"
      }
    }
  }

  origin {
    domain_name = "www.example.com"
    origin_id   = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }
}
`, retainOnDelete))
}
