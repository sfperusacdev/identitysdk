package main

import (
	"fmt"
	"log/slog"
	"os"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/sfperusacdev/identitysdk/utils/sql/sqlproc"
	"github.com/sfperusacdev/identitysdk/utils/sql/sqlreader"
	"github.com/sfperusacdev/identitysdk/utils/sql/sqlutil"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("uso: app <path>")
		os.Exit(1)
	}

	sqlFiles, err := sqlreader.LoadSQLFilesFromPath(os.Args[1])
	if err != nil {
		slog.Error("load sql files", "error", err)
		os.Exit(1)
	}

	total := len(sqlFiles)
	if total == 0 {
		fmt.Println("no sql files found")
		return
	}

	valids := 0
	invalids := 0

	normalizedToPaths := make(map[string][]string)

	for _, file := range sqlFiles {
		if err := sqlproc.ValidateProcedureDefinition(file.Content); err != nil {
			fmt.Printf("invalid: %s: %v\n", file.Path, err)
			invalids++
			continue
		}

		name, err := sqlproc.ExtractProcedureName(file.Content)
		if err != nil {
			fmt.Printf("invalid: %s: %v\n", file.Path, err)
			invalids++
			continue
		}

		normalizedName, err := sqlutil.NormalizeSQLServerIdentifier(name)
		if err != nil {
			fmt.Printf("invalid: %s: %v\n", file.Path, err)
			invalids++
			continue
		}

		normalizedToPaths[normalizedName] = append(normalizedToPaths[normalizedName], file.Path)
		valids++
	}

	duplicates := make(map[string][]string)
	for normalizedName, paths := range normalizedToPaths {
		if len(paths) > 1 {
			sort.Strings(paths)
			duplicates[normalizedName] = paths
		}
	}

	title := color.New(color.Bold, color.FgCyan).SprintFunc()
	ok := color.New(color.Bold, color.FgGreen).SprintFunc()
	bad := color.New(color.Bold, color.FgRed).SprintFunc()
	warn := color.New(color.Bold, color.FgYellow).SprintFunc()

	fmt.Printf("\n%s\n", title("summary"))
	fmt.Printf("total: %d\n", total)
	fmt.Printf("%s %d\n", ok("ok:"), valids)
	fmt.Printf("%s %d\n", bad("invalid:"), invalids)
	fmt.Printf("%s %d\n", warn("duplicates:"), len(duplicates))

	if len(duplicates) > 0 {
		names := make([]string, 0, len(duplicates))
		for name := range duplicates {
			names = append(names, name)
		}
		sort.Strings(names)

		fmt.Printf("\n%s\n", title("duplicate procedures"))
		for _, name := range names {
			fmt.Printf("%s %s\n", bad("duplicate:"), name)
			for _, path := range duplicates[name] {
				fmt.Printf("  - %s\n", path)
			}
		}
	}

	if invalids == 0 && len(duplicates) == 0 {
		fmt.Println(ok("result: Success"))
		return
	}

	var reasons []string
	if invalids > 0 {
		reasons = append(reasons, "invalid files found")
	}
	if len(duplicates) > 0 {
		reasons = append(reasons, "duplicate procedures found")
	}

	fmt.Println(warn("result: " + strings.Join(reasons, ", ")))
	os.Exit(1)
}
