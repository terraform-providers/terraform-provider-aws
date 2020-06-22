package aws

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
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
			"network_interface_ids": {
				Type:     schema.TypeSet,
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
		},
	}
}

func resourceAwsFsxLustreFileSystemCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).fsxconn

	input := &fsx.CreateFileSystemInput{
		ClientRequestToken: aws.String(resource.UniqueId()),
		FileSystemType:     aws.String(fsx.FileSystemTypeLustre),
		StorageCapacity:    aws.Int64(int64(d.Get("storage_capacity").(int))),
		SubnetIds:          expandStringList(d.Get("subnet_ids").([]interface{})),
	}

	if v, ok := d.GetOk("export_path"); ok {
		if input.LustreConfiguration == nil {
			input.LustreConfiguration = &fsx.CreateFileSystemLustreConfiguration{}
		}

		input.LustreConfiguration.ExportPath = aws.String(v.(string))
	}

	if v, ok := d.GetOk("import_path"); ok {
		if input.LustreConfiguration == nil {
			input.LustreConfiguration = &fsx.CreateFileSystemLustreConfiguration{}
		}

		input.LustreConfiguration.ImportPath = aws.String(v.(string))
	}

	if v, ok := d.GetOk("imported_file_chunk_size"); ok {
		if input.LustreConfiguration == nil {
			input.LustreConfiguration = &fsx.CreateFileSystemLustreConfiguration{}
		}

		input.LustreConfiguration.ImportedFileChunkSize = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("security_group_ids"); ok {
		input.SecurityGroupIds = expandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("tags"); ok {
		input.Tags = keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().FsxTags()
	}

	if v, ok := d.GetOk("weekly_maintenance_start_time"); ok {
		if input.LustreConfiguration == nil {
			input.LustreConfiguration = &fsx.CreateFileSystemLustreConfiguration{}
		}

		input.LustreConfiguration.WeeklyMaintenanceStartTime = aws.String(v.(string))
	}

	result, err := conn.CreateFileSystem(input)
	if err != nil {
		return fmt.Errorf("Error creating FSx filesystem: %s", err)
	}

	d.SetId(*result.FileSystem.FileSystemId)

	log.Println("[DEBUG] Waiting for filesystem to become available")

	if err := waitForFsxFileSystemCreation(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("Error waiting for filesystem (%s) to become available: %s", d.Id(), err)
	}

	return resourceAwsFsxLustreFileSystemRead(d, meta)
}

func resourceAwsFsxLustreFileSystemUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).fsxconn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.FsxUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating FSx Windows File System (%s) tags: %s", d.Get("arn").(string), err)
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

	if requestUpdate {
		_, err := conn.UpdateFileSystem(input)
		if err != nil {
			return fmt.Errorf("error updating FSX File System (%s): %s", d.Id(), err)
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
		return fmt.Errorf("Error reading FSx File System (%s): %s", d.Id(), err)
	}

	if filesystem == nil {
		log.Printf("[WARN] FSx File System (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if filesystem.WindowsConfiguration != nil {
		return fmt.Errorf("expected FSx Lustre File System, found FSx Windows File System: %s", d.Id())
	}

	if filesystem.LustreConfiguration == nil {
		return fmt.Errorf("error describing FSx Lustre File System (%s): empty Lustre configuration", d.Id())
	}

	if filesystem.LustreConfiguration.DataRepositoryConfiguration == nil {
		// Initialize an empty structure to simplify d.Set() handling
		filesystem.LustreConfiguration.DataRepositoryConfiguration = &fsx.DataRepositoryConfiguration{}
	}

	d.Set("arn", filesystem.ResourceARN)
	d.Set("dns_name", filesystem.DNSName)
	d.Set("export_path", filesystem.LustreConfiguration.DataRepositoryConfiguration.ExportPath)
	d.Set("import_path", filesystem.LustreConfiguration.DataRepositoryConfiguration.ImportPath)
	d.Set("imported_file_chunk_size", filesystem.LustreConfiguration.DataRepositoryConfiguration.ImportedFileChunkSize)

	if err := d.Set("network_interface_ids", aws.StringValueSlice(filesystem.NetworkInterfaceIds)); err != nil {
		return fmt.Errorf("error setting network_interface_ids: %s", err)
	}

	d.Set("owner_id", filesystem.OwnerId)
	d.Set("storage_capacity", filesystem.StorageCapacity)

	if err := d.Set("subnet_ids", aws.StringValueSlice(filesystem.SubnetIds)); err != nil {
		return fmt.Errorf("error setting subnet_ids: %s", err)
	}

	if err := d.Set("tags", keyvaluetags.FsxKeyValueTags(filesystem.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	d.Set("vpc_id", filesystem.VpcId)
	d.Set("weekly_maintenance_start_time", filesystem.LustreConfiguration.WeeklyMaintenanceStartTime)

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
		return fmt.Errorf("Error deleting FSx filesystem: %s", err)
	}

	log.Println("[DEBUG] Waiting for filesystem to delete")

	if err := waitForFsxFileSystemDeletion(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("Error waiting for filesystem (%s) to delete: %s", d.Id(), err)
	}

	return nil
}
