package main

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

func main() {
	logFile, err := os.OpenFile("patch_installation_log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logAndPrint(nil, "Failed to create log file: %v\n", err)
		return
	}
	defer logFile.Close()

	logger := log.New(logFile, "", log.LstdFlags)

	logAndPrint(logger, "\n[STARTING PROGRAM]\n")

	err = os.MkdirAll("c:\\temp\\patchinstalls", os.ModePerm)
	if err != nil {
		logAndPrint(logger, "Failed to create directory: %v\n", err)
		return
	}

	args := os.Args[1:]
	if contains(args, "-silent") {
		if !contains(args, "-links") {
			logAndPrint(logger, "-silent parameter requires -links parameter.\n")
			return
		}
		linksIndex := indexOf(args, "-links")
		if linksIndex == -1 || linksIndex+1 >= len(args) {
			logAndPrint(logger, "No download links provided.\n")
			return
		}
		linksFile := args[linksIndex+1]
		links, err := readLinksFromFile(linksFile)
		if err != nil {
			logAndPrint(logger, "Failed to read links from file: %v\n", err)
			return
		}
		downloadAndInstall(links, logger)
		return
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter 'zip' for ZIP file or 'msu' for direct MSU file download: ")
	fileType, _ := reader.ReadString('\n')
	fileType = strings.TrimSpace(fileType)

	switch fileType {
	case "zip":
		handleZipDownload(reader, logger)
	case "msu":
		handleMSUDownload(reader, logger)
	default:
		logAndPrint(logger, "Invalid option selected.\n")
	}
}

func downloadAndInstall(links []string, logger *log.Logger) {
	for _, url := range links {
		fileName := filepath.Base(url)
		logAndPrint(logger, "Starting download from: %s\n", url)
		err := downloadFile(fileName, url)
		if err != nil {
			logAndPrint(logger, "Download failed: %v\n", err)
			continue
		}
		logAndPrint(logger, "Download completed.\n")

		if strings.HasSuffix(fileName, ".zip") {
			files, err := unzip(fileName, "c:\\temp\\patchinstalls")
			if err != nil {
				logAndPrint(logger, "Unzip failed: %v\n", err)
				continue
			}
			installMSUFiles(files, logger)
		} else if strings.HasSuffix(fileName, ".msu") {
			installMSUFiles([]string{fileName}, logger)
		}

		err = os.Remove(fileName)
		if err != nil {
			logAndPrint(logger, "Failed to delete the file: %v\n", err)
		} else {
			logAndPrint(logger, "Cleanup completed. Downloaded file removed.\n")
		}
	}
}

func handleZipDownload(reader *bufio.Reader, logger *log.Logger) {
	fmt.Print("Enter the download link for the ZIP file: ")
	url, _ := reader.ReadString('\n')
	url = strings.TrimSpace(url)

	logAndPrint(logger, "Starting download from: %s\n", url)
	err := downloadFile("download.zip", url)
	if err != nil {
		logAndPrint(logger, "Download failed: %v\n", err)
		return
	}
	logAndPrint(logger, "Download completed.\n")

	files, err := unzip("download.zip", "c:\\temp\\patchinstalls")
	if err != nil {
		logAndPrint(logger, "Unzip failed: %v\n", err)
		return
	}

	installMSUFiles(files, logger)

	err = os.Remove("download.zip")
	if err != nil {
		logAndPrint(logger, "Failed to delete the zip file: %v\n", err)
	} else {
		logAndPrint(logger, "Cleanup completed. Downloaded ZIP file removed.\n")
	}
}

func handleMSUDownload(reader *bufio.Reader, logger *log.Logger) {
	fmt.Print("Enter the download link for the MSU file: ")
	url, _ := reader.ReadString('\n')
	url = strings.TrimSpace(url)

	fileName := filepath.Base(url)
	logAndPrint(logger, "Starting download from: %s\n", url)
	err := downloadFile(fileName, url)
	if err != nil {
		logAndPrint(logger, "Download failed: %v\n", err)
		return
	}
	logAndPrint(logger, "Download completed.\n")

	installMSUFiles([]string{fileName}, logger)

	err = os.Remove(fileName)
	if err != nil {
		logAndPrint(logger, "Failed to delete the MSU file: %v\n", err)
	} else {
		logAndPrint(logger, "Cleanup completed. Downloaded MSU file removed.\n")
	}
}

func installMSUFiles(files []string, logger *log.Logger) {
	for _, file := range files {
		if filepath.Ext(file) == ".msu" {
			logAndPrint(logger, "Starting installation of: %s\n", file)
			startTime := time.Now()

			commandString := fmt.Sprintf("wusa.exe %s /quiet /norestart", file)
			cmd := exec.Command("cmd", "/C", commandString)
			err := cmd.Run()

			duration := time.Since(startTime)
			if err != nil {
				if exiterr, ok := err.(*exec.ExitError); ok {
					if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
						errorCode := status.ExitStatus()
						if errorCode == 0 {
							logAndPrint(logger, "Successfully installed %s. Duration: %s\n", file, duration)
						} else if errorCode == 3010 {
							logAndPrint(logger, "Successfully installed %s (needs reboot). Duration: %s\n", file, duration)
						} else {
							logAndPrint(logger, "Failed to install %s: %v. Error code: %d. Duration: %s\n", file, err, errorCode, duration)
						}
					}
				} else {
					logAndPrint(logger, "Failed to install %s: %v. Duration: %s\n", file, err, duration)
				}
			} else {
				logAndPrint(logger, "Successfully installed %s. Duration: %s\n", file, duration)
			}
		}
	}
}

func downloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func unzip(src string, dest string) ([]string, error) {
	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
		} else {
			if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				return filenames, err
			}

			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return filenames, err
			}

			rc, err := f.Open()
			if err != nil {
				return filenames, err
			}

			_, err = io.Copy(outFile, rc)

			outFile.Close()
			rc.Close()

			if err != nil {
				return filenames, err
			}
			filenames = append(filenames, fpath)
		}
	}
	return filenames, nil
}

func logAndPrint(logger *log.Logger, format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Print(msg)
	if logger != nil {
		logger.Printf(msg)
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func indexOf(slice []string, item string) int {
	for i, s := range slice {
		if s == item {
			return i
		}
	}
	return -1
}

func readLinksFromFile(filePath string) ([]string, error) {
	var links []string
	file, err := os.Open(filePath)
	if err != nil {
		return links, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		links = append(links, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return links, err
	}

	return links, nil
}
