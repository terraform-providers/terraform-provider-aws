---
subcategory: "AppStream"
layout: "aws"
page_title: "AWS: aws_appstream_stack"
description: |-
  Provides an AppStream stack
---

# Resource: aws_appstream_stack

Provides an AppStream stack.

## Example Usage

```hcl
resource "aws_appstream_stack" "appstream_stack" {
  name         = "stack name"
  description  = "stack description"
  display_name = "stack display name"
  feedback_url = "http://your-domain/feedback"
  redirect_url = "http://your-domain/redirect"
  storage_connectors {
    connector_type = "HOMEFOLDERS"
  }
  user_settings {
    action  = "CLIPBOARD_COPY_FROM_LOCAL_DEVICE"
    enabled = true
  }
  user_settings {
    action  = "CLIPBOARD_COPY_TO_LOCAL_DEVICE"
    enabled = true
  }
  user_settings {
    action  = "FILE_UPLOAD"
    enabled = true
  }
  user_settings {
    action  = "FILE_DOWNLOAD"
    enabled = true
  }
  application_settings {
    enabled        = true
    settings_group = "SettingsGroup"
  }

  tags = {
    TagName = "TagValue"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the AppStream stack, used as the stack's identifier.  Only allows alphanumeric, hypen, underscore and period.
* `description` - (Optional) Description for the AppStream stack.
* `display_name` - (Optional) Human-readable friendly name for the AppStream stack.
* `redirect_url` - (Optional) URL to redirect at end of session.
* `feedback_url` - (Optional) URL for users to submit feedback.
* `storage_connectors` - (Optional) Nested block of storage connectors.  Each one requires `connector_type`, e.g., `HOMEFOLDERS`
* `user_settings` - (Optional) Nested block of AppStream user settings.  Each one requires `action` and `enabled`.
* `application_settings` - (Optional) settings for application settings persistence.  Requires `enabled` and a `settings_group`.

## Attributes Reference

* `name` - Identifier for this AppStream stack, as specified in input
