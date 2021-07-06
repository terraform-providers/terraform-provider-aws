package tfprotov6

import "github.com/hashicorp/terraform-plugin-go/tftypes"

const (
	// SchemaNestedBlockNestingModeInvalid indicates that the nesting mode
	// for a nested block in the schema is invalid. This generally
	// indicates a nested block that was created incorrectly.
	SchemaNestedBlockNestingModeInvalid SchemaNestedBlockNestingMode = 0

	// SchemaNestedBlockNestingModeSingle indicates that the nested block
	// should be treated as a single block with no labels, and there should
	// not be more than one of these blocks in the containing block. The
	// block will appear in config and state values as a tftypes.Object.
	SchemaNestedBlockNestingModeSingle SchemaNestedBlockNestingMode = 1

	// SchemaNestedBlockNestingModeList indicates that multiple instances
	// of the nested block should be permitted, with no labels, and that
	// the instances of the block should appear in config and state values
	// as a tftypes.List, with an ElementType of tftypes.Object.
	SchemaNestedBlockNestingModeList SchemaNestedBlockNestingMode = 2

	// SchemaNestedBlockNestingModeSet indicates that multiple instances
	// of the nested block should be permitted, with no labels, and that
	// the instances of the block should appear in config and state values
	// as a tftypes.Set, with an ElementType of tftypes.Object.
	SchemaNestedBlockNestingModeSet SchemaNestedBlockNestingMode = 3

	// SchemaNestedBlockNestingModeMap indicates that multiple instances of
	// the nested block should be permitted, each with a single label, and
	// that they should be represented in state and config values as a
	// tftypes.Map, with an AttributeType of tftypes.Object. The labels on
	// the blocks will be used as the map keys. It is an error, therefore,
	// to use the same label value on multiple block instances.
	SchemaNestedBlockNestingModeMap SchemaNestedBlockNestingMode = 4

	// SchemaNestedBlockNestingModeGroup indicates that the nested block
	// should be treated as a single block with no labels, and there should
	// not be more than one of these blocks in the containing block. The
	// block will appear in config and state values as a tftypes.Object.
	//
	// SchemaNestedBlockNestingModeGroup is distinct from
	// SchemaNestedBlockNestingModeSingle in that it guarantees that the
	// block will never be null. If it is omitted from a config, the block
	// will still be set, but its attributes and nested blocks will all be
	// null. This is an exception to the rule that any block not set in the
	// configuration cannot be set in config by the provider; this ensures
	// the block is always considered "set" in the configuration, and is
	// therefore settable in state by the provider.
	SchemaNestedBlockNestingModeGroup SchemaNestedBlockNestingMode = 5

	// SchemaObjectNestingModeInvalid indicates that the nesting mode
	// for a nested type in the schema is invalid. This generally
	// indicates a nested type that was created incorrectly.
	SchemaObjectNestingModeInvalid SchemaObjectNestingMode = 0

	// SchemaObjectNestingModeSingle indicates that the nested type should
	// be treated as a single object. The block will appear in config and state
	// values as a tftypes.Object.
	SchemaObjectNestingModeSingle SchemaObjectNestingMode = 1

	// SchemaObjectNestingModeList indicates that multiple instances of the
	// nested type should be permitted, and that the nested type should appear
	// in config and state values as a tftypes.List, with an ElementType of
	// tftypes.Object.
	SchemaObjectNestingModeList SchemaObjectNestingMode = 2

	// SchemaObjectNestingModeSet indicates that multiple instances of the
	// nested type should be permitted, and that the nested type should appear in
	// config and state values as a tftypes.Set, with an ElementType of
	// tftypes.Object.
	SchemaObjectNestingModeSet SchemaObjectNestingMode = 3

	// SchemaObjectNestingModeMap indicates that multiple instances of the
	// nested type should be permitted, and that they should be appear in state
	// and config values as a tftypes.Map, with an AttributeType of
	// tftypes.Object.
	SchemaObjectNestingModeMap SchemaObjectNestingMode = 4
)

// Schema is how Terraform defines the shape of data. It can be thought of as
// the type information for resources, data sources, provider configuration,
// and all the other data that Terraform sends to providers. It is how
// providers express their requirements for that data.
type Schema struct {
	// Version indicates which version of the schema this is. Versions
	// should be monotonically incrementing numbers. When Terraform
	// encounters a resource stored in state with a schema version lower
	// that the schema version the provider advertises for that resource,
	// Terraform requests the provider upgrade the resource's state.
	Version int64

	// Block is the root level of the schema, the collection of attributes
	// and blocks that make up a resource, data source, provider, or other
	// configuration block.
	Block *SchemaBlock
}

// SchemaBlock represents a block in a schema. Blocks are how Terraform creates
// groupings of attributes. In configurations, they don't use the equals sign
// and use dynamic instead of list comprehensions.
//
// Blocks will show up in state and config Values as a tftypes.Object, with the
// attributes and nested blocks defining the tftypes.Object's AttributeTypes.
type SchemaBlock struct {
	// TODO: why do we have version in the block, too?
	Version int64

	// Attributes are the attributes defined within the block. These are
	// the fields that users can set using the equals sign or reference in
	// interpolations.
	Attributes []*SchemaAttribute

	// BlockTypes are the nested blocks within the block. These are used to
	// have blocks within blocks.
	BlockTypes []*SchemaNestedBlock

	// Description offers an end-user friendly description of what the
	// block is for. This will be surfaced to users through editor
	// integrations, documentation generation, and other settings.
	Description string

	// DescriptionKind indicates the formatting and encoding that the
	// Description field is using.
	DescriptionKind StringKind

	// Deprecated, when set to true, indicates that a block should no
	// longer be used and users should migrate away from it. At the moment
	// it is unused and will have no impact, but it will be used in future
	// tooling that is powered by provider schemas to enable richer user
	// experiences. Providers should set it when deprecating blocks in
	// preparation for these tools.
	Deprecated bool
}

