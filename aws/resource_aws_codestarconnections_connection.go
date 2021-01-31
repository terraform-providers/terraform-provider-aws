package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codestarconnections"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsCodeStarConnectionsConnection() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCodeStarConnectionsConnectionCreate,
		Read:   resourceAwsCodeStarConnectionsConnectionRead,
		Delete: resourceAwsCodeStarConnectionsConnectionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"connection_status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"provider_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(codestarconnections.ProviderType_Values(), false),
			},
		},
	}
}

func resourceAwsCodeStarConnectionsConnectionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codestarconnectionsconn

	params := &codestarconnections.CreateConnectionInput{
		ConnectionName: aws.String(d.Get("name").(string)),
		ProviderType:   aws.String(d.Get("provider_type").(string)),
	}

	resp, err := conn.CreateConnection(params)
	if err != nil {
		return fmt.Errorf("error creating CodeStar connection: %w", err)
	}

	d.SetId(aws.StringValue(resp.ConnectionArn))

	return resourceAwsCodeStarConnectionsConnectionRead(d, meta)
}

func resourceAwsCodeStarConnectionsConnectionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codestarconnectionsconn

	resp, err := conn.GetConnection(&codestarconnections.GetConnectionInput{
		ConnectionArn: aws.String(d.Id()),
	})
	if tfawserr.ErrCodeEquals(err, codestarconnections.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] CodeStar connection (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading CodeStar connection: %w", err)
	}

	if resp == nil || resp.Connection == nil {
		return fmt.Errorf("error reading CodeStar connection (%s): empty response", d.Id())
	}

	d.SetId(aws.StringValue(resp.Connection.ConnectionArn))
	d.Set("arn", resp.Connection.ConnectionArn)
	d.Set("name", resp.Connection.ConnectionName)
	d.Set("connection_status", resp.Connection.ConnectionStatus)
	d.Set("provider_type", resp.Connection.ProviderType)

	return nil
}

func resourceAwsCodeStarConnectionsConnectionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codestarconnectionsconn

	_, err := conn.DeleteConnection(&codestarconnections.DeleteConnectionInput{
		ConnectionArn: aws.String(d.Id()),
	})
	if tfawserr.ErrCodeEquals(err, codestarconnections.ErrCodeResourceNotFoundException) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting CodeStar connection: %w", err)
	}

	return nil
}
