// Copyright (c) https://github.com/fd008, all rights reserved.
// SPDX-License-Identifier: MPL-2.0

package awscloud

import (
	"context"
	"fmt"
	"log"
	"mime"
	"os"
	"path"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	mimeMap = map[string]string{
		".cjs":         "text/javascript",
		".ico":         "image/x-icon",
		".mp4":         "video/mp4",
		".webm":        "video/webm",
		".mp3":         "audio/mpeg",
		".ogg":         "audio/ogg",
		".txt":         "text/plain",
		".woff":        "font/woff",
		".woff2":       "font/woff2",
		".ttf":         "font/ttf",
		".otf":         "font/otf",
		".eot":         "application/vnd.ms-fontobject",
		".webmanifest": "application/manifest+json",
	}
)

func init() {
	// Register custom MIME types for specific file extensions
	for ext, mimeType := range mimeMap {
		mime.AddExtensionType(ext, mimeType)
	}
}

func getContentType(path string) string {
	// Get the MIME type based on the file extension
	if contentType := mime.TypeByExtension(filepath.Ext(path)); contentType != "" {
		return contentType
	}
	// Default to application/octet-stream if no specific type is found
	return "application/octet-stream"
}

// isExcluded checks if a file path matches any pattern in the exclusion list.
func isExcluded(filePath string, exclusionList []string) bool {
	base := filepath.Base(filePath)
	for _, pattern := range exclusionList {
		matched, err := path.Match(pattern, base)
		if err != nil {
			log.Printf("Error matching pattern %s with file path %s: %v\n", pattern, filePath, err)
			continue
		}
		if matched {
			return true
		}
	}
	return false
}

func Upload(cfg aws.Config, localPath string, bucket string, prefix *string, exclusionList *[]string, additional_mimes *map[string]string) {

	// check if localPath is a directory
	if info, err := os.Stat(localPath); err != nil {
		tflog.Error(context.TODO(), fmt.Sprintf("Error stating local path %s: %v", localPath, err))
		return
	} else if !info.IsDir() {
		tflog.Error(context.TODO(), fmt.Sprintf("Local path %s is not a directory", localPath))
		return
	}

	// If additional MIME types are provided, merge them into the mimeMap
	if additional_mimes != nil || len(*additional_mimes) > 0 {
		for ext, mimeType := range *additional_mimes {
			mime.AddExtensionType(ext, mimeType)
		}
	}

	const numWorkers = 16
	const chanBuffer = 256

	fileChan := make(chan string, chanBuffer)

	// Producer: walks the directory and sends file paths to fileChan
	go func() {
		err := filepath.WalkDir(localPath, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if exclusionList != nil && len(*exclusionList) > 0 && isExcluded(path, *exclusionList) {
				log.Printf("Skipping excluded file: %s\n", path)
				tflog.Info(context.TODO(), fmt.Sprintf("Skipping excluded file: %s", path))
				return nil
			}

			if !d.IsDir() {
				fileChan <- path
			}
			return nil
		})
		if err != nil {
			log.Fatalln("WalkDir failed:", err)
		}
		close(fileChan)
	}()

	uploader := manager.NewUploader(s3.NewFromConfig(cfg), func(u *manager.Uploader) {
		u.PartSize = 5 * 1024 * 1024
		u.Concurrency = 10
	})

	// Worker pool: each worker uploads files from fileChan
	done := make(chan struct{})
	for i := 0; i < numWorkers; i++ {
		go func() {
			for path := range fileChan {
				rel, err := filepath.Rel(localPath, path)
				if err != nil {
					log.Println("Unable to get relative path:", path, err)
					continue
				}
				file, err := os.Open(path)
				if err != nil {
					log.Println("Failed opening file", path, err)
					continue
				}
				key := rel

				// add prefix if provided
				if prefix != nil || *prefix != "" {
					key = filepath.Join(*prefix, rel)
				}

				_, err = uploader.Upload(context.TODO(), &s3.PutObjectInput{
					Bucket:      &bucket,
					Key:         aws.String(key),
					ContentType: aws.String(getContentType(path)),
					Body:        file,
				})

				file.Close()
				if err != nil {
					log.Println("Failed to upload", path, err)
					continue
				}
				// log.Println("Uploaded", path)
				tflog.Info(context.TODO(), "Uploaded "+path)
			}
			done <- struct{}{}
		}()
	}

	// Wait for all workers to finish
	for i := 0; i < numWorkers; i++ {
		<-done
	}
}
