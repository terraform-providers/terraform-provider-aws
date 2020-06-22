package aws

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codestarnotifications"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsCodeStarNotificationsNotificationRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCodeStarNotificationsNotificationRuleCreate,
		Read:   resourceAwsCodeStarNotificationsNotificationRuleRead,
		Update: resourceAwsCodeStarNotificationsNotificationRuleUpdate,
		Delete: resourceAwsCodeStarNotificationsNotificationRuleDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"detail_type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					codestarnotifications.DetailTypeBasic,
					codestarnotifications.DetailTypeFull,
				}, false),
			},

			"event_type_ids": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				MinItems: 1,
				MaxItems: 200,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexp.MustCompile(`^[A-Za-z0-9\-_ ]+$`), "must be one or more alphanumeric, hyphen, underscore or space characters"),
				),
			},

			"resource": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},

			"status": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  codestarnotifications.NotificationRuleStatusEnabled,
				ValidateFunc: validation.StringInSlice([]string{
					codestarnotifications.NotificationRuleStatusEnabled,
					codestarnotifications.NotificationRuleStatusDisabled,
				}, false),
			},

			"tags": tagsSchema(),

			"target": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 10,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateArn,
						},
						"type": {
							Type:     schema.TypeString,
							Default:  "SNS",
							Optional: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func expandCodeStarNotificationsNotificationRuleTargets(targetsData []interface{}) []*codestarnotifications.Target {
	targets := make([]*codestarnotifications.Target, 0, len(targetsData))
	for _, t := range targetsData {
		target := t.(map[string]interface{})
		targets = append(targets, &codestarnotifications.Target{
			TargetAddress: aws.String(target["address"].(string)),
			TargetType:    aws.String(target["type"].(string)),
		})
	}
	return targets
}

func resourceAwsCodeStarNotificationsNotificationRuleCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codestarnotificationsconn

	params := &codestarnotifications.CreateNotificationRuleInput{
		DetailType:   aws.String(d.Get("detail_type").(string)),
		EventTypeIds: expandStringList(d.Get("event_type_ids").(*schema.Set).List()),
		Name:         aws.String(d.Get("name").(string)),
		Resource:     aws.String(d.Get("resource").(string)),
		Status:       aws.String(d.Get("status").(string)),
		Targets:      expandCodeStarNotificationsNotificationRuleTargets(d.Get("target").(*schema.Set).List()),
	}

	if v, ok := d.GetOk("tags"); ok {
		params.Tags = keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().CodestarnotificationsTags()
	}

	res, err := conn.CreateNotificationRule(params)
	if err != nil {
		return fmt.Errorf("error creating codestar notification rule: %s", err)
	}

	d.SetId(aws.StringValue(res.Arn))

	return resourceAwsCodeStarNotificationsNotificationRuleRead(d, meta)
}

