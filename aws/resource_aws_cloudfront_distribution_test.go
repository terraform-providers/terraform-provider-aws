package aws

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
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

func TestAccAWSCloudFrontDistribution_importBasic(t *testing.T) {
	ri := acctest.RandInt()
	testConfig := fmt.Sprintf(testAccAWSCloudFrontDistributionS3Config, ri, originBucket, logBucket, testAccAWSCloudFrontDistributionRetainConfig())

	resourceName := "aws_cloudfront_distribution.s3_distribution"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testConfig,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
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
	ri := acctest.RandInt()
	testConfig := fmt.Sprintf(testAccAWSCloudFrontDistributionS3Config, ri, originBucket, logBucket, testAccAWSCloudFrontDistributionRetainConfig())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontDistributionExistence(
						"aws_cloudfront_distribution.s3_distribution",
					),
					resource.TestCheckResourceAttr(
						"aws_cloudfront_distribution.s3_distribution",
						"hosted_zone_id",
						"Z2FDTNDATAQYW2",
					),
				),
			},
		},
	})
}

func TestAccAWSCloudFrontDistribution_S3OriginWithTags(t *testing.T) {
	ri := acctest.RandInt()
	preConfig := fmt.Sprintf(testAccAWSCloudFrontDistributionS3ConfigWithTags, ri, originBucket, logBucket, testAccAWSCloudFrontDistributionRetainConfig())
	postConfig := fmt.Sprintf(testAccAWSCloudFrontDistributionS3ConfigWithTagsUpdated, ri, originBucket, logBucket, testAccAWSCloudFrontDistributionRetainConfig())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontDistributionExistence(
						"aws_cloudfront_distribution.s3_distribution",
					),
					resource.TestCheckResourceAttr(
						"aws_cloudfront_distribution.s3_distribution", "tags.%", "2"),
					resource.TestCheckResourceAttr(
						"aws_cloudfront_distribution.s3_distribution", "tags.environment", "production"),
					resource.TestCheckResourceAttr(
						"aws_cloudfront_distribution.s3_distribution", "tags.account", "main"),
				),
			},
			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontDistributionExistence(
						"aws_cloudfront_distribution.s3_distribution",
					),
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
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontDistributionCustomConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontDistributionExistence(
						"aws_cloudfront_distribution.custom_distribution",
					),
				),
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
	resourceName := "aws_cloudfront_distribution.multi_origin_distribution"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontDistributionMultiOriginConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontDistributionExistence(resourceName),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior.2106187217.default_ttl", "50"),
					resource.TestCheckResourceAttr(resourceName, "cache_behavior.2106187217.path_pattern", "images1/*.jpg"),
				),
			},
		},
	})
}

// https://github.com/terraform-providers/terraform-provider-aws/issues/188
// TestAccAWSCloudFrontDistribution_orderedCacheBehavior runs an
// aws_cloudfront_distribution acceptance test with multiple and ordered cache behaviors.
//
// If you are testing manually and can't wait for deletion, set the
// TF_TEST_CLOUDFRONT_RETAIN environment variable.
func TestAccAWSCloudFrontDistribution_orderedCacheBehavior(t *testing.T) {
	resourceName := "aws_cloudfront_distribution.main"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontDistributionOrderedCacheBehavior,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontDistributionExistence(resourceName),
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
			},
		},
	})
}

