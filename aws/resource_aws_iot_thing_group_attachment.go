package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsIotThingGroupAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsIotThingGroupAttachmentCreate,
		Read:   resourceAwsIotThingGroupAttachmentRead,
		Delete: resourceAwsIotThingGroupAttachmentDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsIotThingGroupAttachmentImport,
		},

		Schema: map[string]*schema.Schema{
			"thing_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"thing_group_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"override_dynamics_group": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsIotThingGroupAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotconn

	params := &iot.AddThingToThingGroupInput{}
	params.ThingName = aws.String(d.Get("thing_name").(string))
	params.ThingGroupName = aws.String(d.Get("thing_group_name").(string))

	if v, ok := d.GetOk("override_dynamics_group"); ok {
		params.OverrideDynamicGroups = aws.Bool(v.(bool))
	}

	_, err := conn.AddThingToThingGroup(params)

	if err != nil {
		return err
	}

	d.SetId(resource.PrefixedUniqueId(fmt.Sprintf("%s-%s", *params.ThingName, *params.ThingGroupName)))

	return resourceAwsIotThingGroupAttachmentRead(d, meta)
}

func resourceAwsIotThingGroupAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotconn

	thingName := d.Get("thing_name").(string)
	thingGroupName := d.Get("thing_group_name").(string)

	hasThingGroup, err := iotThingHasThingGroup(conn, thingName, thingGroupName, "")

	if isAWSErr(err, iot.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] IoT Thing (%s) is not found", thingName)
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error finding IoT Thing Group (%s) of thing  (%s): %s", thingGroupName, thingName, err)
	}

	if !hasThingGroup {
		log.Printf("[WARN] IoT Thing Group (%s) is not found in Thing (%s) group list", thingGroupName, thingName)
		d.SetId("")
		return nil
	}

	d.Set("thing_name", thingName)
	d.Set("thing_group_name", thingGroupName)
	if v, ok := d.GetOk("override_dynamics_group"); ok {
		d.Set("override_dynamics_group", v.(bool))
	}

	return nil
}

func resourceAwsIotThingGroupAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotconn

	params := &iot.RemoveThingFromThingGroupInput{}
	params.ThingName = aws.String(d.Get("thing_name").(string))
	params.ThingGroupName = aws.String(d.Get("thing_group_name").(string))

	_, err := conn.RemoveThingFromThingGroup(params)

	if err != nil {
		return err
	}

	return nil
}

func resourceAwsIotThingGroupAttachmentImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.SplitN(d.Id(), "/", 2)
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return nil, fmt.Errorf("unexpected format of ID (%q), expected <thing-name>/<thing-group>", d.Id())
	}

	thingName := idParts[0]
	thingGroupName := idParts[1]

	d.Set("thing_name", thingName)
	d.Set("thing_group_name", thingGroupName)

	d.SetId(fmt.Sprintf("%s-%s", thingName, thingGroupName))

	return []*schema.ResourceData{d}, nil
}

func iotThingHasThingGroup(conn *iot.IoT, thingName string, thingGroupName string, nextToken string) (bool, error) {
	maxResults := int64(20)

	params := &iot.ListThingGroupsForThingInput{
		MaxResults: aws.Int64(maxResults),
		ThingName:  aws.String(thingName),
	}

	if len(nextToken) > 0 {
		params.NextToken = aws.String(nextToken)
	}

	out, err := conn.ListThingGroupsForThing(params)
	if err != nil {
		return false, err
	}

	// Check if searched group is in current collection
	// If it is return true
	for _, group := range out.ThingGroups {
		if thingGroupName == *group.GroupName {
			return true, nil
		}
	}

	// If group that we searched for not appear in current list
	// then check if NextToken exists. If it is so call hasThingGroup
	// recursively to search in next part of list. Otherwise return false
	if out.NextToken != nil {
		return iotThingHasThingGroup(conn, thingName, thingGroupName, *out.NextToken)
	} else {
		return false, nil
	}
}
