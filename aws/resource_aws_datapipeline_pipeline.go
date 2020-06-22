package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datapipeline"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsDataPipelinePipeline() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDataPipelinePipelineCreate,
		Read:   resourceAwsDataPipelinePipelineRead,
		Update: resourceAwsDataPipelinePipelineUpdate,
		Delete: resourceAwsDataPipelinePipelineDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"tags": tagsSchema(),
		},
	}
}

func resourceAwsDataPipelinePipelineCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).datapipelineconn

	uniqueID := resource.UniqueId()

	input := datapipeline.CreatePipelineInput{
		Name:     aws.String(d.Get("name").(string)),
		UniqueId: aws.String(uniqueID),
		Tags:     keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().DatapipelineTags(),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	resp, err := conn.CreatePipeline(&input)

	if err != nil {
		return fmt.Errorf("Error creating datapipeline: %s", err)
	}

	d.SetId(aws.StringValue(resp.PipelineId))

	return resourceAwsDataPipelinePipelineRead(d, meta)
}

func resourceAwsDataPipelinePipelineRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).datapipelineconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	v, err := resourceAwsDataPipelinePipelineRetrieve(d.Id(), conn)
	if isAWSErr(err, datapipeline.ErrCodePipelineNotFoundException, "") || isAWSErr(err, datapipeline.ErrCodePipelineDeletedException, "") || v == nil {
		log.Printf("[WARN] DataPipeline (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error describing DataPipeline (%s): %s", d.Id(), err)
	}

	d.Set("name", v.Name)
	d.Set("description", v.Description)
	if err := d.Set("tags", keyvaluetags.DatapipelineKeyValueTags(v.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsDataPipelinePipelineUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).datapipelineconn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.DatapipelineUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating Datapipeline Pipeline (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceAwsDataPipelinePipelineRead(d, meta)
}

func resourceAwsDataPipelinePipelineDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).datapipelineconn

	opts := datapipeline.DeletePipelineInput{
		PipelineId: aws.String(d.Id()),
	}

	_, err := conn.DeletePipeline(&opts)
	if isAWSErr(err, datapipeline.ErrCodePipelineNotFoundException, "") || isAWSErr(err, datapipeline.ErrCodePipelineDeletedException, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error deleting Data Pipeline %s: %s", d.Id(), err.Error())
	}

	return waitForDataPipelineDeletion(conn, d.Id())
}

func resourceAwsDataPipelinePipelineRetrieve(id string, conn *datapipeline.DataPipeline) (*datapipeline.PipelineDescription, error) {
	opts := datapipeline.DescribePipelinesInput{
		PipelineIds: []*string{aws.String(id)},
	}

	resp, err := conn.DescribePipelines(&opts)
	if err != nil {
		return nil, err
	}

	var pipeline *datapipeline.PipelineDescription

	for _, p := range resp.PipelineDescriptionList {
		if p == nil {
			continue
		}

		if aws.StringValue(p.PipelineId) == id {
			pipeline = p
			break
		}
	}

	return pipeline, nil
}

func waitForDataPipelineDeletion(conn *datapipeline.DataPipeline, pipelineID string) error {
	params := &datapipeline.DescribePipelinesInput{
		PipelineIds: []*string{aws.String(pipelineID)},
	}
	return resource.Retry(10*time.Minute, func() *resource.RetryError {
		_, err := conn.DescribePipelines(params)
		if isAWSErr(err, datapipeline.ErrCodePipelineNotFoundException, "") || isAWSErr(err, datapipeline.ErrCodePipelineDeletedException, "") {
			return nil
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}
		return resource.RetryableError(fmt.Errorf("DataPipeline (%s) still exists", pipelineID))
	})
}
