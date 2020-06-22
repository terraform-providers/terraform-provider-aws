package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAwsSesNotificationTopic() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSesNotificationTopicSet,
		Read:   resourceAwsSesNotificationTopicRead,
		Update: resourceAwsSesNotificationTopicSet,
		Delete: resourceAwsSesNotificationTopicDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"topic_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateArn,
			},

			"notification_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					ses.NotificationTypeBounce,
					ses.NotificationTypeComplaint,
					ses.NotificationTypeDelivery,
				}, false),
			},

			"identity": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			"include_original_headers": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func resourceAwsSesNotificationTopicSet(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesconn
	notification := d.Get("notification_type").(string)
	identity := d.Get("identity").(string)
	includeOriginalHeaders := d.Get("include_original_headers").(bool)

	setOpts := &ses.SetIdentityNotificationTopicInput{
		Identity:         aws.String(identity),
		NotificationType: aws.String(notification),
	}

	if v, ok := d.GetOk("topic_arn"); ok && v.(string) != "" {
		setOpts.SnsTopic = aws.String(v.(string))
	}

	d.SetId(fmt.Sprintf("%s|%s", identity, notification))

	log.Printf("[DEBUG] Setting SES Identity Notification Topic: %#v", setOpts)

	if _, err := conn.SetIdentityNotificationTopic(setOpts); err != nil {
		return fmt.Errorf("Error setting SES Identity Notification Topic: %s", err)
	}

	setHeadersOpts := &ses.SetIdentityHeadersInNotificationsEnabledInput{
		Identity:         aws.String(identity),
		NotificationType: aws.String(notification),
		Enabled:          aws.Bool(includeOriginalHeaders),
	}

	log.Printf("[DEBUG] Setting SES Identity Notification Topic Headers: %#v", setHeadersOpts)

	if _, err := conn.SetIdentityHeadersInNotificationsEnabled(setHeadersOpts); err != nil {
		return fmt.Errorf("Error setting SES Identity Notification Topic Headers Forwarding: %s", err)
	}

	return resourceAwsSesNotificationTopicRead(d, meta)
}

func resourceAwsSesNotificationTopicRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesconn

	identity, notificationType, err := decodeSesIdentityNotificationTopicId(d.Id())
	if err != nil {
		return err
	}

	d.Set("identity", identity)
	d.Set("notification_type", notificationType)

	getOpts := &ses.GetIdentityNotificationAttributesInput{
		Identities: []*string{aws.String(identity)},
	}

	log.Printf("[DEBUG] Reading SES Identity Notification Topic Attributes: %#v", getOpts)

	response, err := conn.GetIdentityNotificationAttributes(getOpts)

	if err != nil {
		return fmt.Errorf("Error reading SES Identity Notification Topic: %s", err)
	}

	d.Set("topic_arn", "")
	if response == nil {
		return nil
	}

	notificationAttributes, notificationAttributesOk := response.NotificationAttributes[identity]
	if !notificationAttributesOk {
		return nil
	}

	switch notificationType {
	case ses.NotificationTypeBounce:
		d.Set("topic_arn", notificationAttributes.BounceTopic)
		d.Set("include_original_headers", notificationAttributes.HeadersInBounceNotificationsEnabled)
	case ses.NotificationTypeComplaint:
		d.Set("topic_arn", notificationAttributes.ComplaintTopic)
		d.Set("include_original_headers", notificationAttributes.HeadersInComplaintNotificationsEnabled)
	case ses.NotificationTypeDelivery:
		d.Set("topic_arn", notificationAttributes.DeliveryTopic)
		d.Set("include_original_headers", notificationAttributes.HeadersInDeliveryNotificationsEnabled)
	}

	return nil
}

func resourceAwsSesNotificationTopicDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesconn

	identity, notificationType, err := decodeSesIdentityNotificationTopicId(d.Id())
	if err != nil {
		return err
	}

	setOpts := &ses.SetIdentityNotificationTopicInput{
		Identity:         aws.String(identity),
		NotificationType: aws.String(notificationType),
		SnsTopic:         nil,
	}

	log.Printf("[DEBUG] Deleting SES Identity Notification Topic: %#v", setOpts)

	if _, err := conn.SetIdentityNotificationTopic(setOpts); err != nil {
		return fmt.Errorf("Error deleting SES Identity Notification Topic: %s", err)
	}

	return resourceAwsSesNotificationTopicRead(d, meta)
}

func decodeSesIdentityNotificationTopicId(id string) (string, string, error) {
	parts := strings.Split(id, "|")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("Unexpected format of ID (%q), expected IDENTITY|TYPE", id)
	}
	return parts[0], parts[1], nil
}
