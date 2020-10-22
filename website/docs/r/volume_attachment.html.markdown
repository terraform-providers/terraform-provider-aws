---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_volume_attachment"
description: |-
  Provides an AWS EBS Volume Attachment
---

# Resource: aws_volume_attachment

Provides an AWS EBS Volume Attachment as a top level resource, to attach and
detach volumes from AWS Instances.

~> **NOTE on EBS block devices:** If you use `ebs_block_device` on an `aws_instance`, Terraform will assume management over the full set of non-root EBS block devices for the instance, and treats additional block devices as drift. For this reason, `ebs_block_device` cannot be mixed with external `aws_ebs_volume` + `aws_ebs_volume_attachment` resources for a given instance.

## Example Usage

```hcl
resource "aws_volume_attachment" "ebs_att" {
  device_name = "/dev/sdh"
  volume_id   = aws_ebs_volume.example.id
  instance_id = aws_instance.web.id
}

resource "aws_instance" "web" {
  ami               = "ami-21f78e11"
  availability_zone = "us-west-2a"
  instance_type     = "t2.micro"

  tags = {
    Name = "HelloWorld"
  }
}

resource "aws_ebs_volume" "example" {
  availability_zone = "us-west-2a"
  size              = 1
}
```

## Argument Reference

The following arguments are supported:

* `device_name` - (Required) The device name to expose to the instance (for
example, `/dev/sdh` or `xvdh`).  See [Device Naming on Linux Instances][1] and [Device Naming on Windows Instances][2] for more information.
* `instance_id` - (Required) ID of the Instance to attach to
* `volume_id` - (Required) ID of the Volume to be attached
* `force_detach` - (Optional, Boolean) Set to `true` if you want to force the
volume to detach. Useful if previous attempts failed, but use this option only
as a last resort, as this can result in **data loss**. See
[Detaching an Amazon EBS Volume from an Instance][3] for more information.
* `skip_destroy` - (Optional, Boolean) Set this to true if you do not wish
to detach the volume from the instance to which it is attached at destroy
time, and instead just remove the attachment from Terraform state. This is
useful when destroying an instance which has volumes created by some other
means attached.

## Attributes Reference

* `device_name` - The device name exposed to the instance
* `instance_id` - ID of the Instance
* `volume_id` - ID of the Volume

## Import

EBS Volume Attachments can be imported using `DEVICE_NAME:VOLUME_ID:INSTANCE_ID`, e.g.

```
$ terraform import aws_volume_attachment.example /dev/sdh:vol-049df61146c4d7901:i-12345678
```


[1]: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/device_naming.html#available-ec2-device-names
[2]: https://docs.aws.amazon.com/AWSEC2/latest/WindowsGuide/device_naming.html#available-ec2-device-names
[3]: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ebs-detaching-volume.html
