package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codecommit"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsCodeCommitApprovalRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCodeCommitApprovalRuleCreate,
		Read:   resourceAwsCodeCommitApprovalRuleRead,
		Update: resourceAwsCodeCommitApprovalRuleUpdate,
		Delete: resourceAwsCodeCommitApprovalRuleDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"content": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateApprovalRuleContentValue(),
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v.(string))
					return json
				},
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 512),
			},
			"approval_rule_template_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsCodeCommitApprovalRuleCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codecommitconn

	input := &codecommit.CreateApprovalRuleTemplateInput{
		ApprovalRuleTemplateName:        aws.String(d.Get("name").(string)),
		ApprovalRuleTemplateDescription: aws.String(d.Get("description").(string)),
		ApprovalRuleTemplateContent:     aws.String(d.Get("content").(string)),
	}

	resp, err := conn.CreateApprovalRuleTemplate(input)
	if err != nil {
		return fmt.Errorf("Error creating CodeCommit Approval Rule Template: %s", err)
	}

	log.Printf("[INFO] Code Commit Approval Rule template Created %s input %s", resp, input)

	d.SetId(d.Get("name").(string))

	return resourceAwsCodeCommitApprovalRuleRead(d, meta)
}

func resourceAwsCodeCommitApprovalRuleRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codecommitconn

	input := &codecommit.GetApprovalRuleTemplateInput{
		ApprovalRuleTemplateName: aws.String(d.Id()),
	}

	resp, err := conn.GetApprovalRuleTemplate(input)
	if err != nil {
		return fmt.Errorf("Error reading CodeCommit Approval Rule Template: %s", err.Error())
	}

	log.Printf("[DEBUG] CodeCommit Approval Rule Template: %s", resp)

	d.Set("name", resp.ApprovalRuleTemplate.ApprovalRuleTemplateName)
	d.Set("approval_rule_template_id", resp.ApprovalRuleTemplate.ApprovalRuleTemplateId)
	d.Set("description", resp.ApprovalRuleTemplate.ApprovalRuleTemplateDescription)
	templateContent, _ := structure.NormalizeJsonString(*resp.ApprovalRuleTemplate.ApprovalRuleTemplateContent)
	d.Set("content", templateContent)

	return nil
}

func resourceAwsCodeCommitApprovalRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codecommitconn

	if d.HasChange("description") {
		log.Printf("[DEBUG] Updating Code Commit approval rule template %q description", d.Id())
		_, n := d.GetChange("description")
		input := &codecommit.UpdateApprovalRuleTemplateDescriptionInput{
			ApprovalRuleTemplateDescription: aws.String(n.(string)),
			ApprovalRuleTemplateName:        aws.String(d.Get("name").(string)),
		}

		_, err := conn.UpdateApprovalRuleTemplateDescription(input)

		if err != nil {
			return err
		}
		log.Printf("[DEBUG] Code Commit approval rule template (%q) description updated", d.Id())
	}

	if d.HasChange("content") {
		log.Printf("[DEBUG] Updating Code Commit approval rule template %q content", d.Id())
		_, n := d.GetChange("content")
		newContent, _ := structure.NormalizeJsonString(n)

		input := &codecommit.UpdateApprovalRuleTemplateContentInput{
			NewRuleContent:           aws.String(newContent),
			ApprovalRuleTemplateName: aws.String(d.Get("name").(string)),
		}

		_, err := conn.UpdateApprovalRuleTemplateContent(input)

		if err != nil {
			return err
		}
		log.Printf("[DEBUG] Code Commit approval rule template (%q) content updated", d.Id())
	}

	return resourceAwsCodeCommitApprovalRuleRead(d, meta)
}

func resourceAwsCodeCommitApprovalRuleDelete(d *schema.ResourceData, meta interface{}) error {

	conn := meta.(*AWSClient).codecommitconn

	log.Printf("[DEBUG] Deleting Approval Rule Template: %q", d.Id())

	input := &codecommit.DeleteApprovalRuleTemplateInput{
		ApprovalRuleTemplateName: aws.String(d.Get("name").(string)),
	}

	_, err := conn.DeleteApprovalRuleTemplate(input)

	return err
}

func validateApprovalRuleContentValue() schema.SchemaValidateFunc {
	return func(v interface{}, k string) (ws []string, errors []error) {
		json, err := structure.NormalizeJsonString(v)
		if err != nil {
			errors = append(errors, fmt.Errorf("%q contains an invalid JSON: %s", k, err))

			// Invalid JSON? Return immediately,
			// there is no need to collect other
			// errors.
			return
		}

		// Check whether the normalized JSON is within the given length.
		if len(json) > 3000 {
			errors = append(errors, fmt.Errorf(
				"%q cannot be longer than %d characters: %q", k, 3000, json))
		}
		return
	}
}
