package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudfront/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudfront/lister"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func init() {
	RegisterServiceErrorCheckFunc(cloudfront.EndpointsID, testAccErrorCheckSkipFunction)

	resource.AddTestSweepers("aws_cloudfront_function", &resource.Sweeper{
		Name: "aws_cloudfront_function",
		F:    testSweepCloudfrontFunctions,
	})
}

func testAccErrorCheckSkipFunction(t *testing.T) resource.ErrorCheckFunc {
	return testAccErrorCheckSkipMessagesContaining(t,
		"InvalidParameterValueException: Unsupported source arn",
	)
}

func testSweepCloudfrontFunctions(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*AWSClient).cloudfrontconn
	input := &cloudfront.ListFunctionsInput{}
	var sweeperErrs *multierror.Error

	err = lister.ListFunctionsPages(conn, input, func(page *cloudfront.ListFunctionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, item := range page.FunctionList.Items {
			name := aws.StringValue(item.Name)

			output, err := finder.FunctionByNameAndStage(conn, name, cloudfront.FunctionStageDevelopment)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				sweeperErr := fmt.Errorf("error reading CloudFront Function (%s): %w", name, err)
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}

			r := resourceAwsCloudFrontFunction()
			d := r.Data(nil)
			d.SetId(name)
			d.Set("etag", output.ETag)

			err = r.Delete(d, client)

			if err != nil {
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping CloudFront Function sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing CloudFront Functions: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSCloudfrontFunction_basic(t *testing.T) {
	var conf cloudfront.DescribeFunctionOutput
	resourceName := "aws_cloudfront_function.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, cloudfront.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudfrontFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudfrontConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCloudfrontFunctionExists(resourceName, &conf),
					testAccCheckResourceAttrGlobalARN(resourceName, "arn", "cloudfront", fmt.Sprintf("function/%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "code"),
					resource.TestCheckResourceAttr(resourceName, "comment", ""),
					resource.TestCheckResourceAttr(resourceName, "etag", "ETVPDKIKX0DER"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "publish", "true"),
					resource.TestCheckResourceAttr(resourceName, "runtime", "cloudfront-js-1.0"),
					resource.TestCheckResourceAttr(resourceName, "status", "UNASSOCIATED"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"publish"},
			},
		},
	})
}

func TestAccAWSCloudfrontFunction_disappears(t *testing.T) {
	var conf cloudfront.DescribeFunctionOutput

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudfront_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, cloudfront.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudfrontFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudfrontConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCloudfrontFunctionExists(resourceName, &conf),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCloudFrontFunction(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSCloudfrontFunction_Publish(t *testing.T) {
	var conf cloudfront.DescribeFunctionOutput
	resourceName := "aws_cloudfront_function.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, cloudfront.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudfrontFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudfrontConfigPublish(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCloudfrontFunctionExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "publish", "false"),
					resource.TestCheckResourceAttr(resourceName, "status", "UNPUBLISHED"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"publish"},
			},
			{
				Config: testAccAWSCloudfrontConfigPublish(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCloudfrontFunctionExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "publish", "true"),
					resource.TestCheckResourceAttr(resourceName, "status", "UNASSOCIATED"),
				),
			},
		},
	})
}

// If you are testing manually and can't wait for deletion, set the
// TF_TEST_CLOUDFRONT_RETAIN environment variable.
func TestAccAWSCloudfrontFunction_Associated(t *testing.T) {
	var conf cloudfront.DescribeFunctionOutput
	resourceName := "aws_cloudfront_function.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, cloudfront.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudfrontFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudfrontConfigAssociated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCloudfrontFunctionExists(resourceName, &conf),
					// After creation the function will be in UNASSOCIATED status.
					// Apply the same configuration and it will move to DEPLOYED status.
					resource.TestCheckResourceAttr(resourceName, "status", "UNASSOCIATED"),
				),
			},
			{
				Config: testAccAWSCloudfrontConfigAssociated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCloudfrontFunctionExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "status", "DEPLOYED"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"publish"},
			},
			{
				Config: testAccAWSCloudfrontConfigUnassociated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCloudfrontFunctionExists(resourceName, &conf),
				),
			},
		},
	})
}

func TestAccAWSCloudfrontFunction_Update_Code(t *testing.T) {
	var conf cloudfront.DescribeFunctionOutput
	resourceName := "aws_cloudfront_function.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, cloudfront.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudfrontFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudfrontConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCloudfrontFunctionExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "etag", "ETVPDKIKX0DER"),
				),
			},
			{
				Config: testAccAWSCloudfrontConfigCodeUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCloudfrontFunctionExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "etag", "E3UN6WX5RRO2AG"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"publish"},
			},
		},
	})
}

