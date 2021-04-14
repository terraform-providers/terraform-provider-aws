package aws

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsFsxLustreFileSystem() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsFsxLustreFileSystemCreate,
		Read:   resourceAwsFsxLustreFileSystemRead,
		Update: resourceAwsFsxLustreFileSystemUpdate,
		Delete: resourceAwsFsxLustreFileSystemDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"export_path": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(3, 900),
					validation.StringMatch(regexp.MustCompile(`^s3://`), "must begin with s3://"),
				),
			},
			"import_path": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(3, 900),
					validation.StringMatch(regexp.MustCompile(`^s3://`), "must begin with s3://"),
				),
			},
			"imported_file_chunk_size": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(1, 512000),
			},
			"mount_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"network_interface_ids": {
				// As explained in https://docs.aws.amazon.com/fsx/latest/LustreGuide/mounting-on-premises.html, the first
				// network_interface_id is the primary one, so ordering matters. Use TypeList instead of TypeSet to preserve it.
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"security_group_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				MaxItems: 50,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"storage_capacity": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntAtLeast(1200),
			},
			"subnet_ids": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MinItems: 1,
				MaxItems: 1,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags": tagsSchema(),
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"weekly_maintenance_start_time": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(7, 7),
					validation.StringMatch(regexp.MustCompile(`^[1-7]:([01]\d|2[0-3]):?([0-5]\d)$`), "must be in the format d:HH:MM"),
				),
			},
			"deployment_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      fsx.LustreDeploymentTypeScratch1,
				ValidateFunc: validation.StringInSlice(fsx.LustreDeploymentType_Values(), false),
			},
			"kms_key_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"per_unit_storage_throughput": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
				ValidateFunc: validation.IntInSlice([]int{
					12,
					40,
					50,
					100,
					200,
				}),
			},
			"automatic_backup_retention_days": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(0, 90),
			},
			"daily_automatic_backup_start_time": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(5, 5),
					validation.StringMatch(regexp.MustCompile(`^([01]\d|2[0-3]):?([0-5]\d)$`), "must be in the format HH:MM"),
				),
			},
			"storage_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      fsx.StorageTypeSsd,
				ValidateFunc: validation.StringInSlice(fsx.StorageType_Values(), false),
			},
			"drive_cache_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(fsx.DriveCacheType_Values(), false),
			},
			"auto_import_policy": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(fsx.AutoImportPolicyType_Values(), false),
			},
			"copy_tags_to_backups": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
			},
		},
	}
}

func resourceAwsFsxLustreFileSystemCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).fsxconn

	input := &fsx.CreateFileSystemInput{
		ClientRequestToken: aws.String(resource.UniqueId()),
		FileSystemType:     aws.String(fsx.FileSystemTypeLustre),
		StorageCapacity:    aws.Int64(int64(d.Get("storage_capacity").(int))),
		StorageType:        aws.String(d.Get("storage_type").(string)),
		SubnetIds:          expandStringList(d.Get("subnet_ids").([]interface{})),
		LustreConfiguration: &fsx.CreateFileSystemLustreConfiguration{
			DeploymentType: aws.String(d.Get("deployment_type").(string)),
		},
	}

	//Applicable only for TypePersistent1
	if v, ok := d.GetOk("kms_key_id"); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("automatic_backup_retention_days"); ok {
		input.LustreConfiguration.AutomaticBackupRetentionDays = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("daily_automatic_backup_start_time"); ok {
		input.LustreConfiguration.DailyAutomaticBackupStartTime = aws.String(v.(string))
	}

	if v, ok := d.GetOk("export_path"); ok {
		input.LustreConfiguration.ExportPath = aws.String(v.(string))
	}

	if v, ok := d.GetOk("import_path"); ok {
		input.LustreConfiguration.ImportPath = aws.String(v.(string))
	}

	if v, ok := d.GetOk("imported_file_chunk_size"); ok {
		input.LustreConfiguration.ImportedFileChunkSize = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("security_group_ids"); ok {
		input.SecurityGroupIds = expandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("tags"); ok {
		input.Tags = keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().FsxTags()
	}

	if v, ok := d.GetOk("weekly_maintenance_start_time"); ok {
		input.LustreConfiguration.WeeklyMaintenanceStartTime = aws.String(v.(string))
	}

	if v, ok := d.GetOk("per_unit_storage_throughput"); ok {
		input.LustreConfiguration.PerUnitStorageThroughput = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("drive_cache_type"); ok {
		input.LustreConfiguration.DriveCacheType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("auto_import_policy"); ok {
		input.LustreConfiguration.AutoImportPolicy = aws.String(v.(string))
	}

	if v, ok := d.GetOk("copy_tags_to_backups"); ok {
		input.LustreConfiguration.CopyTagsToBackups = aws.Bool(v.(bool))
	}

	result, err := conn.CreateFileSystem(input)
	if err != nil {
		return fmt.Errorf("Error creating FSx Lustre filesystem: %w", err)
	}

	d.SetId(aws.StringValue(result.FileSystem.FileSystemId))

	log.Println("[DEBUG] Waiting for filesystem to become available")

	if err := waitForFsxFileSystemCreation(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("Error waiting for filesystem (%s) to become available: %w", d.Id(), err)
	}

	return resourceAwsFsxLustreFileSystemRead(d, meta)
}

func resourceAwsFsxLustreFileSystemUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).fsxconn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.FsxUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating FSx Lustre File System (%s) tags: %w", d.Get("arn").(string), err)
		}
	}

	requestUpdate := false
	input := &fsx.UpdateFileSystemInput{
		ClientRequestToken:  aws.String(resource.UniqueId()),
		FileSystemId:        aws.String(d.Id()),
		LustreConfiguration: &fsx.UpdateFileSystemLustreConfiguration{},
	}

	if d.HasChange("weekly_maintenance_start_time") {
		input.LustreConfiguration.WeeklyMaintenanceStartTime = aws.String(d.Get("weekly_maintenance_start_time").(string))
		requestUpdate = true
	}

	if d.HasChange("automatic_backup_retention_days") {
		input.LustreConfiguration.AutomaticBackupRetentionDays = aws.Int64(int64(d.Get("automatic_backup_retention_days").(int)))
		requestUpdate = true
	}

	if d.HasChange("daily_automatic_backup_start_time") {
		input.LustreConfiguration.DailyAutomaticBackupStartTime = aws.String(d.Get("daily_automatic_backup_start_time").(string))
		requestUpdate = true
	}

	if d.HasChange("auto_import_policy") {
		input.LustreConfiguration.AutoImportPolicy = aws.String(d.Get("auto_import_policy").(string))
		requestUpdate = true
	}

	if requestUpdate {
		_, err := conn.UpdateFileSystem(input)
		if err != nil {
			return fmt.Errorf("error updating FSX Lustre File System (%s): %w", d.Id(), err)
		}

		log.Println("[DEBUG] Waiting for filesystem to become available")

		if err := waitForFsxFileSystemUpdate(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
			return fmt.Errorf("Error waiting for filesystem (%s) to become available: %w", d.Id(), err)
		}
	}

	return resourceAwsFsxLustreFileSystemRead(d, meta)
}

func resourceAwsFsxLustreFileSystemRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).fsxconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	filesystem, err := describeFsxFileSystem(conn, d.Id())

	if isAWSErr(err, fsx.ErrCodeFileSystemNotFound, "") {
		log.Printf("[WARN] FSx File System (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error reading FSx Lustre File System (%s): %w", d.Id(), err)
	}

	if filesystem == nil {
		log.Printf("[WARN] FSx File System (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	lustreConfig := filesystem.LustreConfiguration

	if filesystem.WindowsConfiguration != nil {
		return fmt.Errorf("expected FSx Lustre File System, found FSx Windows File System: %s", d.Id())
	}

	if lustreConfig == nil {
		return fmt.Errorf("error describing FSx Lustre File System (%s): empty Lustre configuration", d.Id())
	}

	if lustreConfig.DataRepositoryConfiguration == nil {
		// Initialize an empty structure to simplify d.Set() handling
		lustreConfig.DataRepositoryConfiguration = &fsx.DataRepositoryConfiguration{}
	}

	d.Set("arn", filesystem.ResourceARN)
	d.Set("dns_name", filesystem.DNSName)
	d.Set("export_path", lustreConfig.DataRepositoryConfiguration.ExportPath)
	d.Set("import_path", lustreConfig.DataRepositoryConfiguration.ImportPath)
	d.Set("auto_import_policy", lustreConfig.DataRepositoryConfiguration.AutoImportPolicy)
	d.Set("imported_file_chunk_size", lustreConfig.DataRepositoryConfiguration.ImportedFileChunkSize)
	d.Set("deployment_type", lustreConfig.DeploymentType)
	if lustreConfig.PerUnitStorageThroughput != nil {
		d.Set("per_unit_storage_throughput", lustreConfig.PerUnitStorageThroughput)
	}
	d.Set("mount_name", filesystem.LustreConfiguration.MountName)
	d.Set("storage_type", filesystem.StorageType)
	if filesystem.LustreConfiguration.DriveCacheType != nil {
		d.Set("drive_cache_type", filesystem.LustreConfiguration.DriveCacheType)
	}

	if filesystem.KmsKeyId != nil {
		d.Set("kms_key_id", filesystem.KmsKeyId)
	}

	if err := d.Set("network_interface_ids", aws.StringValueSlice(filesystem.NetworkInterfaceIds)); err != nil {
		return fmt.Errorf("error setting network_interface_ids: %w", err)
	}

	d.Set("owner_id", filesystem.OwnerId)
	d.Set("storage_capacity", filesystem.StorageCapacity)

	if err := d.Set("subnet_ids", aws.StringValueSlice(filesystem.SubnetIds)); err != nil {
		return fmt.Errorf("error setting subnet_ids: %w", err)
	}

	if err := d.Set("tags", keyvaluetags.FsxKeyValueTags(filesystem.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	d.Set("vpc_id", filesystem.VpcId)
	d.Set("weekly_maintenance_start_time", lustreConfig.WeeklyMaintenanceStartTime)
	d.Set("automatic_backup_retention_days", lustreConfig.AutomaticBackupRetentionDays)
	d.Set("daily_automatic_backup_start_time", lustreConfig.DailyAutomaticBackupStartTime)
	d.Set("copy_tags_to_backups", filesystem.LustreConfiguration.CopyTagsToBackups)

	return nil
}

func resourceAwsFsxLustreFileSystemDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).fsxconn

	request := &fsx.DeleteFileSystemInput{
		FileSystemId: aws.String(d.Id()),
	}

	_, err := conn.DeleteFileSystem(request)

	if isAWSErr(err, fsx.ErrCodeFileSystemNotFound, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error deleting FSx Lustre filesystem: %w", err)
	}

	log.Println("[DEBUG] Waiting for filesystem to delete")

	if err := waitForFsxFileSystemDeletion(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("Error waiting for filesystem (%s) to delete: %w", d.Id(), err)
	}

	return nil
}
