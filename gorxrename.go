package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"flag"
)

func loadRules(filePath string) (map[*regexp.Regexp]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	rules := make(map[*regexp.Regexp]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		rule_parts := strings.Split(scanner.Text(), "=>")
		if len(rule_parts) == 2 {
			re, err := regexp.Compile(rule_parts[0])
			if err != nil {
				return nil, err
			}
			rules[re] = rule_parts[1]
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return rules, nil
}

func loadIgnoreList(filePath string) ([]*regexp.Regexp, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var ignoreList []*regexp.Regexp
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var ignore = scanner.Text()
		if ignore == "" {
			continue
		}
		re, err := regexp.Compile(ignore)
		if err != nil {
			return nil, err
		}
		ignoreList = append(ignoreList, re)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return ignoreList, nil
}

func renameFilesAndContent(root string, rules map[*regexp.Regexp]string, ignoreList []*regexp.Regexp, dryRun bool) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		for _, ignore := range ignoreList {
			if ignore.MatchString(path) {
				if info.IsDir() {
					return filepath.SkipDir
				} else {
					return nil
				}
			}
		}

		if !info.IsDir() {
			oldContent, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			
			newContent := oldContent
			for re, replacement := range rules {
				newContent = re.ReplaceAll(newContent, []byte(replacement))
			}

			if string(oldContent) != string(newContent) {
				if dryRun {
					fmt.Printf("DRY RUN: Making substitutions in %s\n", path)
				} else {
					if err := ioutil.WriteFile(path, newContent, info.Mode()); err != nil {
						return err
					}
					fmt.Printf("Made substitutions in %s\n", path)
				}
			} 

			newName := info.Name()
			for re, replacement := range rules {
				newName = re.ReplaceAllString(newName, replacement)
			}
			if newName != info.Name() {
				newPath := filepath.Join(filepath.Dir(path), newName)
				if dryRun {
					fmt.Printf("DRY RUN: Would rename %s to %s\n", path, newPath)
				} else {
					if err := os.Rename(path, newPath); err != nil {
						return err
					}
					fmt.Printf("Renamed %s to %s\n", path, newPath)
				}
			}
		}
		return nil
	})
}

func main() {
	var dryRunFlag bool
	flag.BoolVar(&dryRunFlag, "n", false, "Performs a dry run without making actual changes")
	dirFlag := flag.String("dir", "", "Specifies the target directory")
	rulesFlag := flag.String("rules", "", "Specifies the file containing renaming rules")
	ignoreFlag := flag.String("ignore", "", "Specifies the file containing ignore rules")

	flag.Parse()

	if *dirFlag == "" || *rulesFlag == "" {
		fmt.Println("Usage: gorxrename -dir=<directory> -rules=<rules_file> [-ignore=<ignore_file>] [-n]")
		os.Exit(1)
	}

	rules, err := loadRules(*rulesFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load rules: %v\n", err)
		os.Exit(1)
	}

	var ignoreList []*regexp.Regexp
	if *ignoreFlag != "" {
		ignoreList, err = loadIgnoreList(*ignoreFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load ignore list: %v\n", err)
			os.Exit(1)
		}
	}

	if err := renameFilesAndContent(*dirFlag, rules, ignoreList, dryRunFlag); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to rename files and content: %v\n", err)
		os.Exit(1)
	}
}