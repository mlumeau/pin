package storage

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"pin/internal/config"
)

// BackupToZip writes a zip of the database and uploads directory.
func BackupToZip(cfg config.Config, destPath string) (string, error) {
	if destPath == "" {
		destPath = fmt.Sprintf("pin-backup-%s.zip", time.Now().UTC().Format("20060102-150405"))
	}
	if filepath.Ext(destPath) != ".zip" {
		destPath = destPath + ".zip"
	}
	out, err := os.Create(destPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	zipWriter := zip.NewWriter(out)
	defer zipWriter.Close()

	if err := addFileToZip(zipWriter, cfg.DBPath, "identity.db"); err != nil {
		return "", err
	}
	uploadsDir := cfg.UploadsDir
	if uploadsDir != "" {
		if err := addDirToZip(zipWriter, uploadsDir, "uploads"); err != nil {
			return "", err
		}
	}
	return destPath, nil
}

func addFileToZip(zipWriter *zip.Writer, sourcePath, name string) error {
	file, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return err
	}
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = name
	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, file)
	return err
}

func addDirToZip(zipWriter *zip.Writer, sourceDir, prefix string) error {
	return filepath.WalkDir(sourceDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if strings.HasPrefix(rel, ".") {
			return nil
		}
		name := filepath.ToSlash(filepath.Join(prefix, rel))
		return addFileToZip(zipWriter, path, name)
	})
}
