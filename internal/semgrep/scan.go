package semgrep

import (
	"bagel/internal/logger"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
)

// Scan represents a scan uploaded by the user
type Scan struct {
	ID            uuid.UUID `gorm:"type:text;primaryKey;"`            // The UUID of the scan
	ScanName      string    `gorm:"type:text"`                        // The name of the scan defined by the user
	Ruleset       Ruleset   `gorm:"embedded;embeddedPrefix:ruleset_"` // The ruleset used for the scan
	UploadDate    time.Time // The timestamp the scan was uploaded
	UploadName    string    `gorm:"type:text"`    // The name of the uploaded file, used for the front end
	UploadPath    string    `gorm:"type:text"`    // The path to the uploaded file (removed after unpkacing)
	UnpackedPath  string    `gorm:"type:text"`    // The path to the unpacked files
	Finished      bool      `gorm:"type:boolean"` // If the scan has finished
	Error         string    `gorm:"type:text"`    // If there were any errors during unpacking or scanning
	SemgrepOutput string    `gorm:"type:text"`    // The Semgrep output as JSON
	Results       []Result  `gorm:"-"`            // The parsed results (only used temporarily when rendering a page)
}

type semgrepResults struct {
	Results []Result `json:"results"`
}

// AddToQueue adds the given scan to the queue
func (s *Scan) AddToQueue() {
	logger.Info("Adding scan %s to queue", s.ID.String())
	chanJobs <- s
}

// ParseResults parses the Semgrep output into the []Result struct of s.Results
func (s *Scan) ParseResults() (err error) {
	logger.Info("Parsing results for scan %s", s.ID.String())

	semgrepResults := semgrepResults{}
	if err := json.Unmarshal([]byte(s.SemgrepOutput), &semgrepResults); err != nil {
		return fmt.Errorf("error unmarshalling JSON for scan %s: %s", s.ID.String(), err)
	}

	logger.Info("Got %d results for scan %s", len(semgrepResults.Results), s.ID.String())
	s.Results = semgrepResults.Results

	return nil
}

// runSemgrep runs Semgrep on a given scan
func (s *Scan) runSemgrep() (err error) {
	defer s.cleanup()

	// Unpack the file
	if err := s.unpack(); err != nil {
		return err
	}

	// Run semgrep on the unpacked directory
	cmdSemgrep := exec.Command("semgrep", "scan", "-q", "--metrics", "off", "--json", "--config", s.Ruleset.URL(), s.UnpackedPath) // #nosec G204, UnpackedPath does not contain user controllable data
	logger.Info("Running %s", cmdSemgrep.String())
	out, err := cmdSemgrep.Output()
	if err != nil {
		return err
	}

	// Unmarshall the output into a []Result struct
	// to verify the JSON
	semgrepResults := semgrepResults{}
	if err := json.Unmarshal(out, &semgrepResults); err != nil {
		return err
	}

	// Remove the UnpackedPath to normalize the paths
	outStr := strings.ReplaceAll(string(out), s.UnpackedPath+"/", "")

	// Store the JSON, as sqlite does not have support for arrays
	s.SemgrepOutput = outStr

	return nil
}

// unpack unpacks the file based on the file extension
func (s *Scan) unpack() (err error) {
	logger.Info("Unpacking %s", s.UploadPath)

	mtype, err := mimetype.DetectFile(s.UploadPath)
	if err != nil {
		return err
	}
	extension := mtype.Extension()

	// Create the output directory first as tar does not create it
	// #nosec G204, UnpackedPath does not contain user controllable data
	cmdMkdir := exec.Command("mkdir", "-p", s.UnpackedPath)
	if err := cmdMkdir.Run(); err != nil {
		return err
	}

	var cmdUnpack *exec.Cmd
	switch extension {
	case ".zip":
		// #nosec G204, UploadPath and UnpackedPath do not contain user controllable data
		cmdUnpack = exec.Command("unzip", s.UploadPath, "-d", s.UnpackedPath)
	case ".tar":
		// #nosec G204, UploadPath and UnpackedPath do not contain user controllable data
		cmdUnpack = exec.Command("tar", "-xf", s.UploadPath, "--directory", s.UnpackedPath)
	case ".gz":
		// #nosec G204, UploadPath and UnpackedPath do not contain user controllable data
		cmdUnpack = exec.Command("tar", "-xzf", s.UploadPath, "--directory", s.UnpackedPath)
	case ".bz2":
		// #nosec G204, UploadPath and UnpackedPath do not contain user controllable data
		cmdUnpack = exec.Command("tar", "-xjf", s.UploadPath, "--directory", s.UnpackedPath)
	default:
		// Technically not reachable
		return fmt.Errorf("unsupported file extension %s", extension)
	}

	logger.Info("Running %s", cmdUnpack.String())
	if err := cmdUnpack.Run(); err != nil {
		return err
	}

	return nil
}

// cleanup removes the original file and the unpacked directory.
// Do not return any errors as the cleanup should not fail the scan
func (s *Scan) cleanup() {
	// Remove the original file
	logger.Info("Removing %s", s.UploadPath)
	cmdRmOg := exec.Command("rm", s.UploadPath) // #nosec G204, UploadPath does not contain user controllable data
	if err := cmdRmOg.Run(); err != nil {
		logger.ErrorF("error removing file %s: %s", s.UploadPath, err)
	}

	// Remove the unpacked directory
	logger.Info("Removing %s", s.UnpackedPath)
	cmdRmUnpacked := exec.Command("rm", "-rf", s.UnpackedPath) // #nosec G204, UnpackedPath does not contain user controllable data
	if err := cmdRmUnpacked.Run(); err != nil {
		logger.ErrorF("error removing unpacked directory %s: %s", s.UnpackedPath, err)
	}
}
