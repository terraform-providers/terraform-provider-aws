# tagresource

The `tagresource` package is designed to provide a generator and consistent interface for Terraform AWS Provider resources that handle individual resource tags. Most of the heavy lifting is done by the `keyvaluetags` package to smooth over inconsistencies across AWS service APIs, but this generator does implement some final user experience improvements.

## Code Structure

```text
aws/internal/tagresource
├── generator/main.go (generates tag resource)
├── tag_resources_test.go (unit tests for shared tag resource logic)
├── tag_resources.go (shared tag resource logic)
└── service_generation_customizations.go (AWS Go SDK service customizations for generator)
```
