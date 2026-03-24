// Copyright 2019 The vogo Authors. All rights reserved.
// author: wongoo

package gracego

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// unzip will decompress a zip archive, moving all files and folders
// within the zip file (parameter 1) to an output directory (parameter 2).
func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	// Make File
	if err := os.MkdirAll(dest, os.ModePerm); err != nil {
		return err
	}

	for _, zipFile := range r.File {
		// Check for ZipSlip: ensure the file is within the destination directory
		targetPath := filepath.Join(dest, zipFile.Name)
		targetPath = filepath.Clean(targetPath)
		destClean := filepath.Clean(dest)

		if !strings.HasPrefix(targetPath, destClean+string(os.PathSeparator)) && targetPath != destClean {
			graceLog("ignore zip file (ZipSlip): %s", zipFile.Name)
			continue
		}

		if err := writeZipFile(targetPath, zipFile); err != nil {
			return err
		}
	}
	return nil
}

func writeZipFile(targetPath string, f *zip.File) error {
	if f.FileInfo().IsDir() {
		if err := os.MkdirAll(targetPath, os.ModePerm); err != nil {
			return err
		}
		return nil
	}

	outFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}

	rc, err := f.Open()
	if err != nil {
		return err
	}

	_, err = io.Copy(outFile, rc)

	outFile.Close()
	_ = rc.Close()

	return err
}
