package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_run_properties": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     schema.TypeString,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"max_concurrent_runs": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsGlueWorkflowCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn
	name := d.Get("name").(string)

	input := &glue.CreateWorkflowInput{
		Name: aws.String(name),
		Tags: keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().GlueTags(),
	}

	if kv, ok := d.GetOk("default_run_properties"); ok {
		input.DefaultRunProperties = stringMapToPointers(kv.(map[string]interface{}))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("max_concurrent_runs"); ok {
		input.MaxConcurrentRuns = aws.Int64(int64(v.(int)))
	}

	log.Printf("[DEBUG] Creating Glue Workflow: %s", input)
	_, err := conn.CreateWorkflow(input)
	if err != nil {
		return fmt.Errorf("error creating Glue Trigger (%s): %w", name, err)
	}
	d.SetId(name)

	return resourceAwsGlueWorkflowRead(d, meta)
}

func resourceAwsGlueWorkflowRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &glue.GetWorkflowInput{
		Name: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading Glue Workflow: %#v", input)
	output, err := conn.GetWorkflow(input)
	if err != nil {
		if isAWSErr(err, glue.ErrCodeEntityNotFoundException, "") {
			log.Printf("[WARN] Glue Workflow (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading Glue Workflow (%s): %w", d.Id(), err)
	}

	workflow := output.Workflow
	if workflow == nil {
		log.Printf("[WARN] Glue Workflow (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	workFlowArn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "glue",
		Region:    meta.(*AWSClient).region,
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("workflow/%s", d.Id()),
	}.String()
	d.Set("arn", workFlowArn)

	if err := d.Set("default_run_properties", aws.StringValueMap(workflow.DefaultRunProperties)); err != nil {
		return fmt.Errorf("error setting default_run_properties: %w", err)
	}
	d.Set("description", workflow.Description)
	d.Set("max_concurrent_runs", workflow.MaxConcurrentRuns)
	d.Set("name", workflow.Name)

	tags, err := keyvaluetags.GlueListTags(conn, workFlowArn)

	if err != nil {
		return fmt.Errorf("error listing tags for Glue Workflow (%s): %w", workFlowArn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}

func resourceAwsGlueWorkflowUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn

	if d.HasChanges("default_run_properties", "description", "max_concurrent_runs") {
		input := &glue.UpdateWorkflowInput{
			Name: aws.String(d.Get("name").(string)),
		}

		if kv, ok := d.GetOk("default_run_properties"); ok {
			input.DefaultRunProperties = stringMapToPointers(kv.(map[string]interface{}))
		}

		if v, ok := d.GetOk("description"); ok {
			input.Description = aws.String(v.(string))
		}

		if v, ok := d.GetOk("max_concurrent_runs"); ok {
			input.MaxConcurrentRuns = aws.Int64(int64(v.(int)))
		}

		log.Printf("[DEBUG] Updating Glue Workflow: %#v", input)
		_, err := conn.UpdateWorkflow(input)
		if err != nil {
			return fmt.Errorf("error updating Glue Workflow (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.GlueUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsGlueWorkflowRead(d, meta)
}

func resourceAwsGlueWorkflowDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn

	log.Printf("[DEBUG] Deleting Glue Workflow: %s", d.Id())
	err := deleteWorkflow(conn, d.Id())
	if err != nil {
		return fmt.Errorf("error deleting Glue Workflow (%s): %w", d.Id(), err)
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
