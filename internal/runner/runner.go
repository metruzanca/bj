package runner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/metruzanca/bj/internal/config"
	"github.com/metruzanca/bj/internal/tracker"
)

// Runner handles spawning and tracking background jobs
type Runner struct {
	config  *config.Config
	tracker *tracker.Tracker
}

// New creates a new Runner
func New(cfg *config.Config, t *tracker.Tracker) *Runner {
	return &Runner{
		config:  cfg,
		tracker: t,
	}
}

// Run spawns a command in a detached background process
func (r *Runner) Run(command string) (int, error) {
	// Get current working directory
	pwd, err := os.Getwd()
	if err != nil {
		return 0, fmt.Errorf("failed to get working directory: %w", err)
	}

	// Ensure log directory exists
	if err := r.config.EnsureLogDir(); err != nil {
		return 0, fmt.Errorf("failed to create log directory: %w", err)
	}

	logDir, err := r.config.LogDirPath()
	if err != nil {
		return 0, fmt.Errorf("failed to get log directory: %w", err)
	}

	// Add job to tracker first to get ID (needed for log filename)
	jobID, err := r.tracker.Add(command, pwd, "")
	if err != nil {
		return 0, fmt.Errorf("failed to track job: %w", err)
	}

	// Create log file with timestamp and job ID
	timestamp := time.Now().Format("20060102-150405")
	logFileName := fmt.Sprintf("%s-%d.log", timestamp, jobID)
	logPath := filepath.Join(logDir, logFileName)

	// Update tracker with actual log path
	if err := r.tracker.UpdateLogPath(jobID, logPath); err != nil {
		return 0, fmt.Errorf("failed to update log path: %w", err)
	}

	// Create the log file
	logFile, err := os.Create(logPath)
	if err != nil {
		return 0, fmt.Errorf("failed to create log file: %w", err)
	}

	// Get user's shell from environment for running the command
	userShell := os.Getenv("SHELL")
	if userShell == "" {
		userShell = "/bin/sh"
	}

	// Get the path to our own executable
	selfPath, err := os.Executable()
	if err != nil {
		logFile.Close()
		return 0, fmt.Errorf("failed to get executable path: %w", err)
	}

	// Create a wrapper that:
	// 1. Runs the command in the user's shell
	// 2. Captures exit code
	// 3. Calls bj --complete
	// We use /bin/sh for the wrapper since it needs POSIX syntax for variable assignment
	wrapperCmd := fmt.Sprintf(`%s -c %s; exitcode=$?; %s --complete %d $exitcode`,
		userShell, shellQuote(command), shellQuote(selfPath), jobID)

	cmd := exec.Command("/bin/sh", "-c", wrapperCmd)
	cmd.Dir = pwd
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	// Detach the process so it survives parent exit
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}

	if err := cmd.Start(); err != nil {
		logFile.Close()
		return 0, fmt.Errorf("failed to start command: %w", err)
	}

	// Save PID for potential kill later
	if err := r.tracker.UpdatePID(jobID, cmd.Process.Pid); err != nil {
		// Non-fatal - job will still run, just can't be killed
		logFile.Close()
		return jobID, nil
	}

	// Close our handle to the log file - the child process has its own fd
	logFile.Close()

	return jobID, nil
}

