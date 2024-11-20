package files_processor

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ApplicationData represents the application data section of the YAML file
type ApplicationData struct {
	Operation       string            `yaml:"operation"`
	Type            string            `yaml:"type"`
	PackageLocation string            `yaml:"packageLocation"`
	Revision        string            `yaml:"revision"`
	Variables       map[string]string `yaml:"variables"` // Pointer => Value
}

// Metadata represents the metadata section of the YAML file
type Metadata struct {
	Annotations map[string]string `yaml:"annotations"`
	Name        string            `yaml:"name"`
	Namespace   string            `yaml:"namespace"`
}

// Properties represents the properties of a component
type Properties struct {
	PackageLocation string `yaml:"packageLocation"`
	Revision        string `yaml:"revision"`
}

// Component represents a component in the deployment profile
type Component struct {
	Name       string     `yaml:"name"`
	Properties Properties `yaml:"properties"`
}

// DeploymentProfile represents the deployment profile section
type DeploymentProfile struct {
	Type       string      `yaml:"type"`
	Components []Component `yaml:"components"`
}

// Target represents a target in the parameters section
type Target struct {
	Pointer    string   `yaml:"pointer"`
	Components []string `yaml:"components"`
}

// Parameter represents a parameter in the parameters section
type Parameter struct {
	Value   string   `yaml:"value"`
	Targets []Target `yaml:"targets"`
}

// Spec represents the spec section of the YAML file
type Spec struct {
	DeploymentProfile DeploymentProfile    `yaml:"deploymentProfile"`
	Parameters        map[string]Parameter `yaml:"parameters"`
}

// ApplicationDeployment represents the entire YAML structure
type ApplicationDeployment struct {
	APIVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
	Metadata   Metadata `yaml:"metadata"`
	Spec       Spec     `yaml:"spec"`
}

type Files struct {
	InitialFiles map[string][32]byte
}

// listFilesWithHash lists all files in the directory and calculates their SHA-256 hash
func (FileHasher Files) listFilesWithHash(dir string) map[string][32]byte {
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	currentFiles := make(map[string][32]byte)
	for _, file := range files {
		if !file.IsDir() {
			hash, err := FileHasher.calculateFileHash(filepath.Join(dir, file.Name()))
			if err != nil {
				log.Fatal(err)
			}
			currentFiles[file.Name()] = hash
		}
	}

	return currentFiles
}

// calculateFileHash calculates the SHA-256 hash of a file
func (FileHasher Files) calculateFileHash(filePath string) ([32]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return [32]byte{}, err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return [32]byte{}, err
	}

	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result, nil
}

// remove old entries from map one if not exist in map two
func (FileHasher Files) removeOldEntries(mapOne map[string][32]byte, mapTwo map[string][32]byte, filesData map[string]map[string]ApplicationData) {
	for key, value := range mapOne {
		if _, ok := mapTwo[key]; !ok {
			delete(mapOne, key)
			filesData[key] = make(map[string]ApplicationData)
			var appData ApplicationData
			appData.Operation = "uninstall"
			appData.PackageLocation = "uninstall"
			appData.Revision = "uninstall"
			appData.Type = "uninstall"
			appData.Variables = make(map[string]string)
			filesData[key]["uninstall"] = appData
			fmt.Printf("Uninstall: %s, SHA-256: %x\n", key, value)
		}
	}
}

// Read and parse the yaml application deployment definition file
func (FileHasher Files) readApplicationDeploymentDefinitionFile(filePath string) (*ApplicationDeployment, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var config ApplicationDeployment
	err = yaml.Unmarshal(content, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// checkNewFiles checks if new files appear in the directory
func (FileHasher Files) CheckNewFiles(dir string) (map[string]map[string]ApplicationData, error) {
	// Create map with all files in the directory and get their SHA-256 hash
	currentFiles := FileHasher.listFilesWithHash(dir)

	// Remove old entries from the initial files map
	filesData := make(map[string]map[string]ApplicationData)
	FileHasher.removeOldEntries(FileHasher.InitialFiles, currentFiles, filesData)

	// iterate over the current files and check if they are new
	for filename, sha256 := range currentFiles {
		var operation string
		if _, ok := FileHasher.InitialFiles[filename]; !ok {
			// New file detected
			operation = "install"
			fmt.Printf("Install: %s, SHA-256: %x\n", filename, sha256)

		} else {
			if FileHasher.InitialFiles[filename] != sha256 {
				// File has been updated
				operation = "update"
				fmt.Printf("Update: %s, SHA-256: %x\n", filename, sha256)
			} else {
				// File not changed
				continue
			}
		}

		// Store the new file in the initial files map
		FileHasher.InitialFiles[filename] = sha256

		// Read and parse the application deployment definition file
		appConfig, err := FileHasher.readApplicationDeploymentDefinitionFile(filepath.Join(dir, filename))
		if err != nil {
			return nil, err

		}

		// Check if the deployment profile type is docker-compose
		if appConfig.Spec.DeploymentProfile.Type != "docker-compose" {
			break
		}

		// Get the package location and component name for each component in the deployment profile
		filesData[filename] = make(map[string]ApplicationData)
		for _, component := range appConfig.Spec.DeploymentProfile.Components {
			var appData ApplicationData
			appData.Operation = operation
			appData.PackageLocation = component.Properties.PackageLocation
			appData.Revision = component.Properties.Revision
			appData.Type = appConfig.Spec.DeploymentProfile.Type
			appData.Variables = make(map[string]string)
			filesData[filename][component.Name] = appData
		}

		// get the variables for each component
		for _, variable := range appConfig.Spec.Parameters {
			for _, component := range variable.Targets[0].Components {
				filesData[filename][component].Variables[variable.Targets[0].Pointer] = variable.Value
			}
		}

	}
	return filesData, nil
}
