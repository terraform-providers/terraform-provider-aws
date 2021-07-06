package fromproto

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/internal/tfplugin6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func Schema(in *tfplugin6.Schema) (*tfprotov6.Schema, error) {
	var resp tfprotov6.Schema
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

func SchemaBlock(in *tfplugin6.Schema_Block) (*tfprotov6.SchemaBlock, error) {
	resp := &tfprotov6.SchemaBlock{
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

func SchemaAttribute(in *tfplugin6.Schema_Attribute) (*tfprotov6.SchemaAttribute, error) {
	resp := &tfprotov6.SchemaAttribute{
		Name:            in.Name,
		Description:     in.Description,
		Required:        in.Required,
		Optional:        in.Optional,
		Computed:        in.Computed,
		Sensitive:       in.Sensitive,
		DescriptionKind: StringKind(in.DescriptionKind),
		Deprecated:      in.Deprecated,
	}

	if in.Type != nil {
		typ, err := tftypes.ParseJSONType(in.Type) //nolint:staticcheck
		if err != nil {
			return resp, err
		}
		resp.Type = typ
	}

	if in.NestedType != nil {
		nb, err := SchemaObject(in.NestedType)
		if err != nil {
			return resp, err
		}
		resp.NestedType = nb
	}

	return resp, nil
}

func SchemaAttributes(in []*tfplugin6.Schema_Attribute) ([]*tfprotov6.SchemaAttribute, error) {
	resp := make([]*tfprotov6.SchemaAttribute, 0, len(in))
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

func SchemaNestedBlock(in *tfplugin6.Schema_NestedBlock) (*tfprotov6.SchemaNestedBlock, error) {
	resp := &tfprotov6.SchemaNestedBlock{
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

func SchemaNestedBlocks(in []*tfplugin6.Schema_NestedBlock) ([]*tfprotov6.SchemaNestedBlock, error) {
	resp := make([]*tfprotov6.SchemaNestedBlock, 0, len(in))
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

func SchemaNestedBlockNestingMode(in tfplugin6.Schema_NestedBlock_NestingMode) tfprotov6.SchemaNestedBlockNestingMode {
	return tfprotov6.SchemaNestedBlockNestingMode(in)
}

func SchemaObjectNestingMode(in tfplugin6.Schema_Object_NestingMode) tfprotov6.SchemaObjectNestingMode {
	return tfprotov6.SchemaObjectNestingMode(in)
}

func SchemaObject(in *tfplugin6.Schema_Object) (*tfprotov6.SchemaObject, error) {
	resp := &tfprotov6.SchemaObject{
		Nesting:  SchemaObjectNestingMode(in.Nesting),
		MinItems: in.MinItems,
		MaxItems: in.MaxItems,
	}

	attrs, err := SchemaAttributes(in.Attributes)
	if err != nil {
		return nil, err
	}

	resp.Attributes = attrs
	return resp, nil
}
