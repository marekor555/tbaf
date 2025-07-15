package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/klauspost/compress/zstd"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func addFile(archive *os.File, fileName string) {
	if strings.HasPrefix(fileName, "/") {
		fmt.Println("absolute path detected: ", fileName)
		fmt.Println("if you try to unpack with absolute paths, it might break your system")
		log.Fatal("ABSOLUTE PATHS ARE NOT ALLOWED")
	}
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
			log.Fatal(err)
		}
		return
	}
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println("Failed to open file: ", fileName)
		log.Fatal(err)
	}
	defer file.Close()
	fileStat, err := file.Stat()
	if err != nil {
		fmt.Println("Failed to get file stat: ", fileName)
		log.Fatal(err)
	}
	fileBuff := make([]byte, fileStat.Size())
	_, err = file.Read(fileBuff)
	if err != nil {
		fmt.Println("Failed to read file: ", fileName)
		log.Fatal(err)
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
		log.Fatal(err)
	}
}

func main() {
	mainCmd := os.Args[1]
	extraArgs := os.Args[2:]
	switch mainCmd {
	case "build", "b":
		fmt.Println("building tbaf file...")
		archive, err := os.Create(extraArgs[0])
		if err != nil {
			log.Fatal(err)
		}
		defer archive.Close()
		fmt.Println("Created file, copying files to archive...")
		for _, fileName := range extraArgs[1:] {
			addFile(archive, fileName)
		}
	case "list", "l":
		archive, err := os.Open(extraArgs[0])
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
				fmt.Println("Failed to read file name")
				log.Fatal(err)
			}
			fileSizeBuff := make([]byte, 8)
			_, err = io.ReadFull(archive, fileSizeBuff)
			if err != nil {
				fmt.Println("Failed to read file size")
				log.Fatal(err)
			}
			_, err = archive.Seek(int64(binary.BigEndian.Uint64(fileSizeBuff)), io.SeekCurrent)
			if err != nil {
				fmt.Println("Failed to seek to next file")
				log.Fatal(err)
			}
			fmt.Println(string(fileNameBuff))
		}
	case "unpack", "u":
		fmt.Println("unpacking tbaf file...")
		archive, err := os.Open(extraArgs[0])
		if err != nil {
			fmt.Println("Failed to open file: ", extraArgs[0])
			log.Fatal(err)
		}
		defer archive.Close()
		prefix := ""
		if len(extraArgs) > 1 {
			prefix = filepath.Clean(extraArgs[1]) + "/"
		}
		fmt.Println("Opened file, extracting files...")
		for {
			fileNameBuff := make([]byte, 256)
			_, err = io.ReadFull(archive, fileNameBuff)
			if err != nil && err == io.EOF {
				fmt.Println("Finished reading file")
				break
			}
			if err != nil {
				fmt.Println("Failed to read file name")
				log.Fatal(err)
			}
			fileSizeBuff := make([]byte, 8)
			_, err = io.ReadFull(archive, fileSizeBuff)
			if err != nil {
				fmt.Println("Failed to read file size")
				log.Fatal(err)
			}
			fileBuff := make([]byte, binary.BigEndian.Uint64(fileSizeBuff))
			_, err = io.ReadFull(archive, fileBuff)
			if err != nil {
				fmt.Println("Failed to read file")
				log.Fatal(err)
			}

			fileNameBuff = bytes.Trim(fileNameBuff, "\x00")
			filename := path.Join(prefix, string(fileNameBuff))

			fmt.Println("Writing file: ", filename)
			fmt.Println("File size: ", binary.BigEndian.Uint64(fileSizeBuff))

			err = os.MkdirAll(filepath.Dir(filename), 0744)
			if err != nil {
				fmt.Println("Failed to create directory: ", filename)
				log.Fatal(err)
			}
			err = os.WriteFile(filename, fileBuff, 0744)
			if err != nil {
				fmt.Println("Failed to write file: ", filename)
				log.Fatal(err)
			}
		}
		if prefix == "" {
			fmt.Println("Finished extracting files")
		} else {
			fmt.Println("Finished extracting files to: ", prefix)
		}
	case "compress", "c":
		fmt.Println("compressing tbaf file...")
		file, err := os.Open(extraArgs[0])
		if err != nil {
			fmt.Println("Failed to open file: ", extraArgs[0])
			log.Fatal(err)
		}
		defer file.Close()
		compFile, err := os.Create(extraArgs[0] + ".zst")
		if err != nil {
			fmt.Println("Failed to create compressed file: ", extraArgs[0]+".zst")
			log.Fatal(err)
		}
		defer compFile.Close()

		fileStat, err := file.Stat()
		if err != nil {
			fmt.Println("Failed to get file stat: ", extraArgs[0])
			log.Fatal(err)
		}
		fileBuff := make([]byte, fileStat.Size())
		_, err = io.ReadFull(file, fileBuff)
		if err != nil {
			fmt.Println("Failed to read file: ", extraArgs[0])
			log.Fatal(err)
		}

		writer, err := zstd.NewWriter(compFile, zstd.WithEncoderLevel(zstd.EncoderLevelFromZstd(10)))
		if err != nil {
			fmt.Println("Failed to create zstd writer")
			log.Fatal(err)
		}
		defer writer.Close()
		_, err = writer.Write(fileBuff)
		if err != nil {
			fmt.Println("Failed to write encrypted file")
			log.Fatal(err)
		}
		fmt.Println("Compressed file written to: ", extraArgs[0]+".zst")
	case "decompress", "d":
		fmt.Println("decompressing tbaf file...")
		file, err := os.Open(extraArgs[0])
		if err != nil {
			fmt.Println("Failed to open file: ", extraArgs[0])
			log.Fatal(err)
		}
		defer file.Close()

		newFileName := strings.TrimSuffix(extraArgs[0], ".zst")
		decompFile, err := os.Create(newFileName)
		if err != nil {
			fmt.Println("Failed to create decompressed file: ", newFileName)
			log.Fatal(err)
		}
		defer decompFile.Close()

		reader, err := zstd.NewReader(file)
		if err != nil {
			fmt.Println("Failed to create zstd reader")
			log.Fatal(err)
		}
		defer reader.Close()

		fileBuff, err := io.ReadAll(reader)
		_, err = decompFile.Write(fileBuff)
		if err != nil {
			fmt.Println("Failed to write decompressed file")
			log.Fatal(err)
		}
		fmt.Println("Decompressed file written to: ", newFileName)
	default:
		fmt.Println("command not found")
	}
}
