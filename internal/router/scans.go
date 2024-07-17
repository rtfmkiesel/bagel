package router

import (
	"bagel/internal/logger"
	"bagel/internal/semgrep"
	"fmt"
	"net/http"
	"os"
	"path"
	"slices"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var (
	// List of allowed MIME types for the uploaded files
	allowedArchiveMIMETypes = []string{"application/zip", "application/gzip", "application/x-tar", "application/x-bzip2"}
)

// listScans retrieves all scans from the database and displays them
func listScans(c *gin.Context) {
	var scans []semgrep.Scan
	db.Order("upload_date desc").Find(&scans)

	c.HTML(http.StatusOK, "scans.tmpl", gin.H{"Scans": scans, "Rulesets": semgrep.Rulesets})
}

// newScan accepts a POST request with a multipart form containing a file, name and ruleset.
// The file is saved to disk and the scan is added to the database
func newScan(c *gin.Context) {
	// Check if name is empty
	name := c.PostForm("name")
	if name == "" {
		c.String(http.StatusBadRequest, "Name cannot be empty")
		return
	}

	// Check if ruleset is valid
	rulesetStr := c.PostForm("ruleset")
	ruleset, ok := semgrep.Rulesets[rulesetStr]
	if !ok {
		c.String(http.StatusBadRequest, "Invalid ruleset, must be one of '%s'", semgrep.Rulesets)
		return
	}

	// Check if file is empty
	file, err := c.FormFile("file")
	if err != nil {
		c.String(http.StatusBadRequest, "File cannot be empty")
		return
	}

	fileReader, err := file.Open()
	if err != nil {
		c.String(http.StatusInternalServerError, "%s", err)
		return
	}
	defer fileReader.Close()

	// Check if its actually a supported archive
	mtype, err := mimetype.DetectReader(fileReader)
	if err != nil {
		c.String(http.StatusInternalServerError, "%s", err)
		return
	}
	if !slices.Contains(allowedArchiveMIMETypes, mtype.String()) {
		c.String(http.StatusBadRequest, "Filetype must be one of '%s': %s detected as %s", allowedArchiveMIMETypes, file.Filename, mtype.String())
		return
	}

	// Generate a new UUID for the scan
	// Do not use uuid.New() as it can panic
	var id uuid.UUID
	for {
		id, err = uuid.NewRandom()
		if err == nil {
			break
		}
	}

	uploadPath := path.Join(os.TempDir(), id.String()+mtype.Extension())
	unpackPath := path.Join(os.TempDir(), id.String())

	scan := semgrep.Scan{
		ID:           id,
		ScanName:     SanitizeHTML(name),
		Ruleset:      ruleset,
		UploadDate:   time.Now(),
		UploadName:   SanitizeHTML(file.Filename),
		UploadPath:   uploadPath,
		UnpackedPath: unpackPath,
		Finished:     false,
	}

	if err := c.SaveUploadedFile(file, scan.UploadPath); err != nil {
		c.String(http.StatusInternalServerError, "%s", err)
		return
	}
	logger.Info("Saved file %s", scan.UploadPath)

	db.Create(&scan)
	scan.AddToQueue()
	logger.Info("Created and added scan %s to queue", scan.ID.String())

	c.Redirect(http.StatusFound, "/")
}

// getScan retrieves a scan from the database and displays the results
func getScan(c *gin.Context) {
	id := c.Param("id")
	if err := validateID(id); err != nil {
		c.String(http.StatusBadRequest, "%s", err)
		return
	}

	// Retrieve scan from database
	var scan semgrep.Scan
	db.First(&scan, "id = ?", id)

	if scan.ID.String() == "" {
		c.String(http.StatusNotFound, "Scan not found")
		return
	}

	if !scan.Finished {
		c.String(http.StatusForbidden, "Scan not finished")
		return
	}

	// Only try to parse the Semgrep output if there was no error as with an error, there will likely be an unmarshalling error
	if scan.Error == "" {
		// Parse the Semgrep output into the struct temporarily
		if err := scan.ParseResults(); err != nil {
			c.String(http.StatusInternalServerError, "%s", err)
			return
		}
	}

	c.HTML(http.StatusOK, "scan.tmpl", gin.H{"Title": scan.ScanName, "Scan": scan})
}

// getScanJSON retrieves a scan from the database and returns the Semgrep output as JSON
func getScanJSON(c *gin.Context) {
	id := c.Param("id")
	if err := validateID(id); err != nil {
		c.String(http.StatusBadRequest, "%s", err)
		return
	}

	// Retrieve scan from database
	var scan semgrep.Scan
	db.First(&scan, "id = ?", id)

	if scan.ID.String() == "" {
		c.String(http.StatusNotFound, "Scan not found")
		return
	}

	if !scan.Finished {
		c.String(http.StatusForbidden, "Scan not finished")
		return
	}

	if scan.Error != "" {
		c.String(http.StatusInternalServerError, "Scan had an error: %s", scan.Error)
		return
	}

	// Can't use c.JSON as the Semgrep output is not a struct
	c.Data(http.StatusOK, "application/json", []byte(scan.SemgrepOutput))
}

// deleteScan removes a scan from the database
func deleteScan(c *gin.Context) {
	id := c.Param("id")
	if err := validateID(id); err != nil {
		c.String(http.StatusBadRequest, "%s", err)
		return
	}

	// Delete scan from database
	db.Delete(&semgrep.Scan{}, "id = ?", id)

	c.Redirect(http.StatusFound, "/")
}

// validateID checks if the given ID is a valid UUID
func validateID(id string) error {
	if id == "" {
		return fmt.Errorf("ID cannot be empty")
	}

	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("invalid ID: %s", id)
	}

	return nil
}
