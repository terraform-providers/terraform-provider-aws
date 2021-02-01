package aws

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/emrcontainers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/emrcontainers/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/emrcontainers/waiter"
)

func resourceAwsEMRContainersVirtualCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEMRContainersVirtualClusterCreate,
		Read:   resourceAwsEMRContainersVirtualClusterRead,
		Delete: resourceAwsEMRContainersVirtualClusterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"container_provider": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"info": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"eks_info": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										ForceNew: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"namespace": {
													Type:     schema.TypeString,
													Optional: true,
													ForceNew: true,
												},
											},
										},
									},
								},
							},
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(emrcontainers.ContainerProviderType_Values(), false),
						},
					},
				},
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`[.\-_/#A-Za-z0-9]+`), ""),
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsEMRContainersVirtualClusterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).emrcontainersconn

	input := emrcontainers.CreateVirtualClusterInput{
		ContainerProvider: expandEMRContainersContainerProvider(d.Get("container_provider").([]interface{})),
		Name:              aws.String(d.Get("name").(string)),
	}

	log.Printf("[INFO] Creating EMR containers virtual cluster: %s", input)
	out, err := conn.CreateVirtualCluster(&input)
	if err != nil {
		return fmt.Errorf("error creating EMR containers virtual cluster: %w", err)
	}

	d.SetId(aws.StringValue(out.Id))

	if _, err := waiter.VirtualClusterCreated(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for EMR containers virtual cluster (%s) creation: %w", d.Id(), err)
	}

	return resourceAwsEMRContainersVirtualClusterRead(d, meta)
}

func resourceAwsEMRContainersVirtualClusterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).emrcontainersconn

	vc, err := finder.VirtualClusterById(conn, d.Id())

	if err != nil {
		if isAWSErr(err, emrcontainers.ErrCodeResourceNotFoundException, "") && !d.IsNewResource() {
			log.Printf("[WARN] EMR containers virtual cluster (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("error reading EMR containers virtual cluster (%s): %w", d.Id(), err)
	}

	if vc == nil {
		log.Printf("[WARN] EMR containers virtual cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("arn", vc.Arn)
	if err := d.Set("container_provider", flattenEMRContainersContainerProvider(vc.ContainerProvider)); err != nil {
		return fmt.Errorf("error reading EMR containers virtual cluster (%s): %w", d.Id(), err)
	}
	d.Set("created_at", aws.TimeValue(vc.CreatedAt).String())
	d.Set("name", vc.Name)
	d.Set("state", vc.State)

	return nil
}

func resourceAwsEMRContainersVirtualClusterDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).emrcontainersconn

	log.Printf("[INFO] EMR containers virtual cluster: %s", d.Id())
	_, err := conn.DeleteVirtualCluster(&emrcontainers.DeleteVirtualClusterInput{
		Id: aws.String(d.Id()),
	})
	if err != nil {
		if isAWSErr(err, emrcontainers.ErrCodeResourceNotFoundException, "") {
			return nil
		}

		return fmt.Errorf("error deleting EMR containers virtual cluster (%s): %w", d.Id(), err)
	}

	_, err = waiter.VirtualClusterDeleted(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error waiting for EMR containers virtual cluster (%s) deletion: %w", d.Id(), err)
	}

	return nil
}

func expandEMRContainersContainerProvider(l []interface{}) *emrcontainers.ContainerProvider {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	input := emrcontainers.ContainerProvider{
		Id:   aws.String(m["id"].(string)),
		Type: aws.String(m["type"].(string)),
	}

	if v, ok := m["info"]; ok {
		input.Info = expandEMRContainersContainerInfo(v.([]interface{}))
	}

	return &input
}

func expandEMRContainersContainerInfo(l []interface{}) *emrcontainers.ContainerInfo {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	input := emrcontainers.ContainerInfo{}

	if v, ok := m["eks_info"]; ok {
		input.EksInfo = expandEMRContainersEksInfo(v.([]interface{}))
	}

	return &input
}

func expandEMRContainersEksInfo(l []interface{}) *emrcontainers.EksInfo {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	input := emrcontainers.EksInfo{}

	if v, ok := m["namespace"]; ok {
		input.Namespace = aws.String(v.(string))
	}

	return &input
}

func flattenEMRContainersContainerProvider(cp *emrcontainers.ContainerProvider) []interface{} {
	if cp == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	m["id"] = cp.Id
	m["type"] = cp.Type

	if cp.Info != nil {
		m["info"] = flattenEMRContainersContainerInfo(cp.Info)
	}

	return []interface{}{m}
}

func flattenEMRContainersContainerInfo(ci *emrcontainers.ContainerInfo) []interface{} {
	if ci == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	if ci.EksInfo != nil {
		m["eks_info"] = flattenEMRContainersEksInfo(ci.EksInfo)
	}

	return []interface{}{m}
}

func flattenEMRContainersEksInfo(ei *emrcontainers.EksInfo) []interface{} {
	if ei == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	if ei.Namespace != nil {
		m["namespace"] = ei.Namespace
	}

	return []interface{}{m}
}
