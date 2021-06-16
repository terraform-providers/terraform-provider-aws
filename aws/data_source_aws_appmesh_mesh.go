package aws

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsAppmeshMesh() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsAppmeshMeshRead,

		Schema: map[string]*schema.Schema{
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

			"mesh_owner": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"resource_owner": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"spec": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"egress_filter": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},

			"tags": tagsSchemaComputed(),
		},
	}
}

func dataSourceAwsAppmeshMeshRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appmeshconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	meshName := d.Get("name").(string)

	req := &appmesh.DescribeMeshInput{
		MeshName: aws.String(meshName),
	}
	if v, ok := d.GetOk("mesh_owner"); ok {
		req.MeshOwner = aws.String(v.(string))
	}

	resp, err := conn.DescribeMesh(req)
	if err != nil {
		return fmt.Errorf("error reading App Mesh service mesh: %s", err)
	}
	if aws.StringValue(resp.Mesh.Status.Status) == appmesh.MeshStatusCodeDeleted {
		return fmt.Errorf("App Mesh Mesh (%s) has status: 'DELETED'", meshName)
	}

	arn := aws.StringValue(resp.Mesh.Metadata.Arn)

	d.SetId(aws.StringValue(resp.Mesh.MeshName))
	d.Set("arn", arn)
	d.Set("created_date", resp.Mesh.Metadata.CreatedAt.Format(time.RFC3339))
	d.Set("last_updated_date", resp.Mesh.Metadata.LastUpdatedAt.Format(time.RFC3339))
	d.Set("mesh_owner", resp.Mesh.Metadata.MeshOwner)
	d.Set("resource_owner", resp.Mesh.Metadata.ResourceOwner)
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
