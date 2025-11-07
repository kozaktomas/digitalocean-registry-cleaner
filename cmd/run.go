package cmd

import (
	"fmt"
	"os"
	"time"

	"digitalocean-registry-cleaner/pkg/do"

	"github.com/spf13/cobra"
)

var (
	registry     string
	repositories []string
	protected    []string

	keepTags   int
	minAgeDays int

	dryRun bool

	protectedDefault = []string{
		"latest",
		"main",
		"master",
		"prod",
		"production",
	}
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run Cleaner",
	Long:  `Command deletes tags older than [min-age-days] in the registry except the last [keep-tags] tags per repository.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		token := os.Getenv("DO_TOKEN")
		if token == "" {
			return fmt.Errorf("DO_TOKEN is not set")
		}

		doc := do.NewClient(
			token,
			protected,
		)

		if dryRun {
			fmt.Print("==> Dry run mode\n\n")
		}

		for _, repository := range repositories {
			deleted, err := doc.RunCleanup(do.CleanupInput{
				Registry:   registry,
				Repository: repository,
				DryRun:     dryRun,
				KeepTags:   keepTags,
				MinAge:     time.Duration(minAgeDays) * 24 * time.Hour,
			})

			if len(deleted) > 0 {
				fmt.Println(fmt.Sprintf("Registry: %s", registry))
				fmt.Println(fmt.Sprintf("Repository: %s\n", repository))

				for _, tag := range deleted {
					fmt.Printf("Deleted tag: %s\t%s\n", tag.Tag, tag.UpdatedAt.Format(time.RFC3339))
				}
				fmt.Println("=====")
			}

			if err != nil {
				return fmt.Errorf("cleanup failed: %w", err)
			}
		}

		return nil
	},
}

func init() {
	runCmd.Flags().StringVar(&registry, "registry", "", "Registry name")
	runCmd.Flags().StringArrayVar(&repositories, "repository", []string{}, "Repository name")
	runCmd.Flags().StringArrayVar(&protected, "protect", protectedDefault, "Protect tag/branch")
	runCmd.Flags().IntVar(&keepTags, "keep-tags", 5, "How many tags to keep per repository")
	runCmd.Flags().IntVar(&minAgeDays, "min-age-days", 30, "Minimum age of the tags to delete in days")
	runCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Dry run")

	_ = runCmd.MarkFlagRequired("registry")
	_ = runCmd.MarkFlagRequired("repository")
}
