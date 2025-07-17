package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func addFile(archive *os.File, fileName string) (string, error) {
	if strings.HasPrefix(fileName, "/") {
		return "Absolute paths are not allowed", fmt.Errorf("absolute path detected: %s", fileName)
	}

	if stat, err := os.Stat(fileName); err != nil || stat.IsDir() {
		fmt.Println("Adding directory: ", fileName)
		err = filepath.Walk(fileName, func(path string, info os.FileInfo, err error) error {
			if path == fileName || info.IsDir() {
				return nil
			}
			_, err = addFile(archive, path)
			return err
		})
		if err != nil {
			return "Failed to walk directory", err
		}
		return "File is a directory", nil
	}

	file, err := os.Open(fileName)
	if err != nil {
		return "Failed to open file", err
	}
	defer file.Close()

	fileStat, err := file.Stat()
	if err != nil {
		return "Failed to stat file: " + fileName, err
	}

	fileBuff := make([]byte, fileStat.Size())
	_, err = file.Read(fileBuff)
	if err != nil {
		return "Failed to read file: " + fileName, err
	}

	fileNameBuff := make([]byte, 256)
	fileSizeBuff := make([]byte, 8)
	binary.BigEndian.PutUint64(fileSizeBuff, uint64(fileStat.Size()))
	copy(fileNameBuff, fileName)

	fmt.Println("Writing file: ", fileName)
	fmt.Println("File size: ", fileStat.Size())

	_, err = archive.Write(fileNameBuff)
	_, err = archive.Write(fileSizeBuff)
	_, err = archive.Write(fileBuff)
	if err != nil {
		return "Failed to write file to archive: " + fileName, err
	}
	return "Success", nil
}

func listArchive(name string) ([]string, string, error) {
	var output []string
	archive, err := os.Open(name)
	if err != nil {
		log.Fatal(err)
	}
	defer archive.Close()

	for {
		fileNameBuff := make([]byte, 256)
		_, err = io.ReadFull(archive, fileNameBuff)
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			return output, "Failed to read file name", err
		}

		fileSizeBuff := make([]byte, 8)
		_, err = io.ReadFull(archive, fileSizeBuff)
		if err != nil {
			return output, "Failed to read file size", err
		}

		_, err = archive.Seek(int64(binary.BigEndian.Uint64(fileSizeBuff)), io.SeekCurrent)
		if err != nil {
			return output, "Failed to seek to next file", err
		}

		output = append(output, string(fileNameBuff))
	}

	return output, "Success", nil
}

func buildArchive(name string, files []string) (string, error) {
	if _, err := os.Stat(name); err == nil {
		return "Can't overwrite files, exiting", fmt.Errorf("file already exists: %s", name)
	}
	archive, err := os.Create(name)
	if err != nil {
		return "Failed to create archive file: " + name, err
	}
	defer archive.Close()

	for _, fileName := range files {
		msg, err := addFile(archive, fileName)
		if err != nil {
			return msg, err
		}
	}
	return "Successfully built archive: " + name, nil
}

func unpackArchive(name string, prefix string) (string, error) {
	archive, err := os.Open(name)
	if err != nil {
		fmt.Println("Failed to open file: ", name)
		log.Fatal(err)
	}
	defer archive.Close()

	for {
		fileNameBuff := make([]byte, 256)
		_, err = io.ReadFull(archive, fileNameBuff)
		if err != nil && err == io.EOF {
			fmt.Println("Finished reading file")
			break
		}
		if err != nil {
			return "Failed to read file name", err
		}

		fileSizeBuff := make([]byte, 8)
		_, err = io.ReadFull(archive, fileSizeBuff)
		if err != nil {
			return "Failed to read file size", err
		}

		fileBuff := make([]byte, binary.BigEndian.Uint64(fileSizeBuff))
		_, err = io.ReadFull(archive, fileBuff)
		if err != nil {
			return "Failed to read file", err
		}

		fileNameBuff = bytes.Trim(fileNameBuff, "\x00")
		filename := path.Join(prefix, string(fileNameBuff))

		fmt.Println("Writing file: ", filename)
		fmt.Println("File size: ", binary.BigEndian.Uint64(fileSizeBuff))

		err = os.MkdirAll(filepath.Dir(filename), 0744)
		if err != nil {
			return "Failed to create directory: " + filename, err
		}

		err = os.WriteFile(filename, fileBuff, 0744)
		if err != nil {
			return "Failed to write file: " + filename, err
		}
	}

	if prefix == "" {
		return "Finished extracting files", nil
	} else {
		return "Finished extracting files to: " + prefix, nil
	}
}
