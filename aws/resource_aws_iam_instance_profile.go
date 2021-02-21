package aws

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsIamInstanceProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsIamInstanceProfileCreate,
		Read:   resourceAwsIamInstanceProfileRead,
		Update: resourceAwsIamInstanceProfileUpdate,
		Delete: resourceAwsIamInstanceProfileDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"create_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 128),
					validation.StringMatch(regexp.MustCompile(`^[\w+=,.@-]*$`), "must match [\\w+=,.@-]"),
				),
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexp.MustCompile(`^[\w+=,.@-]*$`), "must match [\\w+=,.@-]"),
				),
			},
			"path": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "/",
				ForceNew: true,
			},
			"role": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"unique_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsIamInstanceProfileCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iamconn

	var name string
	if v, ok := d.GetOk("name"); ok {
		name = v.(string)
	} else if v, ok := d.GetOk("name_prefix"); ok {
		name = resource.PrefixedUniqueId(v.(string))
	} else {
		name = resource.UniqueId()
	}

	request := &iam.CreateInstanceProfileInput{
		InstanceProfileName: aws.String(name),
		Path:                aws.String(d.Get("path").(string)),
	}

	var err error
	response, err := conn.CreateInstanceProfile(request)
	if err == nil {
		err = instanceProfileReadResult(d, response.InstanceProfile)
	}
	if err != nil {
		return fmt.Errorf("creating IAM instance profile %s: %w", name, err)
	}

	waiterRequest := &iam.GetInstanceProfileInput{
		InstanceProfileName: aws.String(name),
	}
	// don't return until the IAM service reports that the instance profile is ready.
	// this ensures that terraform resources which rely on the instance profile will 'see'
	// that the instance profile exists.
	err = conn.WaitUntilInstanceProfileExists(waiterRequest)
	if err != nil {
		return fmt.Errorf("timed out while waiting for instance profile %s: %w", name, err)
	}

	return resourceAwsIamInstanceProfileUpdate(d, meta)
}

func instanceProfileAddRole(conn *iam.IAM, profileName, roleName string) error {
	request := &iam.AddRoleToInstanceProfileInput{
		InstanceProfileName: aws.String(profileName),
		RoleName:            aws.String(roleName),
	}

	err := resource.Retry(30*time.Second, func() *resource.RetryError {
		var err error
		_, err = conn.AddRoleToInstanceProfile(request)
		// IAM unfortunately does not provide a better error code or message for eventual consistency
		// InvalidParameterValue: Value (XXX) for parameter iamInstanceProfile.name is invalid. Invalid IAM Instance Profile name
		// NoSuchEntity: The request was rejected because it referenced an entity that does not exist. The error message describes the entity. HTTP Status Code: 404
		if isAWSErr(err, "InvalidParameterValue", "Invalid IAM Instance Profile name") || isAWSErr(err, iam.ErrCodeNoSuchEntityException, "The role with name") {
			return resource.RetryableError(err)
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		_, err = conn.AddRoleToInstanceProfile(request)
	}
	if err != nil {
		return fmt.Errorf("adding IAM Role %s to Instance Profile %s: %w", roleName, profileName, err)
	}

	return err
}

func instanceProfileRemoveRole(conn *iam.IAM, profileName, roleName string) error {
	request := &iam.RemoveRoleFromInstanceProfileInput{
		InstanceProfileName: aws.String(profileName),
		RoleName:            aws.String(roleName),
	}

	_, err := conn.RemoveRoleFromInstanceProfile(request)
	if isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
		return nil
	}
	return err
}

func instanceProfileRemoveAllRoles(d *schema.ResourceData, conn *iam.IAM) error {
	if role, ok := d.GetOk("role"); ok {
		err := instanceProfileRemoveRole(conn, d.Id(), role.(string))
		if err != nil {
			return fmt.Errorf("removing role %s from IAM instance profile %s: %w", role, d.Id(), err)
		}
	}

	return nil
}

func resourceAwsIamInstanceProfileUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iamconn

	if d.HasChange("role") {
		oldRole, newRole := d.GetChange("role")

		if oldRole.(string) != "" {
			err := instanceProfileRemoveRole(conn, d.Id(), oldRole.(string))
			if err != nil {
				return fmt.Errorf("removing role %s to IAM instance profile %s: %w", oldRole.(string), d.Id(), err)
			}
		}

		if newRole.(string) != "" {
			err := instanceProfileAddRole(conn, d.Id(), newRole.(string))
			if err != nil {
				return fmt.Errorf("adding role %s to IAM instance profile %s: %w", newRole.(string), d.Id(), err)
			}
		}
	}

	return nil
}

func resourceAwsIamInstanceProfileRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iamconn

	request := &iam.GetInstanceProfileInput{
		InstanceProfileName: aws.String(d.Id()),
	}

	result, err := conn.GetInstanceProfile(request)
	if isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
		log.Printf("[WARN] IAM Instance Profile %s is already gone", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("reading IAM instance profile %s: %w", d.Id(), err)
	}

	instanceProfile := result.InstanceProfile
	if instanceProfile.Roles != nil && len(instanceProfile.Roles) > 0 {
		roleName := aws.StringValue(instanceProfile.Roles[0].RoleName)
		input := &iam.GetRoleInput{
			RoleName: aws.String(roleName),
		}

		_, err := conn.GetRole(input)
		if err != nil {
			if isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
				err := instanceProfileRemoveRole(conn, d.Id(), roleName)
				if err != nil {
					return fmt.Errorf("removing role %s to IAM instance profile %s: %w", roleName, d.Id(), err)
				}
			}
			return fmt.Errorf("reading IAM Role %s attcahed to IAM Instance Profile %s: %w", roleName, d.Id(), err)
		}
	}

	return instanceProfileReadResult(d, instanceProfile)
}

func resourceAwsIamInstanceProfileDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iamconn

	if err := instanceProfileRemoveAllRoles(d, conn); err != nil {
		return err
	}

	request := &iam.DeleteInstanceProfileInput{
		InstanceProfileName: aws.String(d.Id()),
	}
	_, err := conn.DeleteInstanceProfile(request)
	if err != nil {
		if isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
			return nil
		}
		return fmt.Errorf("deleting IAM instance profile %s: %w", d.Id(), err)
	}

	return nil
}

func instanceProfileReadResult(d *schema.ResourceData, result *iam.InstanceProfile) error {
	d.SetId(aws.StringValue(result.InstanceProfileName))
	if err := d.Set("name", result.InstanceProfileName); err != nil {
		return err
	}
	if err := d.Set("arn", result.Arn); err != nil {
		return err
	}
	if err := d.Set("create_date", result.CreateDate.Format(time.RFC3339)); err != nil {
		return err
	}
	if err := d.Set("path", result.Path); err != nil {
		return err
	}
	d.Set("unique_id", result.InstanceProfileId)

	if result.Roles != nil && len(result.Roles) > 0 {
		d.Set("role", result.Roles[0].RoleName) //there will only be 1 role returned
	}

	return nil
}
