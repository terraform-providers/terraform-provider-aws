package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/worklink"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsWorkLinkWebsiteCertificateAuthorityAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsWorkLinkWebsiteCertificateAuthorityAssociationCreate,
		Read:   resourceAwsWorkLinkWebsiteCertificateAuthorityAssociationRead,
		Delete: resourceAwsWorkLinkWebsiteCertificateAuthorityAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"fleet_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"certificate": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"display_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 100),
			},
			"website_ca_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsWorkLinkWebsiteCertificateAuthorityAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).worklinkconn

	input := &worklink.AssociateWebsiteCertificateAuthorityInput{
		FleetArn:    aws.String(d.Get("fleet_arn").(string)),
		Certificate: aws.String(d.Get("certificate").(string)),
	}

	if v, ok := d.GetOk("display_name"); ok {
		input.DisplayName = aws.String(v.(string))
	}

	resp, err := conn.AssociateWebsiteCertificateAuthority(input)
	if err != nil {
		return fmt.Errorf("Error creating WorkLink Website Certificate Authority Association: %s", err)
	}

	d.SetId(fmt.Sprintf("%s,%s", d.Get("fleet_arn").(string), aws.StringValue(resp.WebsiteCaId)))

	return resourceAwsWorkLinkWebsiteCertificateAuthorityAssociationRead(d, meta)
}

func resourceAwsWorkLinkWebsiteCertificateAuthorityAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).worklinkconn

	fleetArn, websiteCaID, err := decodeWorkLinkWebsiteCertificateAuthorityAssociationResourceID(d.Id())
	if err != nil {
		return err
	}

	input := &worklink.DescribeWebsiteCertificateAuthorityInput{
		FleetArn:    aws.String(fleetArn),
		WebsiteCaId: aws.String(websiteCaID),
	}

	resp, err := conn.DescribeWebsiteCertificateAuthority(input)
	if err != nil {
		if isAWSErr(err, worklink.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] WorkLink Website Certificate Authority Association (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error describing WorkLink Website Certificate Authority Association (%s): %s", d.Id(), err)
	}

	d.Set("website_ca_id", websiteCaID)
	d.Set("fleet_arn", fleetArn)
	d.Set("certificate", resp.Certificate)
	d.Set("display_name", resp.DisplayName)

	return nil
}

func resourceAwsWorkLinkWebsiteCertificateAuthorityAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).worklinkconn

	fleetArn, websiteCaID, err := decodeWorkLinkWebsiteCertificateAuthorityAssociationResourceID(d.Id())
	if err != nil {
		return err
	}

	input := &worklink.DisassociateWebsiteCertificateAuthorityInput{
		FleetArn:    aws.String(fleetArn),
		WebsiteCaId: aws.String(websiteCaID),
	}

	if _, err := conn.DisassociateWebsiteCertificateAuthority(input); err != nil {
		if isAWSErr(err, worklink.ErrCodeResourceNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("Error deleting WorkLink Website Certificate Authority Association (%s): %s", d.Id(), err)
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"DELETING"},
		Target:     []string{"DELETED"},
		Refresh:    worklinkWebsiteCertificateAuthorityAssociationStateRefresh(conn, websiteCaID, fleetArn),
		Timeout:    15 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf(
			"Error waiting for disassociate Worklink Website Certificate Authority (%s) to become deleted: %s",
			d.Id(), err)
	}

	return nil
}

func worklinkWebsiteCertificateAuthorityAssociationStateRefresh(conn *worklink.WorkLink, websiteCaID, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		emptyResp := &worklink.DescribeWebsiteCertificateAuthorityOutput{}

		resp, err := conn.DescribeWebsiteCertificateAuthority(&worklink.DescribeWebsiteCertificateAuthorityInput{
			FleetArn:    aws.String(arn),
			WebsiteCaId: aws.String(websiteCaID),
		})
		if isAWSErr(err, worklink.ErrCodeResourceNotFoundException, "") {
			return emptyResp, "DELETED", nil
		}
		if err != nil {
			return nil, "", err
		}

		return resp, "", nil
	}
}

func decodeWorkLinkWebsiteCertificateAuthorityAssociationResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, ",", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("Unexpected format of ID(%s), expected WebsiteCaId/FleetArn", id)
	}
	fleetArn := parts[0]
	websiteCaID := parts[1]

	return fleetArn, websiteCaID, nil
}
