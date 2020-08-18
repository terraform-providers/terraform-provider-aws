package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iotanalytics"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func generateChannelCustomerManagedS3Schema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
			},
			"key_prefix": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"role_arn": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func generateChannelServiceManagedS3Schema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{},
	}
}

func generateStorageSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"customer_managed_s3": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: []string{"storage.0.service_managed_s3"},
				Elem:          generateChannelCustomerManagedS3Schema(),
			},
			"service_managed_s3": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: []string{"storage.0.customer_managed_s3"},
				Elem:          generateChannelServiceManagedS3Schema(),
			},
		},
	}
}

func generateRetentionPeriodSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"number_of_days": {
				Type:          schema.TypeInt,
				Optional:      true,
				ConflictsWith: []string{"retention_period.0.unlimited"},
				ValidateFunc:  validation.IntAtLeast(1),
			},
			"unlimited": {
				Type:          schema.TypeBool,
				Optional:      true,
				ConflictsWith: []string{"retention_period.0.number_of_days"},
			},
		},
	}
}

func resourceAwsIotAnalyticsChannel() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsIotAnalyticsChannelCreate,
		Read:   resourceAwsIotAnalyticsChannelRead,
		Update: resourceAwsIotAnalyticsChannelUpdate,
		Delete: resourceAwsIotAnalyticsChannelDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"storage": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem:     generateStorageSchema(),
			},
			"retention_period": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem:     generateRetentionPeriodSchema(),
			},
			"tags": tagsSchema(),
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func parseChannelCustomerManagedS3(rawCustomerManagedS3 map[string]interface{}) *iotanalytics.CustomerManagedChannelS3Storage {
	bucket := rawCustomerManagedS3["bucket"].(string)
	roleArn := rawCustomerManagedS3["role_arn"].(string)
	customerManagedS3 := &iotanalytics.CustomerManagedChannelS3Storage{
		Bucket:  aws.String(bucket),
		RoleArn: aws.String(roleArn),
	}

	if v, ok := rawCustomerManagedS3["key_prefix"]; ok && len(v.(string)) >= 1 {
		customerManagedS3.KeyPrefix = aws.String(v.(string))
	}

	return customerManagedS3
}

func parseChannelServiceManagedS3(rawServiceManagedS3 map[string]interface{}) *iotanalytics.ServiceManagedChannelS3Storage {
	return &iotanalytics.ServiceManagedChannelS3Storage{}
}

func parseChannelStorage(rawChannelStorage map[string]interface{}) *iotanalytics.ChannelStorage {

	var customerManagedS3 *iotanalytics.CustomerManagedChannelS3Storage
	if list := rawChannelStorage["customer_managed_s3"].([]interface{}); len(list) > 0 {
		rawCustomerManagedS3 := list[0].(map[string]interface{})
		customerManagedS3 = parseChannelCustomerManagedS3(rawCustomerManagedS3)
	}

	var serviceManagedS3 *iotanalytics.ServiceManagedChannelS3Storage
	if list := rawChannelStorage["service_managed_s3"].([]interface{}); len(list) > 0 {
		switch rawServiceManagedS3 := list[0].(type) {
		case nil:
			serviceManagedS3 = parseChannelServiceManagedS3(make(map[string]interface{}))
		case map[string]interface{}:
			serviceManagedS3 = parseChannelServiceManagedS3(rawServiceManagedS3)
		}
	}

	return &iotanalytics.ChannelStorage{
		CustomerManagedS3: customerManagedS3,
		ServiceManagedS3:  serviceManagedS3,
	}
}

func parseRetentionPeriod(rawRetentionPeriod map[string]interface{}) *iotanalytics.RetentionPeriod {

	var numberOfDays *int64
	if v, ok := rawRetentionPeriod["number_of_days"]; ok && int64(v.(int)) > 1 {
		numberOfDays = aws.Int64(int64(v.(int)))
	}
	var unlimited *bool
	if v, ok := rawRetentionPeriod["unlimited"]; ok {
		unlimited = aws.Bool(v.(bool))
	}
	return &iotanalytics.RetentionPeriod{
		NumberOfDays: numberOfDays,
		Unlimited:    unlimited,
	}
}

func resourceAwsIotAnalyticsChannelCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotanalyticsconn

	params := &iotanalytics.CreateChannelInput{
		ChannelName: aws.String(d.Get("name").(string)),
	}

	if tags := d.Get("tags").(map[string]interface{}); len(tags) > 0 {
		params.Tags = keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().IotanalyticsTags()
	}

	channelStorageSet := d.Get("storage").(*schema.Set).List()
	if len(channelStorageSet) >= 1 {
		rawChannelStorage := channelStorageSet[0].(map[string]interface{})
		params.ChannelStorage = parseChannelStorage(rawChannelStorage)
	}

	retentionPeriodSet := d.Get("retention_period").(*schema.Set).List()
	if len(retentionPeriodSet) >= 1 {
		rawRetentionPeriod := retentionPeriodSet[0].(map[string]interface{})
		params.RetentionPeriod = parseRetentionPeriod(rawRetentionPeriod)
	}

	log.Printf("[DEBUG] Create IoTAnalytics Channel: %s", params)

	retrySecondsList := [6]int{1, 2, 5, 8, 10, 0}

	var err error

	// Primitive retry.
	// During testing channel, problem was detected.
	// When we try to create channel model and role arn that
	// will be assumed by channel during one apply we get:
	// 'Unable to assume role, role ARN' error. However if we run apply
	// second time(when all required resources are created) channel will be created successfully.
	// So we suppose that problem is that AWS return response of successful role arn creation before
	// process of creation is really ended, and then creation of channel model fails.
	for index, sleepSeconds := range retrySecondsList {
		_, err = conn.CreateChannel(params)
		if err == nil {
			break
		} else if err != nil && index != len(retrySecondsList)-1 {
			err = nil
		}
		time.Sleep(time.Duration(sleepSeconds) * time.Second)
	}

	if err != nil {
		return err
	}

	d.SetId(d.Get("name").(string))

	return resourceAwsIotAnalyticsChannelRead(d, meta)
}

