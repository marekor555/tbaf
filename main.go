package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

func handleMsg(msg string, err error, cleanup func()) {
	if err != nil {
		if cleanup != nil {
			cleanup()
		}
		fmt.Println(msg, err)
		log.Fatal(err)
	} else {
		fmt.Println(msg)
	}
}

func main() {
	var extraArgs []string
	var mainCmd string
	if len(os.Args) < 2 {
		mainCmd = "help"
		fmt.Println("No command provided, defaulting to help")
	} else {
		mainCmd = os.Args[1]
		extraArgs = os.Args[2:]
	}

	switch mainCmd {
	case "list", "l":
		files, msg, err := listArchive(extraArgs[0])
		if err != nil {
			fmt.Println(msg)
			log.Fatal(err)
		}
		for _, file := range files {
			fmt.Println(file)
		}
	case "build", "b":
		fmt.Println("building tbaf file...")
		if !strings.HasSuffix(extraArgs[0], ".tbaf") {
			fmt.Println("First argument should be the name of the archive, not a file to be added, add .tbaf suffix")
			fmt.Println("Adding .tbaf suffix to be sure...")
			extraArgs[0] += ".tbaf"
		}
		buildMsg, err := buildArchive(extraArgs[0], extraArgs[1:])
		handleMsg(buildMsg, err, nil)
	case "unpack", "u":
		unpackMsg, err := unpackArchive(extraArgs[0], extraArgs[1])
		handleMsg(unpackMsg, err, nil)
	case "compress", "c":
		fmt.Println("compressing tbaf file...")
		compressMsg, err := compressArchive(extraArgs[0])
		handleMsg(compressMsg, err, nil)
	case "decompress", "d":
		fmt.Println("decompressing tbaf file...")
		decompressMsg, err := decompressArchive(extraArgs[0])
		handleMsg(decompressMsg, err, nil)
	case "build-compress", "bc", "cb":
		cleanup := func() {
			err := os.Remove(strings.TrimSuffix(extraArgs[0], ".zst"))
			if err != nil {
				fmt.Println("Failed to remove archive file")
				log.Fatal(err)
			}
		}
		if !strings.HasSuffix(extraArgs[0], ".tbaf.zst") {
			fmt.Println("First argument should be the name of the compressed archive, not a file to be added, add .tbaf.zst suffix")
			fmt.Println("Adding .tbaf.zst suffix to be sure...")
			extraArgs[0] += ".tbaf.zst"
		}
		fmt.Println("building compressed tbaf file...")
		buildMsg, err := buildArchive(strings.Trim(extraArgs[0], ".zst"), extraArgs[1:])
		handleMsg(buildMsg, err, nil)
		compressMsg, err := compressArchive(strings.Trim(extraArgs[0], ".zst"))
		handleMsg(compressMsg, err, cleanup)
		cleanup()
	case "unpack-decompress", "ud", "du":
		cleanup := func() {
			err := os.Remove(strings.TrimSuffix(extraArgs[0], ".zst"))
			if err != nil {
				fmt.Println("Failed to remove decompressed file")
				log.Fatal(err)
			}
		}
		fmt.Println("unpacking and decompressing tbaf file...")
		decompressMsg, err := decompressArchive(extraArgs[0])
		handleMsg(decompressMsg, err, nil)

		unPackPrefix := ""
		if len(extraArgs) > 1 {
			unPackPrefix = extraArgs[1]
		}
		unpackMsg, err := unpackArchive(strings.TrimSuffix(extraArgs[0], ".zst"), unPackPrefix)
		handleMsg("check if file has .tbaf.zst extension\n"+unpackMsg, err, cleanup)
		cleanup()
	case "help", "h":
		fmt.Println("tbaf usage:")
		fmt.Println("	tbaf list <archive.tbaf> - List files in the archive")
		fmt.Println("	tbaf build <archive.tbaf> <file1> <file2> ... - Build an archive with the specified files")
		fmt.Println("	tbaf unpack <archive.tbaf> <prefix> - Unpack the archive to the specified prefix")
		fmt.Println("	tbaf compress <archive.tbaf> - Compress the archive using zstd")
		fmt.Println("	tbaf decompress <archive.tbaf.zst> - Decompress the archive using zstd")
		fmt.Println("	tbaf build-compress <archive.tbaf.zst> <file1> <file2> ... - Build and compress an archive")
		fmt.Println("	tbaf unpack-decompress <archive.tbaf.zst> <prefix> - Unpack and decompress the archive")
		fmt.Println("	tbaf help - Show this help message")
	default:
		fmt.Println("command not found")
	}
}
