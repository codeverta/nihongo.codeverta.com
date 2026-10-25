package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"nihongo-master-backend/internal/api"
)

type translationKey struct {
	sector string
	field  string
	source string
}

type translations struct {
	materials map[translationKey]string
	questions map[translationKey]string
	options   map[translationKey]string
}

func main() {
	sourceDir := flag.String("source", "../assets/content/ssw", "SSW source JSON directory")
	coverageDir := flag.String("coverage", "../docs/english-localization", "English coverage CSV directory")
	outputDir := flag.String("out", "../assets/content/ssw/en", "English SSW output directory")
	flag.Parse()

	if err := export(*sourceDir, *coverageDir, *outputDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func export(sourceDir, coverageDir, outputDir string) error {
	lookup, err := loadTranslations(coverageDir)
	if err != nil {
		return err
	}
	files, err := filepath.Glob(filepath.Join(sourceDir, "*.json"))
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("no SSW source JSON files found in %s", sourceDir)
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}
	for _, sourcePath := range files {
		sector := strings.TrimSuffix(filepath.Base(sourcePath), filepath.Ext(sourcePath))
		input, err := os.ReadFile(sourcePath)
		if err != nil {
			return err
		}
		var document any
		decoder := json.NewDecoder(strings.NewReader(string(input)))
		decoder.UseNumber()
		if err := decoder.Decode(&document); err != nil {
			return fmt.Errorf("decode %s: %w", sourcePath, err)
		}
		if err := translateValue(document, sector, nil, lookup); err != nil {
			return fmt.Errorf("translate %s: %w", sourcePath, err)
		}
		output, err := json.MarshalIndent(document, "", "  ")
		if err != nil {
			return fmt.Errorf("encode %s: %w", sourcePath, err)
		}
		output = append(output, '\n')
		destination := filepath.Join(outputDir, filepath.Base(sourcePath))
		if err := os.WriteFile(destination, output, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func loadTranslations(coverageDir string) (translations, error) {
	result := translations{
		materials: make(map[translationKey]string),
		questions: make(map[translationKey]string),
		options:   make(map[translationKey]string),
	}
	if err := readCSV(filepath.Join(coverageDir, "materials.csv"), func(row map[string]string) error {
		if row["category"] != "SSW curriculum" {
			return nil
		}
		return putTranslation(result.materials, translationKey{row["sector"], row["field"], row["source_text"]}, row["english_text"], row["status"])
	}); err != nil {
		return translations{}, err
	}
	questionSectors := make(map[string]string)
	if err := readCSV(filepath.Join(coverageDir, "questions.csv"), func(row map[string]string) error {
		if row["category"] != "SSW curriculum" && row["category"] != "SSW practice" {
			return nil
		}
		sector, questionID := normalizeSector(row["sector"]), row["question_id"]
		questionSectors[questionID] = sector
		for _, field := range []string{"prompt", "explanation"} {
			if err := putTranslation(result.questions, translationKey{sector, field, row[field+"_source"]}, row[field+"_english"], row[field+"_status"]); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return translations{}, err
	}
	if err := readCSV(filepath.Join(coverageDir, "options.csv"), func(row map[string]string) error {
		sector := questionSectors[row["question_id"]]
		if sector == "" {
			return nil
		}
		return putTranslation(result.options, translationKey{sector, "option", row["source_text"]}, row["english_text"], row["status"])
	}); err != nil {
		return translations{}, err
	}
	return result, nil
}

func normalizeSector(sector string) string {
	if sector == "nursing-care" {
		return "kaigo"
	}
	return sector
}

func putTranslation(target map[translationKey]string, key translationKey, english, status string) error {
	if key.sector == "" || key.source == "" {
		return nil
	}
	switch status {
	case "Exact mapping", "Already English", "Preserve Japanese":
	case "Needs translation":
		return fmt.Errorf("untranslated SSW source remains: sector=%s field=%s source=%q", key.sector, key.field, key.source)
	default:
		return nil
	}
	if english == "" {
		return fmt.Errorf("coverage row has no English value: sector=%s field=%s source=%q", key.sector, key.field, key.source)
	}
	if old, exists := target[key]; exists && old != english {
		return fmt.Errorf("conflicting English mappings for sector=%s field=%s source=%q", key.sector, key.field, key.source)
	}
	target[key] = english
	return nil
}

func readCSV(path string, visit func(map[string]string) error) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	header, err := reader.Read()
	if err != nil {
		return err
	}
	for {
		values, err := reader.Read()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		row := make(map[string]string, len(header))
		for index, name := range header {
			if index < len(values) {
				row[name] = values[index]
			}
		}
		if err := visit(row); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
	}
}

func translateValue(value any, sector string, parents []string, lookup translations) error {
	switch value := value.(type) {
	case map[string]any:
		for field, child := range value {
			if field == "options" {
				options, ok := child.([]any)
				if !ok {
					return fmt.Errorf("options must be an array, got %T", child)
				}
				for index, option := range options {
					source, ok := option.(string)
					if !ok {
						return fmt.Errorf("option %d must be a string, got %T", index, option)
					}
					translated, found := lookup.options[translationKey{sector, "option", source}]
					if !found {
						return fmt.Errorf("missing English option mapping for %q", source)
					}
					options[index] = translated
				}
				continue
			}
			if text, ok := child.(string); ok {
				translated, err := translateString(text, field, sector, parents, lookup)
				if err != nil {
					return err
				}
				value[field] = translated
				continue
			}
			if err := translateValue(child, sector, append(parents, field), lookup); err != nil {
				return err
			}
		}
	case []any:
		for _, child := range value {
			if err := translateValue(child, sector, parents, lookup); err != nil {
				return err
			}
		}
	}
	return nil
}

func translateString(source, field, sector string, parents []string, lookup translations) (string, error) {
	if field == "japanese" || field == "reading" || field == "url" || field == "id" || field == "kind" || field == "reviewed" {
		return source, nil
	}
	if field == "title" && contains(parents, "sources") {
		translated, ok := api.EnglishSSWSourceTitleTranslation(sector, source)
		if !ok {
			return "", fmt.Errorf("missing English source title mapping for %q", source)
		}
		return translated, nil
	}
	key := translationKey{sector, field, source}
	if field == "prompt" || field == "explanation" {
		if translated, ok := lookup.questions[key]; ok {
			return translated, nil
		}
	}
	if translated, ok := lookup.materials[key]; ok {
		return translated, nil
	}
	if localizedSSWField(field) {
		return "", fmt.Errorf("missing English mapping for field %s source %q", field, source)
	}
	return source, nil
}

func localizedSSWField(field string) bool {
	switch field {
	case "title", "description", "objective", "body", "step", "caution", "scenario", "response", "meaning", "prompt", "explanation":
		return true
	default:
		return false
	}
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