func TestAccAWSCloudfrontFunction_Update_Comment(t *testing.T) {
	var conf cloudfront.DescribeFunctionOutput
	resourceName := "aws_cloudfront_function.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, cloudfront.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudfrontFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudfrontConfigComment(rName, "test 1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCloudfrontFunctionExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "comment", "test 1"),
				),
			},
			{
				Config: testAccAWSCloudfrontConfigComment(rName, "test 2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCloudfrontFunctionExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "comment", "test 2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"publish"},
			},
		},
	})
}

func testAccCheckCloudfrontFunctionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cloudfrontconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudfront_function" {
			continue
		}

		_, err := finder.FunctionByNameAndStage(conn, rs.Primary.ID, cloudfront.FunctionStageDevelopment)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("CloudFront Function %s still exists", rs.Primary.ID)
	}

	return nil

}

func testAccCheckAwsCloudfrontFunctionExists(n string, v *cloudfront.DescribeFunctionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Cloudfront Function not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Cloudfront Function ID not set")
		}

		conn := testAccProvider.Meta().(*AWSClient).cloudfrontconn

		output, err := finder.FunctionByNameAndStage(conn, rs.Primary.ID, cloudfront.FunctionStageDevelopment)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccAWSCloudfrontConfigBasic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_function" "test" {
  name    = %[1]q
  runtime = "cloudfront-js-1.0"
  code    = <<-EOT
function handler(event) {
	var response = {
		statusCode: 302,
		statusDescription: 'Found',
		headers: {
			'cloudfront-functions': { value: 'generated-by-CloudFront-Functions' },
			'location': { value: 'https://aws.amazon.com/cloudfront/' }
		}
	};
	return response;
}
EOT
}
`, rName)
}

func testAccAWSCloudfrontConfigPublish(rName string, publish bool) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_function" "test" {
  name    = %[1]q
  runtime = "cloudfront-js-1.0"
  code    = <<-EOT
function handler(event) {
	var response = {
		statusCode: 302,
		statusDescription: 'Found',
		headers: {
			'cloudfront-functions': { value: 'generated-by-CloudFront-Functions' },
			'location': { value: 'https://aws.amazon.com/cloudfront/' }
		}
	};
	return response;
}
EOT

  publish = %[2]t
}
`, rName, publish)
}

func testAccAWSCloudfrontConfigAssociated(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_distribution" "test" {
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

    function_association {
      event_type   = "viewer-request"
      function_arn = aws_cloudfront_function.test.arn
    }
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

  %[2]s
}

resource "aws_cloudfront_function" "test" {
  name    = %[1]q
  runtime = "cloudfront-js-1.0"
  code    = <<-EOT
function handler(event) {
	var response = {
		statusCode: 302,
		statusDescription: 'Found',
		headers: {
			'cloudfront-functions': { value: 'generated-by-CloudFront-Functions' },
			'location': { value: 'https://aws.amazon.com/cloudfront/' }
		}
	};
	return response;
}
EOT

  publish = true
}
`, rName, testAccAWSCloudFrontDistributionRetainConfig())
}

func testAccAWSCloudfrontConfigUnassociated(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_distribution" "test" {
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

  %[2]s
}

resource "aws_cloudfront_function" "test" {
  name    = %[1]q
  runtime = "cloudfront-js-1.0"
  code    = <<-EOT
function handler(event) {
	var response = {
		statusCode: 302,
		statusDescription: 'Found',
		headers: {
			'cloudfront-functions': { value: 'generated-by-CloudFront-Functions' },
			'location': { value: 'https://aws.amazon.com/cloudfront/' }
		}
	};
	return response;
}
EOT

  publish = true
}
`, rName, testAccAWSCloudFrontDistributionRetainConfig())
}

func testAccAWSCloudfrontConfigCodeUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_function" "test" {
  name    = %[1]q
  runtime = "cloudfront-js-1.0"
  code    = <<-EOT
function handler(event) {
	// updated code
	var response = {
		statusCode: 302,
		statusDescription: 'Found',
		headers: {
			'cloudfront-functions': { value: 'generated-by-CloudFront-Functions' },
			'location': { value: 'https://aws.amazon.com/cloudfront/' }
		}
	};
	return response;
}
EOT
}
`, rName)
}

func testAccAWSCloudfrontConfigComment(rName, comment string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_function" "test" {
  name    = %[1]q
  runtime = "cloudfront-js-1.0"
  comment = %[2]q
  code    = <<-EOT
function handler(event) {
	var response = {
		statusCode: 302,
		statusDescription: 'Found',
		headers: {
			'cloudfront-functions': { value: 'generated-by-CloudFront-Functions' },
			'location': { value: 'https://aws.amazon.com/cloudfront/' }
		}
	};
	return response;
}
EOT
}
`, rName, comment)
}
