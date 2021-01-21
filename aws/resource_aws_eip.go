package aws

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

const (
	// Maximum amount of time to wait for EIP association with EC2-Classic instances
	ec2AddressAssociationClassicTimeout = 2 * time.Minute
)

func resourceAwsEip() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEipCreate,
		Read:   resourceAwsEipRead,
		Update: resourceAwsEipUpdate,
		Delete: resourceAwsEipDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Read:   schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(3 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"vpc": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},

			"instance": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"network_interface": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"allocation_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"association_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"domain": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"public_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"public_dns": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"private_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"private_dns": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"associate_with_private_ip": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"carrier_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"customer_owned_ipv4_pool": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"customer_owned_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"public_ipv4_pool": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},

			"network_border_group": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"tags": tagsSchema(),
		},
	}
}

func resourceAwsEipCreate(d *schema.ResourceData, meta interface{}) error {
	ec2conn := meta.(*AWSClient).ec2conn

	// By default, we're not in a VPC
	domainOpt := ""
	if v := d.Get("vpc"); v != nil && v.(bool) {
		domainOpt = ec2.DomainTypeVpc
	}

	allocOpts := &ec2.AllocateAddressInput{
		Domain: aws.String(domainOpt),
	}

	if v, ok := d.GetOk("public_ipv4_pool"); ok {
		allocOpts.PublicIpv4Pool = aws.String(v.(string))
	}

	if v, ok := d.GetOk("customer_owned_ipv4_pool"); ok {
		allocOpts.CustomerOwnedIpv4Pool = aws.String(v.(string))
	}

	if v, ok := d.GetOk("network_border_group"); ok {
		allocOpts.NetworkBorderGroup = aws.String(v.(string))
	}

	log.Printf("[DEBUG] EIP create configuration: %#v", allocOpts)
	allocResp, err := ec2conn.AllocateAddress(allocOpts)
	if err != nil {
		return fmt.Errorf("Error creating EIP: %s", err)
	}

	// The domain tells us if we're in a VPC or not
	d.Set("domain", allocResp.Domain)

	// Assign the eips (unique) allocation id for use later
	// the EIP api has a conditional unique ID (really), so
	// if we're in a VPC we need to save the ID as such, otherwise
	// it defaults to using the public IP
	log.Printf("[DEBUG] EIP Allocate: %#v", allocResp)
	if d.Get("domain").(string) == ec2.DomainTypeVpc {
		d.SetId(aws.StringValue(allocResp.AllocationId))
	} else {
		d.SetId(aws.StringValue(allocResp.PublicIp))
	}

	log.Printf("[INFO] EIP ID: %s (domain: %v)", d.Id(), *allocResp.Domain)

	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		if d.Get("domain").(string) == ec2.DomainTypeStandard {
			return fmt.Errorf("tags can not be set for an EIP in EC2 Classic")
		}
		if err := keyvaluetags.Ec2CreateTags(ec2conn, d.Id(), v); err != nil {
			return fmt.Errorf("error adding tags: %s", err)
		}
	}

	return resourceAwsEipUpdate(d, meta)
}

