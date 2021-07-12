//go:generate go run -tags generate internal/tagresource/generator/main.go -servicename dynamodb
//go:generate go run -tags generate internal/tagresource/generator/main.go -servicename ec2
//go:generate go run -tags generate internal/tagresource/generator/main.go -servicename ecs

package aws
