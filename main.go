package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/spf13/viper"
)

func main() {

	// read values from configuration file
	configPath := os.Getenv("BACKUP_CONFIG_PATH")
	if configPath == "" {
		log.Fatalf("missing BACKUP_CONFIG_PATH environment variable")
		os.Exit(1)
	}
	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Failed to read configuration file: %v\n", err)
		os.Exit(1)
	}

	// replace these values with your own directory path and server details
	dirPath := viper.GetString("dirPath")
	server := viper.GetString("server")
	backupType := viper.GetString("backupType")

	archiveFileName := viper.GetString("archiveFileName") + "_" +
		time.Now().Format("2006-01-02_15-04") + ".tar.gz"

	// create a tar archive of the directory
	tarCmd := exec.Command("tar", "-czf", archiveFileName, dirPath)
	if err := tarCmd.Run(); err != nil {
		fmt.Printf("Failed to create tar archive: %v\n", err)
		os.Exit(1)
	}

	if backupType == "scp" {
		// copy the archive to the server via scp
		if err := copyArchiveToServer(server, archiveFileName); err != nil {
			fmt.Printf("Failed to copy archive to server: %v\n", err)
			removeArchive(archiveFileName)
			os.Exit(1)
		}
	} else if backupType == "rclone" {
		// copy the archive to the new server via rclone
		rcloneCmd := exec.Command("rclone", "copy", archiveFileName, server)
		if err := rcloneCmd.Run(); err != nil {
			fmt.Printf("Failed to copy archive to object storage: %v\n", err)
			removeArchive(archiveFileName)
			os.Exit(1)
		}
	} else {
		fmt.Printf("Invalid backup type: %v\n, can be scp or rclone", backupType)
		removeArchive(archiveFileName)
		os.Exit(1)
	}

	// delete the local archive file
	if err := removeArchive(archiveFileName); err != nil {
		fmt.Printf("Failed to delete local archive file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Archive created and copied to new server successfully.")
}

func copyArchiveToServer(server string, file string) error {
	scpCmd := exec.Command("scp", file, server)
	if err := scpCmd.Run(); err != nil {
		return fmt.Errorf("failed to copy archive to new server: %v", err)
	}
	return nil
}

func removeArchive(file string) error {
	if err := os.Remove(file); err != nil {
		return fmt.Errorf("failed to delete %v: %v", file, err)
	}
	return nil
}