func resourceAwsCodeStarNotificationsNotificationRuleRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codestarnotificationsconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	rule, err := conn.DescribeNotificationRule(&codestarnotifications.DescribeNotificationRuleInput{
		Arn: aws.String(d.Id()),
	})

	if err != nil {
		if isAWSErr(err, codestarnotifications.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] codestar notification rule (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading codestar notification rule: %s", err)
	}

	d.Set("arn", rule.Arn)
	d.Set("detail_type", rule.DetailType)
	eventTypeIds := make([]string, 0, len(rule.EventTypes))
	for _, et := range rule.EventTypes {
		eventTypeIds = append(eventTypeIds, aws.StringValue(et.EventTypeId))
	}
	if err := d.Set("event_type_ids", eventTypeIds); err != nil {
		return fmt.Errorf("error setting event_type_ids: %s", err)
	}
	d.Set("name", rule.Name)
	d.Set("status", rule.Status)
	d.Set("resource", rule.Resource)
	d.Set("tags", keyvaluetags.New(rule.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map())

	targets := make([]map[string]interface{}, 0, len(rule.Targets))
	for _, t := range rule.Targets {
		targets = append(targets, map[string]interface{}{
			"address": aws.StringValue(t.TargetAddress),
			"type":    aws.StringValue(t.TargetType),
			"status":  aws.StringValue(t.TargetStatus),
		})
	}
	if err = d.Set("target", targets); err != nil {
		return fmt.Errorf("error setting codestar notification target: %s", err)
	}

	return nil
}

const awsCodeStartNotificationsNotificationRuleErrorSubscribed = "The target cannot be deleted because it is subscribed to one or more notification rules."

// cleanupCodeStarNotificationsNotificationRuleTargets tries to remove unused notification targets. AWS API does not
// provide expicit way for creating targets, they are created on first subscription. Here we are trying to remove all
// unused targets which were unsubscribed from this notification rule.
func cleanupCodeStarNotificationsNotificationRuleTargets(conn *codestarnotifications.CodeStarNotifications, oldVal *schema.Set, newVal *schema.Set) error {
	removedTargets := oldVal
	if newVal != nil {
		removedTargets = oldVal.Difference(newVal)
	}

	targets := removedTargets.List()

	err := resource.Retry(time.Duration(5)*time.Second, func() *resource.RetryError {
		var (
			reterr error
			i      int
		)

		for _, t := range targets {
			target := t.(map[string]interface{})
			_, err := conn.DeleteTarget(&codestarnotifications.DeleteTargetInput{
				ForceUnsubscribeAll: aws.Bool(false),
				TargetAddress:       aws.String(target["address"].(string)),
			})
			if isAWSErr(err, codestarnotifications.ErrCodeValidationException, awsCodeStartNotificationsNotificationRuleErrorSubscribed) {
				reterr = err
				targets[i] = target
				i++
			} else if err != nil {
				return resource.NonRetryableError(err)
			}
		}
		targets = targets[:i]

		return resource.RetryableError(reterr)
	})

	if isAWSErr(err, codestarnotifications.ErrCodeValidationException, awsCodeStartNotificationsNotificationRuleErrorSubscribed) {
		err = nil
	}

	return err
}

func resourceAwsCodeStarNotificationsNotificationRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codestarnotificationsconn

	params := &codestarnotifications.UpdateNotificationRuleInput{
		Arn:          aws.String(d.Id()),
		DetailType:   aws.String(d.Get("detail_type").(string)),
		EventTypeIds: expandStringList(d.Get("event_type_ids").(*schema.Set).List()),
		Name:         aws.String(d.Get("name").(string)),
		Status:       aws.String(d.Get("status").(string)),
		Targets:      expandCodeStarNotificationsNotificationRuleTargets(d.Get("target").(*schema.Set).List()),
	}

	if _, err := conn.UpdateNotificationRule(params); err != nil {
		return fmt.Errorf("error updating codestar notification rule: %s", err)
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.CodestarnotificationsUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating codestar notification rule tags: %s", err)
		}
	}

	if d.HasChange("target") {
		o, n := d.GetChange("target")
		if err := cleanupCodeStarNotificationsNotificationRuleTargets(conn, o.(*schema.Set), n.(*schema.Set)); err != nil {
			return err
		}
	}

	return resourceAwsCodeStarNotificationsNotificationRuleRead(d, meta)
}

func resourceAwsCodeStarNotificationsNotificationRuleDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codestarnotificationsconn

	_, err := conn.DeleteNotificationRule(&codestarnotifications.DeleteNotificationRuleInput{
		Arn: aws.String(d.Id()),
	})

	if err != nil {
		return fmt.Errorf("error deleting codestar notification rule: %s", err)
	}

	if err = cleanupCodeStarNotificationsNotificationRuleTargets(conn, d.Get("target").(*schema.Set), nil); err != nil {
		return fmt.Errorf("error deleting codestar notification targets: %s", err)
	}

	return nil
}
