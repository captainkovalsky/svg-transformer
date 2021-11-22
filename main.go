package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

type SVGIconImport struct {
	filePath string
	importAs string
	svgIcon  string
}

// TODO: move to CLI params
func passFile(file string) bool {
	if !strings.HasSuffix(file, ".tsx") {
		return false
	}

	if strings.Contains(file, ".graphql.ts") || strings.Contains(file, ".spec.") || strings.Contains(file, "__mocks__") || strings.HasSuffix(file, ".scss") {
		return false
	}
	return true
}

func findSVGIconImports(
	files <-chan string,
	importsCollection chan<- SVGIconImport, r *regexp.Regexp, wg *sync.WaitGroup) {
	defer wg.Done()

	count := 0

	for file := range files {
		content, err := os.ReadFile(file)

		if err != nil {
			continue
		}

		result := r.FindAllStringSubmatch(string(content), -1)

		if len(result) == 0 {
			continue
		}

		names := r.SubexpNames()
		m := map[string]string{}
		count += len(result)

		fmt.Println(file, len(result))

		for index, _ := range result {
			for i, n := range result[index] {
				m[names[i]] = n
			}
			// fmt.Println("import", m["import"])
			// fmt.Println("svg file", m["svgFile"])
			// fmt.Print(string(content))
		}
	}

	fmt.Println("All matches", count)
}

func listFiles(dir string, files chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Print("working...")
	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}

		if info.IsDir() {
			return nil
		}

		if !passFile(path) {
			return nil
		}

		files <- path

		return nil
	})
	if err != nil {
		fmt.Printf("error walking the path")
		return
	}
	close(files)
}

func main() {
	fmt.Println("Processing files ...")
	var wg sync.WaitGroup
	var sourceDirParam string
	flag.StringVar(&sourceDirParam, "dir", "/home/viktor/projects/packui/src", "a string")
	flag.Parse()

	fmt.Println("source dir", sourceDirParam)

	files := make(chan string, 5)
	importsData := make(chan SVGIconImport, 3)
	r := regexp.MustCompile(`(?m)import { ReactComponent as (?P<import>\w+) } from "(?:.+)/(?P<svgFile>\w+).svg";$`)

	wg.Add(2)
	go listFiles(sourceDirParam, files, &wg)
	go findSVGIconImports(files, importsData, r, &wg)

	wg.Wait()
}
