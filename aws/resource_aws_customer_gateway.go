package aws

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsCustomerGateway() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCustomerGatewayCreate,
		Read:   resourceAwsCustomerGatewayRead,
		Update: resourceAwsCustomerGatewayUpdate,
		Delete: resourceAwsCustomerGatewayDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"bgp_asn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate4ByteAsn,
			},

			"ip_address": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.Any(
					validation.StringIsEmpty,
					validation.IsIPv4Address,
				),
			},

			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					ec2.GatewayTypeIpsec1,
				}, false),
			},

			"tags": tagsSchema(),
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsCustomerGatewayCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	ipAddress := d.Get("ip_address").(string)
	vpnType := d.Get("type").(string)
	bgpAsn := d.Get("bgp_asn").(string)

	alreadyExists, err := resourceAwsCustomerGatewayExists(vpnType, ipAddress, bgpAsn, conn)
	if err != nil {
		return err
	}

	if alreadyExists {
		return fmt.Errorf("An existing customer gateway for IpAddress: %s, VpnType: %s, BGP ASN: %s has been found", ipAddress, vpnType, bgpAsn)
	}

	i64BgpAsn, err := strconv.ParseInt(bgpAsn, 10, 64)
	if err != nil {
		return err
	}

	createOpts := &ec2.CreateCustomerGatewayInput{
		BgpAsn:            aws.Int64(i64BgpAsn),
		PublicIp:          aws.String(ipAddress),
		Type:              aws.String(vpnType),
		TagSpecifications: ec2TagSpecificationsFromMap(d.Get("tags").(map[string]interface{}), ec2.ResourceTypeCustomerGateway),
	}

	// Create the Customer Gateway.
	log.Printf("[DEBUG] Creating customer gateway")
	resp, err := conn.CreateCustomerGateway(createOpts)
	if err != nil {
		return fmt.Errorf("Error creating customer gateway: %s", err)
	}

	// Store the ID
	customerGateway := resp.CustomerGateway
	cgId := aws.StringValue(customerGateway.CustomerGatewayId)
	d.SetId(cgId)
	log.Printf("[INFO] Customer gateway ID: %s", cgId)

	// Wait for the CustomerGateway to be available.
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"pending"},
		Target:     []string{"available"},
		Refresh:    customerGatewayRefreshFunc(conn, cgId),
		Timeout:    10 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, stateErr := stateConf.WaitForState()
	if stateErr != nil {
		return fmt.Errorf(
			"Error waiting for customer gateway (%s) to become ready: %s", cgId, err)
	}

	return resourceAwsCustomerGatewayRead(d, meta)
}

func customerGatewayRefreshFunc(conn *ec2.EC2, gatewayId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		gatewayFilter := &ec2.Filter{
			Name:   aws.String("customer-gateway-id"),
			Values: []*string{aws.String(gatewayId)},
		}

		resp, err := conn.DescribeCustomerGateways(&ec2.DescribeCustomerGatewaysInput{
			Filters: []*ec2.Filter{gatewayFilter},
		})
		if err != nil {
			if isAWSErr(err, "InvalidCustomerGatewayID.NotFound", "") {
				resp = nil
			} else {
				log.Printf("Error on CustomerGatewayRefresh: %s", err)
				return nil, "", err
			}
		}

		if resp == nil || len(resp.CustomerGateways) == 0 {
			// handle consistency issues
			return nil, "", nil
		}

		gateway := resp.CustomerGateways[0]
		return gateway, *gateway.State, nil
	}
}

func resourceAwsCustomerGatewayExists(vpnType, ipAddress, bgpAsn string, conn *ec2.EC2) (bool, error) {
	ipAddressFilter := &ec2.Filter{
		Name:   aws.String("ip-address"),
		Values: []*string{aws.String(ipAddress)},
	}

	typeFilter := &ec2.Filter{
		Name:   aws.String("type"),
		Values: []*string{aws.String(vpnType)},
	}

	bgpAsnFilter := &ec2.Filter{
		Name:   aws.String("bgp-asn"),
		Values: []*string{aws.String(bgpAsn)},
	}

	resp, err := conn.DescribeCustomerGateways(&ec2.DescribeCustomerGatewaysInput{
		Filters: []*ec2.Filter{ipAddressFilter, typeFilter, bgpAsnFilter},
	})
	if err != nil {
		return false, err
	}

	if len(resp.CustomerGateways) > 0 && aws.StringValue(resp.CustomerGateways[0].State) != "deleted" {
		return true, nil
	}

	return false, nil
}

func resourceAwsCustomerGatewayRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	gatewayFilter := &ec2.Filter{
		Name:   aws.String("customer-gateway-id"),
		Values: []*string{aws.String(d.Id())},
	}

	resp, err := conn.DescribeCustomerGateways(&ec2.DescribeCustomerGatewaysInput{
		Filters: []*ec2.Filter{gatewayFilter},
	})
	if err != nil {
		if isAWSErr(err, "InvalidCustomerGatewayID.NotFound", "") {
			log.Printf("[WARN] Customer Gateway (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		} else {
			log.Printf("[ERROR] Error finding CustomerGateway: %s", err)
			return err
		}
	}

	if len(resp.CustomerGateways) != 1 {
		return fmt.Errorf("Error finding CustomerGateway: %s", d.Id())
	}

	if aws.StringValue(resp.CustomerGateways[0].State) == "deleted" {
		log.Printf("[INFO] Customer Gateway is in `deleted` state: %s", d.Id())
		d.SetId("")
		return nil
	}

	customerGateway := resp.CustomerGateways[0]
	d.Set("bgp_asn", customerGateway.BgpAsn)
	d.Set("ip_address", customerGateway.IpAddress)
	d.Set("type", customerGateway.Type)

	if err := d.Set("tags", keyvaluetags.Ec2KeyValueTags(customerGateway.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "ec2",
		Region:    meta.(*AWSClient).region,
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("customer-gateway/%s", d.Id()),
	}.String()

	d.Set("arn", arn)

	return nil
}

func resourceAwsCustomerGatewayUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.Ec2UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EC2 Customer Gateway (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceAwsCustomerGatewayRead(d, meta)
}

func resourceAwsCustomerGatewayDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	_, err := conn.DeleteCustomerGateway(&ec2.DeleteCustomerGatewayInput{
		CustomerGatewayId: aws.String(d.Id()),
	})
	if err != nil {
		if isAWSErr(err, "InvalidCustomerGatewayID.NotFound", "") {
			return nil
		} else {
			return fmt.Errorf("[ERROR] Error deleting CustomerGateway: %s", err)
		}
	}

	gatewayFilter := &ec2.Filter{
		Name:   aws.String("customer-gateway-id"),
		Values: []*string{aws.String(d.Id())},
	}

	input := &ec2.DescribeCustomerGatewaysInput{
		Filters: []*ec2.Filter{gatewayFilter},
	}
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		resp, err := conn.DescribeCustomerGateways(input)

		if err != nil {
			if isAWSErr(err, "InvalidCustomerGatewayID.NotFound", "") {
				return nil
			}
			return resource.NonRetryableError(err)
		}

		err = checkGatewayDeleteResponse(resp, d.Id())
		if err != nil {
			return resource.RetryableError(err)
		}
		return nil
	})

	if isResourceTimeoutError(err) {
		var resp *ec2.DescribeCustomerGatewaysOutput
		resp, err = conn.DescribeCustomerGateways(input)

		if err != nil {
			return checkGatewayDeleteResponse(resp, d.Id())
		}
	}

	if err != nil {
		return fmt.Errorf("Error deleting customer gateway: %s", err)
	}
	return nil

}

func checkGatewayDeleteResponse(resp *ec2.DescribeCustomerGatewaysOutput, id string) error {
	if len(resp.CustomerGateways) != 1 {
		return fmt.Errorf("Error finding CustomerGateway for delete: %s", id)
	}

	cgState := aws.StringValue(resp.CustomerGateways[0].State)
	switch cgState {
	case "pending", "available", "deleting":
		return fmt.Errorf("Gateway (%s) in state (%s), retrying", id, cgState)
	case "deleted":
		return nil
	default:
		return fmt.Errorf("Unrecognized state (%s) for Customer Gateway delete on (%s)", *resp.CustomerGateways[0].State, id)
	}
}