func TestAccAWSCloudFrontDistribution_Origin_EmptyDomainName(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
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
		PreCheck:     func() { testAccPreCheck(t) },
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
	resourceName := "aws_cloudfront_distribution.no_optional_items"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontDistributionNoOptionalItemsConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontDistributionExistence(resourceName),
					resource.TestCheckResourceAttr(resourceName, "aliases.#", "0"),
					resource.TestMatchResourceAttr(resourceName, "arn", regexp.MustCompile(`^arn:[^:]+:cloudfront::[^:]+:distribution/[A-Z0-9]+$`)),
					resource.TestCheckResourceAttr(resourceName, "custom_error_response.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.2871742778.allowed_methods.#", "7"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.2871742778.allowed_methods.0", "HEAD"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.2871742778.allowed_methods.1", "DELETE"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.2871742778.allowed_methods.2", "POST"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.2871742778.allowed_methods.3", "GET"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.2871742778.allowed_methods.4", "OPTIONS"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.2871742778.allowed_methods.5", "PUT"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.2871742778.allowed_methods.6", "PATCH"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.2871742778.cached_methods.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.2871742778.cached_methods.0", "HEAD"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.2871742778.cached_methods.1", "GET"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.2871742778.compress", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.2871742778.default_ttl", "86400"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.2871742778.forwarded_values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.2871742778.forwarded_values.2510654351.cookies.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.2871742778.forwarded_values.2510654351.cookies.1870923232.forward", "all"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.2871742778.forwarded_values.2510654351.cookies.1870923232.whitelisted_names.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.2871742778.forwarded_values.2510654351.headers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.2871742778.forwarded_values.2510654351.query_string", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.2871742778.forwarded_values.2510654351.query_string_cache_keys.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.2871742778.lambda_function_association.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.2871742778.max_ttl", "31536000"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.2871742778.min_ttl", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.2871742778.smooth_streaming", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.2871742778.target_origin_id", "myCustomOrigin"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.2871742778.trusted_signers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_cache_behavior.2871742778.viewer_protocol_policy", "allow-all"),
					resource.TestMatchResourceAttr(resourceName, "domain_name", regexp.MustCompile(`^[a-z0-9]+\.cloudfront\.net$`)),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestMatchResourceAttr(resourceName, "etag", regexp.MustCompile(`^[A-Z0-9]+$`)),
					resource.TestCheckResourceAttr(resourceName, "hosted_zone_id", "Z2FDTNDATAQYW2"),
					resource.TestCheckResourceAttrSet(resourceName, "http_version"),
					resource.TestCheckResourceAttr(resourceName, "is_ipv6_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "logging_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "origin.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "origin.1857972443.custom_header.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "origin.1857972443.custom_origin_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "origin.1857972443.custom_origin_config.3832794885.http_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "origin.1857972443.custom_origin_config.3832794885.https_port", "443"),
					resource.TestCheckResourceAttr(resourceName, "origin.1857972443.custom_origin_config.3832794885.origin_keepalive_timeout", "5"),
					resource.TestCheckResourceAttr(resourceName, "origin.1857972443.custom_origin_config.3832794885.origin_protocol_policy", "http-only"),
					resource.TestCheckResourceAttr(resourceName, "origin.1857972443.custom_origin_config.3832794885.origin_read_timeout", "30"),
					resource.TestCheckResourceAttr(resourceName, "origin.1857972443.custom_origin_config.3832794885.origin_ssl_protocols.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "origin.1857972443.custom_origin_config.3832794885.origin_ssl_protocols.0", "SSLv3"),
					resource.TestCheckResourceAttr(resourceName, "origin.1857972443.custom_origin_config.3832794885.origin_ssl_protocols.1", "TLSv1"),
					resource.TestCheckResourceAttr(resourceName, "origin.1857972443.domain_name", "www.example.com"),
					resource.TestCheckResourceAttr(resourceName, "price_class", "PriceClass_All"),
					resource.TestCheckResourceAttr(resourceName, "restrictions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "restrictions.811065190.geo_restriction.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "restrictions.811065190.geo_restriction.1303118592.locations.#", "4"),
					resource.TestCheckResourceAttr(resourceName, "restrictions.811065190.geo_restriction.1303118592.locations.0", "CA"),
					resource.TestCheckResourceAttr(resourceName, "restrictions.811065190.geo_restriction.1303118592.locations.1", "DE"),
					resource.TestCheckResourceAttr(resourceName, "restrictions.811065190.geo_restriction.1303118592.locations.2", "GB"),
					resource.TestCheckResourceAttr(resourceName, "restrictions.811065190.geo_restriction.1303118592.locations.3", "US"),
					resource.TestCheckResourceAttr(resourceName, "restrictions.811065190.geo_restriction.1303118592.restriction_type", "whitelist"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "viewer_certificate.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "viewer_certificate.69840937.cloudfront_default_certificate", "true"),
				),
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
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontDistributionHTTP11Config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontDistributionExistence(
						"aws_cloudfront_distribution.http_1_1",
					),
				),
			},
		},
	})
}

func TestAccAWSCloudFrontDistribution_IsIPV6EnabledConfig(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontDistributionIsIPV6EnabledConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontDistributionExistence(
						"aws_cloudfront_distribution.is_ipv6_enabled",
					),
					resource.TestCheckResourceAttr(
						"aws_cloudfront_distribution.is_ipv6_enabled", "is_ipv6_enabled", "true"),
				),
			},
		},
	})
}

func TestAccAWSCloudFrontDistribution_noCustomErrorResponseConfig(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontDistributionNoCustomErroResponseInfo,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontDistributionExistence(
						"aws_cloudfront_distribution.no_custom_error_responses",
					),
				),
			},
		},
	})
}

