package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsDxHostedConnection() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDxHostedConnectionCreate,
		Read:   resourceAwsDxHostedConnectionRead,
		Delete: resourceAwsDxHostedConnectionDelete,

		Schema: map[string]*schema.Schema{
			"aws_device": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bandwidth": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateDxConnectionBandWidth(),
			},
			"connection_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"has_logical_redundancy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"jumbo_frame_capable": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"lag_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"loa_issue_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"location": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"owner_account_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateAwsAccountId,
			},
			"partner_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"provider_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"region": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchemaForceNew(),
			"vlan": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(1, 4094),
			},
		},
	}
}

func resourceAwsDxHostedConnectionCreate(d *schema.ResourceData, meta interface{}) error {
	dxconn := meta.(*AWSClient).dxconn

	req := &directconnect.AllocateHostedConnectionInput{
		Bandwidth:      aws.String(d.Get("bandwidth").(string)),
		ConnectionId:   aws.String(d.Get("connection_id").(string)),
		ConnectionName: aws.String(d.Get("name").(string)),
		OwnerAccount:   aws.String(d.Get("owner_account_id").(string)),
		Vlan:           aws.Int64(int64(d.Get("vlan").(int))),
	}
	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		req.Tags = keyvaluetags.New(v).IgnoreAws().DirectconnectTags()
	}

	log.Printf("[DEBUG] Creating Direct Connect hosted connection: %s", req)
	resp, err := dxconn.AllocateHostedConnection(req)
	if err != nil {
		return fmt.Errorf("error creating Direct Connect hosted connection: %s", err)
	}

	d.SetId(aws.StringValue(resp.ConnectionId))

	return resourceAwsDxHostedConnectionRead(d, meta)
}

func resourceAwsDxHostedConnectionRead(d *schema.ResourceData, meta interface{}) error {
	dxconn := meta.(*AWSClient).dxconn

	conn, err := dxHostedConnectionRead(d.Id(), d.Get("connection_id").(string), dxconn)
	if err != nil {
		return err
	}
	if conn == nil {
		log.Printf("[WARN] Direct Connect hosted connection (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	// Cannot set the following attributes from the response:
	// - connection_id: conn.ConnectionId is this resource's ID, not the ID of the interconnect or LAG
	// - tags: conn.Tags seems to always come back empty and DescribeTags needs to be called from the owner account
	d.Set("aws_device", conn.AwsDeviceV2)
	d.Set("has_logical_redundancy", conn.HasLogicalRedundancy)
	d.Set("jumbo_frame_capable", conn.JumboFrameCapable)
	d.Set("lag_id", conn.LagId)
	d.Set("loa_issue_time", aws.TimeValue(conn.LoaIssueTime).Format(time.RFC3339))
	d.Set("location", conn.Location)
	d.Set("name", conn.ConnectionName)
	d.Set("owner_account_id", conn.OwnerAccount)
	d.Set("partner_name", conn.PartnerName)
	d.Set("provider_name", conn.ProviderName)
	d.Set("region", conn.Region)
	d.Set("state", conn.ConnectionState)
	d.Set("vlan", conn.Vlan)

	return nil
}

func resourceAwsDxHostedConnectionDelete(d *schema.ResourceData, meta interface{}) error {
	dxconn := meta.(*AWSClient).dxconn

	return dxConnectionDelete(d, dxconn, func() (*directconnect.Connection, error) {
		return dxHostedConnectionRead(d.Id(), d.Get("connection_id").(string), dxconn)
	})
}

func dxHostedConnectionRead(id string, connId string, dxconn *directconnect.DirectConnect) (*directconnect.Connection, error) {
	resp, err := dxconn.DescribeHostedConnections(&directconnect.DescribeHostedConnectionsInput{
		ConnectionId: aws.String(connId),
	})
	if err != nil {
		return nil, fmt.Errorf("error reading Direct Connect hosted connection (%s): %s", id, err)
	}

	for _, conn := range resp.Connections {
		if conn.ConnectionId != nil && *conn.ConnectionId == id {
			return conn, nil
		}
	}

	return nil, nil
}
