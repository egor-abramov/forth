package main

import (
	"bytes"
	"flag"
	"forth/machine"
	"forth/translator"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yaml.in/yaml/v4"
)

var update = flag.Bool("update", false, "update golden files")

type GoldenFile struct {
	InSource   string `yaml:"in_source"`
	Input      string `yaml:"input"`
	OutLog     string `yaml:"out_log"`
	OutCodeLog string `yaml:"out_code_log"`
}

func cleanLog(input string) string {
	lines := strings.Split(input, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t\r")
	}
	return strings.Join(lines, "\n")
}

func TestTranslatorAndMachine(t *testing.T) {
	matches, err := filepath.Glob("golden/*.yml")
	if err != nil || len(matches) == 0 {
		t.Fatalf("can't find golden files: %v", err)
	}

	for _, goldenPath := range matches {
		t.Run(filepath.Base(goldenPath), func(t *testing.T) {
			yamlData, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("file reading error: %v", err)
			}

			var golden GoldenFile
			if err := yaml.Unmarshal(yamlData, &golden); err != nil {
				t.Fatalf("yaml parsing error: %v", err)
			}

			tmpDir := t.TempDir()
			source := filepath.Join(tmpDir, "source.fth")
			inp := filepath.Join(tmpDir, "input.txt")
			target := filepath.Join(tmpDir, "target.bin")

			if err := os.WriteFile(source, []byte(golden.InSource), 0644); err != nil {
				t.Fatalf("source writing error: %v", err)
			}
			if err := os.WriteFile(inp, []byte(golden.Input), 0644); err != nil {
				t.Fatalf("input writing error: %v", err)
			}

			var logBuf bytes.Buffer
			originalLogOutput := log.Writer()
			originalLogFlags := log.Flags()

			log.SetFlags(0)
			log.SetOutput(&logBuf)

			t.Cleanup(func() {
				log.SetOutput(originalLogOutput)
				log.SetFlags(originalLogFlags)
			})

			translator.Translate(source, target)
			machine.Simulate(target, inp, true, false)

			codeLogData, err := os.ReadFile(target + ".hex")
			if err != nil {
				t.Fatalf("can't read hex file: %v", err)
			}

			actualCodeLog := cleanLog(string(codeLogData))
			if actualCodeLog != golden.OutCodeLog {
				t.Errorf("mismatch in code log.\nExpected:\n%s\nGot:\n%s", golden.OutCodeLog, actualCodeLog)
			}

			actualLog := cleanLog(strings.TrimSpace(logBuf.String()))
			expectedLog := strings.TrimSpace(golden.OutLog)

			if *update {
				golden.OutCodeLog = cleanLog(actualCodeLog)
				golden.OutLog = cleanLog(actualLog)

				updatedYAML, err := yaml.Marshal(&golden)
				if err != nil {
					t.Fatalf("yaml updating error: %v", err)
				}
				if err := os.WriteFile(goldenPath, updatedYAML, 0644); err != nil {
					t.Fatalf("yaml saving error %s: %v", goldenPath, err)
				}
				t.Logf("test updated: %s", goldenPath)
				return
			}

			if actualLog != expectedLog {
				t.Errorf("mismatch in logs.\nExpected:\n%s\nGot:\n%s", expectedLog, actualLog)
			}
		})
	}
}
