package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

const (
	Wafv2WebACLAssociationCreateTimeout = 2 * time.Minute
)

func resourceAwsWafv2WebACLAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsWafv2WebACLAssociationCreate,
		Read:   resourceAwsWafv2WebACLAssociationRead,
		Delete: resourceAwsWafv2WebACLAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				webAclArn, resourceArn, err := resourceAwsWafv2ACLAssociationDecodeId(d.Id())
				if err != nil {
					return nil, fmt.Errorf("Error reading resource ID: %s", err)
				}
				d.Set("resource_arn", resourceArn)
				d.Set("web_acl_arn", webAclArn)
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"resource_arn": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validateArn,
			},
			"web_acl_arn": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validateArn,
			},
		},
	}
}

func resourceAwsWafv2WebACLAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafv2conn
	resourceArn := d.Get("resource_arn").(string)
	webAclArn := d.Get("web_acl_arn").(string)
	params := &wafv2.AssociateWebACLInput{
		ResourceArn: aws.String(resourceArn),
		WebACLArn:   aws.String(webAclArn),
	}

	err := resource.Retry(Wafv2WebACLAssociationCreateTimeout, func() *resource.RetryError {
		var err error
		_, err = conn.AssociateWebACL(params)
		if err != nil {
			if isAWSErr(err, wafv2.ErrCodeWAFUnavailableEntityException, "") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.AssociateWebACL(params)
	}

	if err != nil {
		return err
	}
	d.SetId(fmt.Sprintf("%s,%s", webAclArn, resourceArn))

	return resourceAwsWafv2WebACLAssociationRead(d, meta)
}

func resourceAwsWafv2WebACLAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafv2conn
	resourceArn := d.Get("resource_arn").(string)
	webAclArn := d.Get("web_acl_arn").(string)
	params := &wafv2.GetWebACLForResourceInput{
		ResourceArn: aws.String(resourceArn),
	}

	resp, err := conn.GetWebACLForResource(params)
	if err != nil {
		if isAWSErr(err, wafv2.ErrCodeWAFNonexistentItemException, "") {
			log.Printf("[WARN] WAFv2 Web ACL (%s) not found, removing from state", webAclArn)
			d.SetId("")
			return nil
		}
		return err
	}

	if resp == nil || resp.WebACL == nil {
		log.Printf("[WARN] WAFv2 Web ACL associated resource (%s) not found, removing from state", resourceArn)
		d.SetId("")
	}

	return nil
}

func resourceAwsWafv2WebACLAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafv2conn

	log.Printf("[INFO] Deleting WAFv2 Web ACL Association %s", d.Id())

	params := &wafv2.DisassociateWebACLInput{
		ResourceArn: aws.String(d.Get("resource_arn").(string)),
	}

	_, err := conn.DisassociateWebACL(params)
	if err != nil {
		return fmt.Errorf("Error disassociating WAFv2 Web ACL: %s", err)
	}

	return nil
}

func resourceAwsWafv2ACLAssociationDecodeId(id string) (string, string, error) {
	parts := strings.SplitN(id, ",", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("Unexpected format of ID (%s), expected WEB-ACL-ARN,RESOURCE-ARN", id)
	}

	return parts[0], parts[1], nil
}
