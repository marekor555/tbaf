package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func addFile(archive *os.File, fileName string) {
	if stat, err := os.Stat(fileName); err != nil || stat.IsDir() {
		fmt.Println("Adding directory: ", fileName)
		err = filepath.Walk(fileName, func(path string, info os.FileInfo, err error) error {
			if path == fileName || info.IsDir() {
				return nil
			}
			addFile(archive, path)
			return nil
		})
		if err != nil {
			fmt.Println("Failed to walk directory: ", fileName)
			fmt.Println(err)
		}
		return
	}
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println("Failed to open file: ", fileName)
		fmt.Println(err)
	}
	defer file.Close()
	fileStat, err := file.Stat()
	if err != nil {
		fmt.Println("Failed to get file stat: ", fileName)
		fmt.Println(err)
	}
	fileBuff := make([]byte, fileStat.Size())
	_, err = file.Read(fileBuff)
	if err != nil {
		fmt.Println("Failed to read file: ", fileName)
		fmt.Println(err)
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
		fmt.Println("Failed to write file: ", fileName)
	}
}

func main() {
	mainCmd := os.Args[1]
	extraArgs := os.Args[2:]
	switch mainCmd {
	case "build":
		fmt.Println("building tabf file...")
		archive, err := os.Create(extraArgs[0])
		if err != nil {
			fmt.Println(err)
		}
		defer archive.Close()
		fmt.Println("Created file, copying files to archive...")
		for _, fileName := range extraArgs[1:] {
			addFile(archive, fileName)
		}
	case "list":
		fmt.Println("listing tabf file...")
		fmt.Println("not implemented")
	case "unpack":
		fmt.Println("unpacking tabf file...")
		archive, err := os.Open(extraArgs[0])
		if err != nil {
			fmt.Println(err)
		}
		defer archive.Close()
		fmt.Println("Opened file, extracting files...")
		for {
			fileNameBuff := make([]byte, 256)
			_, err = archive.Read(fileNameBuff)
			if err != nil && err == io.EOF {
				fmt.Println("Finished reading file")
				break
			}
			if err != nil {
				fmt.Println("Failed to read file name: ", err)
				break
			}
			fileSizeBuff := make([]byte, 8)
			_, err = io.ReadFull(archive, fileSizeBuff)
			if err != nil {
				fmt.Println("Failed to read file size: ", err)
			}
			fileBuff := make([]byte, binary.BigEndian.Uint64(fileSizeBuff))
			_, err = io.ReadFull(archive, fileBuff)
			if err != nil {
				fmt.Println("Failed to read file: ", err)
			}
			fmt.Println("Writing file: ", string(fileNameBuff))
			fmt.Println("File size: ", binary.BigEndian.Uint64(fileSizeBuff))

			fileNameBuff = bytes.Trim(fileNameBuff, "\x00")

			err = os.MkdirAll(filepath.Dir(string(fileNameBuff)), 0744)
			if err != nil {
				fmt.Println("Failed to create directory: ", string(fileNameBuff))
				fmt.Println(err)
				break
			}
			err = os.WriteFile(string(fileNameBuff), fileBuff, 0744)
			if err != nil {
				fmt.Println("Failed to write file: ", string(fileNameBuff))
				fmt.Println(err)
				break
			}
		}
	default:
		fmt.Println("command not found")
	}
}
