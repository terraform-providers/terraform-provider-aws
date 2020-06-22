package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsDxConnection() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDxConnectionCreate,
		Read:   resourceAwsDxConnectionRead,
		Update: resourceAwsDxConnectionUpdate,
		Delete: resourceAwsDxConnectionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"bandwidth": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateDxConnectionBandWidth(),
			},
			"location": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"jumbo_frame_capable": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"tags": tagsSchema(),
			"has_logical_redundancy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"aws_device": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsDxConnectionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	req := &directconnect.CreateConnectionInput{
		Bandwidth:      aws.String(d.Get("bandwidth").(string)),
		ConnectionName: aws.String(d.Get("name").(string)),
		Location:       aws.String(d.Get("location").(string)),
	}

	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		req.Tags = keyvaluetags.New(v).IgnoreAws().DirectconnectTags()
	}

	log.Printf("[DEBUG] Creating Direct Connect connection: %#v", req)
	resp, err := conn.CreateConnection(req)
	if err != nil {
		return err
	}

	d.SetId(aws.StringValue(resp.ConnectionId))

	return resourceAwsDxConnectionRead(d, meta)
}

func resourceAwsDxConnectionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	resp, err := conn.DescribeConnections(&directconnect.DescribeConnectionsInput{
		ConnectionId: aws.String(d.Id()),
	})
	if err != nil {
		if isNoSuchDxConnectionErr(err) {
			log.Printf("[WARN] Direct Connect connection (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	if len(resp.Connections) < 1 {
		log.Printf("[WARN] Direct Connect connection (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if len(resp.Connections) != 1 {
		return fmt.Errorf("Number of Direct Connect connections (%s) isn't one, got %d", d.Id(), len(resp.Connections))
	}
	connection := resp.Connections[0]
	if d.Id() != aws.StringValue(connection.ConnectionId) {
		return fmt.Errorf("Direct Connect connection (%s) not found", d.Id())
	}
	if aws.StringValue(connection.ConnectionState) == directconnect.ConnectionStateDeleted {
		log.Printf("[WARN] Direct Connect connection (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Service:   "directconnect",
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("dxcon/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("name", connection.ConnectionName)
	d.Set("bandwidth", connection.Bandwidth)
	d.Set("location", connection.Location)
	d.Set("jumbo_frame_capable", connection.JumboFrameCapable)
	d.Set("has_logical_redundancy", connection.HasLogicalRedundancy)
	d.Set("aws_device", connection.AwsDeviceV2)

	tags, err := keyvaluetags.DirectconnectListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for Direct Connect connection (%s): %s", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsDxConnectionUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	arn := d.Get("arn").(string)
	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.DirectconnectUpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating Direct Connect connection (%s) tags: %s", arn, err)
		}
	}

	return resourceAwsDxConnectionRead(d, meta)
}

func resourceAwsDxConnectionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	log.Printf("[DEBUG] Deleting Direct Connect connection: %s", d.Id())
	_, err := conn.DeleteConnection(&directconnect.DeleteConnectionInput{
		ConnectionId: aws.String(d.Id()),
	})
	if err != nil {
		if isNoSuchDxConnectionErr(err) {
			return nil
		}
		return err
	}

	deleteStateConf := &resource.StateChangeConf{
		Pending:    []string{directconnect.ConnectionStatePending, directconnect.ConnectionStateOrdering, directconnect.ConnectionStateAvailable, directconnect.ConnectionStateRequested, directconnect.ConnectionStateDeleting},
		Target:     []string{directconnect.ConnectionStateDeleted},
		Refresh:    dxConnectionRefreshStateFunc(conn, d.Id()),
		Timeout:    10 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	_, err = deleteStateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for Direct Connect connection (%s) to be deleted: %s", d.Id(), err)
	}

	return nil
}

func dxConnectionRefreshStateFunc(conn *directconnect.DirectConnect, connId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &directconnect.DescribeConnectionsInput{
			ConnectionId: aws.String(connId),
		}
		resp, err := conn.DescribeConnections(input)
		if err != nil {
			return nil, "failed", err
		}
		if len(resp.Connections) < 1 {
			return resp, directconnect.ConnectionStateDeleted, nil
		}
		return resp, *resp.Connections[0].ConnectionState, nil
	}
}

func isNoSuchDxConnectionErr(err error) bool {
	return isAWSErr(err, "DirectConnectClientException", "Could not find Connection with ID")
}
