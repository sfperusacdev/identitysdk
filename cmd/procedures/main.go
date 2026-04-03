package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/fatih/color"
	"github.com/sfperusacdev/identitysdk/utils/sqlproc"
	"github.com/sfperusacdev/identitysdk/utils/sqlreader"
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

	for _, file := range sqlFiles {
		if err := sqlproc.ValidateProcedureDefinition(file.Content); err != nil {
			fmt.Printf("invalid: %s: %v\n", file.Path, err)
			invalids++
			continue
		}
		valids++
	}

	title := color.New(color.Bold, color.FgCyan).SprintFunc()
	ok := color.New(color.Bold, color.FgGreen).SprintFunc()
	bad := color.New(color.Bold, color.FgRed).SprintFunc()
	warn := color.New(color.Bold, color.FgYellow).SprintFunc()

	fmt.Printf("\n%s\n", title("summary"))
	fmt.Printf("total: %d\n", total)
	fmt.Printf("%s %d\n", ok("ok:"), valids)
	fmt.Printf("%s %d\n", bad("invalid:"), invalids)

	if invalids == 0 {
		fmt.Println(ok("result: Success"))
		return
	}

	fmt.Println(warn("result: Invalid files found"))
}
