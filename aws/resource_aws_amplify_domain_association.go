package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/amplify"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	tfamplify "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/amplify"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/amplify/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/amplify/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsAmplifyDomainAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAmplifyDomainAssociationCreate,
		Read:   resourceAwsAmplifyDomainAssociationRead,
		Update: resourceAwsAmplifyDomainAssociationUpdate,
		Delete: resourceAwsAmplifyDomainAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"app_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"certificate_verification_dns_record": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"domain_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},

			"sub_domain": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"branch_name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
						"dns_record": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"prefix": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(0, 255),
						},
						"verified": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},

			"wait_for_verification": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
		},
	}
}

func resourceAwsAmplifyDomainAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).amplifyconn

	appID := d.Get("app_id").(string)
	domainName := d.Get("domain_name").(string)
	id := tfamplify.DomainAssociationCreateResourceID(appID, domainName)

	input := &amplify.CreateDomainAssociationInput{
		AppId:             aws.String(appID),
		DomainName:        aws.String(domainName),
		SubDomainSettings: expandAmplifySubDomainSettings(d.Get("sub_domain").(*schema.Set).List()),
	}

	log.Printf("[DEBUG] Creating Amplify Domain Association: %s", input)
	_, err := conn.CreateDomainAssociation(input)

	if err != nil {
		return fmt.Errorf("error creating Amplify Domain Association (%s): %w", id, err)
	}

	d.SetId(id)

	if _, err := waiter.DomainAssociationCreated(conn, appID, domainName); err != nil {
		return fmt.Errorf("error waiting for Amplify Domain Association (%s) to create: %w", d.Id(), err)
	}

	if d.Get("wait_for_verification").(bool) {
		if _, err := waiter.DomainAssociationVerified(conn, appID, domainName); err != nil {
			return fmt.Errorf("error waiting for Amplify Domain Association (%s) to verify: %w", d.Id(), err)
		}
	}

	return resourceAwsAmplifyDomainAssociationRead(d, meta)
}

func resourceAwsAmplifyDomainAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).amplifyconn

	appID, domainName, err := tfamplify.DomainAssociationParseResourceID(d.Id())

	if err != nil {
		return fmt.Errorf("error parsing Amplify Domain Association ID: %w", err)
	}

	domainAssociation, err := finder.DomainAssociationByAppIDAndDomainName(conn, appID, domainName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Amplify Domain Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Amplify Domain Association (%s): %w", d.Id(), err)
	}

	d.Set("app_id", appID)
	d.Set("arn", domainAssociation.DomainAssociationArn)
	d.Set("certificate_verification_dns_record", domainAssociation.CertificateVerificationDNSRecord)
	d.Set("domain_name", domainAssociation.DomainName)
	if err := d.Set("sub_domain", flattenAmplifySubDomains(domainAssociation.SubDomains)); err != nil {
		return fmt.Errorf("error setting sub_domain: %w", err)
	}

	return nil
}

func resourceAwsAmplifyDomainAssociationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).amplifyconn

	appID, domainName, err := tfamplify.DomainAssociationParseResourceID(d.Id())

	if err != nil {
		return fmt.Errorf("error parsing Amplify Domain Association ID: %w", err)
	}

	if d.HasChange("sub_domain") {
		input := &amplify.UpdateDomainAssociationInput{
			AppId:             aws.String(appID),
			DomainName:        aws.String(domainName),
			SubDomainSettings: expandAmplifySubDomainSettings(d.Get("sub_domain").(*schema.Set).List()),
		}

		log.Printf("[DEBUG] Creating Amplify Domain Association: %s", input)
		_, err := conn.UpdateDomainAssociation(input)

		if err != nil {
			return fmt.Errorf("error updating Amplify Domain Association (%s): %w", d.Id(), err)
		}
	}

	if d.Get("wait_for_verification").(bool) {
		if _, err := waiter.DomainAssociationVerified(conn, appID, domainName); err != nil {
			return fmt.Errorf("error waiting for Amplify Domain Association (%s) to verify: %w", d.Id(), err)
		}
	}

	return resourceAwsAmplifyDomainAssociationRead(d, meta)
}

func resourceAwsAmplifyDomainAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).amplifyconn

	appID, domainName, err := tfamplify.DomainAssociationParseResourceID(d.Id())

	if err != nil {
		return fmt.Errorf("error parsing Amplify Domain Association ID: %w", err)
	}

	log.Printf("[DEBUG] Deleting Amplify Domain Association: %s", d.Id())
	_, err = conn.DeleteDomainAssociation(&amplify.DeleteDomainAssociationInput{
		AppId:      aws.String(appID),
		DomainName: aws.String(domainName),
	})

	if tfawserr.ErrCodeEquals(err, amplify.ErrCodeNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Amplify Domain Association (%s): %w", d.Id(), err)
	}

	return nil
}

func expandAmplifySubDomainSetting(tfMap map[string]interface{}) *amplify.SubDomainSetting {
	if tfMap == nil {
		return nil
	}

	apiObject := &amplify.SubDomainSetting{}

	if v, ok := tfMap["branch_name"].(string); ok && v != "" {
		apiObject.BranchName = aws.String(v)
	}

	// Empty prefix is allowed.
	if v, ok := tfMap["prefix"].(string); ok {
		apiObject.Prefix = aws.String(v)
	}

	return apiObject
}

func expandAmplifySubDomainSettings(tfList []interface{}) []*amplify.SubDomainSetting {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*amplify.SubDomainSetting

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandAmplifySubDomainSetting(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenAmplifySubDomain(apiObject *amplify.SubDomain) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.DnsRecord; v != nil {
		tfMap["dns_record"] = aws.StringValue(v)
	}

	if v := apiObject.SubDomainSetting; v != nil {
		apiObject := v

		if v := apiObject.BranchName; v != nil {
			tfMap["branch_name"] = aws.StringValue(v)
		}

		if v := apiObject.Prefix; v != nil {
			tfMap["prefix"] = aws.StringValue(v)
		}
	}

	if v := apiObject.Verified; v != nil {
		tfMap["verified"] = aws.BoolValue(v)
	}

	return tfMap
}

func flattenAmplifySubDomains(apiObjects []*amplify.SubDomain) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenAmplifySubDomain(apiObject))
	}

	return tfList
}
