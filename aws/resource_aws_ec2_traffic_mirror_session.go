package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsEc2TrafficMirrorSession() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEc2TrafficMirrorSessionCreate,
		Update: resourceAwsEc2TrafficMirrorSessionUpdate,
		Read:   resourceAwsEc2TrafficMirrorSessionRead,
		Delete: resourceAwsEc2TrafficMirrorSessionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"network_interface_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"packet_length": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"session_number": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(1, 32766),
			},
			"traffic_mirror_filter_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"traffic_mirror_target_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"virtual_network_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(1, 16777216),
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsEc2TrafficMirrorSessionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	input := &ec2.CreateTrafficMirrorSessionInput{
		NetworkInterfaceId:    aws.String(d.Get("network_interface_id").(string)),
		TrafficMirrorFilterId: aws.String(d.Get("traffic_mirror_filter_id").(string)),
		TrafficMirrorTargetId: aws.String(d.Get("traffic_mirror_target_id").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("packet_length"); ok {
		input.PacketLength = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("session_number"); ok {
		input.SessionNumber = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("virtual_network_id"); ok {
		input.VirtualNetworkId = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("tags"); ok {
		input.TagSpecifications = ec2TagSpecificationsFromMap(v.(map[string]interface{}), ec2.ResourceTypeTrafficMirrorSession)
	}

	out, err := conn.CreateTrafficMirrorSession(input)
	if nil != err {
		return fmt.Errorf("Error creating traffic mirror session %v", err)
	}

	d.SetId(*out.TrafficMirrorSession.TrafficMirrorSessionId)
	return resourceAwsEc2TrafficMirrorSessionRead(d, meta)
}

func resourceAwsEc2TrafficMirrorSessionUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	sessionId := d.Id()
	input := &ec2.ModifyTrafficMirrorSessionInput{
		TrafficMirrorSessionId: &sessionId,
	}

	if d.HasChange("session_number") {
		n := d.Get("session_number")
		input.SessionNumber = aws.Int64(int64(n.(int)))
	}

	if d.HasChange("traffic_mirror_filter_id") {
		n := d.Get("traffic_mirror_filter_id")
		input.TrafficMirrorFilterId = aws.String(n.(string))
	}

	if d.HasChange("traffic_mirror_target_id") {
		n := d.Get("traffic_mirror_target_id")
		input.TrafficMirrorTargetId = aws.String(n.(string))
	}

	var removeFields []*string
	if d.HasChange("description") {
		n := d.Get("description")
		if "" != n {
			input.Description = aws.String(n.(string))
		} else {
			removeFields = append(removeFields, aws.String("description"))
		}
	}

	if d.HasChange("packet_length") {
		n := d.Get("packet_length")
		if nil != n && n.(int) > 0 {
			input.PacketLength = aws.Int64(int64(n.(int)))
		} else {
			removeFields = append(removeFields, aws.String("packet-length"))
		}
	}
	log.Printf("[DEBUG] removeFields %v", removeFields)

	if d.HasChange("virtual_network_id") {
		n := d.Get("virtual_network_id")
		log.Printf("[DEBUG] VNI has change %v", n)
		if nil != n && n.(int) > 0 {
			input.VirtualNetworkId = aws.Int64(int64(n.(int)))
		} else {
			removeFields = append(removeFields, aws.String("virtual-network-id"))
		}
	}

	log.Printf("[DEBUG] removeFields %v", removeFields)
	if len(removeFields) > 0 {
		input.SetRemoveFields(removeFields)
	}

	_, err := conn.ModifyTrafficMirrorSession(input)
	if nil != err {
		return fmt.Errorf("Error updating traffic mirror session %v", err)
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.Ec2UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EC2 Traffic Mirror Session (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceAwsEc2TrafficMirrorSessionRead(d, meta)
}

func resourceAwsEc2TrafficMirrorSessionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	sessionId := d.Id()
	input := &ec2.DescribeTrafficMirrorSessionsInput{
		TrafficMirrorSessionIds: []*string{
			&sessionId,
		},
	}

	out, err := conn.DescribeTrafficMirrorSessions(input)

	if isAWSErr(err, "InvalidTrafficMirrorSessionId.NotFound", "") {
		log.Printf("[WARN] EC2 Traffic Mirror Session (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if nil != err {
		return fmt.Errorf("error describing EC2 Traffic Mirror Session (%s): %w", sessionId, err)
	}

	if 0 == len(out.TrafficMirrorSessions) {
		log.Printf("[WARN] EC2 Traffic Mirror Session (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	session := out.TrafficMirrorSessions[0]
	d.Set("network_interface_id", session.NetworkInterfaceId)
	d.Set("session_number", session.SessionNumber)
	d.Set("traffic_mirror_filter_id", session.TrafficMirrorFilterId)
	d.Set("traffic_mirror_target_id", session.TrafficMirrorTargetId)
	d.Set("description", session.Description)
	d.Set("packet_length", session.PacketLength)
	d.Set("virtual_network_id", session.VirtualNetworkId)

	if err := d.Set("tags", keyvaluetags.Ec2KeyValueTags(session.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsEc2TrafficMirrorSessionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	sessionId := d.Id()
	input := &ec2.DeleteTrafficMirrorSessionInput{
		TrafficMirrorSessionId: &sessionId,
	}

	_, err := conn.DeleteTrafficMirrorSession(input)
	if nil != err {
		return fmt.Errorf("error deleting EC2 Traffic Mirror Session (%s): %w", sessionId, err)
	}

	return nil
}
