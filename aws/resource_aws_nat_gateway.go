package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsNatGateway() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsNatGatewayCreate,
		Read:   resourceAwsNatGatewayRead,
		Update: resourceAwsNatGatewayUpdate,
		Delete: resourceAwsNatGatewayDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"allocation_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"subnet_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"connectivity_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      ec2.ConnectivityTypePublic,
				ValidateFunc: validation.StringInSlice(ec2.ConnectivityType_Values(), false),
			},

			"network_interface_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"private_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"public_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
		},

		CustomizeDiff: SetTagsDiff,
	}
}

func resourceAwsNatGatewayCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	// Create the NAT Gateway
	createOpts := &ec2.CreateNatGatewayInput{
		TagSpecifications: ec2TagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeNatgateway),
	}

	if v, ok := d.GetOk("allocation_id"); ok {
		createOpts.AllocationId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("connectivity_type"); ok {
		createOpts.ConnectivityType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("subnet_id"); ok {
		createOpts.SubnetId = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Create NAT Gateway: %s", *createOpts)
	natResp, err := conn.CreateNatGateway(createOpts)
	if err != nil {
		return fmt.Errorf("Error creating NAT Gateway: %s", err)
	}

	// Get the ID and store it
	ng := natResp.NatGateway
	d.SetId(aws.StringValue(ng.NatGatewayId))
	log.Printf("[INFO] NAT Gateway ID: %s", d.Id())

	// Wait for the NAT Gateway to become available
	log.Printf("[DEBUG] Waiting for NAT Gateway (%s) to become available", d.Id())
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.NatGatewayStatePending},
		Target:  []string{ec2.NatGatewayStateAvailable},
		Refresh: NGStateRefreshFunc(conn, d.Id()),
		Timeout: 10 * time.Minute,
	}

	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("Error waiting for NAT Gateway (%s) to become available: %s", d.Id(), err)
	}

	return resourceAwsNatGatewayRead(d, meta)
}

func resourceAwsNatGatewayRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	// Refresh the NAT Gateway state
	ngRaw, state, err := NGStateRefreshFunc(conn, d.Id())()
	if err != nil {
		return err
	}

	status := map[string]bool{
		ec2.NatGatewayStateDeleted:  true,
		ec2.NatGatewayStateDeleting: true,
		ec2.NatGatewayStateFailed:   true,
	}

	if _, ok := status[strings.ToLower(state)]; ngRaw == nil || ok {
		log.Printf("[INFO] Removing %s from Terraform state as it is not found or in the deleted state.", d.Id())
		d.SetId("")
		return nil
	}

	// Set NAT Gateway attributes
	ng := ngRaw.(*ec2.NatGateway)
	d.Set("connectivity_type", ng.ConnectivityType)
	d.Set("subnet_id", ng.SubnetId)

	// Address
	address := ng.NatGatewayAddresses[0]
	d.Set("allocation_id", address.AllocationId)
	d.Set("network_interface_id", address.NetworkInterfaceId)
	d.Set("private_ip", address.PrivateIp)
	d.Set("public_ip", address.PublicIp)

	tags := keyvaluetags.Ec2KeyValueTags(ng.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceAwsNatGatewayUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := keyvaluetags.Ec2UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EC2 NAT Gateway (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceAwsNatGatewayRead(d, meta)
}

func resourceAwsNatGatewayDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	deleteOpts := &ec2.DeleteNatGatewayInput{
		NatGatewayId: aws.String(d.Id()),
	}
	log.Printf("[INFO] Deleting NAT Gateway: %s", d.Id())

	_, err := conn.DeleteNatGateway(deleteOpts)
	if err != nil {
		ec2err, ok := err.(awserr.Error)
		if !ok {
			return err
		}

		if ec2err.Code() == "NatGatewayNotFound" {
			return nil
		}

		return err
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{ec2.NatGatewayStateDeleting},
		Target:     []string{ec2.NatGatewayStateDeleted},
		Refresh:    NGStateRefreshFunc(conn, d.Id()),
		Timeout:    30 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	_, stateErr := stateConf.WaitForState()
	if stateErr != nil {
		return fmt.Errorf("Error waiting for NAT Gateway (%s) to delete: %s", d.Id(), err)
	}

	return nil
}

// NGStateRefreshFunc returns a resource.StateRefreshFunc that is used to watch
// a NAT Gateway.
func NGStateRefreshFunc(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		opts := &ec2.DescribeNatGatewaysInput{
			NatGatewayIds: []*string{aws.String(id)},
		}
		resp, err := conn.DescribeNatGateways(opts)
		if err != nil {
			if isAWSErr(err, "NatGatewayNotFound", "") {
				resp = nil
			} else {
				log.Printf("Error on NGStateRefresh: %s", err)
				return nil, "", err
			}
		}

		if resp == nil {
			// Sometimes AWS just has consistency issues and doesn't see
			// our instance yet. Return an empty state.
			return nil, "", nil
		}

		ng := resp.NatGateways[0]
		return ng, *ng.State, nil
	}
}
