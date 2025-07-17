package main

import (
	"fmt"
	"github.com/klauspost/compress/zstd"
	"os"
	"strings"
)

func compressArchive(name string) (string, error) {
	if _, err := os.Stat(name + ".zst"); err == nil {
		return "Can't overwrite files, exiting", fmt.Errorf("file already exists: %s", name)
	}

	file, err := os.Open(name)
	if err != nil {
		return "Failed to open file: " + name, err
	}
	defer file.Close()

	compFile, err := os.Create(name + ".zst")
	if err != nil {
		return "Failed to create compressed file: " + name, err
	}
	defer compFile.Close()

	writer, err := zstd.NewWriter(compFile, zstd.WithEncoderLevel(zstd.EncoderLevelFromZstd(9)))
	if err != nil {
		return "Failed to create zstd writer", err
	}
	defer writer.Close()

	_, err = writer.ReadFrom(file)
	if err != nil {
		return "Failed to write compressed file", err
	}

	return "Compressed file written to: " + name + ".zst", nil
}

func decompressArchive(name string) (string, error) {
	file, err := os.Open(name)
	if err != nil {
		return "Failed to open file: " + name, err
	}
	defer file.Close()

	newFileName := strings.TrimSuffix(name, ".zst")
	decompFile, err := os.Create(newFileName)
	if err != nil {
		return "Failed to create decompressed file: " + newFileName, err
	}
	defer decompFile.Close()

	reader, err := zstd.NewReader(file)
	if err != nil {
		return "Failed to create zstd reader", err
	}
	defer reader.Close()

	_, err = decompFile.ReadFrom(reader)
	if err != nil {
		return "Failed to write decompressed file", err
	}

	return "Decompressed file written to: " + newFileName, nil
}
