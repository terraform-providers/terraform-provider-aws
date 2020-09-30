package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codecommit"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsCodeCommitApprovalRuleAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCodeCommitApprovalRuleAssociationCreate,
		Update: resourceAwsCodeCommitApprovalRuleAssociationUpdate,
		Read:   resourceAwsCodeCommitApprovalRuleAssociationRead,
		Delete: resourceAwsCodeCommitApprovalRuleAssociationDelete,

		Schema: map[string]*schema.Schema{
			"template_name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"repository_names": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Required: true,
			},
		},
	}
}

func resourceAwsCodeCommitApprovalRuleAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codecommitconn

	templateName := d.Get("template_name").(string)
	repositoryNames := d.Get("repository_names")

	for _, repositoryName := range repositoryNames.(*schema.Set).List() {
		input := &codecommit.AssociateApprovalRuleTemplateWithRepositoryInput{
			ApprovalRuleTemplateName: aws.String(templateName),
			RepositoryName:           aws.String(repositoryName.(string)),
		}

		log.Printf("[DEBUG] Associating approval rule %s with: %s", templateName, repositoryName)
		_, err := conn.AssociateApprovalRuleTemplateWithRepository(input)

		if err != nil {
			return fmt.Errorf("error associating approval rule (%s) with repository (%s): %s", templateName, repositoryName, err)
		}
	}

	d.SetId(templateName)

	return resourceAwsCodeCommitApprovalRuleAssociationRead(d, meta)
}

func resourceAwsCodeCommitApprovalRuleAssociationUpdate(d *schema.ResourceData, meta interface{}) error {
	if d.HasChange("repository_names") {
		templateName := d.Get("template_name")
		conn := meta.(*AWSClient).codecommitconn

		o, n := d.GetChange("repository_names")
		oldR, newR := o.(*schema.Set).List(), n.(*schema.Set).List()

		for _, oldRepo := range oldR {
			found := false
			for _, newRepo := range newR {
				if oldRepo == newRepo {
					found = true
				}
			}

			if !found {
				input := &codecommit.DisassociateApprovalRuleTemplateFromRepositoryInput{
					ApprovalRuleTemplateName: aws.String(templateName.(string)),
					RepositoryName:           aws.String(oldRepo.(string)),
				}
				_, err := conn.DisassociateApprovalRuleTemplateFromRepository(input)
				if err != nil {
					return err
				}
			}
		}

		for _, newRepo := range newR {
			found := false
			for _, oldRepo := range oldR {
				if newRepo == oldRepo {
					found = true
				}
			}

			if !found {
				input := &codecommit.AssociateApprovalRuleTemplateWithRepositoryInput{
					ApprovalRuleTemplateName: aws.String(templateName.(string)),
					RepositoryName:           aws.String(newRepo.(string)),
				}
				_, err := conn.AssociateApprovalRuleTemplateWithRepository(input)
				if err != nil {
					return err
				}
			}
		}
	}

	return resourceAwsCodeCommitApprovalRuleAssociationRead(d, meta)
}

func resourceAwsCodeCommitApprovalRuleAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codecommitconn

	templateName := d.Get("template_name")

	input := &codecommit.ListRepositoriesForApprovalRuleTemplateInput{
		ApprovalRuleTemplateName: aws.String(templateName.(string)),
	}

	_, err := conn.ListRepositoriesForApprovalRuleTemplate(input)

	if err != nil {
		return err
	}

	return nil
}

func resourceAwsCodeCommitApprovalRuleAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codecommitconn

	templateName := d.Get("template_name")
	repositoryNames := d.Get("repository_names")

	for _, repositoryName := range repositoryNames.(*schema.Set).List() {
		input := &codecommit.DisassociateApprovalRuleTemplateFromRepositoryInput{
			ApprovalRuleTemplateName: aws.String(templateName.(string)),
			RepositoryName:           aws.String(repositoryName.(string)),
		}

		_, err := conn.DisassociateApprovalRuleTemplateFromRepository(input)

		if err != nil {
			return err
		}
	}
	return nil
}
