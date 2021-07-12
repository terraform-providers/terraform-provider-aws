package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsElasticacheUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsElasticacheUserCreate,
		Read:   resourceAwsElasticacheUserRead,
		Update: resourceAwsElasticacheUserUpdate,
		Delete: resourceAwsElasticacheUserDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"access_string": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringMatch(regexp.MustCompile(`^[o]`), "must begin with an o"), //on or off
					validation.StringDoesNotMatch(regexp.MustCompile(`--`), "cannot contain two consecutive hyphens"),
					validation.StringDoesNotMatch(regexp.MustCompile(`-$`), "cannot end with a hyphen"),
				),
			},
			"engine": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "redis",
				ValidateFunc: validateAwsElastiCacheUserEngine,
			},
			"no_password_required": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"passwords": {
				Type:     schema.TypeSet,
				Optional: true,
				MinItems: 1,
				MaxItems: 2,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"user_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 40),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9._-]+$`), "must contain only alphanumeric characters, periods, underscores, and hyphens"),
					validation.StringDoesNotMatch(regexp.MustCompile(`--`), "cannot contain two consecutive hyphens"),
					validation.StringDoesNotMatch(regexp.MustCompile(`-$`), "cannot end with a hyphen"),
				),
			},
			"user_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 120),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9._-]+$`), "must contain only alphanumeric characters, periods, underscores, and hyphens"),
					validation.StringDoesNotMatch(regexp.MustCompile(`--`), "cannot contain two consecutive hyphens"),
					validation.StringDoesNotMatch(regexp.MustCompile(`-$`), "cannot end with a hyphen"),
				),
			},
		},
	}
}

func resourceAwsElasticacheUserCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elasticacheconn

	userId := d.Get("user_id").(string)

	// Get ElastiCache User Properties
	params := &elasticache.CreateUserInput{
		AccessString:       aws.String(d.Get("access_string").(string)),
		Engine:             aws.String(d.Get("engine").(string)),
		NoPasswordRequired: aws.Bool(d.Get("no_password_required").(bool)),
		UserId:             aws.String(userId),
		UserName:           aws.String(d.Get("user_name").(string)),
	}

	if passwords := d.Get("passwords").(*schema.Set); passwords.Len() > 0 {
		params.Passwords = expandStringSet(passwords)
	}

	_, err := conn.CreateUser(params)
	if err != nil {
		return fmt.Errorf("[ERROR] Error creating ElastiCache User: %s", err)
	}

	d.SetId(userId)

	return resourceAwsElasticacheUserRead(d, meta)
}

func resourceAwsElasticacheUserRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elasticacheconn

	params := &elasticache.DescribeUsersInput{
		UserId: aws.String(d.Get("user_id").(string)),
	}

	response, err := conn.DescribeUsers(params)

	if isAWSErr(err, elasticache.ErrCodeUserNotFoundFault, "") {
		log.Printf("[WARN] ElastiCache User (%s) not found", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("[ERROR] ElastiCache User Reading: %s", err)
	}

	if response == nil || len(response.Users) == 0 || response.Users[0] == nil {
		log.Printf("[WARN] ElastiCache User (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	user := response.Users[0]

	d.Set("access_string", user.AccessString)
	d.Set("engine", user.Engine)
	d.Set("user_id", user.UserId)
	d.Set("user_name", user.UserName)

	return nil
}

func resourceAwsElasticacheUserUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elasticacheconn

	params := &elasticache.ModifyUserInput{
		UserId: aws.String(d.Get("user_id").(string)),
	}

	if d.HasChange("access_string") {
		params.AccessString = aws.String(d.Get("access_string").(string))

		_, err := conn.ModifyUser(params)
		if err != nil {
			return err
		}

		stateConf := &resource.StateChangeConf{
			Pending:    []string{"modifying"},
			Target:     []string{"active"},
			Refresh:    resourceAwsElasticacheUserStateRefreshFunc(conn, d.Id()),
			Timeout:    d.Timeout(schema.TimeoutDelete),
			MinTimeout: 10 * time.Second,
			Delay:      10 * time.Second,
		}

		log.Printf("[DEBUG] Waiting for Elasticache User (%s) to become available", d.Id())

		_, err = stateConf.WaitForState()
		if err != nil {
			return fmt.Errorf("[ERROR] Elasticache User (%s) during modification: %w", d.Id(), err)
		}
	}

	if d.HasChanges("passwords") {
		o, n := d.GetChange("passwords")
		if o == nil {
			o = new(schema.Set)
		}
		if n == nil {
			n = new(schema.Set)
		}

		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		removePasswords := os.Difference(ns).List()
		addPasswords := ns.Difference(os).List()

		if len(addPasswords) > 0 {
			passwords := expandStringSet(d.Get("passwords").(*schema.Set))
			noPasswordRequired := false
			params.NoPasswordRequired = &noPasswordRequired
			params.Passwords = passwords
			_, err := conn.ModifyUser(params)
			if err != nil {
				return err
			}

			stateConf := &resource.StateChangeConf{
				Pending:    []string{"modifying"},
				Target:     []string{"active"},
				Refresh:    resourceAwsElasticacheUserStateRefreshFunc(conn, d.Id()),
				Timeout:    d.Timeout(schema.TimeoutDelete),
				MinTimeout: 10 * time.Second,
				Delay:      10 * time.Second,
			}

			log.Printf("[DEBUG] Waiting for Elasticache User (%s) to become available", d.Id())

			_, err = stateConf.WaitForState()
			if err != nil {
				return fmt.Errorf("[ERROR] Elasticache User (%s) during modification: %w", d.Id(), err)
			}
		} else if len(removePasswords) > 0 {
			noPasswordRequired := true
			req := &elasticache.ModifyUserInput{
				UserId:             aws.String(d.Id()),
				NoPasswordRequired: &noPasswordRequired,
			}
			_, err := conn.ModifyUser(req)
			if err != nil {
				return err
			}

			stateConf := &resource.StateChangeConf{
				Pending:    []string{"modifying"},
				Target:     []string{"active"},
				Refresh:    resourceAwsElasticacheUserStateRefreshFunc(conn, d.Id()),
				Timeout:    d.Timeout(schema.TimeoutDelete),
				MinTimeout: 10 * time.Second,
				Delay:      10 * time.Second,
			}

			log.Printf("[DEBUG] Waiting for Elasticache User (%s) to become available", d.Id())

			_, err = stateConf.WaitForState()
			if err != nil {
				return fmt.Errorf("[ERROR] Elasticache User (%s) during modification: %w", d.Id(), err)
			}
		}
	}
	return resourceAwsElasticacheUserRead(d, meta)
}

func resourceAwsElasticacheUserDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elasticacheconn

	params := &elasticache.DeleteUserInput{
		UserId: aws.String(d.Get("user_id").(string)),
	}

	log.Printf("[DEBUG] ElastiCache User Delete: %#v", params)

	_, err := conn.DeleteUser(params)

	if isAWSErr(err, elasticache.ErrCodeUserNotFoundFault, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("[ERROR] Deleting ElastiCache User (%s): %s", d.Id(), err)
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"deleting"},
		Target:     []string{},
		Refresh:    resourceAwsElasticacheUserStateRefreshFunc(conn, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		MinTimeout: 10 * time.Second,
		Delay:      10 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("[ERROR] ElastiCache User (%s) during Deletion: %w", d.Id(), err)
	}

	return nil
}

func resourceAwsElasticacheUserStateRefreshFunc(conn *elasticache.ElastiCache, userID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		params := &elasticache.DescribeUsersInput{
			UserId: aws.String(userID),
		}

		response, err := conn.DescribeUsers(params)

		if isAWSErr(err, elasticache.ErrCodeUserNotFoundFault, "") {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if response == nil || len(response.Users) == 0 || response.Users[0] == nil {
			return nil, "", nil
		}

		return response, aws.StringValue(response.Users[0].Status), nil
	}
}

func validateAwsElastiCacheUserEngine(v interface{}, k string) (ws []string, errors []error) {
	if strings.ToLower(v.(string)) != "redis" {
		errors = append(errors, fmt.Errorf("The only acceptable Engine type for ElastiCache Users is Redis"))
	}
	return
}
