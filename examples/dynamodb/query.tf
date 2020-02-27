# =========================================================================== #
#
# Proposed arguments for DynamoDB query data source, using an HCL example.
#
# Status: Initial Design
#
# --------------------------------------------------------------------------- #
#
# We would want to stick pretty closely to DynamoDB's Query API for this:
# https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_Query.html
#
# =========================================================================== #
#
#
#
#
data "aws_dynamodb_query" "example" {
  # Use strong consistentcy instead of eventual if true. Won't work with global
  # tables.
  #
  # Type: bool
  # Required: false
  consistent_read = false
  
  # Specify what primary key to start reading data from. Can be used to process
  # paginated batches of items.
  #
  # API takes an AttributeValue mapping (str -> Any):
  # https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_AttributeValue.html
  #
  # TODO (design): Think more about if this would be more user friendly to
  # accept block or map instead of an escaped JSON string. The API just takes a
  # string. But we need to escape any quotes...
  #
  # This way passes the JSON as a string with all the quotes escaped:
  # 
  # Type: str, number, or bool
  # Required: false
  exclusive_start_key = "{\"S\": \"example-hash-key-12345\"}"
  #
  # (continued from above)
  #
  # Heredoc syntax might let us remove the quote escaping (I think?)
  # exclusive_start_key = <<-EOT
  #   {"S": "example-hash-key-12345"}
  #   EOT
  #
  # This does more magic than passing the start key as just a string, but may
  # be a cleaner interface if designed right. I'm not sure what that right way
  # is. Leaning towards just using a string to keep it simple. But doing it
  # this way might enable us to do intelligent type checking before sending
  # queries.
  #
  # exclusive_start_key {
  #   type   = "S"
  #   values = ["example-hash-key-12345"]
  # }
  
  # Used for substituting attribute names in expressions, e.g., when an
  # attribute name conflicts with a DynamoDB reserved keyword. The API accepts
  # a map(str) that maps alias->attribute.
  #
  # Type: map(str)
  # Required: false
  expression_attribute_names = {
    "#A": "example-attribute-A"
    "#B": "example-attribute-B"
  }

  # Create aliases for attributes that can be evaluated in an expression. The
  # API takes a map of attr values like the "exclusive_start_key" arg.
  #
  # Type: map(str)
  # Required: false
  expression_attribute_values = {
    # A colon is required for the API and the value is referenced in
    # expressions the same way. Maybe we can strip it from the key?
    ":avail" = "{\"S\":\"Available\"}"
    ":back"  = "{\"S\":\"Backordered\"}"
  }

  # Filter expression is conditional expression that is applied AFTER the query
  # is performed, but before data is returned. Use the key_condition_expression
  # arg to filter items at query time.
  #
  # https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Query.html#Query.FilterExpression
  #
  # Type: str
  # Required: false
  filter_expression = "Price > :limit"

  # The name of the local or global secondary index to query.
  #   
  # Type: str
  # Required: false
  index_name = "ExampleIndex"

  # key_condition_expression is a condition expression that DynamoDB uses to
  # filter items at query time.
  #
  # https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Query.html#Query.KeyConditionExpressions
  # https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Expressions.ConditionExpressions.html
  #
  # Type: str
  # Required: false
  key_condition_expression = "attribute_not_exists(ExampleAttribute)"

  # Limit the number of items returned in a query.
  #
  # Type: int
  # Required: false
  limit = 100 

  # An expression that defines what attributes to return from the item(s).
  #
  # https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Expressions.ProjectionExpressions.html
  # https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Expressions.Attributes.html
  #
  # TODO (design): It might be nice for this to be a list(str) type that we
  # join together into a string.
  #
  # Type: str
  # Required: false
  projection_expression = "Description, RelatedItems[0], ProductReviews.FiveStar"

  # Level of detail about consumed capacity to return in the response. Can be
  # Valid values: INDEX | TOTAL | NONE
  # 
  # 
  # TODO: The docs don't list the default for this parameter. Figure it out.
  # https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_Query.html#API_Query_RequestSyntax
  #
  # Type: bool
  # Required: false
  # Default: ?
  return_consumed_capacity = "INDEXES"

  # Direction to traverse the index. Default is true, which means ascending
  # order, so false scans in descending order.
  #
  # Type: bool
  # Required: false
  # Default: true
  scan_index_forward = true

  # Configure the attributes to return in the response. When using a projection
  # expression, must be set to SPECIFIC_ATTRIBUTES
  #
  # Type: str
  # Required: false
  # Valid Values: ALL_ATTRIBUTES | ALL_PROJECTED_ATTRIBUTES | SPECIFIC_ATTRIBUTES | COUNT
  # Default: ALL_ATTRIBUTES
  select = "SPECIFIC_ATTRIBUTES"

  # Name of the table to query.
  #
  # Type: str
  # Required: true
  table_name = "ExampleTable"
}
