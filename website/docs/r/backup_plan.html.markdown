---
subcategory: "Backup"
layout: "aws"
page_title: "AWS: aws_backup_plan"
description: |-
  Provides an AWS Backup plan resource.
---

# Resource: aws_backup_plan

Provides an AWS Backup plan resource.

## Example Usage

```hcl
resource "aws_backup_plan" "example" {
  name = "tf_example_backup_plan"

  rule {
    rule_name         = "tf_example_backup_rule"
    target_vault_name = aws_backup_vault.test.name
    schedule          = "cron(0 12 * * ? *)"
  }

  advanced_backup_setting {
    backup_options = {
      WindowsVSS = "enabled"
    }
    resource_type = "EC2"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The display name of a backup plan.
* `rule` - (Required) A rule object that specifies a scheduled task that is used to back up a selection of resources.
* `advanced_backup_setting` - (Optional) An object that specifies backup options for each resource type.
* `tags` - (Optional) Metadata that you can assign to help organize the plans you create.

### Rule Arguments
For **rule** the following attributes are supported:

* `rule_name` - (Required) An display name for a backup rule.
* `target_vault_name` - (Required) The name of a logical container where backups are stored.
* `schedule` - (Optional) A CRON expression specifying when AWS Backup initiates a backup job.
* `start_window` - (Optional) The amount of time in minutes before beginning a backup.
* `completion_window` - (Optional) The amount of time AWS Backup attempts a backup before canceling the job and returning an error.
* `lifecycle` - (Optional) The lifecycle defines when a protected resource is transitioned to cold storage and when it expires.  Fields documented below.
* `recovery_point_tags` - (Optional) Metadata that you can assign to help organize the resources that you create.
* `copy_action` - (Optional) Configuration block(s) with copy operation settings. Detailed below.

### Lifecycle Arguments
For **lifecycle** the following attributes are supported:

* `cold_storage_after` - (Optional) Specifies the number of days after creation that a recovery point is moved to cold storage.
* `delete_after` - (Optional) Specifies the number of days after creation that a recovery point is deleted. Must be 90 days greater than `cold_storage_after`.

### Copy Action Arguments
For **copy_action** the following attributes are supported:

* `lifecycle` - (Optional) The lifecycle defines when a protected resource is copied over to a backup vault and when it expires.  Fields documented above.
* `destination_vault_arn` - (Required) An Amazon Resource Name (ARN) that uniquely identifies the destination backup vault for the copied backup.

### Advanced Backup Setting Arguments
For `advanced_backup_setting` the following attibutes are supported:

* `backup_options` - (Optional) Specifies the backup option for a selected resource. This option is only available for Windows VSS backup jobs. Set to `{ WindowsVSS = "enabled" }` to enable Windows VSS backup option and create a VSS Windows backup.
* `resource_type` - (Optional) The type of AWS resource to be backed up. For VSS Windows backups, the only supported resource type is Amazon EC2. Valid values: `EC2`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The id of the backup plan.
* `arn` - The ARN of the backup plan.
* `version` - Unique, randomly generated, Unicode, UTF-8 encoded string that serves as the version ID of the backup plan.

## Import

Backup Plan can be imported using the `id`, e.g.

```
$ terraform import aws_backup_plan.test <id>
```
