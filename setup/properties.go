package setup

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"

	"github.com/gosimple/slug"
	"github.com/sfperusacdev/identitysdk/helpers/properties/models"
)

func findXMLFiles(fsys fs.FS) ([]string, error) {
	var xmlFiles []string
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Ext(path) == ".xml" {
			xmlFiles = append(xmlFiles, path)
		}
		return nil
	})
	return xmlFiles, err
}

func (s *Service) readProperties(fsys fs.FS) ([]models.DetailedSystemProperty, error) {
	files, err := findXMLFiles(fsys)
	if err != nil {
		return nil, err
	}

	var results = []models.DetailedSystemProperty{}

	for _, file := range files {
		data, err := fs.ReadFile(fsys, file)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", file, err)
		}
		var xmldata struct {
			SystemProperties []models.DetailedSystemProperty `xml:"SystemProperty"`
		}
		if err := xml.Unmarshal(data, &xmldata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal XML from %s: %w", file, err)
		}
		results = append(results, xmldata.SystemProperties...)
	}

	var valids = []string{
		"string",
		"number",
		"boolean",
		"array",
		"object",
	}
	for i := range results {
		results[i].ID = strings.TrimSpace(results[i].ID)
		results[i].Title = strings.TrimSpace(results[i].Title)
		results[i].Description = strings.TrimSpace(results[i].Description)
		results[i].Group = strings.TrimSpace(results[i].Group)
		results[i].Value = strings.TrimSpace(results[i].Value)
		results[i].Type = strings.TrimSpace(results[i].Type)
		if !slices.Contains(valids, results[i].Type) {
			results[i].Type = "string"
		}
	}
	return results, nil
}

func (s *Service) GenerateSystemProps(fsys fs.FS, pkg string, out io.Writer) error {
	entries, err := s.readProperties(fsys)
	if err != nil {
		return err
	}

	var str strings.Builder

	fmt.Fprintf(&str, "package %s\n\n", pkg)
	str.WriteString(`import "github.com/sfperusacdev/identitysdk/helpers/properties"`)
	str.WriteByte('\n')
	str.WriteByte('\n')

	length := len(entries)

	for i, entry := range entries {
		name := strings.ReplaceAll(slug.Make(entry.ID), "-", "_")
		name = strings.ToUpper(name)

		if entry.Description != "" {
			fmt.Fprintf(&str, "// %s\n", entry.Description)
		}
		if entry.Type != "" {
			fmt.Fprintf(&str, "// type: %s\n", entry.Type)
		}

		fmt.Fprintf(&str, `const %s properties.SystemProperty = "%s"`, name, entry.ID)

		if i != length-1 {
			str.WriteByte('\n')
			str.WriteByte('\n')
		}
	}

	_, err = fmt.Fprint(out, str.String())
	return err
}
