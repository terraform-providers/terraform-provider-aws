package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsSagemakerNotebookInstanceLifeCycleConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSagemakerNotebookInstanceLifeCycleConfigurationCreate,
		Read:   resourceAwsSagemakerNotebookInstanceLifeCycleConfigurationRead,
		Update: resourceAwsSagemakerNotebookInstanceLifeCycleConfigurationUpdate,
		Delete: resourceAwsSagemakerNotebookInstanceLifeCycleConfigurationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateSagemakerName,
			},

			"on_create": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 16384),
			},

			"on_start": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 16384),
			},
		},
	}
}

func resourceAwsSagemakerNotebookInstanceLifeCycleConfigurationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	var name string
	if v, ok := d.GetOk("name"); ok {
		name = v.(string)
	} else {
		name = resource.UniqueId()
	}

	createOpts := &sagemaker.CreateNotebookInstanceLifecycleConfigInput{
		NotebookInstanceLifecycleConfigName: aws.String(name),
	}

	// on_create is technically a list of NotebookInstanceLifecycleHook elements, but the list has to be length 1
	// (same for on_start)
	if v, ok := d.GetOk("on_create"); ok {
		hook := &sagemaker.NotebookInstanceLifecycleHook{Content: aws.String(v.(string))}
		createOpts.OnCreate = []*sagemaker.NotebookInstanceLifecycleHook{hook}
	}

	if v, ok := d.GetOk("on_start"); ok {
		hook := &sagemaker.NotebookInstanceLifecycleHook{Content: aws.String(v.(string))}
		createOpts.OnStart = []*sagemaker.NotebookInstanceLifecycleHook{hook}
	}

	log.Printf("[DEBUG] SageMaker notebook instance lifecycle configuration create config: %#v", *createOpts)
	_, err := conn.CreateNotebookInstanceLifecycleConfig(createOpts)
	if err != nil {
		return fmt.Errorf("error creating SageMaker notebook instance lifecycle configuration: %s", err)
	}
	d.SetId(name)

	return resourceAwsSagemakerNotebookInstanceLifeCycleConfigurationRead(d, meta)
}

func resourceAwsSagemakerNotebookInstanceLifeCycleConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	request := &sagemaker.DescribeNotebookInstanceLifecycleConfigInput{
		NotebookInstanceLifecycleConfigName: aws.String(d.Id()),
	}

	lifecycleConfig, err := conn.DescribeNotebookInstanceLifecycleConfig(request)
	if err != nil {
		if isAWSErr(err, "ValidationException", "") {
			log.Printf("[INFO] unable to find the SageMaker notebook instance lifecycle configuration (%s); therefore it is removed from the state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading SageMaker notebook instance lifecycle configuration %s: %s", d.Id(), err)
	}

	if err := d.Set("name", lifecycleConfig.NotebookInstanceLifecycleConfigName); err != nil {
		return fmt.Errorf("error setting name for SageMaker notebook instance lifecycle configuration (%s): %s", d.Id(), err)
	}

	if len(lifecycleConfig.OnCreate) > 0 && lifecycleConfig.OnCreate[0] != nil {
		if err := d.Set("on_create", lifecycleConfig.OnCreate[0].Content); err != nil {
			return fmt.Errorf("error setting on_create for SageMaker notebook instance lifecycle configuration (%s): %s", d.Id(), err)
		}
	}

	if len(lifecycleConfig.OnStart) > 0 && lifecycleConfig.OnStart[0] != nil {
		if err := d.Set("on_start", lifecycleConfig.OnStart[0].Content); err != nil {
			return fmt.Errorf("error setting on_start for SageMaker notebook instance lifecycle configuration (%s): %s", d.Id(), err)
		}
	}

	if err := d.Set("arn", lifecycleConfig.NotebookInstanceLifecycleConfigArn); err != nil {
		return fmt.Errorf("error setting arn for SageMaker notebook instance lifecycle configuration (%s): %s", d.Id(), err)
	}

	return nil
}

func resourceAwsSagemakerNotebookInstanceLifeCycleConfigurationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	updateOpts := &sagemaker.UpdateNotebookInstanceLifecycleConfigInput{
		NotebookInstanceLifecycleConfigName: aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("on_create"); ok {
		onCreateHook := &sagemaker.NotebookInstanceLifecycleHook{Content: aws.String(v.(string))}
		updateOpts.OnCreate = []*sagemaker.NotebookInstanceLifecycleHook{onCreateHook}
	}

	if v, ok := d.GetOk("on_start"); ok {
		onStartHook := &sagemaker.NotebookInstanceLifecycleHook{Content: aws.String(v.(string))}
		updateOpts.OnStart = []*sagemaker.NotebookInstanceLifecycleHook{onStartHook}
	}

	_, err := conn.UpdateNotebookInstanceLifecycleConfig(updateOpts)
	if err != nil {
		return fmt.Errorf("error updating SageMaker Notebook Instance Lifecycle Configuration: %s", err)
	}
	return resourceAwsSagemakerNotebookInstanceLifeCycleConfigurationRead(d, meta)
}

func resourceAwsSagemakerNotebookInstanceLifeCycleConfigurationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	deleteOpts := &sagemaker.DeleteNotebookInstanceLifecycleConfigInput{
		NotebookInstanceLifecycleConfigName: aws.String(d.Id()),
	}
	log.Printf("[INFO] Deleting SageMaker Notebook Instance Lifecycle Configuration: %s", d.Id())

	_, err := conn.DeleteNotebookInstanceLifecycleConfig(deleteOpts)
	if err != nil {

		if isAWSErr(err, "ValidationException", "") {
			return nil
		}

		return fmt.Errorf("error deleting SageMaker Notebook Instance Lifecycle Configuration: %s", err)
	}
	return nil
}
