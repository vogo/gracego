// Copyright 2019 The vogo Authors. All rights reserved.

package gracego

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// Upgrade gracefully upgrade server
// - version: the new version of the server
// - path: the relative path of the command in the upgrade compress file
// - upgradeUrl: the url of the upgrade file, which must be a zip format file with a suffix `.jar` or `.zip`
func Upgrade(version, path, upgradeURL string) error {
	if err := upgradeServerCmd(version, path, upgradeURL); err != nil {
		return err
	}
	go restart()
	return nil
}

// upgradeServerCmd grace up server bin file
func upgradeServerCmd(version, path, upgradeURL string) error {
	if server == nil || serverBin == "" {
		return errors.New("server not started")
	}

	versionDir := filepath.Join(serverDir, version)
	err := os.Mkdir(versionDir, 0770)
	if err != nil && !strings.Contains(err.Error(), "file exists") {
		return err
	}

	fileName, err := parseFileName(upgradeURL)
	if err != nil {
		return err
	}

	upgradeCmd := filepath.Join(versionDir, path)
	_, err = os.Open(upgradeCmd)
	if err == nil {
		graceLog("found upgrade command file: %s", upgradeCmd)
		return link(upgradeCmd, serverCmdPath)
	}

	downloadPath := filepath.Join(versionDir, fileName)
	err = downloadFile(downloadPath, upgradeURL)
	if err != nil {
		return err
	}

	err = unzip(downloadPath, versionDir)
	if err != nil {
		return err
	}

	return link(upgradeCmd, serverCmdPath)
}

func parseFileName(upgradeURL string) (string, error) {
	u, err := url.Parse(upgradeURL)
	if err != nil {
		return "", err
	}
	uri := u.RequestURI()
	index := strings.LastIndex(uri, "/")
	if index < 0 {
		return "", fmt.Errorf("invalid download url: %s", u)
	}

	fileName := uri[index+1:]
	if !acceptFileSuffix(fileName) {
		return "", fmt.Errorf("invalid suffix for download url: %s", u)
	}
	return fileName, nil
}

func link(src, dest string) error {
	_ = os.Remove(dest)
	graceLog("link %s to %s", src, dest)
	return os.Link(src, dest)
}

func acceptFileSuffix(f string) bool {
	return strings.HasSuffix(f, ".jar") || strings.HasSuffix(f, ".zip")
}

// downloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func downloadFile(filePath, upgradeURL string) error {
	u, err := url.Parse(upgradeURL)
	if err != nil {
		return err
	}

	graceLog("download %s to %s", upgradeURL, filePath)
	_ = os.Remove(filePath)

	// Get the data
	resp, err := http.Get(u.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
	case 404:
		return fmt.Errorf("file not found: %s", upgradeURL)
	default:
		buf := make([]byte, 1024)
		result := ""
		if n, readErr := resp.Body.Read(buf); n > 0 && readErr == nil {
			result = string(buf[:n])
		}

		return fmt.Errorf("download failed, status code: %d, result: %s", resp.StatusCode, result)
	}

	// Create the file
	out, err := os.Create(filePath)
	if err != nil {
		graceLog("can't create file: %v", err)
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}