// SchemaAttribute represents a single attribute within a schema block.
// Attributes are the fields users can set in configuration using the equals
// sign, can assign to variables, can interpolate, and can use list
// comprehensions on.
type SchemaAttribute struct {
	// Name is the name of the attribute. This is what the user will put
	// before the equals sign to assign a value to this attribute.
	Name string

	// Type indicates the type of data the attribute expects. See the
	// documentation for the tftypes package for information on what types
	// are supported and their behaviors.
	Type tftypes.Type

	// NestedType indicates that this is a NestedBlock-style object masquerading
	// as an attribute. This field conflicts with Type.
	NestedType *SchemaObject

	// Description offers an end-user friendly description of what the
	// attribute is for. This will be surfaced to users through editor
	// integrations, documentation generation, and other settings.
	Description string

	// Required, when set to true, indicates that this attribute must have
	// a value assigned to it by the user or Terraform will throw an error.
	Required bool

	// Optional, when set to true, indicates that the user does not need to
	// supply a value for this attribute, but may.
	Optional bool

	// Computed, when set to true, indicates the the provider will supply a
	// value for this field. If Optional and Required are false and
	// Computed is true, the user will not be able to specify a value for
	// this field without Terraform throwing an error. If Optional is true
	// and Computed is true, the user can specify a value for this field,
	// but the provider may supply a value if the user does not. It is
	// always a violation of Terraform's protocol to substitute a value for
	// what the user entered, even if Computed is true.
	Computed bool

	// Sensitive, when set to true, indicates that the contents of this
	// attribute should be considered sensitive and not included in output.
	// This does not encrypt or otherwise protect these values in state, it
	// only offers protection from them showing up in plans or other
	// output.
	Sensitive bool

	// DescriptionKind indicates the formatting and encoding that the
	// Description field is using.
	DescriptionKind StringKind

	// Deprecated, when set to true, indicates that a attribute should no
	// longer be used and users should migrate away from it. At the moment
	// it is unused and will have no impact, but it will be used in future
	// tooling that is powered by provider schemas to enable richer user
	// experiences. Providers should set it when deprecating attributes in
	// preparation for these tools.
	Deprecated bool
}

// SchemaNestedBlock is a nested block within another block. See SchemaBlock
// for more information on blocks.
type SchemaNestedBlock struct {
	// TypeName is the name of the block. It is what the user will specify
	// when using the block in configuration.
	TypeName string

	// Block is the block being nested inside another block. See the
	// SchemaBlock documentation for more information on blocks.
	Block *SchemaBlock

	// Nesting is the kind of nesting the block is using. Different nesting
	// modes have different behaviors and imply different kinds of data.
	Nesting SchemaNestedBlockNestingMode

	// MinItems is the minimum number of instances of this block that a
	// user must specify or Terraform will return an error.
	//
	// MinItems can only be set for SchemaNestedBlockNestingModeList and
	// SchemaNestedBlockNestingModeSet. SchemaNestedBlockNestingModeSingle
	// can also set MinItems and MaxItems both to 1 to indicate that the
	// block is required to be set. All other SchemaNestedBlockNestingModes
	// must leave MinItems set to 0.
	MinItems int64

	// MaxItems is the maximum number of instances of this block that a
	// user may specify before Terraform returns an error.
	//
	// MaxItems can only be set for SchemaNestedBlockNestingModeList and
	// SchemaNestedBlockNestingModeSet. SchemaNestedBlockNestingModeSingle
	// can also set MinItems and MaxItems both to 1 to indicate that the
	// block is required to be set. All other SchemaNestedBlockNestingModes
	// must leave MaxItems set to 0.
	MaxItems int64
}

// SchemaNestedBlockNestingMode indicates the nesting mode for
// SchemaNestedBlocks. The nesting mode determines the number of instances of
// the block allowed, how many labels the block expects, and the data structure
// used for the block in config and state values.
type SchemaNestedBlockNestingMode int32

func (s SchemaNestedBlockNestingMode) String() string {
	switch s {
	case 0:
		return "INVALID"
	case 1:
		return "SINGLE"
	case 2:
		return "LIST"
	case 3:
		return "SET"
	case 4:
		return "MAP"
	case 5:
		return "GROUP"
	}
	return "UNKNOWN"
}

// SchemaObject represents a nested-block-stype object in an Attribute.
type SchemaObject struct {
	// Attributes are the attributes defined within the Object.
	Attributes []*SchemaAttribute

	Nesting SchemaObjectNestingMode

	// MinItems is the minimum number of instances of this type that a
	// user must specify or Terraform will return an error.
	MinItems int64

	// MaxItems is the maximum number of instances of this type that a
	// user may specify before Terraform returns an error.
	MaxItems int64
}

// SchemaObjectNestingMode indicates the nesting mode for
// SchemaNestedBlocks. The nesting mode determines the number of instances of
// the nested type allowed and the data structure used for the block in config
// and state values.
type SchemaObjectNestingMode int32

func (s SchemaObjectNestingMode) String() string {
	switch s {
	case 0:
		return "INVALID"
	case 1:
		return "SINGLE"
	case 2:
		return "LIST"
	case 3:
		return "SET"
	case 4:
		return "MAP"
	}
	return "UNKNOWN"
}
