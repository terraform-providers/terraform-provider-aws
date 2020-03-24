package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsAcmpcaPermission() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAcmpcaPermissionCreate,
		Read:   resourceAwsAcmpcaPermissionRead,
		Delete: resourceAwsAcmpcaPermissionDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(1 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"actions": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						acmpca.ActionTypeIssueCertificate,
						acmpca.ActionTypeGetCertificate,
						acmpca.ActionTypeListPermissions,
					}, false),
				},
			},
			"certificate_authority_arn": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validateArn,
			},
			"policy": {
				Type:     schema.TypeString,
				ForceNew: true,
				Computed: true,
			},
			"principal": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"acm.amazonaws.com",
				}, false),
			},
			"source_account": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceAwsAcmpcaPermissionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).acmpcaconn

	ca_arn := d.Get("certificate_authority_arn").(string)
	principal := d.Get("principal").(string)

	input := &acmpca.CreatePermissionInput{
		Actions:                 expandStringSet(d.Get("actions").(*schema.Set)),
		CertificateAuthorityArn: aws.String(ca_arn),
		Principal:               aws.String(principal),
	}

	source_account := d.Get("source_account").(string)
	if source_account != "" {
		input.SetSourceAccount(source_account)
	}

	log.Printf("[DEBUG] Creating ACMPCA Permission: %s", input)

	var err error
	_, err = conn.CreatePermission(input)

	if err != nil {
		return fmt.Errorf("error creating ACMPCA Permission: %s", err)
	}

	d.SetId(resource.PrefixedUniqueId(fmt.Sprintf("%s-%s-", ca_arn, principal)))

	return resourceAwsAcmpcaPermissionRead(d, meta)
}

func describePermissions(conn *acmpca.ACMPCA, certificateAuthorityArn string, principal string, sourceAccount string) (*acmpca.Permission, error) {

	out, err := conn.ListPermissions(&acmpca.ListPermissionsInput{
		CertificateAuthorityArn: &certificateAuthorityArn,
	})

	if err != nil {
		log.Printf("[WARN] Error retrieving ACMPCA Permissions (%s) when waiting: %s", certificateAuthorityArn, err)
		return nil, err
	}

	var permission *acmpca.Permission

	for _, p := range out.Permissions {
		if aws.StringValue(p.CertificateAuthorityArn) == certificateAuthorityArn && aws.StringValue(p.Principal) == principal && (sourceAccount == "" || aws.StringValue(p.SourceAccount) == sourceAccount) {
			permission = p
			break
		}
	}
	return permission, nil
}

func resourceAwsAcmpcaPermissionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).acmpcaconn

	permission, err := describePermissions(conn, d.Get("certificate_authority_arn").(string), d.Get("principal").(string), d.Get("source_account").(string))

	if permission == nil {
		log.Printf("[WARN] ACMPCA Permission (%s) not found", d.Get("certificate_authority_arn"))
		d.SetId("")
		return err
	}

	d.Set("source_account", permission.SourceAccount)
	d.Set("policy", permission.Policy)

	return nil
}

func resourceAwsAcmpcaPermissionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).acmpcaconn

	input := &acmpca.DeletePermissionInput{
		CertificateAuthorityArn: aws.String(d.Get("certificate_authority_arn").(string)),
		Principal:               aws.String(d.Get("principal").(string)),
		SourceAccount:           aws.String(d.Get("source_account").(string)),
	}

	log.Printf("[DEBUG] Deleting ACMPCA Permission: %s", input)
	_, err := conn.DeletePermission(input)
	if err != nil {
		if isAWSErr(err, acmpca.ErrCodeResourceNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("error deleting ACMPCA Permission: %s", err)
	}

	return nil
}
