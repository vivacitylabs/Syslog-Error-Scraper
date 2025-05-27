package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

type PatternCheck struct {
	Pattern      *regexp.Regexp
	ErrorMessage string
	Count        int
	ExampleLine  string
}

const (
	currentLogFilePattern      = "syslog"
	uncompressedLogFilePattern = "syslog.[0-9]"
	compressedLogFilePattern   = "syslog.[0-9].gz"
	multivpRegex               = `^\w{3} \d{2} \d{2}:\d{2}:\d{2} hardware-\d+ docker-compose.*vision_pipeline.*address already in use`
	undistortRegex             = `^\w{3} \d{2} \d{2}:\d{2}:\d{2} hardware-\d+ docker-compose.*vision_pipeline.*perspective_transform config error. TransformationType needs to be DISTORTED OR UNDISTORTED`
	cartesianRegex             = `^\w{3} \d{2} \d{2}:\d{2}:\d{2} hardware-\d+ docker-compose.*vision_pipeline.*near miss calculator requires perspective transform's`
	gpsRegex                   = `^\w{3} \d{2} \d{2}:\d{2}:\d{2} hardware-\d+ docker-compose.*vision_pipeline.*error when excuting the template with pipelineValues: template: gstreamer_inference_sub_pipeline:26:52: executing "gstreamer_inference_sub_pipeline" at <.Sub.latitude>`
	kafkaRegex                 = `^.*Timed out \d+ in-flight, \d+ retry-queued, \d+ out-queue, \d+ partially-sent requests$`
)

var regexes = []string{
	multivpRegex,
	undistortRegex,
	cartesianRegex,
	gpsRegex,
}

type Scraper struct {
	dirPath      string
	logFilenames []string
}

func NewScraper(path string) *Scraper {
	return &Scraper{dirPath: path}
}

func (s *Scraper) ScrapeFiles() error {
	// list the log files
	compressedLogFiles, err := filepath.Glob(filepath.Join(s.dirPath, compressedLogFilePattern))
	if err != nil {
		return fmt.Errorf("error listing compressed log files: %w", err)
	}
	s.logFilenames = append(s.logFilenames, compressedLogFiles...)

	uncompressedLogFiles, err := filepath.Glob(filepath.Join(s.dirPath, uncompressedLogFilePattern))
	if err != nil {
		return fmt.Errorf("error listing uncompressed log files: %w", err)
	}
	s.logFilenames = append(s.logFilenames, uncompressedLogFiles...)

	currentLogFileFile, err := filepath.Glob(filepath.Join(s.dirPath, currentLogFilePattern))
	if err != nil {
		return fmt.Errorf("error listing current log file: %w", err)
	}
	s.logFilenames = append(s.logFilenames, currentLogFileFile...)

	// check that we have log files
	if len(s.logFilenames) == 0 {
		return fmt.Errorf("no log files found in %s", s.dirPath)
	}

	// sort in reverse order to have the oldest log file first
	slices.SortFunc(s.logFilenames, cmpLogFiles)
	return nil
}

func cmpLogFiles(a, b string) int {
	return -strings.Compare(strings.ToLower(a), strings.ToLower(b))
}

func (s *Scraper) LoadLogs() error {
	// load the compressed log files
	for _, filename := range s.logFilenames {
		err := s.loadLog(filename)
		if err != nil {
			return fmt.Errorf("error loading log file %s: %w", filename, err)
		}
	}
	return nil
}

func (s *Scraper) loadLog(filename string) error {
	// check that we can read the file
	_, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("cannot read file %s: %w", filename, err)
	}
	return nil
}

func filterFile2(filePath string, checks []PatternCheck) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		for i := range checks {
			if checks[i].Pattern.MatchString(line) {
				checks[i].Count++
				if checks[i].ExampleLine == "" {
					checks[i].ExampleLine = line
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	// After processing all lines, accumulate errors and matches
	return nil
}

func main() {

	checks := []PatternCheck{
		{Pattern: regexp.MustCompile(multivpRegex), ErrorMessage: "Port already in use."},
		{Pattern: regexp.MustCompile(undistortRegex), ErrorMessage: "Perspective transform config error."},
		{Pattern: regexp.MustCompile(cartesianRegex), ErrorMessage: "Near miss calculator error."},
		{Pattern: regexp.MustCompile(gpsRegex), ErrorMessage: "GPS template execution failed."},
		{Pattern: regexp.MustCompile(kafkaRegex), ErrorMessage: "Kafka error - try restarting the VP. if the issue persists, raise with tech"},
	}

	s := NewScraper(logDirPath) // Correct the logDirPath to your actual path

	err := s.ScrapeFiles()
	if err != nil {
		log.Fatalf("error scraping files: %v", err)
	}

	fmt.Println(s.logFilenames)

	// Process each file
	for _, filename := range s.logFilenames {
		err := filterFile2(filename, checks)
		if err != nil {
			log.Printf("Error processing file %s: %v", filename, err)
		}
	}

	// After processing all files, print the results
	var totalMatches int
	var totalErrors []string

	// Print the results for each error type
	for _, check := range checks {
		totalMatches += check.Count
		if check.Count > 1 {
			totalErrors = append(totalErrors, check.ErrorMessage)
		}

		// Print the total matches for this pattern
		fmt.Printf("\nTotal matches for pattern: %s\n", check.ErrorMessage)
		fmt.Printf("Matches found: %d\n", check.Count)

		// Print the first matching example if any
		if check.Count > 0 {
			fmt.Printf("Example match: %s\n", check.ExampleLine)
		}

		// Print error message if more than 1 match
		if check.Count > 1 {
			fmt.Printf("Error: %s\n", check.ErrorMessage)
		}
	}

	// Print total matches and errors across all files
	fmt.Printf("\nTotal matches across all files: %d\n", totalMatches)
	fmt.Printf("\nError messages across all files:\n")
	for _, errorMsg := range totalErrors {
		fmt.Println(errorMsg)
	}
}
