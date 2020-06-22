package check

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/bmatcuk/doublestar"
)

const (
	DocumentationGlobPattern = `{docs/index.md,docs/{data-sources,guides,resources},website/docs}/**/*`

	LegacyIndexDirectory       = `website/docs`
	LegacyDataSourcesDirectory = `website/docs/d`
	LegacyGuidesDirectory      = `website/docs/guides`
	LegacyResourcesDirectory   = `website/docs/r`

	RegistryIndexDirectory       = `docs`
	RegistryDataSourcesDirectory = `docs/data-sources`
	RegistryGuidesDirectory      = `docs/guides`
	RegistryResourcesDirectory   = `docs/resources`
)

var ValidLegacyDirectories = []string{
	LegacyIndexDirectory,
	LegacyDataSourcesDirectory,
	LegacyGuidesDirectory,
	LegacyResourcesDirectory,
}

var ValidRegistryDirectories = []string{
	RegistryIndexDirectory,
	RegistryDataSourcesDirectory,
	RegistryGuidesDirectory,
	RegistryResourcesDirectory,
}

func InvalidDirectoriesCheck(directories map[string][]string) error {
	for directory := range directories {
		if IsValidRegistryDirectory(directory) {
			continue
		}

		if IsValidLegacyDirectory(directory) {
			continue
		}

		return fmt.Errorf("invalid Terraform Provider documentation directory found: %s", directory)
	}

	return nil
}

func MixedDirectoriesCheck(directories map[string][]string) error {
	var legacyDirectoryFound bool
	var registryDirectoryFound bool
	err := fmt.Errorf("mixed Terraform Provider documentation directory layouts found, must use only legacy or registry layout")

	for directory := range directories {
		// Allow docs/ with other files
		if IsValidRegistryDirectory(directory) && directory != RegistryIndexDirectory {
			registryDirectoryFound = true

			if legacyDirectoryFound {
				return err
			}
		}

		if IsValidLegacyDirectory(directory) {
			legacyDirectoryFound = true

			if registryDirectoryFound {
				return err
			}
		}
	}

	return nil
}

// NumberOfFilesCheck verifies that documentation is below the Terraform Registry storage limit.
// This check presumes that all provided directories are valid, e.g. that directory checking
// for invalid or mixed directory structures was previously completed.
func NumberOfFilesCheck(directories map[string][]string) error {
	var numberOfFiles int

	for directory, files := range directories {
		directoryNumberOfFiles := len(files)
		log.Printf("[TRACE] Found %d documentation files in directory: %s", directoryNumberOfFiles, directory)
		numberOfFiles = numberOfFiles + directoryNumberOfFiles
	}

	log.Printf("[DEBUG] Found %d documentation files with limit of %d", numberOfFiles, RegistryMaximumNumberOfFiles)
	if numberOfFiles >= RegistryMaximumNumberOfFiles {
		return fmt.Errorf("exceeded maximum (%d) number of documentation files for Terraform Registry: %d", RegistryMaximumNumberOfFiles, numberOfFiles)
	}

	return nil
}

func GetDirectories(basepath string) (map[string][]string, error) {
	globPattern := DocumentationGlobPattern

	if basepath != "" {
		globPattern = fmt.Sprintf("%s/%s", basepath, globPattern)
	}

	files, err := doublestar.Glob(globPattern)

	if err != nil {
		return nil, fmt.Errorf("error globbing Terraform Provider documentation directories: %w", err)
	}

	if basepath != "" {
		for index, file := range files {
			files[index], _ = filepath.Rel(basepath, file)
		}
	}

	directories := make(map[string][]string)

	for _, file := range files {
		// Simple skip of glob matches that are known directories
		if IsValidRegistryDirectory(file) || IsValidLegacyDirectory(file) {
			continue
		}

		directory := filepath.Dir(file)

		// Skip handling of docs/ files except index.md
		// if directory == RegistryIndexDirectory && filepath.Base(file) != "index.md" {
		// 	continue
		// }

		// Skip handling of docs/** outside valid Registry Directories

		directories[directory] = append(directories[directory], file)
	}

	return directories, nil
}

func IsValidLegacyDirectory(directory string) bool {
	for _, validLegacyDirectory := range ValidLegacyDirectories {
		if directory == validLegacyDirectory {
			return true
		}
	}

	return false
}

func IsValidRegistryDirectory(directory string) bool {
	for _, validRegistryDirectory := range ValidRegistryDirectories {
		if directory == validRegistryDirectory {
			return true
		}
	}

	return false
}
