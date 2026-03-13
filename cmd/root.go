package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "arazzo-mcp-gen",
	Short: "Generate MCP servers from Arazzo specifications",
	Long:  `arazzo-mcp-gen is a standalone CLI tool for generating Dockerized Python MCP servers directly from an Arazzo specification and its referenced OpenAPI spec files.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
