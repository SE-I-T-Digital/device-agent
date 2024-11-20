package processor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/SE-I-T-Digital/device-agent/pkg/common"
	"github.com/SE-I-T-Digital/device-agent/pkg/downloader"
	"github.com/SE-I-T-Digital/device-agent/pkg/files_processor"
)

type Processor struct {
	InitialFiles map[string][32]byte
}

// deployCompose deploys the docker-compose file
func (main_process Processor) deployCompose(docker_compose_path string, app_name string) error {
	cmd := exec.Command("docker-compose", "-f", docker_compose_path, "-p", app_name, "up", "-d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// removeCompose removes the docker-compose deployment
func (main_process Processor) removeCompose(docker_compose_path string) error {
	cmd := exec.Command("docker-compose", "-f", docker_compose_path, "down", "-v")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// writeEnvFile writes the environment variables to a .env file
func (main_process Processor) writeEnvFile(file string, appname string, data map[string]string) map[string]string {
	_ = os.MkdirAll(common.DataDir+"/"+file+"/"+appname, 0755)
	f, err := os.Create(common.DataDir + "/" + file + "/" + appname + "/.env")
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
	defer f.Close()

	filesToCopy := make(map[string]string)
	for key, value := range data {
		if key == "==FILE==" {
			filesToCopy[common.AppData+"/"+appname+"/"+value] = common.SharedVolume + "/" + appname + "_app/_data/" + value
			continue
		}
		_, err := f.WriteString(key + "=" + value + "\n")
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		}
	}
	return filesToCopy
}

// findYamlFilesInDir finds all the YAML files in a directory
func (main_process Processor) findYamlFilesInDir(path string) []string {
	files, err := os.ReadDir(path)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}

	var yamlFiles []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		match, err := filepath.Match("*.yaml", file.Name())
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		}
		if match {
			yamlFiles = append(yamlFiles, file.Name())
		}
		match, err = filepath.Match("*.yml", file.Name())
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		}
		if match {
			yamlFiles = append(yamlFiles, file.Name())
		}
	}
	return yamlFiles
}

// ProcessNewRequest processes the new request Install/Uninstall/Update of the applications
func (main_process Processor) ProcessNewRequest(deploymentData map[string]map[string]files_processor.ApplicationData) {
	for filename, data := range deploymentData {
		for appName, appData := range data {
			if appData.Operation == "uninstall" {
				fmt.Printf("Uninstalling: %s\n", filename)
				// find all directories in the file directory
				directories, err := common.GetDirectories(common.DataDir + "/" + filename)
				if err != nil {
					fmt.Printf("Error: %s\n", err)
				}
				for _, directory := range directories {
					yamlFiles := main_process.findYamlFilesInDir(common.DataDir + "/" + filename + "/" + directory + "/")
					if len(yamlFiles) != 1 {
						fmt.Printf("Error: Expected 1 YAML filename, found %d\n", len(yamlFiles))
						continue
					}
					main_process.removeCompose(common.DataDir + "/" + filename + "/" + directory + "/" + yamlFiles[0])
					common.DeleteDirectory(common.DataDir + "/" + filename + "/" + directory)
				}
				os.RemoveAll(common.DataDir + "/" + filename)
				fmt.Println("----------------")
			} else if appData.Operation == "install" {
				fmt.Printf("Installing: %s -> %s\n", filename, appName)
				os.MkdirAll(common.SharedVolume+"/"+appName+"_app/_data/", 0755)
				os.MkdirAll(common.DataDir+"/"+filename+"/"+appName, 0755)
				downloader.Download(appData.PackageLocation, common.DataDir+"/"+filename+"/"+appName)
				filesToCopy := main_process.writeEnvFile(filename, appName, appData.Variables)
				downloadedFilename := downloader.ExtractFilename(appData.PackageLocation)
				main_process.deployCompose(common.DataDir+"/"+filename+"/"+appName+"/"+downloadedFilename, appName)
				for src, dst := range filesToCopy {
					err := common.CopyFile(src, dst)
					if err != nil {
						fmt.Printf("Error: %s\n", err)
					}
				}
				fmt.Println("----------------")
			} else if appData.Operation == "update" {
				fmt.Printf("Updating: %s -> %s\n", filename, appName)
				downloader.Download(appData.PackageLocation, common.DataDir+"/"+filename+"/"+appName)
				filesToCopy := main_process.writeEnvFile(filename, appName, appData.Variables)
				downloadedFilename := downloader.ExtractFilename(appData.PackageLocation)
				main_process.deployCompose(common.DataDir+"/"+filename+"/"+appName+"/"+downloadedFilename, appName)
				for src, dst := range filesToCopy {
					err := common.CopyFile(src, dst)
					if err != nil {
						fmt.Printf("Error: %s\n", err)
					}
				}
				fmt.Println("----------------")
			} else {
				fmt.Printf("Operation not supported: %s\n", appData.Operation)
				continue
			}
		}
	}
}
