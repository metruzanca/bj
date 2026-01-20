package tracker

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"syscall"
	"time"

	"github.com/metruzanca/bj/internal/config"
)

// ErrJobNotFound is returned when a job ID doesn't exist
var ErrJobNotFound = errors.New("job not found")

// Job represents a background job
type Job struct {
	ID        int        `json:"id"`
	Command   string     `json:"cmd"`
	PWD       string     `json:"pwd"`
	StartTime time.Time  `json:"start_time"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	ExitCode  *int       `json:"exit_code,omitempty"`
	LogFile   string     `json:"log_file"`
	PID       int        `json:"pid,omitempty"`
}

// Tracker manages job metadata
type Tracker struct {
	path     string
	lockPath string
}

// MaxJobHistory is the maximum number of completed jobs to retain
const MaxJobHistory = 100

// New creates a new Tracker
func New() (*Tracker, error) {
	configDir, err := config.ConfigDir()
	if err != nil {
		return nil, err
	}

	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, err
	}

	return &Tracker{
		path:     filepath.Join(configDir, "jobs.json"),
		lockPath: filepath.Join(configDir, "jobs.lock"),
	}, nil
}

// lock acquires a file lock for cross-process safety
func (t *Tracker) lock() (*os.File, error) {
	f, err := os.OpenFile(t.lockPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		f.Close()
		return nil, err
	}

	return f, nil
}

// unlock releases the file lock
func (t *Tracker) unlock(f *os.File) {
	syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	f.Close()
}

// load reads jobs from disk
func (t *Tracker) load() ([]Job, error) {
	data, err := os.ReadFile(t.path)
	if os.IsNotExist(err) {
		return []Job{}, nil
	}
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return []Job{}, nil
	}

	var jobs []Job
	if err := json.Unmarshal(data, &jobs); err != nil {
		return nil, err
	}
	return jobs, nil
}

// save writes jobs to disk
func (t *Tracker) save(jobs []Job) error {
	data, err := json.MarshalIndent(jobs, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(t.path, data, 0644)
}

// nextID returns the next available job ID
func (t *Tracker) nextID(jobs []Job) int {
	maxID := 0
	for _, j := range jobs {
		if j.ID > maxID {
			maxID = j.ID
		}
	}
	return maxID + 1
}

// Add creates a new job entry and returns its ID
func (t *Tracker) Add(cmd, pwd, logFile string) (int, error) {
	lockFile, err := t.lock()
	if err != nil {
		return 0, fmt.Errorf("failed to acquire lock: %w", err)
	}
	defer t.unlock(lockFile)

	jobs, err := t.load()
	if err != nil {
		return 0, fmt.Errorf("failed to load jobs: %w", err)
	}

	job := Job{
		ID:        t.nextID(jobs),
		Command:   cmd,
		PWD:       pwd,
		StartTime: time.Now(),
		LogFile:   logFile,
	}

	jobs = append(jobs, job)
	if err := t.save(jobs); err != nil {
		return 0, fmt.Errorf("failed to save jobs: %w", err)
	}

	return job.ID, nil
}

// Complete marks a job as completed with exit code and end time
func (t *Tracker) Complete(id int, exitCode int) error {
	lockFile, err := t.lock()
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	defer t.unlock(lockFile)

	jobs, err := t.load()
	if err != nil {
		return fmt.Errorf("failed to load jobs: %w", err)
	}

	for i := range jobs {
		if jobs[i].ID == id {
			now := time.Now()
			jobs[i].EndTime = &now
			jobs[i].ExitCode = &exitCode

			// Prune old completed jobs if we exceed MaxJobHistory
			jobs = t.pruneOldJobs(jobs)

			return t.save(jobs)
		}
	}

	return ErrJobNotFound
}

// List returns all jobs sorted by start time (newest first)
func (t *Tracker) List() ([]Job, error) {
	lockFile, err := t.lock()
	if err != nil {
		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}
	defer t.unlock(lockFile)

	jobs, err := t.load()
	if err != nil {
		return nil, fmt.Errorf("failed to load jobs: %w", err)
	}

	// Sort by start time descending
	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].StartTime.After(jobs[j].StartTime)
	})

	return jobs, nil
}

// Get returns a job by ID
func (t *Tracker) Get(id int) (*Job, error) {
	lockFile, err := t.lock()
	if err != nil {
		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}
	defer t.unlock(lockFile)

	jobs, err := t.load()
	if err != nil {
		return nil, fmt.Errorf("failed to load jobs: %w", err)
	}

	for _, j := range jobs {
		if j.ID == id {
			job := j // avoid returning pointer to loop variable
			return &job, nil
		}
	}

	return nil, nil
}

// Latest returns the most recently started job
func (t *Tracker) Latest() (*Job, error) {
	jobs, err := t.List()
	if err != nil {
		return nil, err
	}

	if len(jobs) == 0 {
		return nil, nil
	}

	return &jobs[0], nil
}

// UpdateLogPath updates the log file path for a job
func (t *Tracker) UpdateLogPath(id int, logPath string) error {
	lockFile, err := t.lock()
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	defer t.unlock(lockFile)

	jobs, err := t.load()
	if err != nil {
		return fmt.Errorf("failed to load jobs: %w", err)
	}

	for i := range jobs {
		if jobs[i].ID == id {
			jobs[i].LogFile = logPath
			if err := t.save(jobs); err != nil {
				return fmt.Errorf("failed to save jobs: %w", err)
			}
			return nil
		}
	}

	return ErrJobNotFound
}

// UpdatePID updates the process ID for a job
func (t *Tracker) UpdatePID(id int, pid int) error {
	lockFile, err := t.lock()
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	defer t.unlock(lockFile)

	jobs, err := t.load()
	if err != nil {
		return fmt.Errorf("failed to load jobs: %w", err)
	}

	for i := range jobs {
		if jobs[i].ID == id {
			jobs[i].PID = pid
			if err := t.save(jobs); err != nil {
				return fmt.Errorf("failed to save jobs: %w", err)
			}
			return nil
		}
	}

	return ErrJobNotFound
}

// Kill terminates a running job by sending SIGTERM to its process group
// Returns the job that was killed, or error if not found/not running
func (t *Tracker) Kill(id int) (*Job, error) {
	lockFile, err := t.lock()
	if err != nil {
		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}
	defer t.unlock(lockFile)

	jobs, err := t.load()
	if err != nil {
		return nil, fmt.Errorf("failed to load jobs: %w", err)
	}

	for i := range jobs {
		if jobs[i].ID == id {
			// Check if job is still running
			if jobs[i].ExitCode != nil {
				return nil, fmt.Errorf("job %d already finished", id)
			}

			if jobs[i].PID == 0 {
				return nil, fmt.Errorf("job %d has no PID recorded", id)
			}

			// Send SIGTERM to the process group (negative PID)
			// This kills the entire process tree since we use Setsid
			if err := syscall.Kill(-jobs[i].PID, syscall.SIGTERM); err != nil {
				return nil, fmt.Errorf("failed to terminate process: %w", err)
			}

			// Mark job as killed (exit code -15 = killed by SIGTERM)
			exitCode := -15
			now := time.Now()
			jobs[i].ExitCode = &exitCode
			jobs[i].EndTime = &now

			if err := t.save(jobs); err != nil {
				return nil, fmt.Errorf("failed to save jobs: %w", err)
			}

			return &jobs[i], nil
		}
	}

	return nil, ErrJobNotFound
}

// LatestRunning returns the most recently started job that is still running
func (t *Tracker) LatestRunning() (*Job, error) {
	jobs, err := t.List()
	if err != nil {
		return nil, err
	}

	for _, j := range jobs {
		if j.ExitCode == nil {
			job := j
			return &job, nil
		}
	}

	return nil, nil
}

// pruneOldJobs removes old completed jobs to keep history bounded
func (t *Tracker) pruneOldJobs(jobs []Job) []Job {
	// Count completed jobs
	var running, completed []Job
	for _, j := range jobs {
		if j.ExitCode == nil {
			running = append(running, j)
		} else {
			completed = append(completed, j)
		}
	}

	// If completed jobs exceed max, keep only the most recent
	if len(completed) > MaxJobHistory {
		// Sort completed by start time descending
		sort.Slice(completed, func(i, j int) bool {
			return completed[i].StartTime.After(completed[j].StartTime)
		})
		completed = completed[:MaxJobHistory]
	}

	// Combine running + retained completed
	result := append(running, completed...)
	return result
}

// Prune removes all done jobs (exit code 0), deletes their log files, and returns count pruned
func (t *Tracker) Prune() (int, error) {
	lockFile, err := t.lock()
	if err != nil {
		return 0, fmt.Errorf("failed to acquire lock: %w", err)
	}
	defer t.unlock(lockFile)

	jobs, err := t.load()
	if err != nil {
		return 0, fmt.Errorf("failed to load jobs: %w", err)
	}

	var kept []Job
	pruned := 0
	for _, j := range jobs {
		if j.ExitCode != nil && *j.ExitCode == 0 {
			// Delete the log file (ignore errors - file may already be gone)
			os.Remove(j.LogFile)
			pruned++
		} else {
			kept = append(kept, j)
		}
	}

	if err := t.save(kept); err != nil {
		return 0, fmt.Errorf("failed to save jobs: %w", err)
	}

	return pruned, nil
}

// GarbageCollect finds orphaned jobs (running but process is gone) and marks them as failed
// Returns the number of jobs cleaned up
func (t *Tracker) GarbageCollect() (int, error) {
	lockFile, err := t.lock()
	if err != nil {
		return 0, fmt.Errorf("failed to acquire lock: %w", err)
	}
	defer t.unlock(lockFile)

	jobs, err := t.load()
	if err != nil {
		return 0, fmt.Errorf("failed to load jobs: %w", err)
	}

	// Grace period: don't GC jobs started less than 5 seconds ago
	// This avoids false positives during the window between Add() and UpdatePID()
	gracePeriod := 5 * time.Second
	now := time.Now()

	collected := 0
	for i := range jobs {
		// Skip completed jobs
		if jobs[i].ExitCode != nil {
			continue
		}

		// Skip very recent jobs (might still be setting up)
		if now.Sub(jobs[i].StartTime) < gracePeriod {
			continue
		}

		// Check if process is still alive
		if jobs[i].PID > 0 {
			// Try to send signal 0 to check if process exists
			err := syscall.Kill(jobs[i].PID, 0)
			if err == nil {
				// Process still exists
				continue
			}
			// Process is gone - mark as failed (ruined)
		}

		// Mark as failed with exit code -1 (indicates abnormal termination)
		exitCode := -1
		jobs[i].ExitCode = &exitCode
		jobs[i].EndTime = &now
		collected++
	}

	if collected > 0 {
		if err := t.save(jobs); err != nil {
			return 0, fmt.Errorf("failed to save jobs: %w", err)
		}
	}

	return collected, nil
}

// PruneOlderThan removes done jobs (exit code 0) older than the given duration, deletes their log files
func (t *Tracker) PruneOlderThan(d time.Duration) (int, error) {
	lockFile, err := t.lock()
	if err != nil {
		return 0, fmt.Errorf("failed to acquire lock: %w", err)
	}
	defer t.unlock(lockFile)

	jobs, err := t.load()
	if err != nil {
		return 0, fmt.Errorf("failed to load jobs: %w", err)
	}

	cutoff := time.Now().Add(-d)
	var kept []Job
	pruned := 0
	for _, j := range jobs {
		// Prune if done (exit 0) and ended before cutoff
		if j.ExitCode != nil && *j.ExitCode == 0 && j.EndTime != nil && j.EndTime.Before(cutoff) {
			// Delete the log file (ignore errors - file may already be gone)
			os.Remove(j.LogFile)
			pruned++
		} else {
			kept = append(kept, j)
		}
	}

	if err := t.save(kept); err != nil {
		return 0, fmt.Errorf("failed to save jobs: %w", err)
	}

	return pruned, nil
}
