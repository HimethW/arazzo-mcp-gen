package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/wso2/arazzo-mcp-gen/internal/generator"
)

const GenerateCmdExample = `# Generate an MCP server Docker image from an Arazzo spec folder
arazzo-mcp-gen mcp-server generate -d ./my-arazzo-folder

# Generate with a custom port
arazzo-mcp-gen mcp-server generate -d ./my-arazzo-folder -p 8080

# Generate and save build artifacts to a directory for inspection or manual editing
arazzo-mcp-gen mcp-server generate -d ./my-arazzo-folder --output-dir ./my-output`

var (
	generateFolder    string
	generatePort      int
	generateOutputDir string
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate an MCP server Docker image from an Arazzo specification",
	Long: `Generate a Docker image containing a Python MCP server from an Arazzo specification.

The command reads an Arazzo file and its referenced OpenAPI spec files from the
provided folder, generates a Python MCP server that exposes each workflow as an
MCP tool, and builds a Docker image ready to run.

Input Folder Requirements:
  - Must contain exactly one Arazzo specification file (.yaml or .yml)
  - All OpenAPI files referenced in sourceDescriptions must be present
  - The Arazzo file must have a valid 'arazzo' version key, 'info.title',
    and at least one workflow defined

Flags:
  -d, --folder string       (required) Path to folder containing the Arazzo and
                             OpenAPI spec files
  -p, --port int            Port the MCP server will listen on inside the
                             container and mapped to localhost (default: 5000)
      --output-dir string   Directory to save generated build artifacts
                             (Dockerfile, mcp_server.py, arazzo specs). Files
                             persist after the build for inspection or manual
                             editing. If not set, a temporary directory is used
                             and cleaned up automatically.`,
	Example: GenerateCmdExample,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runGenerateCommand(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	generateCmd.Flags().StringVarP(&generateFolder, "folder", "d", "", "Path to folder containing Arazzo and OpenAPI spec files (required)")
	generateCmd.Flags().IntVarP(&generatePort, "port", "p", 5000, "Port the MCP server will listen on")
	generateCmd.Flags().StringVar(&generateOutputDir, "output-dir", "", "Output directory to save generated files (Dockerfile, server code, specs)")

	generateCmd.MarkFlagRequired("folder")

	mcpServerCmd.AddCommand(generateCmd)
}

func runGenerateCommand() error {
	// Resolve folder to absolute path
	absFolder, err := filepath.Abs(generateFolder)
	if err != nil {
		return fmt.Errorf("failed to resolve folder path: %w", err)
	}

	// Verify folder exists and is a directory
	info, err := os.Stat(absFolder)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("folder does not exist: %s", absFolder)
		}
		return fmt.Errorf("failed to access folder: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", absFolder)
	}

	fmt.Println("Validating input folder...")
	arazzoFilePath, err := generator.FindArazzoFile(absFolder)
	if err != nil {
		return err
	}
	arazzoFileName := filepath.Base(arazzoFilePath)

	spec, err := generator.ParseArazzoFile(arazzoFilePath)
	if err != nil {
		return err
	}

	if err := generator.ValidateSourceDescriptions(spec, absFolder); err != nil {
		return err
	}
	fmt.Printf("Found Arazzo spec: %s with %d workflow(s)\n", spec.Info.Title, len(spec.Workflows))

	fmt.Println("Generating MCP server code...")
	serverCode, err := generator.GenerateServerCode(spec, arazzoFileName, generatePort)
	if err != nil {
		return fmt.Errorf("failed to generate server code: %w", err)
	}

	dockerfileCode := generator.GenerateDockerfile(generatePort)

	fmt.Println("Building Docker image...")
	config := generator.MCPServerBuildConfig{
		FolderPath:     absFolder,
		Port:           generatePort,
		ArazzoSpec:     spec,
		ArazzoFileName: arazzoFileName,
		ServerCode:     serverCode,
		DockerfileCode: dockerfileCode,
		OutputDir:      generateOutputDir,
	}

	if err := generator.BuildMCPServerImage(config); err != nil {
		return err
	}

	return nil
}
