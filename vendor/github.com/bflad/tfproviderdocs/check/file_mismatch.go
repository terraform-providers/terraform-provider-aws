package check

import (
	"fmt"
	"sort"

	"github.com/hashicorp/go-multierror"
	tfjson "github.com/hashicorp/terraform-json"
)

func ResourceFileMismatchCheck(providerName string, resourceType string, schemaResources map[string]*tfjson.Schema, files []string) error {
	var extraFiles []string
	var missingFiles []string

	for _, file := range files {
		if fileHasResource(schemaResources, providerName, file) {
			continue
		}

		extraFiles = append(extraFiles, file)
	}

	for _, resourceName := range resourceNames(schemaResources) {
		if resourceHasFile(files, providerName, resourceName) {
			continue
		}

		missingFiles = append(missingFiles, resourceName)
	}

	var result *multierror.Error

	for _, extraFile := range extraFiles {
		err := fmt.Errorf("matching %s for documentation file (%s) not found, file is extraneous or incorrectly named", resourceType, extraFile)
		result = multierror.Append(result, err)
	}

	for _, missingFile := range missingFiles {
		err := fmt.Errorf("missing documentation file for %s: %s", resourceType, missingFile)
		result = multierror.Append(result, err)
	}

	return result.ErrorOrNil()
}

func fileHasResource(schemaResources map[string]*tfjson.Schema, providerName, file string) bool {
	if _, ok := schemaResources[fileResourceName(providerName, file)]; ok {
		return true
	}

	return false
}

func fileResourceName(providerName, fileName string) string {
	resourceSuffix := TrimFileExtension(fileName)

	return fmt.Sprintf("%s_%s", providerName, resourceSuffix)
}

func resourceHasFile(files []string, providerName, resourceName string) bool {
	var found bool

	for _, file := range files {
		if fileResourceName(providerName, file) == resourceName {
			found = true
			break
		}
	}

	return found
}

func resourceNames(resources map[string]*tfjson.Schema) []string {
	names := make([]string, 0, len(resources))

	for name := range resources {
		names = append(names, name)
	}

	sort.Strings(names)

	return names
}
