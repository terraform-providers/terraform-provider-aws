package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codecommit"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
)

func dataSourceAwsCodeCommitApprovalRule() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsCodeCommitApprovalRuleRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"content": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"approval_rule_template_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsCodeCommitApprovalRuleRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codecommitconn

	templateName := d.Get("name").(string)
	input := &codecommit.GetApprovalRuleTemplateInput{
		ApprovalRuleTemplateName: aws.String(templateName),
	}

	out, err := conn.GetApprovalRuleTemplate(input)
	if err != nil {
		if isAWSErr(err, codecommit.ErrCodeApprovalRuleTemplateDoesNotExistException, "") {
			log.Printf("[WARN] CodeCommit Approval Rule (%s) not found, removing from state", d.Id())
			d.SetId("")
			return fmt.Errorf("Resource codecommit approval rule not found for %s", templateName)
		} else {
			return fmt.Errorf("Error reading CodeCommit Approval Rule: %s", err.Error())
		}
	}

	if out.ApprovalRuleTemplate == nil {
		return fmt.Errorf("no matches found for approval rule name: %s", templateName)
	}

	d.SetId(aws.StringValue(out.ApprovalRuleTemplate.ApprovalRuleTemplateName))
	d.Set("name", out.ApprovalRuleTemplate.ApprovalRuleTemplateName)
	d.Set("approval_rule_template_id", out.ApprovalRuleTemplate.ApprovalRuleTemplateId)
	d.Set("description", out.ApprovalRuleTemplate.ApprovalRuleTemplateDescription)
	templateContent, _ := structure.NormalizeJsonString(*out.ApprovalRuleTemplate.ApprovalRuleTemplateContent)
	d.Set("content", templateContent)

	return nil
}
