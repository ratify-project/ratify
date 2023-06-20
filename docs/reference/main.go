package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
    "io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
    // TODO add arg for directory to scan
    dir, err := os.Getwd()
    if err != nil {
        fmt.Println("Error getting current directory", err)
        os.Exit(1)
    }

    markdownFiles := findMarkdownFiles(dir)

    for _, markdownFile := range markdownFiles {
        // Read the Markdown file
        file, err := os.Open(markdownFile)
        if err != nil {
            fmt.Println("Error opening file:", err)
            os.Exit(1)
        }
        defer file.Close()

        // Scan the file line by line
        scanner := bufio.NewScanner(file)
        lineNumber := 1
        for scanner.Scan() {
            line := scanner.Text()
            checkLinksInLine(markdownFile, line, lineNumber)
            lineNumber++
        }

        if err := scanner.Err(); err != nil {
            fmt.Println("Error scanning file:", err)
            os.Exit(1)
        }
    }
}

func checkLinksInLine(file,line string, lineNumber int) {
	// Find all the links in the line
	linkPattern := "\\[.*?\\]\\((.*?)\\)"
	links := findMatches(line, linkPattern)

	// Check each link for validity
	for _, link := range links {
		if !isLinkValid(link) {
			fmt.Printf("Broken link in %s at line %d: %s\n", file, lineNumber, link)
		}
	}
}

func findMatches(content string, pattern string) []string {
	re := regexp.MustCompile(pattern)
	matches := re.FindAllStringSubmatch(content, -1)
	results := make([]string, len(matches))
	for i, match := range matches {
		results[i] = match[1]
	}
	return results
}

func isLinkValid(link string) bool {
	if strings.HasPrefix(link, "http://") || strings.HasPrefix(link, "https://") {
		resp, err := http.Head(link)
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}

	// Check if the local file exists
	_, err := os.Stat(link)
	if err == nil {
		return true
	}

	// Check if the relative file path exists
	baseDir, _ := os.Getwd()
	filePath := filepath.Join(baseDir, link)
	_, err = os.Stat(filePath)
	return err == nil
}


func findMarkdownFiles(dir string) []string {
	var markdownFiles []string

	// Read the directory entries
	entries, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Println("Error reading directory:", err)
		os.Exit(1)
	}

	// Iterate over the entries
	for _, entry := range entries {
		if entry.IsDir() {
			// Recursive call for subdirectories
			subdir := filepath.Join(dir, entry.Name())
			subdirFiles := findMarkdownFiles(subdir)
			markdownFiles = append(markdownFiles, subdirFiles...)
		} else {
			// Check if the file has a .md or .markdown extension
			ext := filepath.Ext(entry.Name())
			if strings.EqualFold(ext, ".md") || strings.EqualFold(ext, ".markdown") {
				markdownFiles = append(markdownFiles, entry.Name())
			}
		}
	}

	return markdownFiles
}
