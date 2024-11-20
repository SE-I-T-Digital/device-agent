package main

import (
	"fmt"
	"time"

	"github.com/SE-I-T-Digital/device-agent/pkg/common"
	"github.com/SE-I-T-Digital/device-agent/pkg/files_processor"
	"github.com/SE-I-T-Digital/device-agent/pkg/processor"
	"gopkg.in/yaml.v3"
)

// printDeploymentData prints the deployment data in a human-readable format
func printDeploymentData(deploymentData map[string]map[string]files_processor.ApplicationData) {
	b, err := yaml.Marshal(deploymentData)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
	fmt.Println("Deployment Data:")
	fmt.Println(string(b))
}

// main is the entry point of the program
func main() {
	fmt.Println("--------------------")
	fmt.Println("Device Agent Started...")
	fmt.Println("--------------------")

	// Initialize the directories
	common.Initialize()

	var fileHasher files_processor.Files
	fileHasher.InitialFiles = make(map[string][32]byte)
	inputDir := common.DeploymentDir
	for {
		deploymentData, err := fileHasher.CheckNewFiles(inputDir)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		}

		if len(deploymentData) > 0 {
			printDeploymentData(deploymentData)
			var process processor.Processor
			process.ProcessNewRequest(deploymentData)
		}
		time.Sleep(10 * time.Second) // Check every 10 seconds
	}
}
