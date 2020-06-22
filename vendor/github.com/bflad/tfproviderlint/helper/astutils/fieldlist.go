package astutils

import (
	"go/ast"
	"go/types"
)

// FieldListName returns field name at field position and name position if found
func FieldListName(fieldList *ast.FieldList, fieldPosition int, namePosition int) *string {
	names := FieldListNames(fieldList, fieldPosition)

	if names == nil || len(names) <= namePosition {
		return nil
	}

	name := names[namePosition]

	if name == nil {
		return nil
	}

	return &name.Name
}

// FieldListNames returns field names at field position if found
func FieldListNames(fieldList *ast.FieldList, position int) []*ast.Ident {
	if fieldList == nil {
		return nil
	}

	if len(fieldList.List) <= position {
		return nil
	}

	field := fieldList.List[position]

	if field == nil {
		return nil
	}

	return field.Names
}

// FieldListType returns type at field position if found
func FieldListType(fieldList *ast.FieldList, position int) *ast.Expr {
	if fieldList == nil {
		return nil
	}

	if len(fieldList.List) <= position {
		return nil
	}

	field := fieldList.List[position]

	if field == nil {
		return nil
	}

	return &field.Type
}

// HasFieldListLength returns true if the FieldList has the expected length
// If FieldList is nil, checks against expected length of 0.
func HasFieldListLength(fieldList *ast.FieldList, expectedLength int) bool {
	if fieldList == nil {
		return expectedLength == 0
	}

	return len(fieldList.List) == expectedLength
}

// IsFieldListType returns true if the field at position is present and matches expected ast.Expr
func IsFieldListType(fieldList *ast.FieldList, position int, exprFunc func(ast.Expr) bool) bool {
	t := FieldListType(fieldList, position)

	return t != nil && exprFunc(*t)
}

// IsFieldListTypePackageType returns true if the field at position is present and matches expected package type
func IsFieldListTypePackageType(fieldList *ast.FieldList, position int, info *types.Info, packageSuffix string, typeName string) bool {
	t := FieldListType(fieldList, position)

	return t != nil && IsPackageFunctionFieldListType(*t, info, packageSuffix, typeName)
}
