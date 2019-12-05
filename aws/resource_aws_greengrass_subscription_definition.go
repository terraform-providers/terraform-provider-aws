package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/greengrass"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsGreengrassSubscriptionDefinition() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGreengrassSubscriptionDefinitionCreate,
		Read:   resourceAwsGreengrassSubscriptionDefinitionRead,
		Update: resourceAwsGreengrassSubscriptionDefinitionUpdate,
		Delete: resourceAwsGreengrassSubscriptionDefinitionDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"amzn_client_token": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"tags": tagsSchema(),
			"latest_definition_version_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subscription_definition_version": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subscription": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:     schema.TypeString,
										Required: true,
									},
									"source": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validateArn,
									},
									"target": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validateArn,
									},
									"subject": {
										Type:     schema.TypeString,
										Required: true,
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

func createSubscriptionDefinitionVersion(d *schema.ResourceData, conn *greengrass.Greengrass) error {
	var rawData map[string]interface{}
	if v := d.Get("subscription_definition_version").(*schema.Set).List(); len(v) == 0 {
		return nil
	} else {
		rawData = v[0].(map[string]interface{})
	}

	params := &greengrass.CreateSubscriptionDefinitionVersionInput{
		SubscriptionDefinitionId: aws.String(d.Id()),
	}

	if v := d.Get("amzn_client_token").(string); v != "" {
		params.AmznClientToken = aws.String(v)
	}

	subscriptions := make([]*greengrass.Subscription, 0)
	for _, subscriptionToCast := range rawData["subscription"].(*schema.Set).List() {
		rawSubscription := subscriptionToCast.(map[string]interface{})
		subscription := &greengrass.Subscription{
			Id:      aws.String(rawSubscription["id"].(string)),
			Source:  aws.String(rawSubscription["source"].(string)),
			Target:  aws.String(rawSubscription["target"].(string)),
			Subject: aws.String(rawSubscription["subject"].(string)),
		}
		subscriptions = append(subscriptions, subscription)
	}
	params.Subscriptions = subscriptions

	log.Printf("[DEBUG] Creating Greengrass Subscription Definition Version: %s", params)
	_, err := conn.CreateSubscriptionDefinitionVersion(params)

	if err != nil {
		return err
	}

	return nil
}

func resourceAwsGreengrassSubscriptionDefinitionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).greengrassconn

	params := &greengrass.CreateSubscriptionDefinitionInput{
		Name: aws.String(d.Get("name").(string)),
	}

	if tags := d.Get("tags").(map[string]interface{}); len(tags) > 0 {
		params.Tags = keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().GreengrassTags()
	}

	log.Printf("[DEBUG] Creating Greengrass Subscription Definition: %s", params)
	out, err := conn.CreateSubscriptionDefinition(params)
	if err != nil {
		return err
	}

	d.SetId(*out.Id)

	err = createSubscriptionDefinitionVersion(d, conn)

	if err != nil {
		return err
	}

	return resourceAwsGreengrassSubscriptionDefinitionRead(d, meta)
}

func setSubscriptionDefinitionVersion(latestVersion string, d *schema.ResourceData, conn *greengrass.Greengrass) error {
	params := &greengrass.GetSubscriptionDefinitionVersionInput{
		SubscriptionDefinitionId:        aws.String(d.Id()),
		SubscriptionDefinitionVersionId: aws.String(latestVersion),
	}

	out, err := conn.GetSubscriptionDefinitionVersion(params)

	if err != nil {
		return err
	}

	rawVersion := make(map[string]interface{})
	d.Set("latest_definition_version_arn", *out.Arn)

	rawSubscriptionList := make([]map[string]interface{}, 0)
	for _, subscription := range out.Definition.Subscriptions {
		rawSubscription := make(map[string]interface{})
		rawSubscription["id"] = *subscription.Id
		rawSubscription["source"] = *subscription.Source
		rawSubscription["target"] = *subscription.Target
		rawSubscription["subject"] = *subscription.Subject
		rawSubscriptionList = append(rawSubscriptionList, rawSubscription)
	}

	rawVersion["subscription"] = rawSubscriptionList

	d.Set("subscription_definition_version", []map[string]interface{}{rawVersion})

	return nil
}

func resourceAwsGreengrassSubscriptionDefinitionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).greengrassconn

	params := &greengrass.GetSubscriptionDefinitionInput{
		SubscriptionDefinitionId: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Reading Greengrass Subscription Definition: %s", params)
	out, err := conn.GetSubscriptionDefinition(params)

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Received Greengrass Subscription Definition: %s", out)

	d.Set("arn", out.Arn)
	d.Set("name", out.Name)

	arn := *out.Arn
	tags, err := keyvaluetags.GreengrassListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for resource (%s): %s", arn, err)
	}
	if err := d.Set("tags", tags.IgnoreAws().Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	if out.LatestVersion != nil {
		err = setSubscriptionDefinitionVersion(*out.LatestVersion, d, conn)

		if err != nil {
			return err
		}
	}

	return nil
}

func resourceAwsGreengrassSubscriptionDefinitionUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).greengrassconn

	params := &greengrass.UpdateSubscriptionDefinitionInput{
		Name:                     aws.String(d.Get("name").(string)),
		SubscriptionDefinitionId: aws.String(d.Id()),
	}

	_, err := conn.UpdateSubscriptionDefinition(params)
	if err != nil {
		return err
	}

	if d.HasChange("subscription_definition_version") {
		err = createSubscriptionDefinitionVersion(d, conn)
		if err != nil {
			return err
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.GreengrassUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsGreengrassSubscriptionDefinitionRead(d, meta)
}

func resourceAwsGreengrassSubscriptionDefinitionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).greengrassconn

	params := &greengrass.DeleteSubscriptionDefinitionInput{
		SubscriptionDefinitionId: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Deleting Greengrass Subscription Definition: %s", params)

	_, err := conn.DeleteSubscriptionDefinition(params)

	if err != nil {
		return err
	}

	return nil
}
