package aws

import (
	"fmt"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsIotPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsIotPolicyCreate,
		Read:   resourceAwsIotPolicyRead,
		Update: resourceAwsIotPolicyUpdate,
		Delete: resourceAwsIotPolicyDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"policy": {
				Type:     schema.TypeString,
				Required: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_version_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsIotPolicyCreate(d *schema.ResourceData, meta interface{}) error {

	conn := meta.(*AWSClient).iotconn

	out, err := conn.CreatePolicy(&iot.CreatePolicyInput{
		PolicyName:     aws.String(d.Get("name").(string)),
		PolicyDocument: aws.String(d.Get("policy").(string)),
	})

	if err != nil {
		log.Printf("[ERROR] %s", err)
		return err
	}

	d.SetId(*out.PolicyName)

	return resourceAwsIotPolicyRead(d, meta)
}

func resourceAwsIotPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotconn

	out, err := conn.GetPolicy(&iot.GetPolicyInput{
		PolicyName: aws.String(d.Id()),
	})

	if err != nil {
		log.Printf("[ERROR] %s", err)
		return err
	}

	d.Set("arn", out.PolicyArn)
	d.Set("default_version_id", out.DefaultVersionId)

	return nil
}

func resourceAwsIotPolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotconn

	if err := iotPolicyPruneVersions(d.Id(), conn); err != nil {
		return err
	}

	if d.HasChange("policy") {
		_, err := conn.CreatePolicyVersion(&iot.CreatePolicyVersionInput{
			PolicyName:     aws.String(d.Id()),
			PolicyDocument: aws.String(d.Get("policy").(string)),
			SetAsDefault:   aws.Bool(true),
		})

		if err != nil {
			log.Printf("[ERROR] %s", err)
			return err
		}
	}

	return resourceAwsIotPolicyRead(d, meta)
}

func resourceAwsIotPolicyDelete(d *schema.ResourceData, meta interface{}) error {

	conn := meta.(*AWSClient).iotconn

	out, err := conn.ListPolicyVersions(&iot.ListPolicyVersionsInput{
		PolicyName: aws.String(d.Id()),
	})

	if err != nil {
		return err
	}

	// Delete all non-default versions of the policy
	for _, ver := range out.PolicyVersions {
		if !*ver.IsDefaultVersion {
			_, err = conn.DeletePolicyVersion(&iot.DeletePolicyVersionInput{
				PolicyName:      aws.String(d.Id()),
				PolicyVersionId: ver.VersionId,
			})
			if err != nil {
				log.Printf("[ERROR] %s", err)
				return err
			}
		}
	}

	//Delete default policy version
	_, err = conn.DeletePolicy(&iot.DeletePolicyInput{
		PolicyName: aws.String(d.Id()),
	})

	if err != nil {
		log.Printf("[ERROR] %s", err)
		return err
	}

	return nil
}

// iotPolicyPruneVersions deletes the oldest non-default version if the maximum
// number of versions (5) has been reached.
func iotPolicyPruneVersions(name string, iotconn *iot.IoT) error {
	versions, err := iotPolicyListVersions(name, iotconn)
	if err != nil {
		return err
	}
	if len(versions) < 5 {
		return nil
	}

	var oldestVersion *iot.PolicyVersion
	var oldestVersionId int64

	for _, version := range versions {
		if *version.IsDefaultVersion {
			continue
		}

		versionId, err := strconv.ParseInt(*version.VersionId, 10, 64)
		if err != nil {
			log.Printf("[ERROR] Unexpected version id cannot be parsed to an integer. %s", err)
			return err
		}

		if oldestVersion == nil || versionId < oldestVersionId {
			oldestVersion = version
			oldestVersionId = versionId
		}
	}

	err = iotPolicyDeleteVersion(name, *oldestVersion.VersionId, iotconn)
	return err
}

func iotPolicyListVersions(name string, iotconn *iot.IoT) ([]*iot.PolicyVersion, error) {
	request := &iot.ListPolicyVersionsInput{
		PolicyName: aws.String(name),
	}

	response, err := iotconn.ListPolicyVersions(request)
	if err != nil {
		return nil, fmt.Errorf("Error listing versions for IoT policy %s: %s", name, err)
	}
	return response.PolicyVersions, nil
}

func iotPolicyDeleteVersion(name, versionID string, iotconn *iot.IoT) error {
	request := &iot.DeletePolicyVersionInput{
		PolicyName:      aws.String(name),
		PolicyVersionId: aws.String(versionID),
	}

	_, err := iotconn.DeletePolicyVersion(request)
	if err != nil {
		return fmt.Errorf("Error deleting version %s from IoT policy %s: %s", versionID, name, err)
	}
	return nil
}
