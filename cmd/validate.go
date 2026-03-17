/*
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wso2/arazzo-mcp-gen/internal/generator"
	"github.com/wso2/arazzo-mcp-gen/internal/validator"
)

const ValidateCmdExample = `# Validate an Arazzo spec folder (auto-detects the Arazzo file)
arazzo-mcp-gen validate -d ./my-arazzo-folder

# Validate a single Arazzo file
arazzo-mcp-gen validate -f ./workflow.yaml

# Validate and also check that remote source URLs are accessible
arazzo-mcp-gen validate -d ./my-arazzo-folder --check-remote

# Treat warnings as errors (useful for CI pipelines)
arazzo-mcp-gen validate -d ./my-arazzo-folder --strict`

var (
	validateFolder      string
	validateFile        string
	validateCheckRemote bool
	validateStrict      bool
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate an Arazzo specification file",
	Long: `Validate an Arazzo specification file for correctness and completeness.

Performs comprehensive checks including:
  - YAML syntax validation
  - Arazzo version compatibility (1.0.x)
  - Required fields (info.title, info.version, sourceDescriptions, workflows)
  - Source description completeness and accessibility
  - Workflow structure (workflowId uniqueness, inputs, outputs)
  - Step structure (stepId uniqueness, operationId/workflowId, parameters)
  - Success criteria expression syntax
  - onSuccess/onFailure action validation (goto targets, end, retry)
  - Runtime expression validation ($statusCode, $inputs, $steps, etc.)
  - Cross-reference checks (goto targets, output references)

Use --check-remote to also verify that remote source URLs are accessible.
Use --strict to treat warnings as errors (non-zero exit code on warnings).`,
	Example: ValidateCmdExample,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runValidateCommand(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	validateCmd.Flags().StringVarP(&validateFolder, "folder", "d", "",
		"Path to folder containing Arazzo and OpenAPI spec files")
	validateCmd.Flags().StringVarP(&validateFile, "file", "f", "",
		"Path to a single Arazzo specification file")
	validateCmd.Flags().BoolVar(&validateCheckRemote, "check-remote", false,
		"Check that remote source description URLs are accessible")
	validateCmd.Flags().BoolVar(&validateStrict, "strict", false,
		"Treat warnings as errors (exit code 1 on warnings)")

	rootCmd.AddCommand(validateCmd)
}

func runValidateCommand() error {
	// Determine what to validate
	if validateFolder == "" && validateFile == "" {
		return fmt.Errorf("either --folder (-d) or --file (-f) must be specified\n\n" +
			"Examples:\n" +
			"  arazzo-mcp-gen validate -d ./my-arazzo-folder\n" +
			"  arazzo-mcp-gen validate -f ./workflow.yaml")
	}

	if validateFolder != "" && validateFile != "" {
		return fmt.Errorf("cannot use both --folder (-d) and --file (-f) at the same time")
	}

	var filePath string
	var folderPath string

	if validateFile != "" {
		// Validate a single file
		absFile, err := filepath.Abs(validateFile)
		if err != nil {
			return fmt.Errorf("failed to resolve file path: %w", err)
		}
		if _, err := os.Stat(absFile); os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %s", absFile)
		}
		filePath = absFile
		folderPath = filepath.Dir(absFile)
	} else {
		// Find Arazzo file in folder
		absFolder, err := filepath.Abs(validateFolder)
		if err != nil {
			return fmt.Errorf("failed to resolve folder path: %w", err)
		}
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

		found, err := generator.FindArazzoFile(absFolder)
		if err != nil {
			return err
		}
		filePath = found
		folderPath = absFolder
	}

	// Run validation
	result := validator.ValidateFile(filePath, folderPath, validateCheckRemote)
	result.PrintReport()

	// Determine exit code
	if result.HasErrors() {
		os.Exit(1)
	}
	if validateStrict && result.WarningCount() > 0 {
		fmt.Printf("\n%s--strict mode: %d warning(s) treated as errors%s\n",
			"\033[33m", result.WarningCount(), "\033[0m")
		os.Exit(1)
	}

	// Print helpful suggestion if validation passed
	if !result.HasErrors() {
		fmt.Println("💡 Tip: Run 'arazzo-mcp-gen mcp-server generate -d " +
			formatFolderHint(folderPath) + "' to build an MCP server from this spec.")
	}

	return nil
}

// formatFolderHint returns a relative path if short, otherwise the absolute path
func formatFolderHint(absPath string) string {
	cwd, err := os.Getwd()
	if err != nil {
		return absPath
	}
	rel, err := filepath.Rel(cwd, absPath)
	if err != nil || strings.HasPrefix(rel, ".."+string(filepath.Separator)+".."+string(filepath.Separator)+"..") {
		return absPath
	}
	return "./" + filepath.ToSlash(rel)
}
