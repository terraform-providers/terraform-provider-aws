package fromproto

import (
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/internal/tfplugin5"
)

func DynamicValue(in *tfplugin5.DynamicValue) *tfprotov5.DynamicValue {
	return &tfprotov5.DynamicValue{
		MsgPack: in.Msgpack,
		JSON:    in.Json,
	}
}
