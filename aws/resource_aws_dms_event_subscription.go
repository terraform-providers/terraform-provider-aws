package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	dms "github.com/aws/aws-sdk-go/service/databasemigrationservice"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsDmsEventSubscription() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDmsEventSubscriptionCreate,
		Read:   resourceAwsDmsEventSubscriptionRead,
		Update: resourceAwsDmsEventSubscriptionUpdate,
		Delete: resourceAwsDmsEventSubscriptionDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
		},

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"event_categories": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Required: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"sns_topic_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},
			"source_ids": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				ForceNew: true,
				Optional: true,
			},
			"source_type": {
				Type:     schema.TypeString,
				Optional: true,
				// The API suppors modification but doing so loses all source_ids
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"replication-instance",
					"replication-task",
				}, false),
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsDmsEventSubscriptionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dmsconn

	request := &dms.CreateEventSubscriptionInput{
		Enabled:          aws.Bool(d.Get("enabled").(bool)),
		SnsTopicArn:      aws.String(d.Get("sns_topic_arn").(string)),
		SubscriptionName: aws.String(d.Get("name").(string)),
		SourceType:       aws.String(d.Get("source_type").(string)),
		Tags:             keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().DatabasemigrationserviceTags(),
	}

	if v, ok := d.GetOk("event_categories"); ok {
		request.EventCategories = expandStringList(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("source_ids"); ok {
		request.SourceIds = expandStringList(v.(*schema.Set).List())
	}

	_, err := conn.CreateEventSubscription(request)

	if err != nil {
		return fmt.Errorf("error creating DMS Event Subscription (%s): %w", d.Get("name").(string), err)
	}

	d.SetId(d.Get("name").(string))

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"creating", "modifying"},
		Target:     []string{"active"},
		Refresh:    resourceAwsDmsEventSubscriptionStateRefreshFunc(conn, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		MinTimeout: 10 * time.Second,
		Delay:      10 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for DMS Event Subscription (%s) creation: %w", d.Id(), err)
	}

	return resourceAwsDmsEventSubscriptionRead(d, meta)
}

func resourceAwsDmsEventSubscriptionUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dmsconn

	if d.HasChanges("enabled", "event_categories", "sns_topic_arn", "source_type") {
		request := &dms.ModifyEventSubscriptionInput{
			Enabled:          aws.Bool(d.Get("enabled").(bool)),
			SnsTopicArn:      aws.String(d.Get("sns_topic_arn").(string)),
			SubscriptionName: aws.String(d.Get("name").(string)),
			SourceType:       aws.String(d.Get("source_type").(string)),
		}

		if v, ok := d.GetOk("event_categories"); ok {
			request.EventCategories = expandStringList(v.(*schema.Set).List())
		}

		_, err := conn.ModifyEventSubscription(request)

		if err != nil {
			return fmt.Errorf("error updating DMS Event Subscription (%s): %w", d.Id(), err)
		}

		stateConf := &resource.StateChangeConf{
			Pending:    []string{"modifying"},
			Target:     []string{"active"},
			Refresh:    resourceAwsDmsEventSubscriptionStateRefreshFunc(conn, d.Id()),
			Timeout:    d.Timeout(schema.TimeoutUpdate),
			MinTimeout: 10 * time.Second,
			Delay:      10 * time.Second,
		}

		_, err = stateConf.WaitForState()
		if err != nil {
			return fmt.Errorf("error waiting for DMS Event Subscription (%s) modification: %w", d.Id(), err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.DatabasemigrationserviceUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating DMS Event Subscription (%s) tags: %s", d.Get("arn").(string), err)
		}
	}

	return resourceAwsDmsEventSubscriptionRead(d, meta)
}

func resourceAwsDmsEventSubscriptionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dmsconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	request := &dms.DescribeEventSubscriptionsInput{
		SubscriptionName: aws.String(d.Id()),
	}

	response, err := conn.DescribeEventSubscriptions(request)

	if isAWSErr(err, dms.ErrCodeResourceNotFoundFault, "") {
		log.Printf("[WARN] DMS event subscription (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error reading DMS event subscription: %s", err)
	}

	if response == nil || len(response.EventSubscriptionsList) == 0 || response.EventSubscriptionsList[0] == nil {
		log.Printf("[WARN] DMS event subscription (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	subscription := response.EventSubscriptionsList[0]

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "dms",
		Region:    meta.(*AWSClient).region,
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("es:%s", d.Id()),
	}.String()
	d.Set("arn", arn)

	d.Set("enabled", subscription.Enabled)
	d.Set("sns_topic_arn", subscription.SnsTopicArn)
	d.Set("source_type", subscription.SourceType)
	d.Set("name", d.Id())
	d.Set("event_categories", flattenStringList(subscription.EventCategoriesList))
	d.Set("source_ids", flattenStringList(subscription.SourceIdsList))

	tags, err := keyvaluetags.DatabasemigrationserviceListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for DMS Event Subscription (%s): %s", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsDmsEventSubscriptionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dmsconn

	request := &dms.DeleteEventSubscriptionInput{
		SubscriptionName: aws.String(d.Id()),
	}

	_, err := conn.DeleteEventSubscription(request)

	if isAWSErr(err, dms.ErrCodeResourceNotFoundFault, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting DMS Event Subscription (%s): %w", d.Id(), err)
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"deleting"},
		Target:     []string{},
		Refresh:    resourceAwsDmsEventSubscriptionStateRefreshFunc(conn, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		MinTimeout: 10 * time.Second,
		Delay:      10 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for DMS Event Subscription (%s) deletion: %w", d.Id(), err)
	}

	return nil
}

func resourceAwsDmsEventSubscriptionStateRefreshFunc(conn *dms.DatabaseMigrationService, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		v, err := conn.DescribeEventSubscriptions(&dms.DescribeEventSubscriptionsInput{
			SubscriptionName: aws.String(name),
		})

		if isAWSErr(err, dms.ErrCodeResourceNotFoundFault, "") {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if v == nil || len(v.EventSubscriptionsList) == 0 || v.EventSubscriptionsList[0] == nil {
			return nil, "", nil
		}

		return v, aws.StringValue(v.EventSubscriptionsList[0].Status), nil
	}
}
