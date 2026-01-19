package tracker

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"sync"
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
}

// Tracker manages job metadata
type Tracker struct {
	mu       sync.Mutex
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
	t.mu.Lock()
	defer t.mu.Unlock()

	lockFile, err := t.lock()
	if err != nil {
		return 0, err
	}
	defer t.unlock(lockFile)

	jobs, err := t.load()
	if err != nil {
		return 0, err
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
		return 0, err
	}

	return job.ID, nil
}

// Complete marks a job as completed with exit code and end time
func (t *Tracker) Complete(id int, exitCode int) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	lockFile, err := t.lock()
	if err != nil {
		return err
	}
	defer t.unlock(lockFile)

	jobs, err := t.load()
	if err != nil {
		return err
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
	t.mu.Lock()
	defer t.mu.Unlock()

	lockFile, err := t.lock()
	if err != nil {
		return nil, err
	}
	defer t.unlock(lockFile)

	jobs, err := t.load()
	if err != nil {
		return nil, err
	}

	// Sort by start time descending
	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].StartTime.After(jobs[j].StartTime)
	})

	return jobs, nil
}

// Get returns a job by ID
func (t *Tracker) Get(id int) (*Job, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	lockFile, err := t.lock()
	if err != nil {
		return nil, err
	}
	defer t.unlock(lockFile)

	jobs, err := t.load()
	if err != nil {
		return nil, err
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

// Prune removes all done jobs (exit code 0) and returns count pruned
func (t *Tracker) Prune() (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	lockFile, err := t.lock()
	if err != nil {
		return 0, err
	}
	defer t.unlock(lockFile)

	jobs, err := t.load()
	if err != nil {
		return 0, err
	}

	var kept []Job
	pruned := 0
	for _, j := range jobs {
		if j.ExitCode != nil && *j.ExitCode == 0 {
			pruned++
		} else {
			kept = append(kept, j)
		}
	}

	if err := t.save(kept); err != nil {
		return 0, err
	}

	return pruned, nil
}

// PruneOlderThan removes done jobs (exit code 0) older than the given duration
func (t *Tracker) PruneOlderThan(d time.Duration) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	lockFile, err := t.lock()
	if err != nil {
		return 0, err
	}
	defer t.unlock(lockFile)

	jobs, err := t.load()
	if err != nil {
		return 0, err
	}

	cutoff := time.Now().Add(-d)
	var kept []Job
	pruned := 0
	for _, j := range jobs {
		// Prune if done (exit 0) and ended before cutoff
		if j.ExitCode != nil && *j.ExitCode == 0 && j.EndTime != nil && j.EndTime.Before(cutoff) {
			pruned++
		} else {
			kept = append(kept, j)
		}
	}

	if err := t.save(kept); err != nil {
		return 0, err
	}

	return pruned, nil
}
