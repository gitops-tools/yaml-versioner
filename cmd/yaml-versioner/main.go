package main

import (
	"fmt"

	"github.com/gitops-tools/yaml-versioner/pkg/versioner"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "yaml-versioner",
		Short: "Tooling for manipulating versions in YAML files",
	}

	rootCmd.PersistentFlags().String("filename", "", "Full path of file to modify")
	rootCmd.MarkFlagRequired("filename")
	rootCmd.AddCommand(newIncrementCommand())

	cobra.CheckErr(rootCmd.Execute())
}

func newIncrementCommand() *cobra.Command {
	var (
		key        string
		patch      bool
		minor      bool
		major      bool
		preRelease string
	)

	cmd := &cobra.Command{
		Use:   "increment",
		Short: "Increment a SemVer in a YAML file",
		RunE: func(cmd *cobra.Command, args []string) error {
			filename, err := cmd.Flags().GetString("filename")
			if err != nil {
				return fmt.Errorf("getting filename to change: %s", err)
			}

			return versioner.IncrementVersion(filename, key, versioner.VersionOptions{
				Patch:      patch,
				Minor:      minor,
				Major:      major,
				PreRelease: preRelease,
			})
		},
	}

	cmd.Flags().StringVar(&key, "key", "", "key to modify e.g. version or capi.version")
	cmd.MarkFlagRequired("key")
	cmd.Flags().BoolVar(&patch, "patch", false, "increment the patch version within the semver e.g. 1.0.x")
	cmd.Flags().BoolVar(&minor, "minor", false, "increment the minor version within the semver e.g. 1.x.0")
	cmd.Flags().BoolVar(&major, "major", false, "increment the major version within the semver e.g. x.0.0")
	cmd.Flags().StringVar(&preRelease, "pre-release", "", "set pre-release metadata e.g. alpha.1, beta.2")
	cmd.MarkFlagsMutuallyExclusive("patch", "minor", "major")

	return cmd
}
