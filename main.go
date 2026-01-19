package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/metruzanca/bj/internal/config"
	"github.com/metruzanca/bj/internal/runner"
	"github.com/metruzanca/bj/internal/tracker"
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
  bj -h, --help     Show this help

Examples:
  bj sleep 10       Run 'sleep 10' in background
  bj npm install    Run 'npm install' in background
  bj -l             List all background jobs
  bj --logs         View logs of the latest job
  bj --logs 5       View logs of job #5`)
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

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tSTATUS\tSTART\tDURATION\tCOMMAND")

	for _, job := range jobs {
		status := "running"
		duration := time.Since(job.StartTime).Round(time.Second).String()

		if job.ExitCode != nil {
			if *job.ExitCode == 0 {
				status = "done"
			} else {
				status = fmt.Sprintf("exit(%d)", *job.ExitCode)
			}
			if job.EndTime != nil {
				duration = job.EndTime.Sub(job.StartTime).Round(time.Second).String()
			}
		}

		startStr := job.StartTime.Format("Jan 02 15:04")
		
		// Truncate long commands
		cmd := job.Command
		if len(cmd) > 40 {
			cmd = cmd[:37] + "..."
		}

		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n", job.ID, status, startStr, duration, cmd)
	}
	w.Flush()
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
