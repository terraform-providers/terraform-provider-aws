package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/sagemaker/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/sagemaker/waiter"
)

func resourceAwsSagemakerApp() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSagemakerAppCreate,
		Read:   resourceAwsSagemakerAppRead,
		Update: resourceAwsSagemakerAppUpdate,
		Delete: resourceAwsSagemakerAppDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"app_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 63),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9](-*[a-zA-Z0-9]){0,62}`), "Valid characters are a-z, A-Z, 0-9, and - (hyphen)."),
				),
			},
			"app_type": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validation.StringInSlice(sagemaker.AppType_Values(), false),
			},
			"domain_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"resource_spec": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"instance_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(sagemaker.AppInstanceType_Values(), false),
						},
						"sagemaker_image_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validateArn,
						},
					},
				},
			},
			"tags": tagsSchema(),
			"user_profile_name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
		},
	}
}

func resourceAwsSagemakerAppCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	input := &sagemaker.CreateAppInput{
		AppName:         aws.String(d.Get("app_name").(string)),
		AppType:         aws.String(d.Get("app_type").(string)),
		DomainId:        aws.String(d.Get("domain_id").(string)),
		UserProfileName: aws.String(d.Get("user_profile_name").(string)),
	}

	if v, ok := d.GetOk("tags"); ok {
		input.Tags = keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().SagemakerTags()
	}

	if v, ok := d.GetOk("resource_spec"); ok {
		input.ResourceSpec = expandSagemakerDomainDefaultResourceSpec(v.([]interface{}))
	}

	log.Printf("[DEBUG] Sagemaker App create config: %#v", *input)
	output, err := conn.CreateApp(input)
	if err != nil {
		return fmt.Errorf("error creating SageMaker App: %w", err)
	}

	appArn := aws.StringValue(output.AppArn)
	domainID, userProfileName, appType, appName, err := decodeSagemakerAppID(appArn)
	if err != nil {
		return err
	}

	d.SetId(appArn)

	if _, err := waiter.AppInService(conn, domainID, userProfileName, appType, appName); err != nil {
		return fmt.Errorf("error waiting for SageMaker App (%s) to create: %w", d.Id(), err)
	}

	return resourceAwsSagemakerAppRead(d, meta)
}

func resourceAwsSagemakerAppRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	domainID, userProfileName, appType, appName, err := decodeSagemakerAppID(d.Id())
	if err != nil {
		return err
	}

	app, err := finder.AppByName(conn, domainID, userProfileName, appType, appName)
	if err != nil {
		if isAWSErr(err, sagemaker.ErrCodeResourceNotFound, "") {
			d.SetId("")
			log.Printf("[WARN] Unable to find SageMaker App (%s), removing from state", d.Id())
			return nil
		}
		return fmt.Errorf("error reading SageMaker App (%s): %w", d.Id(), err)
	}

	if aws.StringValue(app.Status) == sagemaker.AppStatusDeleted {
		d.SetId("")
		log.Printf("[WARN] Unable to find SageMaker App (%s), removing from state", d.Id())
		return nil
	}

	arn := aws.StringValue(app.AppArn)
	d.Set("app_name", app.AppName)
	d.Set("app_type", app.AppType)
	d.Set("arn", arn)
	d.Set("domain_id", app.DomainId)
	d.Set("user_profile_name", app.UserProfileName)

	if err := d.Set("resource_spec", flattenSagemakerDomainDefaultResourceSpec(app.ResourceSpec)); err != nil {
		return fmt.Errorf("error setting resource_spec for SageMaker App (%s): %w", d.Id(), err)
	}

	tags, err := keyvaluetags.SagemakerListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for SageMaker App (%s): %w", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}

func resourceAwsSagemakerAppUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.SagemakerUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating SageMaker App (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceAwsSagemakerAppRead(d, meta)
}

func resourceAwsSagemakerAppDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	appName := d.Get("app_name").(string)
	appType := d.Get("app_type").(string)
	domainID := d.Get("domain_id").(string)
	userProfileName := d.Get("user_profile_name").(string)

	input := &sagemaker.DeleteAppInput{
		AppName:         aws.String(appName),
		AppType:         aws.String(appType),
		DomainId:        aws.String(domainID),
		UserProfileName: aws.String(userProfileName),
	}

	if _, err := conn.DeleteApp(input); err != nil {

		if isAWSErr(err, "ValidationException", "has already been deleted") ||
			isAWSErr(err, "ValidationException", "previously failed and was automatically deleted") {
			return nil
		}

		if !isAWSErr(err, sagemaker.ErrCodeResourceNotFound, "") {
			return fmt.Errorf("error deleting SageMaker App (%s): %w", d.Id(), err)
		}
	}

	if _, err := waiter.AppDeleted(conn, domainID, userProfileName, appType, appName); err != nil {
		if !isAWSErr(err, sagemaker.ErrCodeResourceNotFound, "") {
			return fmt.Errorf("error waiting for SageMaker App (%s) to delete: %w", d.Id(), err)
		}
	}

	return nil
}

func decodeSagemakerAppID(id string) (string, string, string, string, error) {
	appArn, err := arn.Parse(id)
	if err != nil {
		return "", "", "", "", err
	}

	appResourceName := strings.TrimPrefix(appArn.Resource, "app/")
	parts := strings.Split(appResourceName, "/")

	if len(parts) != 4 {
		return "", "", "", "", fmt.Errorf("Unexpected format of ID (%q), expected DOMAIN-ID/USER-PROFILE-NAME/APP-TYPE/APP-NAME", appResourceName)
	}

	domainID := parts[0]
	userProfileName := parts[1]
	appType := parts[2]

	if appType == "jupyterserver" {
		appType = sagemaker.AppTypeJupyterServer
	} else if appType == "kernelgateway" {
		appType = sagemaker.AppTypeKernelGateway
	} else if appType == "tensorboard" {
		appType = sagemaker.AppTypeTensorBoard
	}

	appName := parts[3]

	return domainID, userProfileName, appType, appName, nil
}
