package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/metruzanca/bj/internal/config"
	"github.com/metruzanca/bj/internal/runner"
	"github.com/metruzanca/bj/internal/tracker"
)

// ANSI color codes
const (
	colorReset = "\033[0m"
	colorDim   = "\033[2m"
	colorRed   = "\033[31m"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Create tracker
	t, err := tracker.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating tracker: %v\n", err)
		os.Exit(1)
	}

	// Auto-prune if configured
	if cfg.AutoPruneHours > 0 {
		t.PruneOlderThan(time.Duration(cfg.AutoPruneHours) * time.Hour)
	}

	// Parse first argument to determine action
	arg := os.Args[1]

	switch arg {
	case "-h", "--help":
		printUsage()
		os.Exit(0)

	case "-l", "--list":
		listJobs(t)

	case "--logs":
		var jobID int
		if len(os.Args) > 2 {
			id, err := strconv.Atoi(os.Args[2])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Invalid job ID: %s\n", os.Args[2])
				os.Exit(1)
			}
			jobID = id
		}
		viewLogs(cfg, t, jobID)

	case "--prune":
		pruneJobs(t)

	case "--complete":
		// Internal command: mark job as complete
		if len(os.Args) < 4 {
			fmt.Fprintf(os.Stderr, "Usage: bj --complete <job_id> <exit_code>\n")
			os.Exit(1)
		}
		jobID, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid job ID: %s\n", os.Args[2])
			os.Exit(1)
		}
		exitCode, err := strconv.Atoi(os.Args[3])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid exit code: %s\n", os.Args[3])
			os.Exit(1)
		}
		r := runner.New(cfg, t)
		if err := r.Complete(jobID, exitCode); err != nil {
			fmt.Fprintf(os.Stderr, "Error completing job: %v\n", err)
			os.Exit(1)
		}

	default:
		// Everything else is treated as a command to run
		command := strings.Join(os.Args[1:], " ")
		runCommand(cfg, t, command)
	}
}

func printUsage() {
	fmt.Println(`bj - Background Jobs

Usage:
  bj <command>      Run command in background
  bj -l, --list     List all jobs
  bj --logs [id]    View logs (latest job if no id specified)
  bj --prune        Clear all done jobs
  bj -h, --help     Show this help

Examples:
  bj sleep 10       Run 'sleep 10' in background
  bj npm install    Run 'npm install' in background
  bj -l             List all background jobs
  bj --logs         View logs of the latest job
  bj --logs 5       View logs of job #5
  bj --prune        Remove all successfully completed jobs`)
}

func runCommand(cfg *config.Config, t *tracker.Tracker, command string) {
	r := runner.New(cfg, t)
	jobID, err := r.Run(command)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error starting job: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("[%d] Started: %s\n", jobID, command)
}

type jobRow struct {
	id       int
	status   string
	start    string
	duration string
	cmd      string
	isError  bool
	isDone   bool
}

func listJobs(t *tracker.Tracker) {
	jobs, err := t.List()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing jobs: %v\n", err)
		os.Exit(1)
	}

	if len(jobs) == 0 {
		fmt.Println("No jobs found.")
		return
	}

	// Build rows first
	var rows []jobRow
	for _, job := range jobs {
		row := jobRow{id: job.ID}
		row.status = "running"
		row.duration = time.Since(job.StartTime).Round(time.Second).String()

		if job.ExitCode != nil {
			if *job.ExitCode == 0 {
				row.status = "done"
				row.isDone = true
			} else {
				row.status = fmt.Sprintf("exit(%d)", *job.ExitCode)
				row.isError = true
			}
			if job.EndTime != nil {
				row.duration = job.EndTime.Sub(job.StartTime).Round(time.Second).String()
			}
		}

		row.start = job.StartTime.Format("Jan 02 15:04")

		// Truncate long commands
		row.cmd = job.Command
		if len(row.cmd) > 40 {
			row.cmd = row.cmd[:37] + "..."
		}

		rows = append(rows, row)
	}

	// Calculate column widths
	idW, statusW, startW, durW := 2, 6, 5, 8 // header widths
	for _, r := range rows {
		if w := len(fmt.Sprintf("%d", r.id)); w > idW {
			idW = w
		}
		if w := len(r.status); w > statusW {
			statusW = w
		}
		if w := len(r.start); w > startW {
			startW = w
		}
		if w := len(r.duration); w > durW {
			durW = w
		}
	}

	// Print header
	fmt.Printf("%-*s  %-*s  %-*s  %-*s  %s\n", idW, "ID", statusW, "STATUS", startW, "START", durW, "DURATION", "COMMAND")

	// Print rows with colors
	for _, r := range rows {
		line := fmt.Sprintf("%-*d  %-*s  %-*s  %-*s  %s", idW, r.id, statusW, r.status, startW, r.start, durW, r.duration, r.cmd)
		if r.isError {
			// Dim row with red status
			statusStart := idW + 2
			statusEnd := statusStart + statusW
			fmt.Printf("%s%s%s%s%s%s%s\n",
				colorDim, line[:statusStart],
				colorRed, line[statusStart:statusEnd],
				colorReset, colorDim, line[statusEnd:]+colorReset)
		} else if r.isDone {
			fmt.Printf("%s%s%s\n", colorDim, line, colorReset)
		} else {
			fmt.Println(line)
		}
	}
}

func pruneJobs(t *tracker.Tracker) {
	count, err := t.Prune()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error pruning jobs: %v\n", err)
		os.Exit(1)
	}
	if count == 0 {
		fmt.Println("No done jobs to prune.")
	} else {
		fmt.Printf("Pruned %d done job(s).\n", count)
	}
}

func viewLogs(cfg *config.Config, t *tracker.Tracker, jobID int) {
	var job *tracker.Job
	var err error

	if jobID == 0 {
		job, err = t.Latest()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting latest job: %v\n", err)
			os.Exit(1)
		}
		if job == nil {
			fmt.Println("No jobs found.")
			os.Exit(0)
		}
	} else {
		job, err = t.Get(jobID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting job: %v\n", err)
			os.Exit(1)
		}
		if job == nil {
			fmt.Fprintf(os.Stderr, "Job %d not found.\n", jobID)
			os.Exit(1)
		}
	}

	// Check if log file exists
	if _, err := os.Stat(job.LogFile); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Log file not found: %s\n", job.LogFile)
		os.Exit(1)
	}

	// Open with configured viewer
	cmd := exec.Command(cfg.Viewer, job.LogFile)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error opening logs: %v\n", err)
		os.Exit(1)
	}
}