func resourceAwsEipRead(d *schema.ResourceData, meta interface{}) error {
	ec2conn := meta.(*AWSClient).ec2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	domain := resourceAwsEipDomain(d)
	id := d.Id()

	req := &ec2.DescribeAddressesInput{}

	if domain == ec2.DomainTypeVpc {
		req.AllocationIds = []*string{aws.String(id)}
	} else {
		req.PublicIps = []*string{aws.String(id)}
	}

	log.Printf(
		"[DEBUG] EIP describe configuration: %s (domain: %s)",
		req, domain)

	var err error
	var describeAddresses *ec2.DescribeAddressesOutput

	if d.IsNewResource() {
		err := resource.Retry(d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
			describeAddresses, err = ec2conn.DescribeAddresses(req)
			if err != nil {
				awsErr, ok := err.(awserr.Error)
				if ok && (awsErr.Code() == "InvalidAllocationID.NotFound" ||
					awsErr.Code() == "InvalidAddress.NotFound") {
					return resource.RetryableError(err)
				}

				return resource.NonRetryableError(err)
			}
			return nil
		})
		if isResourceTimeoutError(err) {
			describeAddresses, err = ec2conn.DescribeAddresses(req)
		}
		if err != nil {
			return fmt.Errorf("Error retrieving EIP: %s", err)
		}
	} else {
		describeAddresses, err = ec2conn.DescribeAddresses(req)
		if err != nil {
			awsErr, ok := err.(awserr.Error)
			if ok && (awsErr.Code() == "InvalidAllocationID.NotFound" ||
				awsErr.Code() == "InvalidAddress.NotFound") {
				log.Printf("[WARN] EIP not found, removing from state: %s", req)
				d.SetId("")
				return nil
			}
			return err
		}
	}

	var address *ec2.Address

	// In the case that AWS returns more EIPs than we intend it to, we loop
	// over the returned addresses to see if it's in the list of results
	for _, addr := range describeAddresses.Addresses {
		if (domain == ec2.DomainTypeVpc && aws.StringValue(addr.AllocationId) == id) || aws.StringValue(addr.PublicIp) == id {
			address = addr
			break
		}
	}

	if address == nil {
		log.Printf("[WARN] EIP %q not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("association_id", address.AssociationId)
	if address.InstanceId != nil {
		d.Set("instance", address.InstanceId)
	} else {
		d.Set("instance", "")
	}
	if address.NetworkInterfaceId != nil {
		d.Set("network_interface", address.NetworkInterfaceId)
	} else {
		d.Set("network_interface", "")
	}

	region := *ec2conn.Config.Region
	d.Set("private_ip", address.PrivateIpAddress)
	if address.PrivateIpAddress != nil {
		d.Set("private_dns", fmt.Sprintf("ip-%s.%s", resourceAwsEc2DashIP(*address.PrivateIpAddress), resourceAwsEc2RegionalPrivateDnsSuffix(region)))
	}

	d.Set("public_ip", address.PublicIp)
	if address.PublicIp != nil {
		d.Set("public_dns", meta.(*AWSClient).PartitionHostname(fmt.Sprintf("ec2-%s.%s", resourceAwsEc2DashIP(*address.PublicIp), resourceAwsEc2RegionalPublicDnsSuffix(region))))
	}

	d.Set("public_ipv4_pool", address.PublicIpv4Pool)
	d.Set("carrier_ip", address.CarrierIp)
	d.Set("customer_owned_ipv4_pool", address.CustomerOwnedIpv4Pool)
	d.Set("customer_owned_ip", address.CustomerOwnedIp)
	d.Set("network_border_group", address.NetworkBorderGroup)

	// On import (domain never set, which it must've been if we created),
	// set the 'vpc' attribute depending on if we're in a VPC.
	if address.Domain != nil {
		d.Set("vpc", *address.Domain == ec2.DomainTypeVpc)
	}

	d.Set("domain", address.Domain)

	// Force ID to be an Allocation ID if we're on a VPC
	// This allows users to import the EIP based on the IP if they are in a VPC
	if *address.Domain == ec2.DomainTypeVpc && net.ParseIP(id) != nil {
		log.Printf("[DEBUG] Re-assigning EIP ID (%s) to it's Allocation ID (%s)", d.Id(), *address.AllocationId)
		d.SetId(aws.StringValue(address.AllocationId))
	}

	if err := d.Set("tags", keyvaluetags.Ec2KeyValueTags(address.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsEipUpdate(d *schema.ResourceData, meta interface{}) error {
	ec2conn := meta.(*AWSClient).ec2conn

	domain := resourceAwsEipDomain(d)

	// If we are updating an EIP that is not newly created, and we are attached to
	// an instance or interface, detach first.
	disassociate := false
	if !d.IsNewResource() {
		if d.HasChange("instance") && d.Get("instance").(string) != "" {
			disassociate = true
		} else if (d.HasChanges("network_interface", "associate_with_private_ip")) && d.Get("association_id").(string) != "" {
			disassociate = true
		}
	}
	if disassociate {
		if err := disassociateEip(d, meta); err != nil {
			return err
		}
	}

	// Associate to instance or interface if specified
	associate := false
	v_instance, ok_instance := d.GetOk("instance")
	v_interface, ok_interface := d.GetOk("network_interface")

	if d.HasChange("instance") && ok_instance {
		associate = true
	} else if (d.HasChanges("network_interface", "associate_with_private_ip")) && ok_interface {
		associate = true
	}
	if associate {
		instanceId := v_instance.(string)
		networkInterfaceId := v_interface.(string)

		assocOpts := &ec2.AssociateAddressInput{
			InstanceId: aws.String(instanceId),
			PublicIp:   aws.String(d.Id()),
		}

		// more unique ID conditionals
		if domain == ec2.DomainTypeVpc {
			var privateIpAddress *string
			if v := d.Get("associate_with_private_ip").(string); v != "" {
				privateIpAddress = aws.String(v)
			}
			assocOpts = &ec2.AssociateAddressInput{
				NetworkInterfaceId: aws.String(networkInterfaceId),
				InstanceId:         aws.String(instanceId),
				AllocationId:       aws.String(d.Id()),
				PrivateIpAddress:   privateIpAddress,
			}
		}

		log.Printf("[DEBUG] EIP associate configuration: %s (domain: %s)", assocOpts, domain)

		err := resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
			_, err := ec2conn.AssociateAddress(assocOpts)
			if err != nil {
				if isAWSErr(err, "InvalidAllocationID.NotFound", "") {
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			return nil
		})
		if isResourceTimeoutError(err) {
			_, err = ec2conn.AssociateAddress(assocOpts)
		}
		if err != nil {
			// Prevent saving instance if association failed
			// e.g. missing internet gateway in VPC
			d.Set("instance", "")
			d.Set("network_interface", "")
			return fmt.Errorf("Failure associating EIP: %s", err)
		}

		if assocOpts.AllocationId == nil {
			if err := waitForEc2AddressAssociationClassic(ec2conn, aws.StringValue(assocOpts.PublicIp), aws.StringValue(assocOpts.InstanceId)); err != nil {
				return fmt.Errorf("error waiting for EC2 Address (%s) to associate with EC2-Classic Instance (%s): %w", aws.StringValue(assocOpts.PublicIp), aws.StringValue(assocOpts.InstanceId), err)
			}
		}
	}

	if d.HasChange("tags") && !d.IsNewResource() {
		if d.Get("domain").(string) == ec2.DomainTypeStandard {
			return fmt.Errorf("tags can not be set for an EIP in EC2 Classic")
		}
		o, n := d.GetChange("tags")
		if err := keyvaluetags.Ec2UpdateTags(ec2conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EIP (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceAwsEipRead(d, meta)
}

func resourceAwsEipDelete(d *schema.ResourceData, meta interface{}) error {
	ec2conn := meta.(*AWSClient).ec2conn

	if err := resourceAwsEipRead(d, meta); err != nil {
		return err
	}
	if d.Id() == "" {
		// This might happen from the read
		return nil
	}

	// If we are attached to an instance or interface, detach first.
	if d.Get("instance").(string) != "" || d.Get("association_id").(string) != "" {
		if err := disassociateEip(d, meta); err != nil {
			return err
		}
	}

	domain := resourceAwsEipDomain(d)

	var input *ec2.ReleaseAddressInput
	switch domain {
	case ec2.DomainTypeVpc:
		log.Printf("[DEBUG] EIP release (destroy) address allocation: %v", d.Id())
		input = &ec2.ReleaseAddressInput{
			AllocationId:       aws.String(d.Id()),
			NetworkBorderGroup: aws.String(d.Get("network_border_group").(string)),
		}
	case ec2.DomainTypeStandard:
		log.Printf("[DEBUG] EIP release (destroy) address: %v", d.Id())
		input = &ec2.ReleaseAddressInput{
			PublicIp: aws.String(d.Id()),
		}
	}

	err := resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		var err error
		_, err = ec2conn.ReleaseAddress(input)

		if err == nil {
			return nil
		}
		if _, ok := err.(awserr.Error); !ok {
			return resource.NonRetryableError(err)
		}

		return resource.RetryableError(err)
	})
	if isResourceTimeoutError(err) {
		_, err = ec2conn.ReleaseAddress(input)
	}
	if err != nil {
		return fmt.Errorf("Error releasing EIP address: %s", err)
	}
	return nil
}

func resourceAwsEipDomain(d *schema.ResourceData) string {
	if v, ok := d.GetOk("domain"); ok {
		return v.(string)
	} else if strings.Contains(d.Id(), "eipalloc") {
		// We have to do this for backwards compatibility since TF 0.1
		// didn't have the "domain" computed attribute.
		return ec2.DomainTypeVpc
	}

	return ec2.DomainTypeStandard
}

func disassociateEip(d *schema.ResourceData, meta interface{}) error {
	ec2conn := meta.(*AWSClient).ec2conn
	log.Printf("[DEBUG] Disassociating EIP: %s", d.Id())
	var err error
	switch resourceAwsEipDomain(d) {
	case ec2.DomainTypeVpc:
		associationID := d.Get("association_id").(string)
		if associationID == "" {
			// If assiciationID is empty, it means there's no association.
			// Hence this disassociation can be skipped.
			return nil
		}
		_, err = ec2conn.DisassociateAddress(&ec2.DisassociateAddressInput{
			AssociationId: aws.String(associationID),
		})
	case ec2.DomainTypeStandard:
		_, err = ec2conn.DisassociateAddress(&ec2.DisassociateAddressInput{
			PublicIp: aws.String(d.Get("public_ip").(string)),
		})
	}

	// First check if the association ID is not found. If this
	// is the case, then it was already disassociated somehow,
	// and that is okay. The most commmon reason for this is that
	// the instance or ENI it was attached it was destroyed.
	if ec2err, ok := err.(awserr.Error); ok && ec2err.Code() == "InvalidAssociationID.NotFound" {
		err = nil
	}
	return err
}

// waitForEc2AddressAssociationClassic ensures the correct Instance is associated with an Address
//
// This can take a few seconds to appear correctly for EC2-Classic addresses.
func waitForEc2AddressAssociationClassic(conn *ec2.EC2, publicIP string, instanceID string) error {
	input := &ec2.DescribeAddressesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("public-ip"),
				Values: []*string{aws.String(publicIP)},
			},
			{
				Name:   aws.String("domain"),
				Values: []*string{aws.String(ec2.DomainTypeStandard)},
			},
		},
	}

	err := resource.Retry(ec2AddressAssociationClassicTimeout, func() *resource.RetryError {
		output, err := conn.DescribeAddresses(input)

		if tfawserr.ErrCodeEquals(err, "InvalidAddress.NotFound") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		if len(output.Addresses) == 0 || output.Addresses[0] == nil {
			return resource.RetryableError(fmt.Errorf("not found"))
		}

		if aws.StringValue(output.Addresses[0].InstanceId) != instanceID {
			return resource.RetryableError(fmt.Errorf("not associated"))
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.DescribeAddresses(input)
	}

	if err != nil {
		return fmt.Errorf("error describing EC2 Address (%s) association: %w", publicIP, err)
	}

	return nil
}
