package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/workdocs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsAdUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAdUserCreate,
		Read:   resourceAwsAdUserRead,
		Update: resourceAwsAdUserUpdate,
		Delete: resourceAwsAdUserDelete,

		Schema: map[string]*schema.Schema{
			"email_address": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"given_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"organization_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"password": {
				Type:         schema.TypeString,
				Required:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringLenBetween(4, 2048),
			},
			"surname": {
				Type:     schema.TypeString,
				Required: true,
			},
			"username": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceAwsAdUserCreate(d *schema.ResourceData, meta interface{}) error {
	workdocsconn := meta.(*AWSClient).workdocsconn

	emailAddress := d.Get("email_address").(string)
	givenName := d.Get("given_name").(string)
	organizationId := d.Get("organization_id").(string)
	password := d.Get("password").(string)
	surname := d.Get("surname").(string)
	username := d.Get("username").(string)

	request := &workdocs.CreateUserInput{
		EmailAddress:   aws.String(emailAddress),
		GivenName:      aws.String(givenName),
		OrganizationId: aws.String(organizationId),
		Password:       aws.String(password),
		Surname:        aws.String(surname),
		Username:       aws.String(username),
	}

	log.Println("[DEBUG] Create AD User request:", request)
	createResp, err := workdocsconn.CreateUser(request)
	if err != nil {
		return fmt.Errorf("Error creating AD User %s: %s", username, err)
	}

	d.SetId(aws.StringValue(createResp.User.Id))

	return resourceAwsAdUserRead(d, meta)
}

func resourceAwsAdUserRead(d *schema.ResourceData, meta interface{}) error {
	workdocsconn := meta.(*AWSClient).workdocsconn

	request := &workdocs.DescribeUsersInput{
		UserIds: aws.String(d.Id()),
	}

	var output *workdocs.DescribeUsersOutput

	var err error

	output, err = workdocsconn.DescribeUsers(request)

	if !d.IsNewResource() && len(output.Users) == 0 {
		log.Printf("[WARN] AD User (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading AD User (%s): %w", d.Id(), err)
	}

	if output == nil || len(output.Users) == 0 {
		return fmt.Errorf("error reading AD User (%s): empty response", d.Id())
	}

	user := output.Users[0]

	d.Set("email_address", user.EmailAddress)
	d.Set("given_name", user.GivenName)
	d.Set("organization_id", user.OrganizationId)
	d.Set("surname", user.Surname)
	d.Set("username", user.Username)
	d.Set("unique_id", user.Id)

	return nil
}

func resourceAwsAdUserUpdate(d *schema.ResourceData, meta interface{}) error {
	workdocsconn := meta.(*AWSClient).workdocsconn

	if d.HasChanges("given_name", "surname") {
		_, ng := d.GetChange("given_name")
		_, ns := d.GetChange("surname")

		request := &workdocs.UpdateUserInput{
			GivenName: aws.String(ng.(string)),
			Surname:   aws.String(ns.(string)),
			UserId:    aws.String(d.Id()),
		}

		log.Println("[DEBUG] Update AD User request:", request)
		_, err := workdocsconn.UpdateUser(request)
		if err != nil {
			return fmt.Errorf("Error updating AD User %s: %s", d.Id(), err)
		}
	}

	return resourceAwsAdUserRead(d, meta)
}

func resourceAwsAdUserDelete(d *schema.ResourceData, meta interface{}) error {
	workdocsconn := meta.(*AWSClient).workdocsconn

	deleteUserInput := &workdocs.DeleteUserInput{
		UserId: aws.String(d.Id()),
	}

	log.Println("[DEBUG] Delete AD User request:", deleteUserInput)
	_, err := workdocsconn.DeleteUser(deleteUserInput)

	if err != nil {
		return fmt.Errorf("Error deleting AD User %s: %s", d.Id(), err)
	}

	return nil
}
