package aws

import (
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/opsworkscm"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

const EngineModel = "Single"
const EngineVersion = "12"

func resourceAwsOpsworksChef() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsOpsworksChefCreate,
		Read:   resourceAwsOpsworksChefRead,
		Update: resourceAwsOpsworksChefUpdate,
		Delete: resourceAwsOpsworksChefDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			// TODO: do we want this? I think this is another "magic" thing that happens
			// I think we can still associate a public IP later if we want? :shrug:
			"associate_public_ip_address": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"backup_retention_count": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
				// https://docs.aws.amazon.com/opsworks/latest/userguide/opscm-chef-backup.html
				// I could swear I found this 1-30 limitation elsewhere, but I can't seem to find it now other than there
				ValidateFunc: validation.IntBetween(1, 30),
			},

			"disable_automated_backup": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			// TODO:
			// I kinda want to make this required because how else are we going to get at that info?
			// Also, since these are write-only and not returned by the API, we'll probably need a custom diff function that ignores them?
			"chef_pivotal_key": {
				Type:     schema.TypeString,
				Required: true,
				// TODO: should we?
				// ValidateFunc: validateRsaKey,
			},

			"chef_delivery_admin_password": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
				// TODO: should we?
				// ValidateFunc: validatePasswordOk,
			},

			// TODO: document the instance role required
			// https://s3.amazonaws.com/opsworks-cm-us-east-1-prod-default-assets/misc/opsworks-cm-roles.yaml
			"instance_profile_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},

			"instance_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"ssh_key_pair": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"preferred_backup_window": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateStartTimeFormat,
			},

			"preferred_maintenance_window": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateStartTimeFormat,
			},

			// TODO:
			// if you don't specify this, it'll create one. I'm of the opinion that users should
			// create it themselves. Or maybe some sort of wrapper module? I dunno.
			// if it's optional, will they be computed? if so, is that something we can express here?
			"security_group_ids": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				MinItems: 1,
			},

			// TODO: document this role creation
			// https://s3.amazonaws.com/opsworks-cm-us-east-1-prod-default-assets/misc/opsworks-cm-roles.yaml
			"service_role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},

			// TODO: huh?
			// The IDs of subnets in which to launch the server EC2 instance.
			//
			// Amazon EC2-Classic customers: This field is required. All servers must run
			// within a VPC. The VPC must have "Auto Assign Public IP" enabled.
			//
			// EC2-VPC customers: This field is optional. If you do not specify subnet IDs,
			// your EC2 instances are created in a default subnet that is selected by Amazon
			// EC2. If you specify subnet IDs, the VPC must have "Auto Assign Public IP"
			// enabled.
			// if it's optional, will they be computed? if so, is that something we can express here?
			"subnet_ids": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				MinItems: 1,
			},

			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsOpsworksChefValidate(d *schema.ResourceData) error {
	// TODO: validation function
	// chefPivotalKey validate rsa? or is that too much?
	// chefDeliveryAdminPassword validate the password meets the API's requirements or is that too much?
	// backupRetentionCount validate that it's positive
	//     but also maybe validate that it's in the valid range? or is that too much?
	//     I also wonder if there are constants in the API connector that maybe we can leverage here
	return nil
}

func resourceAwsOpsworksChefRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AWSClient).opsworkscmconn

	req := &opsworkscm.DescribeServersInput{
		ServerName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading OpsWorks Chef server: %s", d.Id())

	var resp *opsworkscm.DescribeServersOutput
	var dErr error

	resp, dErr = client.DescribeServers(req)
	if dErr != nil {
		log.Printf("[DEBUG] OpsWorks Chef (%s) not found", d.Id())
		d.SetId("")
		return dErr
	}

	server := resp.Servers[0]
	d.Set("associate_public_ip_address", server.AssociatePublicIpAddress)
	d.Set("backup_retention_count", server.BackupRetentionCount)
	d.Set("disable_automated_backup", server.DisableAutomatedBackup)
	d.Set("instance_profile_arn", server.InstanceProfileArn)
	d.Set("instance_type", server.InstanceType)
	d.SetId(*server.ServerName)
	d.Set("name", server.ServerName)
	if server.KeyPair != nil {
		d.Set("ssh_key_pair", server.KeyPair)
	}
	d.Set("preferred_backup_window", server.PreferredBackupWindow)
	d.Set("preferred_maintenance_window", server.PreferredMaintenanceWindow)
	d.Set("security_group_ids", flattenStringList(server.SecurityGroupIds))
	d.Set("subnet_ids", flattenStringList(server.SubnetIds))
	d.Set("endpoint", server.Endpoint)
	d.Set("arn", server.ServerArn)

	return nil
}

func resourceAwsOpsworksChefCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AWSClient).opsworkscmconn

	err := resourceAwsOpsworksChefValidate(d)
	if err != nil {
		return err
	}

	req := &opsworkscm.CreateServerInput{
		AssociatePublicIpAddress:   aws.Bool(d.Get("associate_public_ip_address").(bool)),
		BackupRetentionCount:       aws.Int64(int64(d.Get("backup_retention_count").(int))),
		DisableAutomatedBackup:     aws.Bool(d.Get("disable_automated_backup").(bool)),
		Engine:                     aws.String("Chef"),
		EngineModel:                aws.String(EngineModel),
		EngineVersion:              aws.String(EngineVersion),
		InstanceProfileArn:         aws.String(d.Get("instance_profile_arn").(string)),
		InstanceType:               aws.String(d.Get("instance_type").(string)),
		ServerName:                 aws.String(d.Get("name").(string)),
		PreferredBackupWindow:      aws.String(d.Get("preferred_backup_window").(string)),
		PreferredMaintenanceWindow: aws.String(d.Get("preferred_maintenance_window").(string)),
		SecurityGroupIds:           expandStringSet(d.Get("security_group_ids").(*schema.Set)),
		ServiceRoleArn:             aws.String(d.Get("service_role_arn").(string)),
		SubnetIds:                  expandStringSet(d.Get("subnet_ids").(*schema.Set)),
	}

	chefPivotalKeyAttribute := opsworkscm.EngineAttribute{
		Name:  aws.String("CHEF_PIVOTAL_KEY"),
		Value: aws.String(d.Get("chef_pivotal_key").(string)),
	}

	chefDeliveryAdminPassword := opsworkscm.EngineAttribute{
		Name:  aws.String("CHEF_DELIVERY_ADMIN_PASSWORD"),
		Value: aws.String(d.Get("chef_delivery_admin_password").(string)),
	}

	req.SetEngineAttributes([]*opsworkscm.EngineAttribute{&chefPivotalKeyAttribute, &chefDeliveryAdminPassword})

	// TODO: possibly make these optional
	// TODO: security_group_ids, optional
	// TODO: subnet_ids, optional

	if sshKeyPair, ok := d.GetOk("ssh_key_pair"); ok {
		req.KeyPair = aws.String(sshKeyPair.(string))
	}

	log.Printf("[DEBUG] Creating OpsWorks Chef server: %s", req)

	var resp *opsworkscm.CreateServerOutput
	err = resource.Retry(20*time.Minute, func() *resource.RetryError {
		var cerr error
		resp, cerr = client.CreateServer(req)
		if cerr != nil {
			if opserr, ok := cerr.(awserr.Error); ok {
				// TODO: does this also happen with this resource?
				// If Terraform is also managing the service IAM role,
				// it may have just been created and not yet be
				// propagated.
				// AWS doesn't provide a machine-readable code for this
				// specific error, so we're forced to do fragile message
				// matching.
				// The full error we're looking for looks something like
				// the following:
				// Service Role Arn: [...] is not yet propagated, please try again in a couple of minutes
				propErr := "not yet propagated"
				trustErr := "not the necessary trust relationship"
				validateErr := "validate IAM role permission"
				if opserr.Code() == "ValidationException" && (strings.Contains(opserr.Message(), trustErr) || strings.Contains(opserr.Message(), propErr) || strings.Contains(opserr.Message(), validateErr)) {
					log.Printf("[INFO] Waiting for service IAM role to propagate")
					return resource.RetryableError(cerr)
				}
			}
			return resource.NonRetryableError(cerr)
		}
		return nil
	})
	if err != nil {
		return err
	}

	serverName := *resp.Server.ServerName
	d.SetId(serverName)

	return resourceAwsOpsworksChefUpdate(d, meta)
}

func resourceAwsOpsworksChefUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AWSClient).opsworkscmconn

	err := resourceAwsOpsworksChefValidate(d)
	if err != nil {
		return err
	}

	// TODO: these are the only values that are allowed to be changed
	// according to the API, is there a way to express this in the provider somehow?
	req := &opsworkscm.UpdateServerInput{
		BackupRetentionCount:       aws.Int64(int64(d.Get("backup_retention_count").(int))),
		DisableAutomatedBackup:     aws.Bool(d.Get("disable_automated_backup").(bool)),
		PreferredBackupWindow:      aws.String(d.Get("preferred_backup_window").(string)),
		PreferredMaintenanceWindow: aws.String(d.Get("preferred_maintenance_window").(string)),
		ServerName:                 aws.String(d.Get("name").(string)),
	}

	log.Printf("[DEBUG] Updating OpsWorks Chef server: %s", req)

	_, err = client.UpdateServer(req)
	if err != nil {
		return err
	}

	return resourceAwsOpsworksChefRead(d, meta)
}

func resourceAwsOpsworksChefDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AWSClient).opsworkscmconn

	req := &opsworkscm.DeleteServerInput{
		ServerName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting OpsWorks Chef server: %s", d.Id())

	_, err := client.DeleteServer(req)
	if err != nil {
		return err
	}

	return nil
}
