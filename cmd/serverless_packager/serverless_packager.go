package main

import (
	"archive/zip"
	"io"
	"log"
	"os"
)

func main() {
	files := []string{"bin/hello", "bin/world"}
	output := "bin/cryptobill.zip"

	err := ZipFiles(output, files)
	if err != nil {
		log.Fatal(err)
	}
}

func ZipFiles(filename string, files []string) error {
	newZipFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	// Add files to zip
	for _, file := range files {
		zipfile, err := os.Open(file)
		if err != nil {
			return err
		}
		defer zipfile.Close()

		// Get the file information
		info, err := zipfile.Stat()
		if err != nil {
			return err
		}

		info = SetExecutable(info)

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// Using FileInfoHeader() above only uses the basename of the file. If we want
		// to preserve the folder structure we can overwrite this with the full path.
		header.Name = file

		// Change to deflate to gain better compression
		// see http://golang.org/pkg/archive/zip/#pkg-constants
		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}
		if _, err = io.Copy(writer, zipfile); err != nil {
			return err
		}
	}
	return nil
}

// Not sure of a better way to do this, but it works, and I am happy.
func SetExecutable(info os.FileInfo) os.FileInfo {
	return &MyFileInfo{info}
}

type MyFileInfo struct {
	original os.FileInfo
}

func (mfi *MyFileInfo) Name() string {
	return mfi.original.Name()
}

func (mfi *MyFileInfo) Size() int64 {
	return mfi.original.Size()
}

func (mfi *MyFileInfo) Mode() os.FileMode {
	return os.ModePerm
}

func (mfi *MyFileInfo) ModTime() time.Time {
	return mfi.original.ModTime()
}

func (mfi *MyFileInfo) IsDir() bool {
	return mfi.original.IsDir()
}

func (mfi *MyFileInfo) Sys() interface{} {
	return mfi.original.Sys()
}
