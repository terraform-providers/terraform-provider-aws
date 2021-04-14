package configschema

import (
	"github.com/hashicorp/go-cty/cty"
)

// ImpliedType returns the cty.Type that would result from decoding a
// configuration block using the receiving block schema.
//
// ImpliedType always returns a result, even if the given schema is
// inconsistent.
func (b *Block) ImpliedType() cty.Type {
	if b == nil {
		return cty.EmptyObject
	}

	atys := make(map[string]cty.Type)

	for name, attrS := range b.Attributes {
		atys[name] = attrS.Type
	}

	for name, blockS := range b.BlockTypes {
		if _, exists := atys[name]; exists {
			panic("invalid schema, blocks and attributes cannot have the same name")
		}

		childType := blockS.Block.ImpliedType()

		switch blockS.Nesting {
		case NestingSingle, NestingGroup:
			atys[name] = childType
		case NestingList:
			// We prefer to use a list where possible, since it makes our
			// implied type more complete, but if there are any
			// dynamically-typed attributes inside we must use a tuple
			// instead, which means our type _constraint_ must be
			// cty.DynamicPseudoType to allow the tuple type to be decided
			// separately for each value.
			if childType.HasDynamicTypes() {
				atys[name] = cty.DynamicPseudoType
			} else {
				atys[name] = cty.List(childType)
			}
		case NestingSet:
			if childType.HasDynamicTypes() {
				panic("can't use cty.DynamicPseudoType inside a block type with NestingSet")
			}
			atys[name] = cty.Set(childType)
		case NestingMap:
			// We prefer to use a map where possible, since it makes our
			// implied type more complete, but if there are any
			// dynamically-typed attributes inside we must use an object
			// instead, which means our type _constraint_ must be
			// cty.DynamicPseudoType to allow the tuple type to be decided
			// separately for each value.
			if childType.HasDynamicTypes() {
				atys[name] = cty.DynamicPseudoType
			} else {
				atys[name] = cty.Map(childType)
			}
		default:
			panic("invalid nesting type")
		}
	}

	return cty.Object(atys)
}
