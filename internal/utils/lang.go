// Copyright (c) https://github.com/fd008, all rights reserved.
// SPDX-License-Identifier: MPL-2.0

package utils

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
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
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
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

func (s *ULang) toAbsPath(dir string) string {
	if dir != "" && !filepath.IsAbs(dir) {
		absPath, err := filepath.Abs(dir)
		if err != nil {
			tflog.Info(s.Ctx, "Error getting absolute path of working directory:"+err.Error())
			return dir
		}
		return absPath
	}
	return dir
}

func (s *ULang) getCurDir() string {
	dir, err := os.Getwd()

	if err != nil {
		tflog.Info(s.Ctx, "Error getting current directory:"+err.Error())
		return ""
	}

	return s.toAbsPath(dir)
}

// runCommand executes a single command in the interpreter's state.
func (s *ULang) RunCommand(cmd Command) error {
	switch strings.ToUpper(cmd.Name) {
	case "WORKDIR":
		return s.setWorkDir(cmd.Args)
	case "RUN":
		return s.runExternalCommand(cmd.Args)
	case "COPY":
		return s.copy(cmd.Args)
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
	envfile := s.toAbsPath(args[0])

	file, err := os.Open(envfile)
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
	targetDir := s.toAbsPath(args[0])

	absPath := targetDir

	if s.WorkDir != "" {
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

func checkPathType(path string) map[string]bool {
	res := map[string]bool{
		"exists": false,
		"file":   false,
		"dir":    false,
	}

	fileInfo, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Path \"%s\" does not exist.\n", path)
		} else {
			fmt.Printf("Error checking path \"%s\": %v\n", path, err)
		}
		return res
	}

	if fileInfo.IsDir() {
		fmt.Printf("Path \"%s\" is a directory.\n", path)
		res["dir"] = true
		res["exists"] = true
		return res
	} else if fileInfo.Mode().IsRegular() {
		fmt.Printf("Path \"%s\" is a regular file.\n", path)
		res["file"] = true
		res["exists"] = true
		return res
	} else {
		fmt.Printf("Path \"%s\" is neither a regular file nor a directory (e.g., a symbolic link, device file, etc.).\n", path)

		return res
	}
}

func (s *ULang) copy(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("COPY requires two arguments (source, destination)")
	}
	srcPath := s.toAbsPath(args[0])
	destPath := s.toAbsPath(args[1])

	srcRes := checkPathType(srcPath)
	desRes := checkPathType(destPath)

	if !srcRes["exists"] {
		return fmt.Errorf("source path does not exist: %s", srcPath)
	}

	if !srcRes["file"] && !srcRes["dir"] {
		return fmt.Errorf("source path is neither a file nor a directory: %s", srcPath)
	}

	if !desRes["exists"] {
		// destination does not exist, create it
		if srcRes["file"] {
			// create parent directory for file
			parentDir := filepath.Dir(destPath)
			err := os.MkdirAll(parentDir, os.ModePerm)
			if err != nil {
				return fmt.Errorf("failed to create parent directory for destination file: %w", err)
			}
			// copy file
			err = s.copyFile(srcPath, destPath)
			if err != nil {
				return fmt.Errorf("failed to copy file: %w", err)
			}
		} else if srcRes["dir"] {
			// create directory
			err := os.MkdirAll(destPath, os.ModePerm)
			if err != nil {
				return fmt.Errorf("failed to create destination directory: %w", err)
			}
			// copy directory recursively
			err = copyDir(srcPath, destPath)
			if err != nil {
				return fmt.Errorf("failed to copy directory: %w", err)
			}
		}
	}
	return nil
}

// copyFile copies a file from source to destination.
func (s *ULang) copyFile(src, dest string) error {
	tflog.Info(s.Ctx, fmt.Sprintf("Copying file from %s to %s\n", src, dest))
	return copy(src, dest)

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

	cmd.Dir = s.getCurDir()
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
