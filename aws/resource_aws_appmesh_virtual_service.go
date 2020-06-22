package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsAppmeshVirtualService() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAppmeshVirtualServiceCreate,
		Read:   resourceAwsAppmeshVirtualServiceRead,
		Update: resourceAwsAppmeshVirtualServiceUpdate,
		Delete: resourceAwsAppmeshVirtualServiceDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsAppmeshVirtualServiceImport,
		},

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

			"spec": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"provider": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 0,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"virtual_node": {
										Type:          schema.TypeList,
										Optional:      true,
										MinItems:      0,
										MaxItems:      1,
										ConflictsWith: []string{"spec.0.provider.0.virtual_router"},
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"virtual_node_name": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(1, 255),
												},
											},
										},
									},

									"virtual_router": {
										Type:          schema.TypeList,
										Optional:      true,
										MinItems:      0,
										MaxItems:      1,
										ConflictsWith: []string{"spec.0.provider.0.virtual_node"},
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"virtual_router_name": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(1, 255),
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

			"tags": tagsSchema(),
		},
	}
}

func resourceAwsAppmeshVirtualServiceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appmeshconn

	req := &appmesh.CreateVirtualServiceInput{
		MeshName:           aws.String(d.Get("mesh_name").(string)),
		VirtualServiceName: aws.String(d.Get("name").(string)),
		Spec:               expandAppmeshVirtualServiceSpec(d.Get("spec").([]interface{})),
		Tags:               keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().AppmeshTags(),
	}

	log.Printf("[DEBUG] Creating App Mesh virtual service: %#v", req)
	resp, err := conn.CreateVirtualService(req)
	if err != nil {
		return fmt.Errorf("error creating App Mesh virtual service: %s", err)
	}

	d.SetId(aws.StringValue(resp.VirtualService.Metadata.Uid))

	return resourceAwsAppmeshVirtualServiceRead(d, meta)
}

func resourceAwsAppmeshVirtualServiceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appmeshconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	resp, err := conn.DescribeVirtualService(&appmesh.DescribeVirtualServiceInput{
		MeshName:           aws.String(d.Get("mesh_name").(string)),
		VirtualServiceName: aws.String(d.Get("name").(string)),
	})
	if isAWSErr(err, appmesh.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] App Mesh virtual service (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading App Mesh virtual service: %s", err)
	}
	if aws.StringValue(resp.VirtualService.Status.Status) == appmesh.VirtualServiceStatusCodeDeleted {
		log.Printf("[WARN] App Mesh virtual service (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	arn := aws.StringValue(resp.VirtualService.Metadata.Arn)
	d.Set("name", resp.VirtualService.VirtualServiceName)
	d.Set("mesh_name", resp.VirtualService.MeshName)
	d.Set("arn", arn)
	d.Set("created_date", resp.VirtualService.Metadata.CreatedAt.Format(time.RFC3339))
	d.Set("last_updated_date", resp.VirtualService.Metadata.LastUpdatedAt.Format(time.RFC3339))
	err = d.Set("spec", flattenAppmeshVirtualServiceSpec(resp.VirtualService.Spec))
	if err != nil {
		return fmt.Errorf("error setting spec: %s", err)
	}

	tags, err := keyvaluetags.AppmeshListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for App Mesh virtual service (%s): %s", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsAppmeshVirtualServiceUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appmeshconn

	if d.HasChange("spec") {
		_, v := d.GetChange("spec")
		req := &appmesh.UpdateVirtualServiceInput{
			MeshName:           aws.String(d.Get("mesh_name").(string)),
			VirtualServiceName: aws.String(d.Get("name").(string)),
			Spec:               expandAppmeshVirtualServiceSpec(v.([]interface{})),
		}

		log.Printf("[DEBUG] Updating App Mesh virtual service: %#v", req)
		_, err := conn.UpdateVirtualService(req)
		if err != nil {
			return fmt.Errorf("error updating App Mesh virtual service: %s", err)
		}
	}

	arn := d.Get("arn").(string)
	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.AppmeshUpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating App Mesh virtual service (%s) tags: %s", arn, err)
		}
	}

	return resourceAwsAppmeshVirtualServiceRead(d, meta)
}

func resourceAwsAppmeshVirtualServiceDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appmeshconn

	log.Printf("[DEBUG] Deleting App Mesh virtual service: %s", d.Id())
	_, err := conn.DeleteVirtualService(&appmesh.DeleteVirtualServiceInput{
		MeshName:           aws.String(d.Get("mesh_name").(string)),
		VirtualServiceName: aws.String(d.Get("name").(string)),
	})
	if isAWSErr(err, appmesh.ErrCodeNotFoundException, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting App Mesh virtual service: %s", err)
	}

	return nil
}

func resourceAwsAppmeshVirtualServiceImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("Wrong format of resource: %s. Please follow 'mesh-name/virtual-service-name'", d.Id())
	}

	mesh := parts[0]
	name := parts[1]
	log.Printf("[DEBUG] Importing App Mesh virtual service %s from mesh %s", name, mesh)

	conn := meta.(*AWSClient).appmeshconn

	resp, err := conn.DescribeVirtualService(&appmesh.DescribeVirtualServiceInput{
		MeshName:           aws.String(mesh),
		VirtualServiceName: aws.String(name),
	})
	if err != nil {
		return nil, err
	}

	d.SetId(aws.StringValue(resp.VirtualService.Metadata.Uid))
	d.Set("name", resp.VirtualService.VirtualServiceName)
	d.Set("mesh_name", resp.VirtualService.MeshName)

	return []*schema.ResourceData{d}, nil
}
