package aws

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
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

			"unique_id": {
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

			"roles": {
				Type:          schema.TypeSet,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"role"},
				Elem:          &schema.Schema{Type: schema.TypeString},
				Set:           schema.HashString,
				Deprecated:    "Use `role` instead. Only a single role can be passed to an IAM Instance Profile",
			},

			"role": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"roles"},
			},
		},
	}
}

func resourceAwsIamInstanceProfileCreate(d *schema.ResourceData, meta interface{}) error {
	iamconn := meta.(*AWSClient).iamconn

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
	response, err := iamconn.CreateInstanceProfile(request)
	if err == nil {
		err = instanceProfileReadResult(d, response.InstanceProfile)
	}
	if err != nil {
		return fmt.Errorf("Error creating IAM instance profile %s: %s", name, err)
	}

	waiterRequest := &iam.GetInstanceProfileInput{
		InstanceProfileName: aws.String(name),
	}
	// don't return until the IAM service reports that the instance profile is ready.
	// this ensures that terraform resources which rely on the instance profile will 'see'
	// that the instance profile exists.
	err = iamconn.WaitUntilInstanceProfileExists(waiterRequest)
	if err != nil {
		return fmt.Errorf("Timed out while waiting for instance profile %s: %s", name, err)
	}

	return resourceAwsIamInstanceProfileUpdate(d, meta)
}

func instanceProfileAddRole(iamconn *iam.IAM, profileName, roleName string) error {
	request := &iam.AddRoleToInstanceProfileInput{
		InstanceProfileName: aws.String(profileName),
		RoleName:            aws.String(roleName),
	}

	err := resource.Retry(30*time.Second, func() *resource.RetryError {
		var err error
		_, err = iamconn.AddRoleToInstanceProfile(request)
		// IAM unfortunately does not provide a better error code or message for eventual consistency
		// InvalidParameterValue: Value (XXX) for parameter iamInstanceProfile.name is invalid. Invalid IAM Instance Profile name
		// NoSuchEntity: The request was rejected because it referenced an entity that does not exist. The error message describes the entity. HTTP Status Code: 404
		if isAWSErr(err, "InvalidParameterValue", "Invalid IAM Instance Profile name") || isAWSErr(err, "NoSuchEntity", "The role with name") {
			return resource.RetryableError(err)
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		_, err = iamconn.AddRoleToInstanceProfile(request)
	}
	if err != nil {
		return fmt.Errorf("Error adding IAM Role %s to Instance Profile %s: %s", roleName, profileName, err)
	}

	return err
}

func instanceProfileRemoveRole(iamconn *iam.IAM, profileName, roleName string) error {
	request := &iam.RemoveRoleFromInstanceProfileInput{
		InstanceProfileName: aws.String(profileName),
		RoleName:            aws.String(roleName),
	}

	_, err := iamconn.RemoveRoleFromInstanceProfile(request)
	if isAWSErr(err, "NoSuchEntity", "") {
		return nil
	}
	return err
}

func instanceProfileSetRoles(d *schema.ResourceData, iamconn *iam.IAM) error {
	oldInterface, newInterface := d.GetChange("roles")
	oldRoles := oldInterface.(*schema.Set)
	newRoles := newInterface.(*schema.Set)

	currentRoles := schema.CopySet(oldRoles)

	for _, role := range oldRoles.Difference(newRoles).List() {
		err := instanceProfileRemoveRole(iamconn, d.Id(), role.(string))
		if err != nil {
			return fmt.Errorf("Error removing role %s from IAM instance profile %s: %s", role, d.Id(), err)
		}
		currentRoles.Remove(role)
		d.Set("roles", currentRoles)
	}

	for _, role := range newRoles.Difference(oldRoles).List() {
		err := instanceProfileAddRole(iamconn, d.Id(), role.(string))
		if err != nil {
			return fmt.Errorf("Error adding role %s to IAM instance profile %s: %s", role, d.Id(), err)
		}
		currentRoles.Add(role)
		d.Set("roles", currentRoles)
	}

	return nil
}

func instanceProfileRemoveAllRoles(d *schema.ResourceData, iamconn *iam.IAM) error {
	role, hasRole := d.GetOk("role")
	roles, hasRoles := d.GetOk("roles")
	if hasRole && !hasRoles { // "roles" will always be a superset of "role", if set
		err := instanceProfileRemoveRole(iamconn, d.Id(), role.(string))
		if err != nil {
			return fmt.Errorf("Error removing role %s from IAM instance profile %s: %s", role, d.Id(), err)
		}
	} else {
		for _, role := range roles.(*schema.Set).List() {
			err := instanceProfileRemoveRole(iamconn, d.Id(), role.(string))
			if err != nil {
				return fmt.Errorf("Error removing role %s from IAM instance profile %s: %s", role, d.Id(), err)
			}
		}
	}
	return nil
}

func resourceAwsIamInstanceProfileUpdate(d *schema.ResourceData, meta interface{}) error {
	iamconn := meta.(*AWSClient).iamconn

	if d.HasChange("role") {
		oldRole, newRole := d.GetChange("role")

		if oldRole.(string) != "" {
			err := instanceProfileRemoveRole(iamconn, d.Id(), oldRole.(string))
			if err != nil {
				return fmt.Errorf("Error adding role %s to IAM instance profile %s: %s", oldRole.(string), d.Id(), err)
			}
		}

		if newRole.(string) != "" {
			err := instanceProfileAddRole(iamconn, d.Id(), newRole.(string))
			if err != nil {
				return fmt.Errorf("Error adding role %s to IAM instance profile %s: %s", newRole.(string), d.Id(), err)
			}
		}
	}

	if d.HasChange("roles") {
		return instanceProfileSetRoles(d, iamconn)
	}

	return nil
}

func resourceAwsIamInstanceProfileRead(d *schema.ResourceData, meta interface{}) error {
	iamconn := meta.(*AWSClient).iamconn

	request := &iam.GetInstanceProfileInput{
		InstanceProfileName: aws.String(d.Id()),
	}

	result, err := iamconn.GetInstanceProfile(request)
	if isAWSErr(err, "NoSuchEntity", "") {
		log.Printf("[WARN] IAM Instance Profile %s is already gone", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error reading IAM instance profile %s: %s", d.Id(), err)
	}

	return instanceProfileReadResult(d, result.InstanceProfile)
}

func resourceAwsIamInstanceProfileDelete(d *schema.ResourceData, meta interface{}) error {
	iamconn := meta.(*AWSClient).iamconn

	if err := instanceProfileRemoveAllRoles(d, iamconn); err != nil {
		return err
	}

	request := &iam.DeleteInstanceProfileInput{
		InstanceProfileName: aws.String(d.Id()),
	}
	_, err := iamconn.DeleteInstanceProfile(request)
	if err != nil {
		return fmt.Errorf("Error deleting IAM instance profile %s: %s", d.Id(), err)
	}

	return nil
}

func instanceProfileReadResult(d *schema.ResourceData, result *iam.InstanceProfile) error {
	d.SetId(*result.InstanceProfileName)
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

	roles := &schema.Set{F: schema.HashString}
	for _, role := range result.Roles {
		roles.Add(*role.RoleName)
	}
	err := d.Set("roles", roles)
	return err
}
