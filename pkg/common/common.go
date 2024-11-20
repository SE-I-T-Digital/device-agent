package common

import (
	"io"
	"os"
)

const Basedir = "/home/intdigital/dev/margo"
const DeploymentDir = Basedir + "/deployments"
const DataDir = Basedir + "/data"
const AppData = Basedir + "/app"               // [base location]/app/[component name]/[files]
const SharedVolume = "/var/lib/docker/volumes" // /var/lib/docker/volumes/${COMPOSE_PROJECT_NAME}_app/_data

// Initialize creates the directories if they do not exist
func Initialize() {
	_ = os.MkdirAll(DeploymentDir, 0755)
	_ = os.MkdirAll(DataDir, 0755)
	_ = os.MkdirAll(SharedVolume, 0755)
}

// CopyFile copies a file from src to dst.
// If dst does not exist, it will be created.
func CopyFile(src, dst string) error {
	// Open the source file
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Create the destination file
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Copy the contents of the source file to the destination file
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// Sync to ensure data is flushed to disk
	err = destFile.Sync()
	return err
}

func GetDirectories(path string) ([]string, error) {
	var directories []string
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			directories = append(directories, entry.Name())
		}
	}
	return directories, nil
}

func DeleteDirectory(path string) error {
	err := os.RemoveAll(path)
	if err != nil {
		return err
	}
	return nil
}
