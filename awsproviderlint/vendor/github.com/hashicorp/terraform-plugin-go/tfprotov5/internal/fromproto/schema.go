package fromproto

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/internal/tfplugin5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"
)

func Schema(in *tfplugin5.Schema) (*tfprotov5.Schema, error) {
	var resp tfprotov5.Schema
	resp.Version = in.Version
	if in.Block != nil {
		block, err := SchemaBlock(in.Block)
		if err != nil {
			return &resp, err
		}
		resp.Block = block
	}
	return &resp, nil
}

func SchemaBlock(in *tfplugin5.Schema_Block) (*tfprotov5.SchemaBlock, error) {
	resp := &tfprotov5.SchemaBlock{
		Version:         in.Version,
		Description:     in.Description,
		DescriptionKind: StringKind(in.DescriptionKind),
		Deprecated:      in.Deprecated,
	}
	attrs, err := SchemaAttributes(in.Attributes)
	if err != nil {
		return resp, err
	}
	resp.Attributes = attrs
	blocks, err := SchemaNestedBlocks(in.BlockTypes)
	if err != nil {
		return resp, err
	}
	resp.BlockTypes = blocks
	return resp, nil
}

func SchemaAttribute(in *tfplugin5.Schema_Attribute) (*tfprotov5.SchemaAttribute, error) {
	resp := &tfprotov5.SchemaAttribute{
		Name:            in.Name,
		Description:     in.Description,
		Required:        in.Required,
		Optional:        in.Optional,
		Computed:        in.Computed,
		Sensitive:       in.Sensitive,
		DescriptionKind: StringKind(in.DescriptionKind),
		Deprecated:      in.Deprecated,
	}
	typ, err := tftypes.ParseJSONType(in.Type) //nolint:staticcheck
	if err != nil {
		return resp, err
	}
	resp.Type = typ
	return resp, nil
}

func SchemaAttributes(in []*tfplugin5.Schema_Attribute) ([]*tfprotov5.SchemaAttribute, error) {
	resp := make([]*tfprotov5.SchemaAttribute, 0, len(in))
	for pos, a := range in {
		if a == nil {
			resp = append(resp, nil)
			continue
		}
		attr, err := SchemaAttribute(a)
		if err != nil {
			return resp, fmt.Errorf("error converting schema attribute %d: %w", pos, err)
		}
		resp = append(resp, attr)
	}
	return resp, nil
}

func SchemaNestedBlock(in *tfplugin5.Schema_NestedBlock) (*tfprotov5.SchemaNestedBlock, error) {
	resp := &tfprotov5.SchemaNestedBlock{
		TypeName: in.TypeName,
		Nesting:  SchemaNestedBlockNestingMode(in.Nesting),
		MinItems: in.MinItems,
		MaxItems: in.MaxItems,
	}
	if in.Block != nil {
		block, err := SchemaBlock(in.Block)
		if err != nil {
			return resp, err
		}
		resp.Block = block
	}
	return resp, nil
}

func SchemaNestedBlocks(in []*tfplugin5.Schema_NestedBlock) ([]*tfprotov5.SchemaNestedBlock, error) {
	resp := make([]*tfprotov5.SchemaNestedBlock, 0, len(in))
	for pos, b := range in {
		if b == nil {
			resp = append(resp, nil)
			continue
		}
		block, err := SchemaNestedBlock(b)
		if err != nil {
			return resp, fmt.Errorf("error converting nested block %d: %w", pos, err)
		}
		resp = append(resp, block)
	}
	return resp, nil
}

func SchemaNestedBlockNestingMode(in tfplugin5.Schema_NestedBlock_NestingMode) tfprotov5.SchemaNestedBlockNestingMode {
	return tfprotov5.SchemaNestedBlockNestingMode(in)
}
