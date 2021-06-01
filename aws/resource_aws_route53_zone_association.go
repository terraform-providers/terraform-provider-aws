package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsRoute53ZoneAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsRoute53ZoneAssociationCreate,
		Read:   resourceAwsRoute53ZoneAssociationRead,
		Delete: resourceAwsRoute53ZoneAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"zone_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"vpc_region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"owning_account": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsRoute53ZoneAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).r53conn

	vpcRegion := meta.(*AWSClient).region
	vpcID := d.Get("vpc_id").(string)
	zoneID := d.Get("zone_id").(string)

	if v, ok := d.GetOk("vpc_region"); ok {
		vpcRegion = v.(string)
	}

	input := &route53.AssociateVPCWithHostedZoneInput{
		HostedZoneId: aws.String(zoneID),
		VPC: &route53.VPC{
			VPCId:     aws.String(vpcID),
			VPCRegion: aws.String(vpcRegion),
		},
		Comment: aws.String("Managed by Terraform"),
	}

	output, err := conn.AssociateVPCWithHostedZone(input)

	if err != nil {
		return fmt.Errorf("error associating Route 53 Hosted Zone (%s) to EC2 VPC (%s): %w", zoneID, vpcID, err)
	}

	d.SetId(fmt.Sprintf("%s:%s:%s", zoneID, vpcID, vpcRegion))

	if output != nil && output.ChangeInfo != nil && output.ChangeInfo.Id != nil {
		wait := resource.StateChangeConf{
			Delay:      30 * time.Second,
			Pending:    []string{route53.ChangeStatusPending},
			Target:     []string{route53.ChangeStatusInsync},
			Timeout:    10 * time.Minute,
			MinTimeout: 2 * time.Second,
			Refresh:    resourceAwsRoute53ZoneAssociationRefreshFunc(conn, cleanChangeID(aws.StringValue(output.ChangeInfo.Id)), d.Id()),
		}

		if _, err := wait.WaitForState(); err != nil {
			return fmt.Errorf("error waiting for Route 53 Zone Association (%s) synchronization: %w", d.Id(), err)
		}
	}

	return resourceAwsRoute53ZoneAssociationRead(d, meta)
}

func resourceAwsRoute53ZoneAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).r53conn

	zoneID, vpcID, vpcRegion, err := resourceAwsRoute53ZoneAssociationParseId(d.Id())

	if err != nil {
		return err
	}

	// Continue supporting older resources without VPC Region in ID
	if vpcRegion == "" {
		vpcRegion = d.Get("vpc_region").(string)
	}

	if vpcRegion == "" {
		vpcRegion = meta.(*AWSClient).region
	}

	hostedZoneSummary, err := route53GetZoneAssociation(conn, zoneID, vpcID, vpcRegion)

	if isAWSErr(err, "AccessDenied", "is not owned by you") && !d.IsNewResource() {
		log.Printf("[WARN] Route 53 Zone Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting Route 53 Zone Association (%s): %w", d.Id(), err)
	}

	if hostedZoneSummary == nil {
		if d.IsNewResource() {
			return fmt.Errorf("error getting Route 53 Zone Association (%s): missing after creation", d.Id())
		}

		log.Printf("[WARN] Route 53 Hosted Zone (%s) Association (%s) not found, removing from state", zoneID, vpcID)
		d.SetId("")
		return nil
	}

	d.Set("vpc_id", vpcID)
	d.Set("vpc_region", vpcRegion)
	d.Set("zone_id", zoneID)
	d.Set("owning_account", hostedZoneSummary.Owner.OwningAccount)

	return nil
}

func resourceAwsRoute53ZoneAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).r53conn

	zoneID, vpcID, vpcRegion, err := resourceAwsRoute53ZoneAssociationParseId(d.Id())

	if err != nil {
		return err
	}

	// Continue supporting older resources without VPC Region in ID
	if vpcRegion == "" {
		vpcRegion = d.Get("vpc_region").(string)
	}

	if vpcRegion == "" {
		vpcRegion = meta.(*AWSClient).region
	}

	input := &route53.DisassociateVPCFromHostedZoneInput{
		HostedZoneId: aws.String(zoneID),
		VPC: &route53.VPC{
			VPCId:     aws.String(vpcID),
			VPCRegion: aws.String(vpcRegion),
		},
		Comment: aws.String("Managed by Terraform"),
	}

	_, err = conn.DisassociateVPCFromHostedZone(input)

	if tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchHostedZone) {
		return nil
	}

	if tfawserr.ErrCodeEquals(err, route53.ErrCodeVPCAssociationNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error disassociating Route 53 Hosted Zone (%s) from EC2 VPC (%s): %w", zoneID, vpcID, err)
	}

	return nil
}

func resourceAwsRoute53ZoneAssociationParseId(id string) (string, string, string, error) {
	parts := strings.Split(id, ":")

	if len(parts) == 3 && parts[0] != "" && parts[1] != "" && parts[2] != "" {
		return parts[0], parts[1], parts[2], nil
	}

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", "", fmt.Errorf("Unexpected format of ID (%q), expected ZONEID:VPCID or ZONEID:VPCID:VPCREGION", id)
	}

	return parts[0], parts[1], "", nil
}

func resourceAwsRoute53ZoneAssociationRefreshFunc(conn *route53.Route53, changeId, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		changeRequest := &route53.GetChangeInput{
			Id: aws.String(changeId),
		}
		result, state, err := resourceAwsGoRoute53Wait(conn, changeRequest)
		if isAWSErr(err, "AccessDenied", "") {
			log.Printf("[WARN] AccessDenied when trying to get Route 53 change progress for %s - ignoring due to likely cross account issue", id)
			return true, route53.ChangeStatusInsync, nil
		}
		return result, state, err
	}
}

func route53GetZoneAssociation(conn *route53.Route53, zoneID, vpcID, vpcRegion string) (*route53.HostedZoneSummary, error) {
	input := &route53.ListHostedZonesByVPCInput{
		VPCId:     aws.String(vpcID),
		VPCRegion: aws.String(vpcRegion),
	}

	for {
		output, err := conn.ListHostedZonesByVPC(input)

		if err != nil {
			return nil, err
		}

		var associatedHostedZoneSummary *route53.HostedZoneSummary
		for _, hostedZoneSummary := range output.HostedZoneSummaries {
			if zoneID == aws.StringValue(hostedZoneSummary.HostedZoneId) {
				associatedHostedZoneSummary = hostedZoneSummary
				return associatedHostedZoneSummary, nil
			}
		}
		if output.NextToken == nil {
			break
		}
		input.NextToken = output.NextToken
	}

	return nil, nil
}
