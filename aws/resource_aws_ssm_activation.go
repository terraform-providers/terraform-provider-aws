package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsSsmActivation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSsmActivationCreate,
		Read:   resourceAwsSsmActivationRead,
		Delete: resourceAwsSsmActivationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"expired": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"expiration_date": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsRFC3339Time,
			},
			"iam_role": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"registration_limit": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"registration_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"activation_code": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchemaForceNew(),
		},
	}
}

func resourceAwsSsmActivationCreate(d *schema.ResourceData, meta interface{}) error {
	ssmconn := meta.(*AWSClient).ssmconn

	log.Printf("[DEBUG] SSM activation create: %s", d.Id())

	activationInput := &ssm.CreateActivationInput{
		IamRole: aws.String(d.Get("name").(string)),
	}

	if _, ok := d.GetOk("name"); ok {
		activationInput.DefaultInstanceName = aws.String(d.Get("name").(string))
	}

	if _, ok := d.GetOk("description"); ok {
		activationInput.Description = aws.String(d.Get("description").(string))
	}

	if v, ok := d.GetOk("expiration_date"); ok {
		t, _ := time.Parse(time.RFC3339, v.(string))
		activationInput.ExpirationDate = aws.Time(t)
	}

	if _, ok := d.GetOk("iam_role"); ok {
		activationInput.IamRole = aws.String(d.Get("iam_role").(string))
	}

	if _, ok := d.GetOk("registration_limit"); ok {
		activationInput.RegistrationLimit = aws.Int64(int64(d.Get("registration_limit").(int)))
	}
	if v, ok := d.GetOk("tags"); ok {
		activationInput.Tags = keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().SsmTags()
	}

	// Retry to allow iam_role to be created and policy attachment to take place
	var resp *ssm.CreateActivationOutput
	err := resource.Retry(30*time.Second, func() *resource.RetryError {
		var err error

		resp, err = ssmconn.CreateActivation(activationInput)

		if err != nil {
			return resource.RetryableError(err)
		}

		return resource.NonRetryableError(err)
	})

	if isResourceTimeoutError(err) {
		resp, err = ssmconn.CreateActivation(activationInput)
	}

	if err != nil {
		return fmt.Errorf("Error creating SSM activation: %s", err)
	}

	if resp.ActivationId == nil {
		return fmt.Errorf("ActivationId was nil")
	}
	d.SetId(*resp.ActivationId)
	d.Set("activation_code", resp.ActivationCode)

	return resourceAwsSsmActivationRead(d, meta)
}

func resourceAwsSsmActivationRead(d *schema.ResourceData, meta interface{}) error {
	ssmconn := meta.(*AWSClient).ssmconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	log.Printf("[DEBUG] Reading SSM Activation: %s", d.Id())

	params := &ssm.DescribeActivationsInput{
		Filters: []*ssm.DescribeActivationsFilter{
			{
				FilterKey: aws.String("ActivationIds"),
				FilterValues: []*string{
					aws.String(d.Id()),
				},
			},
		},
		MaxResults: aws.Int64(1),
	}

	resp, err := ssmconn.DescribeActivations(params)

	if err != nil {
		return fmt.Errorf("Error reading SSM activation: %s", err)
	}
	if resp.ActivationList == nil || len(resp.ActivationList) == 0 {
		log.Printf("[WARN] SSM Activation (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	activation := resp.ActivationList[0] // Only 1 result as MaxResults is 1 above
	d.Set("name", activation.DefaultInstanceName)
	d.Set("description", activation.Description)
	d.Set("expiration_date", aws.TimeValue(activation.ExpirationDate).Format(time.RFC3339))
	d.Set("expired", activation.Expired)
	d.Set("iam_role", activation.IamRole)
	d.Set("registration_limit", activation.RegistrationLimit)
	d.Set("registration_count", activation.RegistrationsCount)
	if err := d.Set("tags", keyvaluetags.SsmKeyValueTags(activation.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsSsmActivationDelete(d *schema.ResourceData, meta interface{}) error {
	ssmconn := meta.(*AWSClient).ssmconn

	log.Printf("[DEBUG] Deleting SSM Activation: %s", d.Id())

	params := &ssm.DeleteActivationInput{
		ActivationId: aws.String(d.Id()),
	}

	_, err := ssmconn.DeleteActivation(params)

	if err != nil {
		return fmt.Errorf("Error deleting SSM activation: %s", err)
	}

	return nil
}
