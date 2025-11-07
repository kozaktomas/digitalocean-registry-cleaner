package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dorc",
	Short: "DigitalOcean Registry Cleaner",
	Long:  `A CLI tool to clean up unused images in DigitalOcean Container Registry.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println("There was an error:", err)
	}
}

func init() {
	rootCmd.AddCommand(runCmd)
}
