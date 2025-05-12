package setup

import (
	"encoding/xml"
	"fmt"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"

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
