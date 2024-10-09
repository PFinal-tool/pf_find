package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	progressbar "github.com/schollz/progressbar/v3"
	"os"
	"path/filepath"
	"strings"
)

func logo() {
	logo0 := `
	PFS - File Search
`
	fmt.Println(logo0)
}

func searchFiles(directory string, extensions []string) ([]string, error) {
	var files []string
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			for _, extension := range extensions {
				if strings.HasSuffix(info.Name(), extension) {
					files = append(files, path)
					break
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

func searchContent(filePath string, content string) ([][]string, error) {
	var matchingLines [][]string
	file, err := os.Open(filePath)
	if err != nil {
		return nil, errors.New("[-] Error opening file " + filePath)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	scanner := bufio.NewScanner(file)
	lineNum := 1
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, content) {
			matchingLines = append(matchingLines, []string{fmt.Sprint(lineNum), line})
		}
		lineNum++
	}
	if err := scanner.Err(); err != nil {
		return nil, errors.New("[-] Error reading file " + filePath)
	}
	return matchingLines, nil
}

func writeToFile(outputFile string, filePath string, matchingLines [][]string) error {
	file, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	writer := bufio.NewWriter(file)
	_, _ = fmt.Fprintf(writer, "[+] File Path: %s\n", filePath)
	_, _ = fmt.Fprintf(writer, "[=] Line Rows: %d\n", len(matchingLines))
	for _, line := range matchingLines {
		_, _ = fmt.Fprintf(writer, "[~] In Line %s: %s\n", line[0], strings.TrimSpace(line[1]))
	}
	_, _ = fmt.Fprintln(writer)

	return writer.Flush()
}

func main() {
	logo()

	// Parsing command line arguments
	name := flag.String("n", "", "Specify the suffix (required)")
	content := flag.String("c", "", "Specify file content (required)")
	outputFile := flag.String("o", "findout.txt", "Specify output file")
	directory := flag.String("d", "./", "Target directory")

	flag.Parse()

	if *name == "" || *content == "" {
		flag.Usage()
		os.Exit(1)
	}

	extensions := strings.Split(*name, ",")

	fmt.Println("[+] Running Search...")

	files, err := searchFiles(*directory, extensions)
	if err != nil {
		fmt.Println(err)
		return
	}

	bar := progressbar.Default(int64(len(files)))
	for _, filePath := range files {
		matchingLines, err := searchContent(filePath, *content)
		if err != nil {
			// fmt.Println(err)
			continue
		}
		if len(matchingLines) > 0 {
			err := writeToFile(*outputFile, filePath, matchingLines)
			if err != nil {
				fmt.Println("[-] Error writing to file:", err)
			}
		}
		bar.Add(1)
	}

	fmt.Println("[+] Output to findout.txt..")
}
