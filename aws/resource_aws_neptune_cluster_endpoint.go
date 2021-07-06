package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/neptune/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/neptune/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsNeptuneClusterEndpoint() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsNeptuneClusterEndpointCreate,
		Read:   resourceAwsNeptuneClusterEndpointRead,
		Update: resourceAwsNeptuneClusterEndpointUpdate,
		Delete: resourceAwsNeptuneClusterEndpointDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_identifier": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateNeptuneIdentifier,
			},
			"cluster_endpoint_identifier": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateNeptuneIdentifier,
			},
			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"endpoint_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"READER", "WRITER", "ANY"}, false),
			},
			"static_members": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
			"excluded_members": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
		},

		CustomizeDiff: SetTagsDiff,
	}
}

func resourceAwsNeptuneClusterEndpointCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).neptuneconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	input := &neptune.CreateDBClusterEndpointInput{
		DBClusterEndpointIdentifier: aws.String(d.Get("cluster_endpoint_identifier").(string)),
		DBClusterIdentifier:         aws.String(d.Get("cluster_identifier").(string)),
		EndpointType:                aws.String(d.Get("endpoint_type").(string)),
	}

	if attr := d.Get("static_members").(*schema.Set); attr.Len() > 0 {
		input.StaticMembers = expandStringSet(attr)
	}

	if attr := d.Get("excluded_members").(*schema.Set); attr.Len() > 0 {
		input.ExcludedMembers = expandStringSet(attr)
	}

	// Tags are currently only supported in AWS Commercial.
	if len(tags) > 0 && meta.(*AWSClient).partition == endpoints.AwsPartitionID {
		input.Tags = tags.IgnoreAws().NeptuneTags()
	}

	out, err := conn.CreateDBClusterEndpoint(input)
	if err != nil {
		return fmt.Errorf("error creating Neptune Cluster Endpoint: %w", err)
	}

	clusterId := aws.StringValue(out.DBClusterIdentifier)
	endpointId := aws.StringValue(out.DBClusterEndpointIdentifier)
	d.SetId(fmt.Sprintf("%s:%s", clusterId, endpointId))

	_, err = waiter.DBClusterEndpointAvailable(conn, d.Id())
	if err != nil {
		return fmt.Errorf("error waiting for Neptune Cluster Endpoint (%q) to be Available: %w", d.Id(), err)
	}

	return resourceAwsNeptuneClusterEndpointRead(d, meta)

}

func resourceAwsNeptuneClusterEndpointRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).neptuneconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	resp, err := finder.EndpointById(conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		d.SetId("")
		log.Printf("[DEBUG] Neptune Cluster Endpoint (%s) not found", d.Id())
		return nil
	}

	if err != nil {
		return fmt.Errorf("error describing Neptune Cluster Endpoint (%s): %w", d.Id(), err)
	}

	d.Set("cluster_endpoint_identifier", resp.DBClusterEndpointIdentifier)
	d.Set("cluster_identifier", resp.DBClusterIdentifier)
	d.Set("endpoint_type", resp.CustomEndpointType)
	d.Set("endpoint", resp.Endpoint)
	d.Set("excluded_members", flattenStringSet(resp.ExcludedMembers))
	d.Set("static_members", flattenStringSet(resp.StaticMembers))

	arn := aws.StringValue(resp.DBClusterEndpointArn)
	d.Set("arn", arn)

	// Tags are currently only supported in AWS Commercial.
	if meta.(*AWSClient).partition == endpoints.AwsPartitionID {
		tags, err := keyvaluetags.NeptuneListTags(conn, arn)

		if err != nil {
			return fmt.Errorf("error listing tags for Neptune Cluster Endpoint (%s): %w", arn, err)
		}

		tags = tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig)

		//lintignore:AWSR002
		if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
			return fmt.Errorf("error setting tags: %w", err)
		}

		if err := d.Set("tags_all", tags.Map()); err != nil {
			return fmt.Errorf("error setting tags_all: %w", err)
		}
	} else {
		d.Set("tags", nil)
		d.Set("tags_all", nil)
	}

	return nil
}

func resourceAwsNeptuneClusterEndpointUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).neptuneconn

	if d.HasChangeExcept("tags_all") {
		req := &neptune.ModifyDBClusterEndpointInput{
			DBClusterEndpointIdentifier: aws.String(d.Get("cluster_endpoint_identifier").(string)),
		}

		if d.HasChange("endpoint_type") {
			req.EndpointType = aws.String(d.Get("endpoint_type").(string))
		}

		if d.HasChange("static_members") {
			req.StaticMembers = expandStringSet(d.Get("static_members").(*schema.Set))
		}

		if d.HasChange("excluded_members") {
			req.ExcludedMembers = expandStringSet(d.Get("excluded_members").(*schema.Set))
		}

		_, err := conn.ModifyDBClusterEndpoint(req)
		if err != nil {
			return fmt.Errorf("error updating Neptune Cluster Endpoint (%q): %w", d.Id(), err)
		}

		_, err = waiter.DBClusterEndpointAvailable(conn, d.Id())
		if err != nil {
			return fmt.Errorf("error waiting for Neptune Cluster Endpoint (%q) to be Available: %w", d.Id(), err)
		}
	}

	// Tags are currently only supported in AWS Commercial.
	if d.HasChange("tags_all") && meta.(*AWSClient).partition == endpoints.AwsPartitionID {
		o, n := d.GetChange("tags_all")

		if err := keyvaluetags.NeptuneUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating Neptune Cluster Endpoint (%s) tags: %w", d.Get("arn").(string), err)
		}
	}

	return resourceAwsNeptuneClusterEndpointRead(d, meta)
}

func resourceAwsNeptuneClusterEndpointDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).neptuneconn

	endpointId := d.Get("cluster_endpoint_identifier").(string)
	input := &neptune.DeleteDBClusterEndpointInput{
		DBClusterEndpointIdentifier: aws.String(endpointId),
	}

	_, err := conn.DeleteDBClusterEndpoint(input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, neptune.ErrCodeDBClusterEndpointNotFoundFault) ||
			tfawserr.ErrCodeEquals(err, neptune.ErrCodeDBClusterNotFoundFault) {
			return nil
		}
		return fmt.Errorf("Neptune Cluster Endpoint cannot be deleted: %w", err)
	}
	_, err = waiter.DBClusterEndpointDeleted(conn, d.Id())
	if err != nil {
		if tfawserr.ErrCodeEquals(err, neptune.ErrCodeDBClusterEndpointNotFoundFault) {
			return nil
		}
		return fmt.Errorf("error waiting for Neptune Cluster Endpoint (%q) to be Deleted: %w", d.Id(), err)
	}

	return nil
}
