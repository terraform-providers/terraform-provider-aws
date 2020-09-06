package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
)

func resourceAwsGlueWorkflow() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGlueWorkflowCreate,
		Read:   resourceAwsGlueWorkflowRead,
		Update: resourceAwsGlueWorkflowUpdate,
		Delete: resourceAwsGlueWorkflowDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"default_run_properties": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     schema.TypeString,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
		},
	}
}

func resourceAwsGlueWorkflowCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn
	name := d.Get("name").(string)

	input := &glue.CreateWorkflowInput{
		Name: aws.String(name),
	}

	if kv, ok := d.GetOk("default_run_properties"); ok {
		defaultRunPropertiesMap := make(map[string]string)
		for k, v := range kv.(map[string]interface{}) {
			defaultRunPropertiesMap[k] = v.(string)
		}
		input.DefaultRunProperties = aws.StringMap(defaultRunPropertiesMap)
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Glue Workflow: %s", input)
	_, err := conn.CreateWorkflow(input)
	if err != nil {
		return fmt.Errorf("error creating Glue Trigger (%s): %s", name, err)
	}
	d.SetId(name)

	return resourceAwsGlueWorkflowRead(d, meta)
}

func resourceAwsGlueWorkflowRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn

	input := &glue.GetWorkflowInput{
		Name: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading Glue Workflow: %s", input)
	output, err := conn.GetWorkflow(input)
	if err != nil {
		if isAWSErr(err, glue.ErrCodeEntityNotFoundException, "") {
			log.Printf("[WARN] Glue Workflow (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading Glue Workflow (%s): %s", d.Id(), err)
	}

	workflow := output.Workflow
	if workflow == nil {
		log.Printf("[WARN] Glue Workflow (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err := d.Set("default_run_properties", aws.StringValueMap(workflow.DefaultRunProperties)); err != nil {
		return fmt.Errorf("error setting default_run_properties: %s", err)
	}
	d.Set("description", workflow.Description)
	d.Set("name", workflow.Name)

	return nil
}

func resourceAwsGlueWorkflowUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn

	input := &glue.UpdateWorkflowInput{
		Name: aws.String(d.Get("name").(string)),
	}

	if kv, ok := d.GetOk("default_run_properties"); ok {
		defaultRunPropertiesMap := make(map[string]string)
		for k, v := range kv.(map[string]interface{}) {
			defaultRunPropertiesMap[k] = v.(string)
		}
		input.DefaultRunProperties = aws.StringMap(defaultRunPropertiesMap)
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Updating Glue Workflow: %s", input)
	_, err := conn.UpdateWorkflow(input)
	if err != nil {
		return fmt.Errorf("error updating Glue Workflow (%s): %s", d.Id(), err)
	}

	return resourceAwsGlueWorkflowRead(d, meta)
}

func resourceAwsGlueWorkflowDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn

	log.Printf("[DEBUG] Deleting Glue Workflow: %s", d.Id())
	err := deleteWorkflow(conn, d.Id())
	if err != nil {
		return fmt.Errorf("error deleting Glue Workflow (%s): %s", d.Id(), err)
	}

	return nil
}

func deleteWorkflow(conn *glue.Glue, name string) error {
	input := &glue.DeleteWorkflowInput{
		Name: aws.String(name),
	}

	_, err := conn.DeleteWorkflow(input)
	if err != nil {
		if isAWSErr(err, glue.ErrCodeEntityNotFoundException, "") {
			return nil
		}
		return err
	}

	return nil
}
