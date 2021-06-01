package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/s3outposts"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/s3outposts/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/s3outposts/waiter"
)

func resourceAwsS3OutpostsEndpoint() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsS3OutpostsEndpointCreate,
		Read:   resourceAwsS3OutpostsEndpointRead,
		Delete: resourceAwsS3OutpostsEndpointDelete,

		Importer: &schema.ResourceImporter{
			State: resourceAwsS3OutpostsEndpointImportState,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cidr_block": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"creation_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"network_interfaces": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"network_interface_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"outpost_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"security_group_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"subnet_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
		},
	}
}

func resourceAwsS3OutpostsEndpointCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).s3outpostsconn

	input := &s3outposts.CreateEndpointInput{
		OutpostId:       aws.String(d.Get("outpost_id").(string)),
		SecurityGroupId: aws.String(d.Get("security_group_id").(string)),
		SubnetId:        aws.String(d.Get("subnet_id").(string)),
	}

	output, err := conn.CreateEndpoint(input)

	if err != nil {
		return fmt.Errorf("error creating S3 Outposts Endpoint: %w", err)
	}

	if output == nil {
		return fmt.Errorf("error creating S3 Outposts Endpoint: empty response")
	}

	d.SetId(aws.StringValue(output.EndpointArn))

	if _, err := waiter.EndpointStatusCreated(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for S3 Outposts Endpoint (%s) to become available: %w", d.Id(), err)
	}

	return resourceAwsS3OutpostsEndpointRead(d, meta)
}

func resourceAwsS3OutpostsEndpointRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).s3outpostsconn

	endpoint, err := finder.Endpoint(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error reading S3 Outposts Endpoint (%s): %w", d.Id(), err)
	}

	if endpoint == nil {
		if d.IsNewResource() {
			return fmt.Errorf("error reading S3 Outposts Endpoint (%s): not found after creation", d.Id())
		}

		log.Printf("[WARN] S3 Outposts Endpoint (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("arn", endpoint.EndpointArn)
	d.Set("cidr_block", endpoint.CidrBlock)

	if endpoint.CreationTime != nil {
		d.Set("creation_time", aws.TimeValue(endpoint.CreationTime).Format(time.RFC3339))
	}

	if err := d.Set("network_interfaces", flattenS3outpostsNetworkInterfaces(endpoint.NetworkInterfaces)); err != nil {
		return fmt.Errorf("error setting network_interfaces: %w", err)
	}

	d.Set("outpost_id", endpoint.OutpostsId)

	return nil
}

func resourceAwsS3OutpostsEndpointDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).s3outpostsconn

	parsedArn, err := arn.Parse(d.Id())

	if err != nil {
		return fmt.Errorf("error parsing S3 Outposts Endpoint ARN (%s): %w", d.Id(), err)
	}

	// ARN resource format: outpost/<outpost-id>/endpoint/<endpoint-id>
	arnResourceParts := strings.Split(parsedArn.Resource, "/")

	if parsedArn.AccountID == "" || len(arnResourceParts) != 4 {
		return fmt.Errorf("error parsing S3 Outposts Endpoint ARN (%s): unknown format", d.Id())
	}

	input := &s3outposts.DeleteEndpointInput{
		EndpointId: aws.String(arnResourceParts[3]),
		OutpostId:  aws.String(arnResourceParts[1]),
	}

	_, err = conn.DeleteEndpoint(input)

	if err != nil {
		return fmt.Errorf("error deleting S3 Outposts Endpoint (%s): %w", d.Id(), err)
	}

	return nil
}

func resourceAwsS3OutpostsEndpointImportState(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.Split(d.Id(), ",")

	if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
		return nil, fmt.Errorf("unexpected format of ID (%s), expected ENDPOINT-ARN,SECURITY-GROUP-ID,SUBNET-ID", d.Id())
	}

	endpointArn := idParts[0]
	securityGroupId := idParts[1]
	subnetId := idParts[2]

	d.SetId(endpointArn)
	d.Set("security_group_id", securityGroupId)
	d.Set("subnet_id", subnetId)

	return []*schema.ResourceData{d}, nil
}

func flattenS3outpostsNetworkInterfaces(apiObjects []*s3outposts.NetworkInterface) []interface{} {
	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenS3outpostsNetworkInterface(apiObject))
	}

	return tfList
}

func flattenS3outpostsNetworkInterface(apiObject *s3outposts.NetworkInterface) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.NetworkInterfaceId; v != nil {
		tfMap["network_interface_id"] = aws.StringValue(v)
	}

	return tfMap
}
