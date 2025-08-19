// Copyright (c) HashiCorp, Inc.

package utils

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

func ConcurrentFileSet(root string, excludePatterns ...string) ([]string, error) {
	var (
		wg          sync.WaitGroup                             // To wait for all goroutines to finish
		filePaths   = make(chan string)                        // To send file paths from goroutines
		errChan     = make(chan error, 1)                      // To send the first error encountered
		ctx, cancel = context.WithCancel(context.Background()) // For managing cancellation
	)
	defer cancel() // Ensure cancellation is called to clean up goroutines

	// Start the root directory scan in a goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		walkDir(ctx, root, excludePatterns, filePaths, errChan, &wg)
	}()

	// Goroutine to close the filePaths channel once all walkDir goroutines are done
	go func() {
		wg.Wait()
		close(filePaths)
		close(errChan) // Close error channel after all goroutines finish
	}()

	var files []string
	for {
		select {
		case filePath, ok := <-filePaths:
			if !ok { // Channel closed
				select { // Check if an error was sent before closing filePaths
				case err := <-errChan:
					if err != nil {
						return nil, err
					}
				default:
					// No error
				}
				return files, nil
			}
			files = append(files, filePath)
		case err := <-errChan:
			if err != nil {
				return nil, err // Return the first error encountered
			}
		case <-ctx.Done(): // Context cancelled (e.g., due to an error in another goroutine)
			return nil, ctx.Err()
		}
	}
}

// walkDir is the recursive function to walk a directory concurrently.
func walkDir(ctx context.Context, dir string, excludePatterns []string, filePaths chan<- string, errChan chan<- error, wg *sync.WaitGroup) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		select {
		case errChan <- fmt.Errorf("failed to read directory %q: %w", dir, err):
		case <-ctx.Done():
			// Another error might have been sent or context cancelled, don't block
		}
		return
	}

	for _, entry := range entries {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return // Stop processing if cancelled
		default:
			// Continue
		}

		path := filepath.Join(dir, entry.Name())

		// Apply exclusion patterns
		if len(excludePatterns) > 0 {
			for _, pattern := range excludePatterns {
				matched, err := filepath.Match(pattern, entry.Name())
				if err != nil {
					select {
					case errChan <- fmt.Errorf("failed to match pattern %q against %q: %w", pattern, entry.Name(), err):
					case <-ctx.Done():
					}
					return
				}
				if matched {
					if entry.IsDir() {
						// Skip this directory and its contents
						goto nextEntry // Skip to the next entry
					}
					goto nextEntry // Skip this file
				}
			}
		}

		if entry.IsDir() {
			wg.Add(1)
			go func(path string) {
				defer wg.Done()
				walkDir(ctx, path, excludePatterns, filePaths, errChan, wg)
			}(path)
		} else {
			filePaths <- path
		}
	nextEntry:
	}
}
