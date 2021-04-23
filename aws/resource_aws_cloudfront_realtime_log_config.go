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

func resourceAwsCloudFrontRealtimeLogConfig() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCloudFrontRealtimeLogConfigCreate,
		Read:   resourceAwsCloudFrontRealtimeLogConfigRead,
		Update: resourceAwsCloudFrontRealtimeLogConfigUpdate,
		Delete: resourceAwsCloudFrontRealtimeLogConfigDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"endpoint": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"stream_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"Kinesis"}, false),
						},

						"kinesis_stream_config": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"role_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validateArn,
									},

									"stream_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validateArn,
									},
								},
							},
						},
					},
				},
			},

			"fields": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"sampling_rate": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(1, 100),
			},
		},
	}
}

func resourceAwsCloudFrontRealtimeLogConfigCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudfrontconn

	name := d.Get("name").(string)
	input := &cloudfront.CreateRealtimeLogConfigInput{
		EndPoints:    expandCloudFrontEndPoints(d.Get("endpoint").([]interface{})),
		Fields:       expandStringSet(d.Get("fields").(*schema.Set)),
		Name:         aws.String(name),
		SamplingRate: aws.Int64(int64(d.Get("sampling_rate").(int))),
	}

	log.Printf("[DEBUG] Creating CloudFront Real-time Log Config: %s", input)
	output, err := conn.CreateRealtimeLogConfig(input)

	if err != nil {
		return fmt.Errorf("error creating CloudFront Real-time Log Config (%s): %w", name, err)
	}

	d.SetId(aws.StringValue(output.RealtimeLogConfig.ARN))

	return resourceAwsCloudFrontRealtimeLogConfigRead(d, meta)
}

func resourceAwsCloudFrontRealtimeLogConfigRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudfrontconn

	logConfig, err := finder.RealtimeLogConfigByARN(conn, d.Id())

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeNoSuchRealtimeLogConfig) {
		log.Printf("[WARN] CloudFront Real-time Log Config (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading CloudFront Real-time Log Config (%s): %w", d.Id(), err)
	}

	if logConfig == nil {
		if d.IsNewResource() {
			return fmt.Errorf("error reading CloudFront Real-time Log Config (%s): not found", d.Id())
		}
		log.Printf("[WARN] CloudFront Real-time Log Config (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("arn", logConfig.ARN)
	if err := d.Set("endpoint", flattenCloudFrontEndPoints(logConfig.EndPoints)); err != nil {
		return fmt.Errorf("error setting endpoint: %w", err)
	}
	if err := d.Set("fields", flattenStringSet(logConfig.Fields)); err != nil {
		return fmt.Errorf("error setting fields: %w", err)
	}
	d.Set("name", logConfig.Name)
	d.Set("sampling_rate", int(aws.Int64Value(logConfig.SamplingRate)))

	return nil
}

func resourceAwsCloudFrontRealtimeLogConfigUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudfrontconn

	input := &cloudfront.UpdateRealtimeLogConfigInput{
		ARN:          aws.String(d.Id()),
		EndPoints:    expandCloudFrontEndPoints(d.Get("endpoint").([]interface{})),
		Fields:       expandStringSet(d.Get("fields").((*schema.Set))),
		SamplingRate: aws.Int64(int64(d.Get("sampling_rate").(int))),
	}

	log.Printf("[DEBUG] Updating CloudFront Real-time Log Config: %s", input)
	_, err := conn.UpdateRealtimeLogConfig(input)

	if err != nil {
		return fmt.Errorf("error updating CloudFront Real-time Log Config (%s): %s", d.Id(), err)
	}

	return resourceAwsCloudFrontRealtimeLogConfigRead(d, meta)
}

func resourceAwsCloudFrontRealtimeLogConfigDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudfrontconn

	log.Printf("[DEBUG] Deleting CloudFront Real-time Log Config (%s)", d.Id())
	_, err := conn.DeleteRealtimeLogConfig(&cloudfront.DeleteRealtimeLogConfigInput{
		ARN: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeNoSuchRealtimeLogConfig) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Cloudfront Real-time Log Config (%s): %w", d.Id(), err)
	}

	return nil
}

func expandCloudFrontEndPoints(vEndpoints []interface{}) []*cloudfront.EndPoint {
	if len(vEndpoints) == 0 || vEndpoints[0] == nil {
		return nil
	}

	endpoints := []*cloudfront.EndPoint{}

	for _, vEndpoint := range vEndpoints {
		endpoint := &cloudfront.EndPoint{}

		mEndpoint := vEndpoint.(map[string]interface{})

		if vStreamType, ok := mEndpoint["stream_type"].(string); ok && vStreamType != "" {
			endpoint.StreamType = aws.String(vStreamType)
		}
		if vKinesisStreamConfig, ok := mEndpoint["kinesis_stream_config"].([]interface{}); ok && len(vKinesisStreamConfig) > 0 && vKinesisStreamConfig[0] != nil {
			kinesisStreamConfig := &cloudfront.KinesisStreamConfig{}

			mKinesisStreamConfig := vKinesisStreamConfig[0].(map[string]interface{})

			if vRoleArn, ok := mKinesisStreamConfig["role_arn"].(string); ok && vRoleArn != "" {
				kinesisStreamConfig.RoleARN = aws.String(vRoleArn)
			}
			if vStreamArn, ok := mKinesisStreamConfig["stream_arn"].(string); ok && vStreamArn != "" {
				kinesisStreamConfig.StreamARN = aws.String(vStreamArn)
			}

			endpoint.KinesisStreamConfig = kinesisStreamConfig
		}

		endpoints = append(endpoints, endpoint)
	}

	return endpoints
}

func flattenCloudFrontEndPoints(endpoints []*cloudfront.EndPoint) []interface{} {
	if endpoints == nil {
		return []interface{}{}
	}

	vEndpoints := []interface{}{}

	for _, endpoint := range endpoints {
		mEndpoint := map[string]interface{}{
			"stream_type": aws.StringValue(endpoint.StreamType),
		}

		if kinesisStreamConfig := endpoint.KinesisStreamConfig; kinesisStreamConfig != nil {
			mKinesisStreamConfig := map[string]interface{}{
				"role_arn":   aws.StringValue(kinesisStreamConfig.RoleARN),
				"stream_arn": aws.StringValue(kinesisStreamConfig.StreamARN),
			}

			mEndpoint["kinesis_stream_config"] = []interface{}{mKinesisStreamConfig}
		}

		vEndpoints = append(vEndpoints, mEndpoint)
	}

	return vEndpoints
}
