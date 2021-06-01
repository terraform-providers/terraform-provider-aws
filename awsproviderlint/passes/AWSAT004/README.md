# AWSAT004

The AWSAT004 analyzer reports TestCheckResourceAttr() calls with hardcoded
TypeSet state hashes. Hardcoded state hashes are an unreliable way to
specifically address state values since hashes may change over time, be
inconsistent across partitions, and can be inadvertently changed by modifying
configurations.

## Flagged Code

```go
func TestAccAWSELB_basic(t *testing.T) {
  ...
	resource.ParallelTest(t, resource.TestCase{
    ...
		Steps: []resource.TestStep{
			{
				Config: testAccAWSELBConfig,
				Check: resource.ComposeTestCheckFunc(
          ...
					resource.TestCheckResourceAttr(resourceName, "listener.206423021.instance_port", "8000"),
					resource.TestCheckResourceAttr(resourceName, "listener.206423021.instance_protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "listener.206423021.lb_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "listener.206423021.lb_protocol", "http"),
          ...
				),
			},
      ...
		},
	})
}
```

## Passing Code

The flagged code above can be replaced with the following passing code:

```go
func TestAccAWSELB_basic(t *testing.T) {
  ...
	resource.ParallelTest(t, resource.TestCase{
    ...
		Steps: []resource.TestStep{
			{
				Config: testAccAWSELBConfig,
				Check: resource.ComposeTestCheckFunc(
          ...
          resource.TestCheckTypeSetElemNestedAttrs(resourceName, "listener.*", map[string]string{
 						"instance_port":     "8000",
            "instance_protocol": "http",
            "lb_port":           "80",
            "lb_protocol":       "http",
 					}),
          ...
				),
			},
      ...
		},
	})
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:AWSAT004` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
resource.TestCheckResourceAttr(resourceName, "listener.206423021.instance_port", "8000"), //lintignore:AWSAT004
```
