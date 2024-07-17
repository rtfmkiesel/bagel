package semgrep

import (
	"bagel/internal/logger"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"gorm.io/gorm"
)

var (
	chanJobs chan *Scan
	chanStop chan bool
	wgJobs   *sync.WaitGroup
)

const (
	workerCount = 3 // Hardcoded to 3 workers for now
)

// checkIfInstalled checks if Semgrep is installed
func checkIfInstalled() bool {
	logger.Info("Checking if Semgrep is installed")

	cmdHelp := exec.Command("semgrep", "--help")
	out, err := cmdHelp.Output()
	if err != nil {
		return false
	}

	if !strings.Contains(string(out), "Usage: semgrep") {
		return false
	}

	logger.Info("Semgrep is installed")
	return true
}

// checkForPro checks if an ENV variable is present to use Semgrep Pro
func checkForPro() (err error) {
	logger.Info("Checking for SEMGREP_APP_TOKEN")

	_, ok := os.LookupEnv("SEMGREP_APP_TOKEN")
	if !ok {
		// No ENV variable found, do not use Semgrep Pro
		return nil
	}

	logger.Info("Found token, logging into Semgrep using SEMGREP_APP_TOKEN")
	cmdLogin := exec.Command("semgrep", "login")
	if err := cmdLogin.Run(); err != nil {
		return fmt.Errorf("error logging into Semgrep Pro: %s", err)
	}

	logger.Info("Installing Semgrep Pro")
	cmdInstallPro := exec.Command("semgrep", "install-semgrep-pro")
	if err := cmdInstallPro.Run(); err != nil {
		return fmt.Errorf("error installing Semgrep Pro: %s", err)
	}

	logger.Info("Semgrep Pro installed")
	return nil
}

// StartWorkers starts the worker goroutines
func StartWorkers(db *gorm.DB) (err error) {
	if ok := checkIfInstalled(); !ok {
		return fmt.Errorf("semgrep is not installed or in $PATH")
	}

	if err := checkForPro(); err != nil {
		return err
	}

	chanJobs = make(chan *Scan)
	chanStop = make(chan bool, workerCount)
	wgJobs = new(sync.WaitGroup)

	for i := 0; i < workerCount; i++ {
		go func() {
			defer wgJobs.Done()
			logger.Info("Starting worker %d", i)

			for {
				select {
				case <-chanStop:
					// Stop signal received
					logger.Info("Stopping worker %d", i)
					return

				case job := <-chanJobs:
					logger.Info("Starting scan %s", job.ID.String())

					// Run the scan
					if err := job.runSemgrep(); err != nil {
						errMsg := fmt.Errorf("error running scan %s: %s", job.ID.String(), err)
						job.Error = errMsg.Error()
						logger.Error(errMsg)
					}

					job.Finished = true

					// Save the scan to the database
					if err := db.Save(job).Error; err != nil {
						logger.ErrorF("error saving scan %s: %s", job.ID.String(), err)
					}

					logger.Info("Finished scan %s", job.ID.String())
				}
			}
		}()

		wgJobs.Add(1)
	}

	return nil
}

// StopWorkers stops the worker goroutines
func StopWorkers() {
	logger.Info("Stopping workers")

	// Send the stop signal
	for i := 0; i <= workerCount; i++ {
		chanStop <- true
	}

	close(chanJobs)
	close(chanStop)

	// Wait for all workers to finish stopping
	wgJobs.Wait()

	logger.Info("Workers stopped")
}

// WaitForShutdown waits for a stop signal like SIGINT or SIGTERM, returns true when received
func WaitForShutdown() bool {
	// Set up the channel for the stop signal
	chanSignal := make(chan os.Signal, 1)
	signal.Notify(chanSignal, syscall.SIGINT, syscall.SIGTERM)

	<-chanSignal // Waits here until a stop signal is received
	logger.Info("Received stop signal")

	return true
}
