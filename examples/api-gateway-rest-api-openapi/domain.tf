#
# Domain Setup
#

resource "aws_api_gateway_domain_name" "example" {
  domain_name              = aws_acm_certificate.example.domain_name
  regional_certificate_arn = aws_acm_certificate.example.arn

  endpoint_configuration {
    types = ["REGIONAL"]
  }
}

resource "aws_api_gateway_base_path_mapping" "example" {
  api_id      = aws_api_gateway_rest_api.example.id
  domain_name = aws_api_gateway_domain_name.example.domain_name
  stage_name  = aws_api_gateway_stage.example.stage_name
}
