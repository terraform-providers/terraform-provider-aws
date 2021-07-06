package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudfront/finder"
)

func init() {
	resource.AddTestSweepers("aws_cloudfront_monitoring_subscription", &resource.Sweeper{
		Name: "aws_cloudfront_monitoring_subscription",
		F:    testSweepCloudFrontMonitoringSubscriptions,
	})
}

func testSweepCloudFrontMonitoringSubscriptions(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).cloudfrontconn
	var sweeperErrs *multierror.Error

	distributionSummaries := make([]*cloudfront.DistributionSummary, 0)

	input := &cloudfront.ListDistributionsInput{}
	err = conn.ListDistributionsPages(input, func(page *cloudfront.ListDistributionsOutput, lastPage bool) bool {
		distributionSummaries = append(distributionSummaries, page.DistributionList.Items...)
		return !lastPage
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping CloudFront Monitoring Subscriptions sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("error listing CloudFront Distributions: %s", err)
	}

	if len(distributionSummaries) == 0 {
		log.Print("[DEBUG] No CloudFront Distributions found")
		return nil
	}

	for _, distributionSummary := range distributionSummaries {

		_, err := conn.GetMonitoringSubscription(&cloudfront.GetMonitoringSubscriptionInput{
			DistributionId: distributionSummary.Id,
		})
		if err != nil {
			return fmt.Errorf("error reading CloudFront Monitoring Subscription %s: %s", aws.StringValue(distributionSummary.Id), err)
		}

		_, err = conn.DeleteMonitoringSubscription(&cloudfront.DeleteMonitoringSubscriptionInput{
			DistributionId: distributionSummary.Id,
		})
		if err != nil {
			return fmt.Errorf("error deleting CloudFront Monitoring Subscription %s: %s", aws.StringValue(distributionSummary.Id), err)
		}
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSCloudFrontMonitoringSubscription_basic(t *testing.T) {
	var v cloudfront.MonitoringSubscription
	resourceName := "aws_cloudfront_monitoring_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontMonitoringSubscriptionDestroy,
		ErrorCheck:   testAccErrorCheck(t, cloudfront.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontMonitoringSubscriptionConfig("Enabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontMonitoringSubscriptionExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "distribution_id"),
					resource.TestCheckResourceAttr(resourceName, "monitoring_subscription.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "monitoring_subscription.0.realtime_metrics_subscription_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "monitoring_subscription.0.realtime_metrics_subscription_config.0.realtime_metrics_subscription_status", "Enabled"),
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

func TestAccAWSCloudFrontMonitoringSubscription_disappears(t *testing.T) {
	var v cloudfront.MonitoringSubscription
	resourceName := "aws_cloudfront_monitoring_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontMonitoringSubscriptionDestroy,
		ErrorCheck:   testAccErrorCheck(t, cloudfront.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontMonitoringSubscriptionConfig("Enabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontMonitoringSubscriptionExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCloudFrontMonitoringSubscription(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSCloudFrontMonitoringSubscription_update(t *testing.T) {
	var v cloudfront.MonitoringSubscription
	resourceName := "aws_cloudfront_monitoring_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontMonitoringSubscriptionDestroy,
		ErrorCheck:   testAccErrorCheck(t, cloudfront.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontMonitoringSubscriptionConfig("Enabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontMonitoringSubscriptionExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "distribution_id"),
					resource.TestCheckResourceAttr(resourceName, "monitoring_subscription.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "monitoring_subscription.0.realtime_metrics_subscription_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "monitoring_subscription.0.realtime_metrics_subscription_config.0.realtime_metrics_subscription_status", "Enabled"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCloudFrontMonitoringSubscriptionConfig("Disabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontMonitoringSubscriptionExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "distribution_id"),
					resource.TestCheckResourceAttr(resourceName, "monitoring_subscription.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "monitoring_subscription.0.realtime_metrics_subscription_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "monitoring_subscription.0.realtime_metrics_subscription_config.0.realtime_metrics_subscription_status", "Disabled"),
				),
			},
		},
	})
}

func testAccCheckCloudFrontMonitoringSubscriptionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cloudfrontconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudfront_monitoring_subscription" {
			continue
		}

		s, err := finder.MonitoringSubscriptionByDistributionId(conn, rs.Primary.ID)

		if isAWSErr(err, cloudfront.ErrCodeNoSuchDistribution, "") {
			continue
		}
		if err != nil {
			return err
		}
		if s != nil {
			continue
		}
		return fmt.Errorf("CloudFront Monitoring Subscription still exists: %s", rs.Primary.ID)
	}

	return nil
}

func testAccCheckCloudFrontMonitoringSubscriptionExists(n string, v *cloudfront.MonitoringSubscription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No CloudFront Monitoring Subscription ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).cloudfrontconn
		out, err := finder.MonitoringSubscriptionByDistributionId(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*v = *out

		return nil
	}
}

func testAccAWSCloudFrontMonitoringSubscriptionConfigBase() string {
	return `
resource "aws_cloudfront_distribution" "test" {
  enabled          = true
  retain_on_delete = false

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
`
}

func testAccAWSCloudFrontMonitoringSubscriptionConfig(status string) string {
	return composeConfig(
		testAccAWSCloudFrontMonitoringSubscriptionConfigBase(),
		fmt.Sprintf(`
resource "aws_cloudfront_monitoring_subscription" "test" {
  distribution_id = aws_cloudfront_distribution.test.id

  monitoring_subscription {
    realtime_metrics_subscription_config {
      realtime_metrics_subscription_status = %[1]q
    }
  }
}
`, status))
}