func testAccCheckCloudFrontDistributionDestroy(s *terraform.State) error {
	for k, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudfront_distribution" {
			continue
		}
		dist, err := testAccAuxCloudFrontGetDistributionConfig(s, k)
		if err == nil {
			if _, ok := os.LookupEnv("TF_TEST_CLOUDFRONT_RETAIN"); ok {
				if *dist.DistributionConfig.Enabled {
					return fmt.Errorf("CloudFront distribution should be disabled")
				}
				return nil
			}
			return fmt.Errorf("CloudFront distribution did not destroy")
		}
	}
	return nil
}

func testAccCheckCloudFrontDistributionExistence(cloudFrontResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := testAccAuxCloudFrontGetDistributionConfig(s, cloudFrontResource)

		return err
	}
}

func testAccAuxCloudFrontGetDistributionConfig(s *terraform.State, cloudFrontResource string) (*cloudfront.Distribution, error) {
	cf, ok := s.RootModule().Resources[cloudFrontResource]
	if !ok {
		return nil, fmt.Errorf("Not found: %s", cloudFrontResource)
	}

	if cf.Primary.ID == "" {
		return nil, fmt.Errorf("No Id is set")
	}

	cloudfrontconn := testAccProvider.Meta().(*AWSClient).cloudfrontconn

	req := &cloudfront.GetDistributionInput{
		Id: aws.String(cf.Primary.ID),
	}

	res, err := cloudfrontconn.GetDistribution(req)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving CloudFront distribution: %s", err)
	}

	return res.Distribution, nil
}

func testAccAWSCloudFrontDistributionRetainConfig() string {
	if _, ok := os.LookupEnv("TF_TEST_CLOUDFRONT_RETAIN"); ok {
		return "retain_on_delete = true"
	}
	return ""
}

var originBucket = fmt.Sprintf(`
resource "aws_s3_bucket" "s3_bucket_origin" {
	bucket = "mybucket.${var.rand_id}"
	acl = "public-read"
}
`)

