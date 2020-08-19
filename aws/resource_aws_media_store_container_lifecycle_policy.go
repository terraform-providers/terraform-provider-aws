package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mediastore"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsMediaStoreContainerLifecyclePolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsMediaStoreContainerLifecyclePolicyPut,
		Read:   resourceAwsMediaStoreContainerLifecyclePolicyRead,
		Update: resourceAwsMediaStoreContainerLifecyclePolicyPut,
		Delete: resourceAwsMediaStoreContainerLifecyclePolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"container_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"policy": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceAwsMediaStoreContainerLifecyclePolicyPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).mediastoreconn

	containerName := d.Get("container_name").(string)
	input := &mediastore.PutLifecyclePolicyInput{
		ContainerName:   aws.String(containerName),
		LifecyclePolicy: aws.String(d.Get("policy").(string)),
	}

	_, err := conn.PutLifecyclePolicy(input)
	if !d.IsNewResource() && isAWSErr(err, mediastore.ErrCodeContainerNotFoundException, "") {
		log.Printf("[WARN] Media Store Container (%s) not found, removing from state", containerName)
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error putting media store container (%s) lifecycle policy: %s", containerName, err)
	}

	d.SetId(containerName)

	return resourceAwsMediaStoreContainerLifecyclePolicyRead(d, meta)
}

func resourceAwsMediaStoreContainerLifecyclePolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).mediastoreconn

	input := &mediastore.GetLifecyclePolicyInput{
		ContainerName: aws.String(d.Id()),
	}

	resp, err := conn.GetLifecyclePolicy(input)
	if isAWSErr(err, mediastore.ErrCodeContainerNotFoundException, "") {
		log.Printf("[WARN] Media Store Container (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if isAWSErr(err, mediastore.ErrCodePolicyNotFoundException, "") {
		log.Printf("[WARN] Lifecycle Policy for Media Store Container (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error describing media store container (%s) lifecycle policy: %s", d.Id(), err)
	}

	d.Set("container_name", d.Id())
	d.Set("policy", aws.StringValue(resp.LifecyclePolicy))

	return nil
}

func resourceAwsMediaStoreContainerLifecyclePolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).mediastoreconn

	input := &mediastore.DeleteLifecyclePolicyInput{
		ContainerName: aws.String(d.Id()),
	}

	_, err := conn.DeleteLifecyclePolicy(input)
	if isAWSErr(err, mediastore.ErrCodeContainerNotFoundException, "") {
		return nil
	}
	if isAWSErr(err, mediastore.ErrCodePolicyNotFoundException, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error deleting media store container (%s) lifecycle policy: %s", d.Id(), err)
	}

	return nil
}
