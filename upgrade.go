// Copyright 2019 The vogo Authors. All rights reserved.

package gracego

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

//Upgrade gracefully upgrade server
// - version: the new version of the server
// - path: the relative path of the command in the upgrade compress file
// - upgradeUrl: the url of the upgrade file, which must be a zip format file with a suffix `.jar` or `.zip`
func Upgrade(version, path, upgradeUrl string) error {
	if err := upgradeServerCmd(version, path, upgradeUrl); err != nil {
		return err
	}
	go restart()
	return nil
}

//upgradeServerCmd graceup server bin file
func upgradeServerCmd(version, path, upgradeUrl string) error {
	if server == nil || serverBin == "" {
		return errors.New("server not started")
	}

	versionDir := fmt.Sprintf("%s%c%s", serverDir, os.PathSeparator, version)
	err := os.Mkdir(versionDir, 0770)
	if err != nil && !strings.Contains(err.Error(), "file exists") {
		return err
	}

	u, err := url.Parse(upgradeUrl)
	if err != nil {
		return err
	}
	uri := u.RequestURI()
	index := strings.LastIndex(uri, "/")
	if index < 0 {
		return fmt.Errorf("invalid download url: %s", upgradeUrl)
	}

	fileName := uri[index+1:]
	if !acceptFileSuffix(fileName) {
		return fmt.Errorf("invalid suffix for download url: %s", upgradeUrl)
	}

	upgradeCmd := fmt.Sprintf("%s%c%s", versionDir, os.PathSeparator, path)
	_, err = os.Open(upgradeCmd)
	if err == nil {
		fmt.Println("found upgrade command file: ", upgradeCmd)
		return link(upgradeCmd, serverCmdPath)
	}

	downloadPath := fmt.Sprintf("%s%c%s", versionDir, os.PathSeparator, fileName)
	err = downloadFile(downloadPath, upgradeUrl)
	if err != nil {
		return err
	}

	err = unzip(downloadPath, versionDir)
	if err != nil {
		return err
	}

	return link(upgradeCmd, serverCmdPath)
}

func link(src string, dest string) error {
	_ = os.Remove(dest)
	info("link %s to %s", src, dest)
	return os.Link(src, dest)
}

func acceptFileSuffix(f string) bool {
	return strings.HasSuffix(f, ".jar") || strings.HasSuffix(f, ".zip")
}

// downloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func downloadFile(filePath string, url string) error {
	info("download %s to %s", url, filePath)
	_ = os.Remove(filePath)

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
	case 404:
		return fmt.Errorf("file not found: %s", url)
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
		info("can't create file: %v", err)
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}
