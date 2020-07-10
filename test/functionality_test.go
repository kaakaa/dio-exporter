package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestFunctionality(t *testing.T) {
	binaryFile := "../dist/dio-exporter-darwin-amd64"
	binary, err := filepath.Abs(binaryFile)
	if err != nil {
		t.Fatalf("binary file is missing in %s", binaryFile)
	}
	dataDir := "./data"
	dataRoot, err := filepath.Abs(dataDir)
	if err != nil {
		t.Fatalf("data dir is missing in %s.", dataDir)
	}

	for name, test := range map[string]struct {
		caseName string
		format   string
	}{
		"two tabs": {
			caseName: "two-tabs",
			format:   "png",
		},
	} {
		t.Run(name, func(t *testing.T) {
			testRoot := filepath.Join(dataRoot, test.caseName)
			inDir := filepath.Join(testRoot, "input")
			oracleDir := filepath.Join(testRoot, "oracle")
			outDir := filepath.Join(testRoot, "output")
			diffDir := filepath.Join(testRoot, "diff")

			// clean up
			if err := os.RemoveAll(outDir); err != nil {
				log.Fatalf("failed to remove output dir %s", outDir)
			}
			if err := os.RemoveAll(diffDir); err != nil {
				log.Fatalf("failed to remove output dir %s", diffDir)
			}

			// export drawio files as other format
			cmd := exec.Command(binary, "-in", inDir, "-out", outDir, "-format", test.format)
			b, err := cmd.CombinedOutput()
			t.Log(string(b))
			if err != nil {
				t.Fatalf("exporting is failed. details: %v", err)
			}

			t.Logf("success to export %s files", test.format)

			// setup for comparing
			oracles, err := findFiles(oracleDir)
			if err != nil {
				t.Fatalf("an error occurs when finding files in %s", oracleDir)
			}
			outputs, err := findFiles(outDir)
			if err != nil {
				t.Fatalf("an error occurs when finding files in %s", outDir)
			}

			results := compare(testRoot, oracles, outputs)

			// count failures
			failCount := 0
			for k, v := range results {
				t.Logf("%s:\t\t%s", k, v)
				if v != "OK" {
					failCount++
				}
			}
			if failCount != 0 {
				t.Fatalf("%d cases is failed.", failCount)
			}
		})
	}
}

// findFiles returns paths of children files in root
func findFiles(root string) ([]string, error) {
	var results []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		results = append(results, rel)
		return nil
	})
	return results, err
}

// compare func compares oracles and actual outputs, and returns results for comparing as map
func compare(testRoot string, oracles, comparisons []string) map[string]string {
	var results = map[string]string{}
	for _, path := range oracles {
		if !includes(comparisons, path) {
			results[path] = "output file is missing"
			continue
		}

		oracleFile := filepath.Join(testRoot, "oracle", path)
		comparisonFile := filepath.Join(testRoot, "output", path)

		// Calculate how many pixels is difference between oracle and actual output
		cmd := exec.Command("node", "diff.js", "pixel", oracleFile, comparisonFile)
		outputDiffPixelCmd, err := cmd.CombinedOutput()
		if err != nil {
			results[path] = fmt.Sprintf("diff.js[pixel] returns error. details: %s.", string(outputDiffPixelCmd))
			continue
		}

		diffPixel := strings.Trim(string(outputDiffPixelCmd), "\n")
		if diffPixel != "0" {
			// if there is differrence of pixel, generate diff image
			diffFile := filepath.Join(testRoot, "diff", path)
			if err := os.MkdirAll(filepath.Dir(diffFile), os.ModePerm); err != nil {
				results[path] = fmt.Sprintf("failed to create dir for diff image %s", diffFile)
				continue
			}

			cmd = exec.Command("node", "diff.js", "image", oracleFile, comparisonFile, diffFile)
			outputDiffImageCmd, err := cmd.CombinedOutput()
			if err != nil {
				results[path] = fmt.Sprintf("diff.js[image] return error. details: %s.", string(outputDiffImageCmd))
			} else {
				results[path] = fmt.Sprintf("output image didn't match oracle. %s pixel was different. Diff image was generated in %s", diffPixel, diffFile)
			}
			continue
		}
		results[path] = "OK"
	}

	// Add results about the case that there is only output file (, not oracle)
	for _, c := range comparisons {
		if !includeAsKey(results, c) {
			results[c] = "oracle is missing."
		}
	}
	return results
}

func includes(arr []string, target string) bool {
	for _, v := range arr {
		if v == target {
			return true
		}
	}
	return false
}

func includeAsKey(m map[string]string, target string) bool {
	for k, _ := range m {
		if k == target {
			return true
		}
	}
	return false
}
