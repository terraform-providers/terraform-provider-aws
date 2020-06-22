package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsAppmeshMesh() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAppmeshMeshCreate,
		Read:   resourceAwsAppmeshMeshRead,
		Update: resourceAwsAppmeshMeshUpdate,
		Delete: resourceAwsAppmeshMeshDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},

			"spec": {
				Type:             schema.TypeList,
				Optional:         true,
				MinItems:         0,
				MaxItems:         1,
				DiffSuppressFunc: suppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"egress_filter": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 0,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  appmesh.EgressFilterTypeDropAll,
										ValidateFunc: validation.StringInSlice([]string{
											appmesh.EgressFilterTypeAllowAll,
											appmesh.EgressFilterTypeDropAll,
										}, false),
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

func resourceAwsAppmeshMeshCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appmeshconn

	meshName := d.Get("name").(string)
	req := &appmesh.CreateMeshInput{
		MeshName: aws.String(meshName),
		Spec:     expandAppmeshMeshSpec(d.Get("spec").([]interface{})),
		Tags:     keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().AppmeshTags(),
	}

	log.Printf("[DEBUG] Creating App Mesh service mesh: %#v", req)
	_, err := conn.CreateMesh(req)
	if err != nil {
		return fmt.Errorf("error creating App Mesh service mesh: %s", err)
	}

	d.SetId(meshName)

	return resourceAwsAppmeshMeshRead(d, meta)
}

func resourceAwsAppmeshMeshRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appmeshconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	resp, err := conn.DescribeMesh(&appmesh.DescribeMeshInput{
		MeshName: aws.String(d.Id()),
	})
	if isAWSErr(err, appmesh.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] App Mesh service mesh (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading App Mesh service mesh: %s", err)
	}
	if aws.StringValue(resp.Mesh.Status.Status) == appmesh.MeshStatusCodeDeleted {
		log.Printf("[WARN] App Mesh service mesh (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	arn := aws.StringValue(resp.Mesh.Metadata.Arn)
	d.Set("name", resp.Mesh.MeshName)
	d.Set("arn", arn)
	d.Set("created_date", resp.Mesh.Metadata.CreatedAt.Format(time.RFC3339))
	d.Set("last_updated_date", resp.Mesh.Metadata.LastUpdatedAt.Format(time.RFC3339))
	err = d.Set("spec", flattenAppmeshMeshSpec(resp.Mesh.Spec))
	if err != nil {
		return fmt.Errorf("error setting spec: %s", err)
	}

	tags, err := keyvaluetags.AppmeshListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for App Mesh service mesh (%s): %s", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsAppmeshMeshUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appmeshconn

	if d.HasChange("spec") {
		_, v := d.GetChange("spec")
		req := &appmesh.UpdateMeshInput{
			MeshName: aws.String(d.Id()),
			Spec:     expandAppmeshMeshSpec(v.([]interface{})),
		}

		log.Printf("[DEBUG] Updating App Mesh service mesh: %#v", req)
		_, err := conn.UpdateMesh(req)
		if err != nil {
			return fmt.Errorf("error updating App Mesh service mesh: %s", err)
		}
	}

	arn := d.Get("arn").(string)
	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.AppmeshUpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating App Mesh service mesh (%s) tags: %s", arn, err)
		}
	}

	return resourceAwsAppmeshMeshRead(d, meta)
}

func resourceAwsAppmeshMeshDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appmeshconn

	log.Printf("[DEBUG] Deleting App Mesh service mesh: %s", d.Id())
	_, err := conn.DeleteMesh(&appmesh.DeleteMeshInput{
		MeshName: aws.String(d.Id()),
	})
	if isAWSErr(err, appmesh.ErrCodeNotFoundException, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting App Mesh service mesh: %s", err)
	}

	return nil
}
