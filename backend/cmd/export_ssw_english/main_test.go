package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestExportPreservesJapaneseAndCreatesEverySectorAsset(t *testing.T) {
	projectRoot := filepath.Clean(filepath.Join("..", "..", ".."))
	sourceDir := filepath.Join(projectRoot, "assets", "content", "ssw")
	coverageDir := filepath.Join(projectRoot, "docs", "english-localization")
	outputDir := t.TempDir()
	if err := export(sourceDir, coverageDir, outputDir); err != nil {
		t.Fatal(err)
	}

	sources, err := filepath.Glob(filepath.Join(sourceDir, "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	outputs, err := filepath.Glob(filepath.Join(outputDir, "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(sources) != 19 || len(outputs) != len(sources) {
		t.Fatalf("exported %d of %d source sector files; want 19 complete copies", len(outputs), len(sources))
	}
	for _, sourcePath := range sources {
		name := filepath.Base(sourcePath)
		englishPath := filepath.Join(outputDir, name)
		original, err := readJSON(sourcePath)
		if err != nil {
			t.Fatal(err)
		}
		translated, err := readJSON(englishPath)
		if err != nil {
			t.Fatalf("read generated %s: %v", name, err)
		}
		if err := checkPreservedFields(original, translated, nil); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
}

func readJSON(path string) (any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, err
	}
	return value, nil
}

func checkPreservedFields(original, translated any, path []string) error {
	switch original := original.(type) {
	case map[string]any:
		translatedMap, ok := translated.(map[string]any)
		if !ok || len(original) != len(translatedMap) {
			return fmt.Errorf("object shape changed at %v", path)
		}
		for key, sourceValue := range original {
			translatedValue, exists := translatedMap[key]
			if !exists {
				return fmt.Errorf("field %q disappeared at %v", key, path)
			}
			if err := checkPreservedFields(sourceValue, translatedValue, append(path, key)); err != nil {
				return err
			}
		}
	case []any:
		translatedList, ok := translated.([]any)
		if !ok || len(original) != len(translatedList) {
			return fmt.Errorf("array shape changed at %v", path)
		}
		for index := range original {
			if err := checkPreservedFields(original[index], translatedList[index], path); err != nil {
				return err
			}
		}
	case string:
		translatedText, ok := translated.(string)
		if !ok {
			return fmt.Errorf("string changed type at %v", path)
		}
		if len(path) == 0 {
			return nil
		}
		field := path[len(path)-1]
		switch field {
		case "japanese", "reading", "url", "id", "kind", "reviewed":
			if translatedText != original {
				return fmt.Errorf("preserved field %v changed from %q to %q", path, original, translatedText)
			}
		}
	}
	return nil
}
