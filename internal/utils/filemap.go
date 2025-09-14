// Copyright (c) https://github.com/fd008, all rights reserved.
// SPDX-License-Identifier: MPL-2.0

package utils

import (
	"os"
	"path/filepath"
)

func FileTree(root string, depth int, exclusionPatterns []string) (map[string]interface{}, error) {
	// A depth of 0 should default to full traversal.
	if depth <= 0 {
		depth = -1 // Use -1 to represent infinite depth
	}

	resultMap := make(map[string]interface{})

	// Compile the exclusion patterns
	patternMatchers := make([]func(string) bool, len(exclusionPatterns))
	for i, pattern := range exclusionPatterns {
		p := pattern
		patternMatchers[i] = func(path string) bool {
			match, _ := filepath.Match(p, filepath.Base(path))
			return match
		}
	}

	// Use a recursive helper function to build the nested map.
	err := buildMap(root, depth, patternMatchers, resultMap)
	if err != nil {
		return nil, err
	}

	return resultMap, nil
}

// buildMap is a recursive helper function that populates the nested map.
func buildMap(dirPath string, depth int, patternMatchers []func(string) bool, currentMap map[string]interface{}) error {
	// Read the directory contents
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		fullPath := filepath.Join(dirPath, entry.Name())

		// Check for exclusion
		isExcluded := false
		for _, matchFunc := range patternMatchers {
			if matchFunc(fullPath) {
				isExcluded = true
				break
			}
		}
		if isExcluded {
			continue
		}

		// Handle directory entries
		if entry.IsDir() {
			if depth > 1 || depth == -1 {
				nestedMap := make(map[string]interface{})
				currentMap[entry.Name()] = nestedMap

				// Recursively call for nested directory
				newDepth := depth - 1
				if depth == -1 {
					newDepth = -1
				}
				err := buildMap(fullPath, newDepth, patternMatchers, nestedMap)
				if err != nil {
					return err
				}
			} else {
				// If depth limit is reached, just add the directory name
				currentMap[entry.Name()] = nil // A simple placeholder
			}
		} else {
			// Handle file entries
			currentMap[entry.Name()] = nil // Represent a file with a placeholder
		}
	}
	return nil
}
