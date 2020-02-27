# =========================================================================== #
#
# Proposed arguments for DynamoDB table item data source, using an HCL example.
#
# Status: Initial Design
#
# --------------------------------------------------------------------------- #
#
# We would want to stick pretty closely to DynamoDB's GetItem API for this:
# https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_GetItem.html
# 
# The attribute value objects get a bit more complex, with several possible
# types available:
# https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_AttributeValue.html
#
# For more details about argument types, see notes for the aws_dynamodb_query
# data source.
#
# =========================================================================== #
#
#
#
#
data "aws_dynamodb_table_item" "example" {
  # Use strong consistentcy instead of eventual if true. Won't work with global
  # tables.
  #
  # Type: bool
  # Required: false
  consistent_read = false

  # Create aliases for attributes that can be used to reference attribute names
  # (and/or values?) in expressions.
  #
  # Type: map(str, str))
  # Required: false
  expression_attribute_names = {
    "#A": "example-attribute-A"
    "#B": "example-attribute-B"
  }

  # An AttributeValue mapping that maps attribute names to a type = value
  # object. Values can be any of the valid DynamoDB types:
  # https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_AttributeValue.html
  #
  # TODO (design): It might be more user-friendly to define keys as block args
  #
  # Block example:
  #
  #   key {
  #     name  = "ExampleStringAttribute"
  #     type  = "S"
  #     value = "abcdefg"
  #   }
  #
  #   key {
  #     name  = "ExampleNumberAttibute"
  #     type  = "N"
  #     value = 12345
  #   }
  #
  # The nested attribute types/values can be any of the following types:
  #
  # - string
  # - number
  # - set of strings
  # - set of numbers
  # - map of AttributeValue objects (double-nested attributes can use the same
  #   types that are available at the top-level, i.e., any in this list:
  #   map(str, map(str, map(str, any)))
  # - null
  #
  # Type: map(str, map(str, any))
  # Required: true
  key = {
    "ExampleStringAttribute" = {
      "S": "This is a string"
    },
    "ExampleNumberAttribute" = {
      "N": 12345
    }
  }

  # Define what attributes to return from the item(s)
  #
  # Type: str
  # Required: false
  projection_expression = "TopLevelAttribute, FirstAttributeValueFromList[0], NestedAttribute.Example" 

  # Level of detail to return about the capacity consumed by the operation.
  #
  # Type: str
  # Required: false
  return_consumed_capacity = "TOTAL"

  # Obviously, this one defines the table to query. Okay?
  #
  # Type: str
  # Required: true
  table_name = "ExampleTable"
}
