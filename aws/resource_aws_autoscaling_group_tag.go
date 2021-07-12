package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsAutoscalingGroupTag() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAutoscalingGroupTagCreate,
		Read:   resourceAwsAutoscalingGroupTagRead,
		Update: resourceAwsAutoscalingGroupTagUpdate,
		Delete: resourceAwsAutoscalingGroupTagDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"asg_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"tag": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
						"propagate_at_launch": {
							Type:     schema.TypeBool,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func extractAutoscalingGroupNameAndKeyFromAutoscalingGroupTagID(id string) (string, string, error) {
	parts := strings.SplitN(id, ",", 2)

	if len(parts) != 2 {
		return "", "", fmt.Errorf("Invalid resource ID; cannot look up resource: %s", id)
	}

	return parts[0], parts[1], nil
}

func resourceAwsAutoscalingGroupTagCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).autoscalingconn

	asgName := d.Get("asg_name").(string)
	tags := d.Get("tag").([]interface{})

	tag := tags[0].(map[string]interface{})
	key := tag["key"].(string)

	if err := keyvaluetags.AutoscalingUpdateTags(conn, asgName, autoscalingTagResourceTypeAutoScalingGroup, nil, tags); err != nil {
		return fmt.Errorf("error updating Autoscaling Tag (%s) for resource (%s): %w", key, asgName, err)
	}

	d.SetId(fmt.Sprintf("%s,%s", asgName, key))

	return resourceAwsAutoscalingGroupTagRead(d, meta)
}

func resourceAwsAutoscalingGroupTagRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).autoscalingconn
	asgName, key, err := extractAutoscalingGroupNameAndKeyFromAutoscalingGroupTagID(d.Id())

	if err != nil {
		return err
	}

	exists, tagData, err := keyvaluetags.AutoscalingGetTag(conn, asgName, autoscalingTagResourceTypeAutoScalingGroup, key)

	if err != nil {
		return fmt.Errorf("error reading Autoscaling Tag (%s) for resource (%s): %w", key, asgName, err)
	}

	if !exists {
		log.Printf("[WARN] Autoscaling Tag (%s) for resource (%s) not found, removing from state", key, asgName)
		d.SetId("")
		return nil
	}

	d.Set("asg_name", asgName)

	tag := map[string]interface{}{
		"key":   key,
		"value": tagData.Value,

		"propagate_at_launch": tagData.AdditionalBoolFields["PropagateAtLaunch"],
	}
	d.Set("tag", []map[string]interface{}{tag})

	return nil
}

func resourceAwsAutoscalingGroupTagUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).autoscalingconn
	asgName, key, err := extractAutoscalingGroupNameAndKeyFromAutoscalingGroupTagID(d.Id())

	if err != nil {
		return err
	}

	if err := keyvaluetags.AutoscalingUpdateTags(conn, asgName, autoscalingTagResourceTypeAutoScalingGroup, nil, d.Get("tag")); err != nil {
		return fmt.Errorf("error updating Autoscaling Tag (%s) for resource (%s): %w", key, asgName, err)
	}

	return resourceAwsAutoscalingGroupTagRead(d, meta)
}

func resourceAwsAutoscalingGroupTagDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).autoscalingconn
	asgName, key, err := extractAutoscalingGroupNameAndKeyFromAutoscalingGroupTagID(d.Id())

	if err != nil {
		return err
	}

	if err := keyvaluetags.AutoscalingUpdateTags(conn, asgName, autoscalingTagResourceTypeAutoScalingGroup, d.Get("tag"), nil); err != nil {
		return fmt.Errorf("error deleting Autoscaling Tag (%s) for resource (%s): %w", key, asgName, err)
	}

	return nil
}
