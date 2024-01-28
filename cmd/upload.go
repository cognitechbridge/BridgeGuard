/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/

package cmd

import (
	"ctb-cli/manager"
	"fmt"
	"github.com/spf13/cobra"
	"path/filepath"
)

// uploadCmd represents the upload command
var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Uploads a file to cloud",
	Long:  `Uploads a file to cloud`,
	Run: func(cmd *cobra.Command, args []string) {
		path, _ := cmd.Flags().GetString("path")
		absPath, _ := filepath.Abs(path)

		name, _ := cmd.Flags().GetString("name")
		force, _ := cmd.Flags().GetBool("force")
		recursive, _ := cmd.Flags().GetBool("recursive")

		if !recursive {
			uploader := manager.Client.NewUploader(absPath, name, force)
			res, err := uploader.Upload()
			if err != nil {
				fmt.Printf("Error uploading: %v", err)
				return
			}
			fmt.Printf("Upload completed: %s\n", res)
		} else {
			wk := manager.Client.NewFileWalker(absPath, name, force)
			err := wk.Upload()
			if err != nil {
				fmt.Printf("Error uploading: %v", err)
				return
			}
		}

	},
}

func init() {
	rootCmd.AddCommand(uploadCmd)

	uploadCmd.Flags().StringP("name", "n", "", "name on cloud")
	uploadCmd.Flags().StringP("path", "p", "", "path to file to upload")
	uploadCmd.Flags().BoolP("force", "f", false, "force")
	uploadCmd.Flags().BoolP("recursive", "r", false, "recursive")
	_ = uploadCmd.MarkFlagRequired("path")
	_ = uploadCmd.MarkFlagRequired("file")
}
