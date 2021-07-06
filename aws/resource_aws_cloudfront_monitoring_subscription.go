package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudfront/finder"
)

func resourceAwsCloudFrontMonitoringSubscription() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCloudFrontMonitoringSubscriptionCreate,
		Read:   resourceAwsCloudFrontMonitoringSubscriptionRead,
		Update: resourceAwsCloudFrontMonitoringSubscriptionCreate,
		Delete: resourceAwsCloudFrontMonitoringSubscriptionDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsCloudFrontMonitoringSubscriptionImport,
		},

		Schema: map[string]*schema.Schema{
			"distribution_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"monitoring_subscription": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"realtime_metrics_subscription_config": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"realtime_metrics_subscription_status": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(cloudfront.RealtimeMetricsSubscriptionStatus_Values(), false),
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceAwsCloudFrontMonitoringSubscriptionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudfrontconn

	id := d.Get("distribution_id").(string)
	input := &cloudfront.CreateMonitoringSubscriptionInput{
		DistributionId:         aws.String(id),
		MonitoringSubscription: expandCloudFrontMonitoringSubscription(d.Get("monitoring_subscription").([]interface{})),
	}

	log.Printf("[DEBUG] Creating CloudFront Monitoring Subscription: %s", input)
	_, err := conn.CreateMonitoringSubscription(input)

	if err != nil {
		return fmt.Errorf("error creating CloudFront Monitoring Subscription (%s): %w", id, err)
	}

	d.SetId(id)

	return resourceAwsCloudFrontMonitoringSubscriptionRead(d, meta)
}

func resourceAwsCloudFrontMonitoringSubscriptionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudfrontconn

	subscription, err := finder.MonitoringSubscriptionByDistributionId(conn, d.Id())

	if tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeNoSuchDistribution) {
		log.Printf("[WARN] CloudFront Distribution (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading CloudFront Monitoring Subscription (%s): %w", d.Id(), err)
	}

	if subscription == nil {
		if d.IsNewResource() {
			return fmt.Errorf("error reading CloudFront Monitoring Subscription (%s): not found", d.Id())
		}
		log.Printf("[WARN] CloudFront Monitoring Subscription (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err := d.Set("monitoring_subscription", flattenCloudFrontMonitoringSubscription(subscription)); err != nil {
		return fmt.Errorf("error setting monitoring_subscription: %w", err)
	}

	return nil
}

func resourceAwsCloudFrontMonitoringSubscriptionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudfrontconn

	log.Printf("[DEBUG] Deleting CloudFront Monitoring Subscription (%s)", d.Id())
	_, err := conn.DeleteMonitoringSubscription(&cloudfront.DeleteMonitoringSubscriptionInput{
		DistributionId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeNoSuchDistribution) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting CloudFront Monitoring Subscription (%s): %w", d.Id(), err)
	}

	return nil
}

func expandCloudFrontMonitoringSubscription(vSubscription []interface{}) *cloudfront.MonitoringSubscription {
	if len(vSubscription) == 0 || vSubscription[0] == nil {
		return nil
	}

	mSubscription := vSubscription[0].(map[string]interface{})

	return &cloudfront.MonitoringSubscription{
		RealtimeMetricsSubscriptionConfig: expandCloudFrontRealtimeMetricsSubscriptionConfig(mSubscription["realtime_metrics_subscription_config"].([]interface{})),
	}
}

func expandCloudFrontRealtimeMetricsSubscriptionConfig(vConfig []interface{}) *cloudfront.RealtimeMetricsSubscriptionConfig {
	if len(vConfig) == 0 || vConfig[0] == nil {
		return nil
	}

	mConfig := vConfig[0].(map[string]interface{})

	return &cloudfront.RealtimeMetricsSubscriptionConfig{
		RealtimeMetricsSubscriptionStatus: aws.String(mConfig["realtime_metrics_subscription_status"].(string)),
	}
}

func flattenCloudFrontMonitoringSubscription(subscription *cloudfront.MonitoringSubscription) []interface{} {
	return []interface{}{map[string]interface{}{"realtime_metrics_subscription_config": flattenCloudFrontRealtimeMetricsSubscriptionConfig(subscription.RealtimeMetricsSubscriptionConfig)}}
}

func flattenCloudFrontRealtimeMetricsSubscriptionConfig(config *cloudfront.RealtimeMetricsSubscriptionConfig) []interface{} {
	return []interface{}{map[string]interface{}{"realtime_metrics_subscription_status": config.RealtimeMetricsSubscriptionStatus}}
}

func resourceAwsCloudFrontMonitoringSubscriptionImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	d.Set("distribution_id", d.Id())
	return []*schema.ResourceData{d}, nil
}
