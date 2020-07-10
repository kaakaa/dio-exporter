package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Before running, YOU MUST generate latest binaries by `make dist` command
func TestFunctionality(t *testing.T) {
	binaryFile := fmt.Sprintf("../dist/dio-exporter-%s-amd64", runtime.GOOS)
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
		// For testing, only "svg" is supported as output format.
		// If using PNG, the content of the exported images will be dependent on the environment's font,
		// making it impossible to write portable tests.
		format string
	}{
		"single quote": {
			caseName: "single-quote",
			format:   "svg",
		},
		"two tabs": {
			caseName: "two-tabs",
			format:   "svg",
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

			var results map[string]string
			switch test.format {
			case "svg":
				results = compareSVG(testRoot, oracles, outputs)
			case "png":
				results = compareImage(testRoot, oracles, outputs)
			default:
				t.Fatalf("Invalid format: %s", test.format)
			}

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

func compareSVG(testRoot string, oracles, comparisons []string) map[string]string {
	var results = map[string]string{}
	for _, path := range oracles {
		if !includes(comparisons, path) {
			results[path] = "output file is missing"
			continue
		}

		oracleFile := filepath.Join(testRoot, "oracle", path)
		comparisonFile := filepath.Join(testRoot, "output", path)

		b1, err := ioutil.ReadFile(oracleFile)
		if err != nil {
			results[path] = "failed to read oracle file"
			continue
		}
		b2, err := ioutil.ReadFile(comparisonFile)
		if err != nil {
			results[path] = "failed to read comparison file"
			continue
		}

		if bytes.Compare(b1, b2) != 0 {
			results[path] = "two files are different"
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

// compare func compares oracles and actual outputs, and returns results for comparing as map
func compareImage(testRoot string, oracles, comparisons []string) map[string]string {
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