// RunWithRetry spawns a command that will retry on failure
// maxAttempts of 0 means unlimited retries until success
// delaySecs is the delay between retries in seconds
func (r *Runner) RunWithRetry(command string, pwd string, maxAttempts int, delaySecs int) (int, error) {
	// Ensure log directory exists
	if err := r.config.EnsureLogDir(); err != nil {
		return 0, fmt.Errorf("failed to create log directory: %w", err)
	}

	logDir, err := r.config.LogDirPath()
	if err != nil {
		return 0, fmt.Errorf("failed to get log directory: %w", err)
	}

	// Add job to tracker first to get ID (needed for log filename)
	jobID, err := r.tracker.Add(command, pwd, "")
	if err != nil {
		return 0, fmt.Errorf("failed to track job: %w", err)
	}

	// Create log file with timestamp and job ID
	timestamp := time.Now().Format("20060102-150405")
	logFileName := fmt.Sprintf("%s-%d.log", timestamp, jobID)
	logPath := filepath.Join(logDir, logFileName)

	// Update tracker with actual log path
	if err := r.tracker.UpdateLogPath(jobID, logPath); err != nil {
		return 0, fmt.Errorf("failed to update log path: %w", err)
	}

	// Create the log file
	logFile, err := os.Create(logPath)
	if err != nil {
		return 0, fmt.Errorf("failed to create log file: %w", err)
	}

	// Get user's shell from environment for running the command
	userShell := os.Getenv("SHELL")
	if userShell == "" {
		userShell = "/bin/sh"
	}

	// Get the path to our own executable
	selfPath, err := os.Executable()
	if err != nil {
		logFile.Close()
		return 0, fmt.Errorf("failed to get executable path: %w", err)
	}

	// Create a wrapper that retries until success or max attempts
	// The wrapper script handles the retry logic
	var wrapperCmd string
	if maxAttempts == 0 {
		// Unlimited retries - keep going until success
		wrapperCmd = fmt.Sprintf(`
attempt=1
while true; do
  echo "=== Attempt $attempt ===" 
  %s -c %s
  exitcode=$?
  if [ $exitcode -eq 0 ]; then
    %s --complete %d 0
    exit 0
  fi
  echo "=== Attempt $attempt ruined (exit $exitcode), trying again... ===" 
  attempt=$((attempt + 1))
  sleep %d
done`,
			userShell, shellQuote(command), shellQuote(selfPath), jobID, delaySecs)
	} else {
		// Limited retries
		wrapperCmd = fmt.Sprintf(`
attempt=1
max=%d
while [ $attempt -le $max ]; do
  echo "=== Attempt $attempt of $max ===" 
  %s -c %s
  exitcode=$?
  if [ $exitcode -eq 0 ]; then
    %s --complete %d 0
    exit 0
  fi
  if [ $attempt -lt $max ]; then
    echo "=== Attempt $attempt ruined (exit $exitcode), trying again... ===" 
  fi
  attempt=$((attempt + 1))
  sleep %d
done
echo "=== All %d attempts ruined ===" 
%s --complete %d $exitcode`,
			maxAttempts, userShell, shellQuote(command), shellQuote(selfPath), jobID,
			delaySecs, maxAttempts, shellQuote(selfPath), jobID)
	}

	cmd := exec.Command("/bin/sh", "-c", wrapperCmd)
	cmd.Dir = pwd
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	// Detach the process so it survives parent exit
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}

	if err := cmd.Start(); err != nil {
		logFile.Close()
		return 0, fmt.Errorf("failed to start command: %w", err)
	}

	// Save PID for potential kill later
	if err := r.tracker.UpdatePID(jobID, cmd.Process.Pid); err != nil {
		// Non-fatal - job will still run, just can't be killed
		logFile.Close()
		return jobID, nil
	}

	// Close our handle to the log file - the child process has its own fd
	logFile.Close()

	return jobID, nil
}

// Complete marks a job as completed (called by the wrapper)
func (r *Runner) Complete(jobID int, exitCode int) error {
	return r.tracker.Complete(jobID, exitCode)
}

// RunWithRestart spawns a command that will restart on failure after a delay
// The command runs in an infinite loop: on non-zero exit, wait 5s and restart
func (r *Runner) RunWithRestart(command string, pwd string) (int, error) {
	// Ensure log directory exists
	if err := r.config.EnsureLogDir(); err != nil {
		return 0, fmt.Errorf("failed to create log directory: %w", err)
	}

	logDir, err := r.config.LogDirPath()
	if err != nil {
		return 0, fmt.Errorf("failed to get log directory: %w", err)
	}

	// Add job to tracker first to get ID (needed for log filename)
	jobID, err := r.tracker.Add(command, pwd, "")
	if err != nil {
		return 0, fmt.Errorf("failed to track job: %w", err)
	}

	// Create log file with timestamp and job ID
	timestamp := time.Now().Format("20060102-150405")
	logFileName := fmt.Sprintf("%s-%d.log", timestamp, jobID)
	logPath := filepath.Join(logDir, logFileName)

	// Update tracker with actual log path
	if err := r.tracker.UpdateLogPath(jobID, logPath); err != nil {
		return 0, fmt.Errorf("failed to update log path: %w", err)
	}

	// Create the log file
	logFile, err := os.Create(logPath)
	if err != nil {
		return 0, fmt.Errorf("failed to create log file: %w", err)
	}

	// Get user's shell from environment for running the command
	userShell := os.Getenv("SHELL")
	if userShell == "" {
		userShell = "/bin/sh"
	}

	// Get the path to our own executable
	selfPath, err := os.Executable()
	if err != nil {
		logFile.Close()
		return 0, fmt.Errorf("failed to get executable path: %w", err)
	}

	// Create a wrapper that restarts on failure with 5 second delay
	// Runs forever until success, then exits
	wrapperCmd := fmt.Sprintf(`
while true; do
  echo "=== Starting: $(date) ==="
  %s -c %s
  exitcode=$?
  if [ $exitcode -eq 0 ]; then
    %s --complete %d 0
    exit 0
  fi
  echo "=== Failed with exit $exitcode, restarting in 5s... ==="
  sleep 5
done`,
		userShell, shellQuote(command), shellQuote(selfPath), jobID)

	cmd := exec.Command("/bin/sh", "-c", wrapperCmd)
	cmd.Dir = pwd
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	// Detach the process so it survives parent exit
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}

	if err := cmd.Start(); err != nil {
		logFile.Close()
		return 0, fmt.Errorf("failed to start command: %w", err)
	}

	// Save PID for potential kill later
	if err := r.tracker.UpdatePID(jobID, cmd.Process.Pid); err != nil {
		// Non-fatal - job will still run, just can't be killed
		logFile.Close()
		return jobID, nil
	}

	// Close our handle to the log file - the child process has its own fd
	logFile.Close()

	return jobID, nil
}

// shellQuote properly quotes a string for shell execution
func shellQuote(s string) string {
	// Use single quotes and escape any single quotes in the string
	return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
}
