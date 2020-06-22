package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/wafregional"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsWafRegionalWebAclAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsWafRegionalWebAclAssociationCreate,
		Read:   resourceAwsWafRegionalWebAclAssociationRead,
		Delete: resourceAwsWafRegionalWebAclAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"web_acl_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"resource_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsWafRegionalWebAclAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafregionalconn

	log.Printf(
		"[INFO] Creating WAF Regional Web ACL association: %s => %s",
		d.Get("web_acl_id").(string),
		d.Get("resource_arn").(string))

	params := &wafregional.AssociateWebACLInput{
		WebACLId:    aws.String(d.Get("web_acl_id").(string)),
		ResourceArn: aws.String(d.Get("resource_arn").(string)),
	}

	// create association and wait on retryable error
	// no response body
	var err error
	err = resource.Retry(2*time.Minute, func() *resource.RetryError {
		_, err = conn.AssociateWebACL(params)
		if err != nil {
			if isAWSErr(err, wafregional.ErrCodeWAFUnavailableEntityException, "") {
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
		return fmt.Errorf("Error creating WAF Regional Web ACL association: %s", err)
	}

	// Store association id
	d.SetId(fmt.Sprintf("%s:%s", *params.WebACLId, *params.ResourceArn))

	return nil
}

func resourceAwsWafRegionalWebAclAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafregionalconn

	resourceArn := resourceAwsWafRegionalWebAclAssociationParseId(d.Id())

	input := &wafregional.GetWebACLForResourceInput{
		ResourceArn: aws.String(resourceArn),
	}

	output, err := conn.GetWebACLForResource(input)

	if isAWSErr(err, wafregional.ErrCodeWAFNonexistentItemException, "") {
		log.Printf("[WARN] WAF Regional Web ACL for resource (%s) not found, removing from state", resourceArn)
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting WAF Regional Web ACL for resource (%s): %s", resourceArn, err)
	}

	if output == nil || output.WebACLSummary == nil {
		log.Printf("[WARN] WAF Regional Web ACL for resource (%s) not found, removing from state", resourceArn)
		d.SetId("")
		return nil
	}

	d.Set("resource_arn", resourceArn)
	d.Set("web_acl_id", output.WebACLSummary.WebACLId)

	return nil
}

func resourceAwsWafRegionalWebAclAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafregionalconn

	resourceArn := resourceAwsWafRegionalWebAclAssociationParseId(d.Id())

	log.Printf("[INFO] Deleting WAF Regional Web ACL association: %s", resourceArn)

	params := &wafregional.DisassociateWebACLInput{
		ResourceArn: aws.String(resourceArn),
	}

	// If action successful HTTP 200 response with an empty body
	_, err := conn.DisassociateWebACL(params)
	return err
}

func resourceAwsWafRegionalWebAclAssociationParseId(id string) (resourceArn string) {
	parts := strings.SplitN(id, ":", 2)
	resourceArn = parts[1]
	return
}