var logBucket = fmt.Sprintf(`
resource "aws_s3_bucket" "s3_bucket_logs" {
	bucket = "mylogs.${var.rand_id}"
	acl = "public-read"
}
`)

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
		origin_id = "myS3Origin"
	}
	enabled = true
	default_root_object = "index.html"
	logging_config {
		include_cookies = false
		bucket = "${aws_s3_bucket.s3_bucket_logs.id}.s3.amazonaws.com"
		prefix = "myprefix"
	}
	aliases = [ "mysite.${var.rand_id}.example.com", "yoursite.${var.rand_id}.example.com" ]
	default_cache_behavior {
		allowed_methods = [ "DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT" ]
		cached_methods = [ "GET", "HEAD" ]
		target_origin_id = "myS3Origin"
		forwarded_values {
			query_string = false
			cookies {
				forward = "none"
			}
		}
		viewer_protocol_policy = "allow-all"
		min_ttl = 0
		default_ttl = 3600
		max_ttl = 86400
	}
	price_class = "PriceClass_200"
	restrictions {
		geo_restriction {
			restriction_type = "whitelist"
			locations = [ "US", "CA", "GB", "DE" ]
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
		origin_id = "myS3Origin"
	}
	enabled = true
	default_root_object = "index.html"
	aliases = [ "mysite.${var.rand_id}.example.com", "yoursite.${var.rand_id}.example.com" ]
	default_cache_behavior {
		allowed_methods = [ "DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT" ]
		cached_methods = [ "GET", "HEAD" ]
		target_origin_id = "myS3Origin"
		forwarded_values {
			query_string = false
			cookies {
				forward = "none"
			}
		}
		viewer_protocol_policy = "allow-all"
		min_ttl = 0
		default_ttl = 3600
		max_ttl = 86400
	}
	price_class = "PriceClass_200"
	restrictions {
		geo_restriction {
			restriction_type = "whitelist"
			locations = [ "US", "CA", "GB", "DE" ]
		}
	}
	viewer_certificate {
		cloudfront_default_certificate = true
	}
	tags = {
            environment = "production"
            account = "main"
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
		origin_id = "myS3Origin"
	}
	enabled = true
	default_root_object = "index.html"
	aliases = [ "mysite.${var.rand_id}.example.com", "yoursite.${var.rand_id}.example.com" ]
	default_cache_behavior {
		allowed_methods = [ "DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT" ]
		cached_methods = [ "GET", "HEAD" ]
		target_origin_id = "myS3Origin"
		forwarded_values {
			query_string = false
			cookies {
				forward = "none"
			}
		}
		viewer_protocol_policy = "allow-all"
		min_ttl = 0
		default_ttl = 3600
		max_ttl = 86400
	}
	price_class = "PriceClass_200"
	restrictions {
		geo_restriction {
			restriction_type = "whitelist"
			locations = [ "US", "CA", "GB", "DE" ]
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
		origin_id = "myCustomOrigin"
		custom_origin_config {
			http_port = 80
			https_port = 443
			origin_protocol_policy = "http-only"
			origin_ssl_protocols = [ "SSLv3", "TLSv1" ]
			origin_read_timeout = 30
			origin_keepalive_timeout = 5
		}
	}
	enabled = true
	comment = "Some comment"
	default_root_object = "index.html"
	logging_config {
		include_cookies = false
		bucket = "${aws_s3_bucket.s3_bucket_logs.id}.s3.amazonaws.com"
		prefix = "myprefix"
	}
	aliases = [ "mysite.${var.rand_id}.example.com", "*.yoursite.${var.rand_id}.example.com" ]
	default_cache_behavior {
		allowed_methods = [ "DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT" ]
		cached_methods = [ "GET", "HEAD" ]
		target_origin_id = "myCustomOrigin"
		smooth_streaming = false
		forwarded_values {
			query_string = false
			cookies {
				forward = "all"
			}
		}
		viewer_protocol_policy = "allow-all"
		min_ttl = 0
		default_ttl = 3600
		max_ttl = 86400
	}
	price_class = "PriceClass_200"
	restrictions {
		geo_restriction {
			restriction_type = "whitelist"
			locations = [ "US", "CA", "GB", "DE" ]
		}
	}
	viewer_certificate {
		cloudfront_default_certificate = true
	}
	%s
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
		origin_id = "myS3Origin"
	}
	origin {
		domain_name = "www.example.com"
		origin_id = "myCustomOrigin"
		custom_origin_config {
			http_port = 80
			https_port = 443
			origin_protocol_policy = "http-only"
			origin_ssl_protocols = [ "SSLv3", "TLSv1" ]
			origin_keepalive_timeout = 45
		}
	}
	enabled = true
	comment = "Some comment"
	default_root_object = "index.html"
	logging_config {
		include_cookies = false
		bucket = "${aws_s3_bucket.s3_bucket_logs.id}.s3.amazonaws.com"
		prefix = "myprefix"
	}
	aliases = [ "mysite.${var.rand_id}.example.com", "*.yoursite.${var.rand_id}.example.com" ]
	default_cache_behavior {
		allowed_methods = [ "DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT" ]
		cached_methods = [ "GET", "HEAD" ]
		target_origin_id = "myS3Origin"
		smooth_streaming = true
		forwarded_values {
			query_string = false
			cookies {
				forward = "all"
			}
		}
		min_ttl = 100
		default_ttl = 100
		max_ttl = 100
		viewer_protocol_policy = "allow-all"
	}
	cache_behavior {
		allowed_methods = [ "DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT" ]
		cached_methods = [ "GET", "HEAD" ]
		target_origin_id = "myS3Origin"
		forwarded_values {
			query_string = true
			cookies {
				forward = "none"
			}
		}
		min_ttl = 50
		default_ttl = 50
		max_ttl = 50
		viewer_protocol_policy = "allow-all"
		path_pattern = "images1/*.jpg"
	}
	cache_behavior {
		allowed_methods = [ "DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT" ]
		cached_methods = [ "GET", "HEAD" ]
		target_origin_id = "myCustomOrigin"
		forwarded_values {
			query_string = true
			cookies {
				forward = "none"
			}
		}
		min_ttl = 50
		default_ttl = 50
		max_ttl = 50
		viewer_protocol_policy = "allow-all"
		path_pattern = "images2/*.jpg"
	}

	price_class = "PriceClass_All"
	custom_error_response {
		error_code = 404
		response_page_path = "/error-pages/404.html"
		response_code = 200
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
		origin_id = "myCustomOrigin"
		custom_origin_config {
			http_port = 80
			https_port = 443
			origin_protocol_policy = "http-only"
			origin_ssl_protocols = [ "SSLv3", "TLSv1" ]
		}
	}
	enabled = true
	comment = "Some comment"
	default_cache_behavior {
		allowed_methods = [ "DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT" ]
		cached_methods = [ "GET", "HEAD" ]
		target_origin_id = "myCustomOrigin"
		smooth_streaming = false
		forwarded_values {
			query_string = false
			cookies {
				forward = "all"
			}
		}
		viewer_protocol_policy = "allow-all"
		min_ttl = 0
		default_ttl = 3600
		max_ttl = 86400
	}
	custom_error_response {
		error_code = 404
		error_caching_min_ttl = 30
	}
	restrictions {
		geo_restriction {
			restriction_type = "whitelist"
			locations = [ "US", "CA", "GB", "DE" ]
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
		origin_id = "myCustomOrigin"
		custom_origin_config {
			http_port = 80
			https_port = 443
			origin_protocol_policy = "http-only"
			origin_ssl_protocols = [ "SSLv3", "TLSv1" ]
		}
	}
	enabled = true
	default_cache_behavior {
		allowed_methods = [ "DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT" ]
		cached_methods = [ "GET", "HEAD" ]
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
			locations = [ "US", "CA", "GB", "DE" ]
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
      origin_ssl_protocols   = [ "SSLv3", "TLSv1" ]
    }
  }
  enabled = true
  default_cache_behavior {
    allowed_methods  = [ "DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT" ]
    cached_methods   = [ "GET", "HEAD" ]
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
      locations        = [ "US", "CA", "GB", "DE" ]
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
      origin_ssl_protocols   = [ "SSLv3", "TLSv1" ]
    }
  }
  enabled = true
  default_cache_behavior {
    allowed_methods  = [ "DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT" ]
    cached_methods   = [ "GET", "HEAD" ]
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
      locations        = [ "US", "CA", "GB", "DE" ]
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
		origin_id = "myCustomOrigin"
		custom_origin_config {
			http_port = 80
			https_port = 443
			origin_protocol_policy = "http-only"
			origin_ssl_protocols = [ "SSLv3", "TLSv1" ]
		}
	}
	enabled = true
	comment = "Some comment"
	default_cache_behavior {
		allowed_methods = [ "DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT" ]
		cached_methods = [ "GET", "HEAD" ]
		target_origin_id = "myCustomOrigin"
		smooth_streaming = false
		forwarded_values {
			query_string = false
			cookies {
				forward = "all"
			}
		}
		viewer_protocol_policy = "allow-all"
		min_ttl = 0
		default_ttl = 3600
		max_ttl = 86400
	}
	http_version = "http1.1"
	restrictions {
		geo_restriction {
			restriction_type = "whitelist"
			locations = [ "US", "CA", "GB", "DE" ]
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
		origin_id = "myCustomOrigin"
		custom_origin_config {
			http_port = 80
			https_port = 443
			origin_protocol_policy = "http-only"
			origin_ssl_protocols = [ "SSLv3", "TLSv1" ]
		}
	}
	enabled = true
	is_ipv6_enabled = true
	comment = "Some comment"
	default_cache_behavior {
		allowed_methods = [ "DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT" ]
		cached_methods = [ "GET", "HEAD" ]
		target_origin_id = "myCustomOrigin"
		smooth_streaming = false
		forwarded_values {
			query_string = false
			cookies {
				forward = "all"
			}
		}
		viewer_protocol_policy = "allow-all"
		min_ttl = 0
		default_ttl = 3600
		max_ttl = 86400
	}
	http_version = "http1.1"
	restrictions {
		geo_restriction {
			restriction_type = "whitelist"
			locations = [ "US", "CA", "GB", "DE" ]
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
		origin_id = "myCustomOrigin"
		custom_origin_config {
			http_port = 80
			https_port = 443
			origin_protocol_policy = "http-only"
			origin_ssl_protocols = [ "SSLv3", "TLSv1" ]
		}
	}
	enabled = true
	comment = "Some comment"
	aliases = []
	default_cache_behavior {
		allowed_methods = [ "DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT" ]
		cached_methods = [ "GET", "HEAD" ]
		target_origin_id = "myCustomOrigin"
		smooth_streaming = true
		forwarded_values {
			query_string = false
			cookies {
				forward = "all"
			}
		}
		min_ttl = 100
		default_ttl = 100
		max_ttl = 100
		viewer_protocol_policy = "allow-all"
	}
	ordered_cache_behavior {
		allowed_methods = [ "DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT" ]
		cached_methods = [ "GET", "HEAD" ]
		target_origin_id = "myCustomOrigin"
		forwarded_values {
			query_string = true
			cookies {
				forward = "none"
			}
		}
		min_ttl = 50
		default_ttl = 50
		max_ttl = 50
		viewer_protocol_policy = "allow-all"
		path_pattern = "images1/*.jpg"
	}
	ordered_cache_behavior {
		allowed_methods = [ "DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT" ]
		cached_methods = [ "GET", "HEAD" ]
		target_origin_id = "myCustomOrigin"
		forwarded_values {
			query_string = true
			cookies {
				forward = "none"
			}
		}
		min_ttl = 51
		default_ttl = 51
		max_ttl = 51
		viewer_protocol_policy = "allow-all"
		path_pattern = "images2/*.jpg"
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
