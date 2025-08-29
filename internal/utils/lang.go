// Copyright (c) HashiCorp, Inc.

package utils

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type Command struct {
	Name string
	Args []string
}

type ULang struct {
	CmdString string
	WorkDir   string
	CurDir    string
	Env       map[string]string
	Ctx       context.Context
}

func NewUlang(ctx context.Context, cmdString, workDir string) *ULang {
	dir, err := os.Getwd()

	if err != nil {
		tflog.Info(ctx, "Error getting current directory:"+err.Error())
	}

	if workDir == "" {
		workDir = dir
	}

	return &ULang{
		CmdString: cmdString,
		WorkDir:   workDir,
		Env:       make(map[string]string),
		CurDir:    dir,
		Ctx:       ctx,
	}
}

// parseFile reads a file line by line and converts it into a slice of Commands.
func (s *ULang) ParseCommands() ([]Command, error) {
	var commands []Command
	scanner := bufio.NewScanner(strings.NewReader(s.CmdString))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and comments
		}

		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		cmd := Command{
			Name: parts[0],
			Args: parts[1:],
		}
		commands = append(commands, cmd)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return commands, nil
}

// runCommand executes a single command in the interpreter's state.
func (s *ULang) RunCommand(cmd Command) error {
	switch strings.ToUpper(cmd.Name) {
	case "WORKDIR":
		return s.setWorkDir(cmd.Args)
	case "RUN":
		return s.runExternalCommand(cmd.Args)
	case "COPY":
		return s.copyFile(cmd.Args)
	case "ENV":
		return s.setEnv(cmd.Args)
	case "SLEEP":
		return s.sleep(cmd.Args)
	case "ENVFILE":
		return s.loadEnv(cmd.Args)
	default:
		return fmt.Errorf("unknown command: %s", cmd.Name)
	}
}

func (s *ULang) loadEnv(args []string) error {

	if len(args) != 1 {
		return fmt.Errorf("ENVFILE requires exactly one argument like this - ENVFILE <path>")
	}
	envfile := args[0]

	file, err := os.Open(path.Join(s.CurDir, envfile))
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip comments and blank lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Split on first '='
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue // or log error
		}
		key := strings.TrimSpace(parts[0])
		val := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
		s.setEnv([]string{key, val})
	}
	return scanner.Err()

}

func DirExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil // Directory does not exist
	}
	if err != nil {
		return false, err // Other error occurred
	}
	return info.IsDir(), nil // Return true if it's a directory
}

// setWorkDir changes the interpreter's working directory.
func (s *ULang) setWorkDir(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("workdir requires exactly one argument like this - workdir <path>")
	}
	targetDir := args[0]

	absPath := targetDir
	if !filepath.IsAbs(targetDir) {
		absPath = filepath.Join(s.WorkDir, targetDir)
	}

	tflog.Info(s.Ctx, "Setting working directory to: "+absPath)

	exists, err := DirExists(absPath)
	if err != nil {
		return fmt.Errorf("error checking if directory exists: %w", err)
	}
	if !exists {
		err := os.MkdirAll(absPath, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	tflog.Info(s.Ctx, "Working directory is: "+absPath)
	s.WorkDir = absPath
	return nil
}

// copyFile copies a file from source to destination.
func (s *ULang) copyFile(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("COPY requires two arguments (source, destination)")
	}
	srcPath := filepath.Join(s.CurDir, args[0])
	destPath := filepath.Join(s.WorkDir, args[1])

	tflog.Info(s.Ctx, fmt.Sprintf("Copying file from %s to %s\n", srcPath, destPath))

	return copyDir(srcPath, destPath)
}

// copyDir recursively copies a directory from src to dest.
func copyDir(src, dest string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get the relative path from the source directory
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		// Construct the destination path
		destPath := filepath.Join(dest, relPath)

		if info.IsDir() {
			// Create the destination directory, including any parents
			return os.MkdirAll(destPath, info.Mode())
		}

		// It's a file, so copy it
		return copy(path, destPath)
	})
}

// setEnv sets an environment variable in the interpreter's state.
func (s *ULang) setEnv(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("ENV requires two arguments (key, value)")
	}
	key, value := args[0], args[1]

	if value == "=" && len(s.Env) == 2 {
		value = args[2]
	}

	os.Setenv(key, value)
	s.Env[key] = value
	return nil
}

// sleep pauses execution for a specified duration.
func (s *ULang) sleep(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("SLEEP requires one argument (duration, e.g. 3s or 500ms)")
	}
	duration, err := time.ParseDuration(args[0])
	if err != nil {
		return fmt.Errorf("invalid duration: %w", err)
	}
	time.Sleep(duration)
	return nil
}

// Helper function to build the environment slice for os/exec.
func buildEnvSlice(env map[string]string) []string {
	var envSlice []string
	for k, v := range env {
		envSlice = append(envSlice, fmt.Sprintf("%s=%s", k, v))
	}
	return envSlice
}

// Helper function to copy a file.
func copy(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// runExternalCommand executes a command using os/exec, with cross-platform support.
func (s *ULang) runExternalCommand(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("RUN requires arguments")
	}

	var cmd *exec.Cmd

	// Check for shell-specific syntax like pipes, expansion, etc.
	// This approach is simplified and may need refinement for complex cases.
	requiresShell := strings.Contains(strings.Join(args, " "), "|") ||
		strings.Contains(strings.Join(args, " "), ">") ||
		strings.Contains(strings.Join(args, " "), "$")

	if requiresShell {
		// Use a shell to execute the command string
		fullCommand := strings.Join(args, " ")
		switch runtime.GOOS {
		case "windows":
			// On Windows, use cmd.exe and the "/C" flag
			cmd = exec.CommandContext(s.Ctx, "cmd", "/C", fullCommand)
			if errors.Is(cmd.Err, exec.ErrDot) {
				cmd.Err = nil
			}
		default: // Linux, macOS, etc.
			// On Unix-like systems, use a POSIX-compliant shell like sh
			cmd = exec.CommandContext(s.Ctx, "sh", "-c", fullCommand)
			if errors.Is(cmd.Err, exec.ErrDot) {
				cmd.Err = nil
			}
		}
	} else {
		// No special shell features needed; execute directly
		cmd = exec.CommandContext(s.Ctx, args[0], args[1:]...)
		if errors.Is(cmd.Err, exec.ErrDot) {
			cmd.Err = nil
		}
	}

	cmd.Dir = s.WorkDir
	cmd.Env = buildEnvSlice(s.Env)

	var stderr strings.Builder
	cmd.Stderr = &stderr

	output, err := cmd.Output()
	stderrStr := stderr.String()

	tflog.Error(s.Ctx, fmt.Sprintf("Command error: %s\n", stderrStr))
	tflog.Info(s.Ctx, fmt.Sprintf("Command output: %s\n", string(output)))

	// // Configure the command with the interpreter state
	// cmd.Stdin = bytes.NewReader([]byte{})
	// var stderr strings.Builder

	// cmd.Dir = s.WorkDir
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = &stderr
	// cmd.Env = buildEnvSlice(s.Env)

	// return cmd.Run()
	return err
}
