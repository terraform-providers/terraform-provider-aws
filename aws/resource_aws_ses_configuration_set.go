package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsSesConfigurationSet() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSesConfigurationSetCreate,
		Read:   resourceAwsSesConfigurationSetRead,
		Delete: resourceAwsSesConfigurationSetDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsSesConfigurationSetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesconn

	configurationSetName := d.Get("name").(string)

	createOpts := &ses.CreateConfigurationSetInput{
		ConfigurationSet: &ses.ConfigurationSet{
			Name: aws.String(configurationSetName),
		},
	}

	_, err := conn.CreateConfigurationSet(createOpts)
	if err != nil {
		return fmt.Errorf("Error creating SES configuration set: %s", err)
	}

	d.SetId(configurationSetName)

	return resourceAwsSesConfigurationSetRead(d, meta)
}

func resourceAwsSesConfigurationSetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesconn

	configSetInput := &ses.DescribeConfigurationSetInput{
		ConfigurationSetName: aws.String(d.Id()),
	}

	response, err := conn.DescribeConfigurationSet(configSetInput)

	if err != nil {
		if isAWSErr(err, ses.ErrCodeConfigurationSetDoesNotExistException, "") {
			log.Printf("[WARN] SES Configuration Set (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("name", aws.StringValue(response.ConfigurationSet.Name))

	return nil
}

func resourceAwsSesConfigurationSetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesconn

	log.Printf("[DEBUG] SES Delete Configuration Rule Set: %s", d.Id())
	_, err := conn.DeleteConfigurationSet(&ses.DeleteConfigurationSetInput{
		ConfigurationSetName: aws.String(d.Id()),
	})

	return err
}
