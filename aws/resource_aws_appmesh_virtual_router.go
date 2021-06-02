package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/appmesh/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsAppmeshVirtualRouter() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		Create: resourceAwsAppmeshVirtualRouterCreate,
		Read:   resourceAwsAppmeshVirtualRouterRead,
		Update: resourceAwsAppmeshVirtualRouterUpdate,
		Delete: resourceAwsAppmeshVirtualRouterDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsAppmeshVirtualRouterImport,
		},

		SchemaVersion: 1,
		MigrateState:  resourceAwsAppmeshVirtualRouterMigrateState,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},

			"mesh_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},

			"mesh_owner": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validateAwsAccountId,
			},

			"spec": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"listener": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"port_mapping": {
										Type:     schema.TypeList,
										Required: true,
										MinItems: 1,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"port": {
													Type:         schema.TypeInt,
													Required:     true,
													ValidateFunc: validation.IsPortNumber,
												},

												"protocol": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(appmesh.PortProtocol_Values(), false),
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"last_updated_date": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"resource_owner": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tagsSchema(),

			"tags_all": tagsSchemaComputed(),
		},

		CustomizeDiff: SetTagsDiff,
	}
}

func resourceAwsAppmeshVirtualRouterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appmeshconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	req := &appmesh.CreateVirtualRouterInput{
		MeshName:          aws.String(d.Get("mesh_name").(string)),
		VirtualRouterName: aws.String(d.Get("name").(string)),
		Spec:              expandAppmeshVirtualRouterSpec(d.Get("spec").([]interface{})),
		Tags:              tags.IgnoreAws().AppmeshTags(),
	}
	if v, ok := d.GetOk("mesh_owner"); ok {
		req.MeshOwner = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating App Mesh virtual router: %#v", req)
	resp, err := conn.CreateVirtualRouter(req)
	if err != nil {
		return fmt.Errorf("error creating App Mesh virtual router: %s", err)
	}

	d.SetId(aws.StringValue(resp.VirtualRouter.Metadata.Uid))

	return resourceAwsAppmeshVirtualRouterRead(d, meta)
}

func resourceAwsAppmeshVirtualRouterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appmeshconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	req := &appmesh.DescribeVirtualRouterInput{
		MeshName:          aws.String(d.Get("mesh_name").(string)),
		VirtualRouterName: aws.String(d.Get("name").(string)),
	}
	if v, ok := d.GetOk("mesh_owner"); ok {
		req.MeshOwner = aws.String(v.(string))
	}

	var resp *appmesh.DescribeVirtualRouterOutput

	err := resource.Retry(waiter.PropagationTimeout, func() *resource.RetryError {
		var err error

		resp, err = conn.DescribeVirtualRouter(req)

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, appmesh.ErrCodeNotFoundException) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		resp, err = conn.DescribeVirtualRouter(req)
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, appmesh.ErrCodeNotFoundException) {
		log.Printf("[WARN] App Mesh Virtual Router (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading App Mesh Virtual Router: %w", err)
	}

	if resp == nil || resp.VirtualRouter == nil {
		return fmt.Errorf("error reading App Mesh Virtual Router: empty response")
	}

	if aws.StringValue(resp.VirtualRouter.Status.Status) == appmesh.VirtualRouterStatusCodeDeleted {
		if d.IsNewResource() {
			return fmt.Errorf("error reading App Mesh Virtual Router: %s after creation", aws.StringValue(resp.VirtualRouter.Status.Status))
		}

		log.Printf("[WARN] App Mesh Virtual Router (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	arn := aws.StringValue(resp.VirtualRouter.Metadata.Arn)
	d.Set("name", resp.VirtualRouter.VirtualRouterName)
	d.Set("mesh_name", resp.VirtualRouter.MeshName)
	d.Set("mesh_owner", resp.VirtualRouter.Metadata.MeshOwner)
	d.Set("arn", arn)
	d.Set("created_date", resp.VirtualRouter.Metadata.CreatedAt.Format(time.RFC3339))
	d.Set("last_updated_date", resp.VirtualRouter.Metadata.LastUpdatedAt.Format(time.RFC3339))
	d.Set("resource_owner", resp.VirtualRouter.Metadata.ResourceOwner)
	err = d.Set("spec", flattenAppmeshVirtualRouterSpec(resp.VirtualRouter.Spec))
	if err != nil {
		return fmt.Errorf("error setting spec: %s", err)
	}

	tags, err := keyvaluetags.AppmeshListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for App Mesh virtual router (%s): %s", arn, err)
	}

	tags = tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceAwsAppmeshVirtualRouterUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appmeshconn

	if d.HasChange("spec") {
		_, v := d.GetChange("spec")
		req := &appmesh.UpdateVirtualRouterInput{
			MeshName:          aws.String(d.Get("mesh_name").(string)),
			VirtualRouterName: aws.String(d.Get("name").(string)),
			Spec:              expandAppmeshVirtualRouterSpec(v.([]interface{})),
		}
		if v, ok := d.GetOk("mesh_owner"); ok {
			req.MeshOwner = aws.String(v.(string))
		}

		log.Printf("[DEBUG] Updating App Mesh virtual router: %#v", req)
		_, err := conn.UpdateVirtualRouter(req)
		if err != nil {
			return fmt.Errorf("error updating App Mesh virtual router: %s", err)
		}
	}

	arn := d.Get("arn").(string)
	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := keyvaluetags.AppmeshUpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating App Mesh virtual router (%s) tags: %s", arn, err)
		}
	}

	return resourceAwsAppmeshVirtualRouterRead(d, meta)
}

func resourceAwsAppmeshVirtualRouterDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appmeshconn

	log.Printf("[DEBUG] Deleting App Mesh virtual router: %s", d.Id())
	_, err := conn.DeleteVirtualRouter(&appmesh.DeleteVirtualRouterInput{
		MeshName:          aws.String(d.Get("mesh_name").(string)),
		VirtualRouterName: aws.String(d.Get("name").(string)),
	})
	if isAWSErr(err, appmesh.ErrCodeNotFoundException, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting App Mesh virtual router: %s", err)
	}

	return nil
}

func resourceAwsAppmeshVirtualRouterImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("Wrong format of resource: %s. Please follow 'mesh-name/virtual-router-name'", d.Id())
	}

	mesh := parts[0]
	name := parts[1]
	log.Printf("[DEBUG] Importing App Mesh virtual router %s from mesh %s", name, mesh)

	conn := meta.(*AWSClient).appmeshconn

	resp, err := conn.DescribeVirtualRouter(&appmesh.DescribeVirtualRouterInput{
		MeshName:          aws.String(mesh),
		VirtualRouterName: aws.String(name),
	})
	if err != nil {
		return nil, err
	}

	d.SetId(aws.StringValue(resp.VirtualRouter.Metadata.Uid))
	d.Set("name", resp.VirtualRouter.VirtualRouterName)
	d.Set("mesh_name", resp.VirtualRouter.MeshName)

	return []*schema.ResourceData{d}, nil
}
