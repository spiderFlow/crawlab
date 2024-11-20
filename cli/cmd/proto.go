package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var protoCmd = &cobra.Command{
	Use:   "proto",
	Short: "Compile proto files to Python pb files",
	Run:   runProto,
}

var (
	protoDir  string
	outputDir string
)

func init() {
	protoCmd.Flags().StringVarP(&protoDir, "proto-dir", "p", "grpc/proto", "Directory containing proto files")
	protoCmd.Flags().StringVarP(&outputDir, "output-dir", "o", "python/crawlab/grpc", "Output directory for Python pb files")
	RootCmd.AddCommand(protoCmd)
}

func runProto(cmd *cobra.Command, args []string) {
	// Ensure protoc is installed
	if _, err := exec.LookPath("protoc"); err != nil {
		fmt.Println("Error: protoc is not installed. Please install Protocol Buffers compiler first.")
		os.Exit(1)
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Find all proto files
	protoFiles, err := filepath.Glob(filepath.Join(protoDir, "**/*.proto"))
	if err != nil {
		fmt.Printf("Error finding proto files: %v\n", err)
		os.Exit(1)
	}

	for _, protoFile := range protoFiles {
		relPath, _ := filepath.Rel(protoDir, protoFile)
		fmt.Printf("Compiling: %s\n", relPath)

		args := []string{
			"--proto_path=" + protoDir,
			"--python_out=" + outputDir,
			"--grpc_python_out=" + outputDir,
			protoFile,
		}

		cmd := exec.Command("protoc", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Printf("Error compiling %s: %v\n", relPath, err)
			os.Exit(1)
		}

		// Fix Python imports
		fixPythonImports(outputDir)
	}

	fmt.Println("Successfully compiled all proto files to Python")
}

func fixPythonImports(dir string) error {
	pbFiles, err := filepath.Glob(filepath.Join(dir, "**/*_pb2*.py"))
	if err != nil {
		return err
	}

	for _, file := range pbFiles {
		content, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		// Replace import statements
		newContent := strings.ReplaceAll(string(content),
			"from grpc.proto",
			"from crawlab.grpc")

		if err := os.WriteFile(file, []byte(newContent), 0644); err != nil {
			return err
		}
	}
	return nil
}
