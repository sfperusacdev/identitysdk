package main

import (
	"bufio"
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed all:docs
var embeddedStandards embed.FS

func main() {
	outputPath := flag.String("out", ".estandares", "output directory")
	flag.Parse()

	if err := recreateDir(*outputPath); err != nil {
		exit("cleaning destination", err)
	}

	if err := extractDir(embeddedStandards, "docs", *outputPath); err != nil {
		exit("extracting content", err)
	}

	if err := ensureGitignoreEntry(".", *outputPath); err != nil {
		exit("updating .gitignore", err)
	}

	fmt.Printf("content extracted to: %s\n", *outputPath)
}

func recreateDir(path string) error {
	if err := os.RemoveAll(path); err != nil {
		return err
	}
	return os.MkdirAll(path, 0755)
}

func extractDir(source fs.FS, sourceDir, targetDir string) error {
	return fs.WalkDir(source, sourceDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(targetDir, relPath)

		if entry.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}

		content, err := fs.ReadFile(source, path)
		if err != nil {
			return err
		}

		return os.WriteFile(targetPath, content, 0444)
	})
}

func ensureGitignoreEntry(repoRoot, targetPath string) error {
	gitignorePath := filepath.Join(repoRoot, ".gitignore")

	entry := normalizeGitignoreEntry(targetPath)

	exists, err := fileExists(gitignorePath)
	if err != nil {
		return err
	}

	if !exists {
		return os.WriteFile(gitignorePath, []byte(entry+"\n"), 0644)
	}

	found, err := gitignoreContains(gitignorePath, entry)
	if err != nil {
		return err
	}

	if found {
		return nil
	}

	f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(entry + "\n")
	return err
}

func normalizeGitignoreEntry(p string) string {
	p = filepath.ToSlash(p)
	p = strings.TrimPrefix(p, "./")
	p = strings.TrimPrefix(p, "/")
	if !strings.HasSuffix(p, "/") {
		p += "/"
	}
	return p
}

func gitignoreContains(path, entry string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == entry {
			return true, nil
		}
	}

	return false, scanner.Err()
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func exit(action string, err error) {
	fmt.Fprintf(os.Stderr, "error %s: %v\n", action, err)
	os.Exit(1)
}
