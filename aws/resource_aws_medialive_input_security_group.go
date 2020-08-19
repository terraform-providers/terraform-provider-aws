package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/medialive"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsMediaLiveInputSecurityGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsMediaLiveInputSecurityGroupCreate,
		Read:   resourceAwsMediaLiveInputSecurityGroupRead,
		Update: resourceAwsMediaLiveInputSecurityGroupUpdate,
		Delete: resourceAwsMediaLiveInputSecurityGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"whitelist_rule": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cidr": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsMediaLiveInputSecurityGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).medialiveconn

	input := &medialive.CreateInputSecurityGroupInput{}

	if v, ok := d.GetOk("whitelist_rule"); ok && len(v.([]interface{})) > 0 {
		input.WhitelistRules = expandWhitelistRules(
			v.([]interface{}),
		)
	}

	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		input.Tags = keyvaluetags.New(v).IgnoreAws().MedialiveTags()
	}

	resp, err := conn.CreateInputSecurityGroup(input)
	if err != nil {
		return fmt.Errorf("Error creating MediaLive Input Security Group: %s", err)
	}

	d.SetId(aws.StringValue(resp.SecurityGroup.Id))

	return resourceAwsMediaLiveInputSecurityGroupRead(d, meta)
}

func resourceAwsMediaLiveInputSecurityGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).medialiveconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &medialive.DescribeInputSecurityGroupInput{
		InputSecurityGroupId: aws.String(d.Id()),
	}

	resp, err := conn.DescribeInputSecurityGroup(input)
	if err != nil {
		if isAWSErr(err, medialive.ErrCodeNotFoundException, "") {
			log.Printf("[WARN] MediaLive Input Security Group %s not found, error code (404)", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error describing MediaLive Input Security Group(%s): %s", d.Id(), err)
	}

	d.Set("arn", aws.StringValue(resp.Arn))

	if err := d.Set("whitelist_rule", flattenWhitelistRules(resp.WhitelistRules)); err != nil {
		return fmt.Errorf("error setting whitelist_rule: %s", err)
	}

	if err := d.Set("tags", keyvaluetags.MedialiveKeyValueTags(resp.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsMediaLiveInputSecurityGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).medialiveconn

	if d.HasChange("whitelist_rule") {
		input := &medialive.UpdateInputSecurityGroupInput{
			InputSecurityGroupId: aws.String(d.Id()),
			WhitelistRules: expandWhitelistRules(
				d.Get("whitelist_rule").([]interface{}),
			),
		}

		_, err := conn.UpdateInputSecurityGroup(input)
		if err != nil {
			if isAWSErr(err, medialive.ErrCodeNotFoundException, "") {
				log.Printf("[WARN] MediaLive Input Security Group %s not found, error code (404)", d.Id())
				d.SetId("")
				return nil
			}
			return fmt.Errorf("Error updating MediaLive Input Security Group(%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.MedialiveUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	if err := waitForMediaLiveInputSecurityGroupOperation(conn, d.Id()); err != nil {
		return fmt.Errorf("Error waiting for operational MediaLive Input Security Group(%s): %s", d.Id(), err)
	}

	return resourceAwsMediaLiveInputSecurityGroupRead(d, meta)
}

func resourceAwsMediaLiveInputSecurityGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).medialiveconn
	input := &medialive.DeleteInputSecurityGroupInput{
		InputSecurityGroupId: aws.String(d.Id()),
	}

	_, err := conn.DeleteInputSecurityGroup(input)
	if err != nil {
		if isAWSErr(err, medialive.ErrCodeNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("Error deleting MediaLive Input Security Group(%s): %s", d.Id(), err)
	}

	if err := waitForMediaLiveInputSecurityGroupDeletion(conn, d.Id()); err != nil {
		return fmt.Errorf("Error waiting for deleting MediaLive Input Security Group(%s): %s", d.Id(), err)
	}

	return nil
}

func expandWhitelistRules(whitelistRules []interface{}) []*medialive.InputWhitelistRuleCidr {
	var result []*medialive.InputWhitelistRuleCidr
	if len(whitelistRules) == 0 {
		return nil
	}

	for _, whitelistRule := range whitelistRules {
		r := whitelistRule.(map[string]interface{})

		result = append(result, &medialive.InputWhitelistRuleCidr{
			Cidr: aws.String(r["cidr"].(string)),
		})
	}
	return result
}

func flattenWhitelistRules(whitelistRules []*medialive.InputWhitelistRule) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(whitelistRules))
	for _, whitelistRule := range whitelistRules {
		r := map[string]interface{}{
			"cidr": aws.StringValue(whitelistRule.Cidr),
		}
		result = append(result, r)
	}
	return result
}

func mediaLiveInputSecurityGroupRefreshFunc(conn *medialive.MediaLive, inputSecurityGroupId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		inputSecurityGroup, err := conn.DescribeInputSecurityGroup(&medialive.DescribeInputSecurityGroupInput{
			InputSecurityGroupId: aws.String(inputSecurityGroupId),
		})

		if isAWSErr(err, medialive.ErrCodeNotFoundException, "") {
			return nil, medialive.InputSecurityGroupStateDeleted, nil
		}

		if err != nil {
			return nil, "", fmt.Errorf("error reading MediaLive Input Security Group (%s): %s", inputSecurityGroupId, err)
		}

		if inputSecurityGroup == nil {
			return nil, medialive.InputSecurityGroupStateDeleted, nil
		}

		return inputSecurityGroup, aws.StringValue(inputSecurityGroup.State), nil
	}
}

func waitForMediaLiveInputSecurityGroupOperation(conn *medialive.MediaLive, inputSecurityGroupId string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{medialive.InputSecurityGroupStateUpdating},
		Target: []string{
			medialive.InputSecurityGroupStateIdle,
			medialive.InputSecurityGroupStateInUse,
		},
		Refresh: mediaLiveInputSecurityGroupRefreshFunc(conn, inputSecurityGroupId),
		Timeout: 30 * time.Minute,
	}

	log.Printf("[DEBUG] Waiting for Media Live Input Security Group (%s) Operation", inputSecurityGroupId)
	_, err := stateConf.WaitForState()

	return err
}

func waitForMediaLiveInputSecurityGroupDeletion(conn *medialive.MediaLive, inputSecurityGroupId string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			medialive.InputSecurityGroupStateIdle,
			medialive.InputSecurityGroupStateUpdating,
			medialive.InputSecurityGroupStateInUse,
		},
		Target:         []string{medialive.InputSecurityGroupStateDeleted},
		Refresh:        mediaLiveInputSecurityGroupRefreshFunc(conn, inputSecurityGroupId),
		Timeout:        30 * time.Minute,
		NotFoundChecks: 1,
	}

	log.Printf("[DEBUG] Waiting for Media Live Input Security Group (%s) deletion", inputSecurityGroupId)
	_, err := stateConf.WaitForState()

	if isAWSErr(err, medialive.ErrCodeNotFoundException, "") {
		return nil
	}

	return err
}