func flattenChannelCustomerManagedS3(customerManagedS3 *iotanalytics.CustomerManagedChannelS3Storage) map[string]interface{} {
	if customerManagedS3 == nil {
		return nil
	}

	rawCustomerManagedS3 := make(map[string]interface{})

	rawCustomerManagedS3["bucket"] = aws.StringValue(customerManagedS3.Bucket)
	rawCustomerManagedS3["role_arn"] = aws.StringValue(customerManagedS3.RoleArn)

	if customerManagedS3.KeyPrefix != nil {
		rawCustomerManagedS3["key_prefix"] = aws.StringValue(customerManagedS3.KeyPrefix)
	}

	return rawCustomerManagedS3
}

func flattenChannelServiceManagedS3(serviceManagedS3 *iotanalytics.ServiceManagedChannelS3Storage) map[string]interface{} {
	if serviceManagedS3 == nil {
		return nil
	}

	rawServiceManagedS3 := make(map[string]interface{})
	return rawServiceManagedS3
}

func flattenChannelStorage(channelStorage *iotanalytics.ChannelStorage) map[string]interface{} {
	customerManagedS3 := flattenChannelCustomerManagedS3(channelStorage.CustomerManagedS3)
	serviceManagedS3 := flattenChannelServiceManagedS3(channelStorage.ServiceManagedS3)

	if customerManagedS3 == nil && serviceManagedS3 == nil {
		return nil
	}

	rawStorage := make(map[string]interface{})
	rawStorage["customer_managed_s3"] = wrapMapInList(customerManagedS3)
	rawStorage["service_managed_s3"] = wrapMapInList(serviceManagedS3)
	return rawStorage
}

func flattenRetentionPeriod(retentionPeriod *iotanalytics.RetentionPeriod) map[string]interface{} {
	rawRetentionPeriod := make(map[string]interface{})

	if retentionPeriod.NumberOfDays != nil {
		rawRetentionPeriod["number_of_days"] = aws.Int64Value(retentionPeriod.NumberOfDays)
	}
	if retentionPeriod.Unlimited != nil {
		rawRetentionPeriod["unlimited"] = aws.BoolValue(retentionPeriod.Unlimited)
	}

	return rawRetentionPeriod
}

func wrapMapInList(mapping map[string]interface{}) []interface{} {
	if mapping == nil {
		return make([]interface{}, 0)
	} else {
		return []interface{}{mapping}
	}
}

func resourceAwsIotAnalyticsChannelRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotanalyticsconn

	params := &iotanalytics.DescribeChannelInput{
		ChannelName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading IoTAnalytics Channel: %s", params)

	out, err := conn.DescribeChannel(params)

	if err != nil {
		return err
	}

	d.Set("name", out.Channel.Name)
	storage := flattenChannelStorage(out.Channel.Storage)
	d.Set("storage", wrapMapInList(storage))
	retentionPeriod := flattenRetentionPeriod(out.Channel.RetentionPeriod)
	d.Set("retention_period", wrapMapInList(retentionPeriod))
	d.Set("arn", out.Channel.Arn)

	arn := *out.Channel.Arn
	tags, err := keyvaluetags.IotanalyticsListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for resource (%s): %s", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}
	return nil
}

func resourceAwsIotAnalyticsChannelUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotanalyticsconn

	params := &iotanalytics.UpdateChannelInput{
		ChannelName: aws.String(d.Get("name").(string)),
	}

	channelStorageSet := d.Get("storage").(*schema.Set).List()
	if len(channelStorageSet) >= 1 {
		rawChannelStorage := channelStorageSet[0].(map[string]interface{})
		params.ChannelStorage = parseChannelStorage(rawChannelStorage)
	}

	retentionPeriodSet := d.Get("retention_period").(*schema.Set).List()
	if len(retentionPeriodSet) >= 1 {
		rawRetentionPeriod := retentionPeriodSet[0].(map[string]interface{})
		params.RetentionPeriod = parseRetentionPeriod(rawRetentionPeriod)
	}

	log.Printf("[DEBUG] Updating IoTAnalytics Channel: %s", params)

	retrySecondsList := [6]int{1, 2, 5, 8, 10, 0}

	var err error

	// Primitive retry.
	// Full explanation can be found in function `resourceAwsIotAnalyticsChannelCreate`.
	// We suppose that such error can appear during update also, if you update
	// role arn.
	for index, sleepSeconds := range retrySecondsList {
		_, err = conn.UpdateChannel(params)
		if err == nil {
			break
		} else if err != nil && index != len(retrySecondsList)-1 {
			err = nil
		}
		time.Sleep(time.Duration(sleepSeconds) * time.Second)
	}

	if err != nil {
		return err
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.IotanalyticsUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsIotAnalyticsChannelRead(d, meta)
}

func resourceAwsIotAnalyticsChannelDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotanalyticsconn

	params := &iotanalytics.DeleteChannelInput{
		ChannelName: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Delete IoTAnalytics Channel: %s", params)
	_, err := conn.DeleteChannel(params)

	return err
}
