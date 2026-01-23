package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/metruzanca/bj/internal/config"
	"github.com/metruzanca/bj/internal/locales"
	"github.com/metruzanca/bj/internal/runner"
	"github.com/metruzanca/bj/internal/tracker"
)

// ANSI color codes
const (
	colorReset = "\033[0m"
	colorDim   = "\033[2m"
	colorRed   = "\033[31m"
)

// Global flags
var jsonOutput bool
var helpRequested bool
var retryFlag int    // -1 = not set, 0 = unlimited, N = max attempts
var retryJobID int   // 0 = not set (use latest), N = specific job ID
var retryDelay int   // delay in seconds between retries (default 1)

// List filter flags
var listRunning bool
var listFailed bool
var listDone bool

func main() {
	// Initialize retryFlag to -1 (not set) and delay to 1 second
	retryFlag = -1
	retryDelay = 1

	// Load config first so we can initialize locales (needed for help messages)
	cfg, err := config.Load()
	if err != nil {
		// Can't use locales.Msg yet, use a hardcoded message
		exitWithError(fmt.Sprintf("bj couldn't get comfortable: %v", err))
	}

	// Initialize locales based on config
	locales.Init(cfg.NSFW)

	// Check for --json, --help, --retry[=N], and --id flags anywhere in args
	args := filterArgs(os.Args[1:], &jsonOutput, &helpRequested, &retryFlag, &retryJobID)

	// Handle help for --retry
	if helpRequested && retryFlag >= 0 {
		printUsage("--retry")
		os.Exit(0)
	}

	// Handle general help
	if helpRequested || (len(args) < 1 && retryFlag < 0) {
		if len(args) > 0 {
			printUsage(args[0])
		} else {
			printUsage("")
		}
		os.Exit(0)
	}

	// Validate --id is only used with --retry
	if retryJobID != 0 && retryFlag < 0 {
		exitWithError(locales.Msg("err.id_only_with_retry"))
	}

	// Validate --delay is only used with --retry
	if retryDelay != 1 && retryFlag < 0 {
		exitWithError(locales.Msg("err.delay_only_with_retry"))
	}

	// Create tracker
	t, err := tracker.New()
	if err != nil {
		exitWithError(locales.Msg("err.tracker_init", err))
	}

	// Auto-prune if configured
	if cfg.AutoPruneHours > 0 {
		t.PruneOlderThan(time.Duration(cfg.AutoPruneHours) * time.Hour)
	}

	// Handle --retry as a modifier flag
	if retryFlag >= 0 {
		if len(args) > 0 {
			// Run new command with retry
			command := strings.Join(args, " ")
			runCommandWithRetry(cfg, t, command, retryFlag, retryDelay)
		} else {
			// Retry existing job
			retryExistingJob(cfg, t, retryJobID, retryFlag, retryDelay)
		}
		return
	}

	// Parse first argument to determine action
	arg := args[0]

	switch {
	case arg == "--man":
		printManPage()

	case arg == "--completion":
		if len(args) < 2 {
			exitWithError(locales.Msg("err.completion_usage"))
		}
		printCompletion(args[1])

	case arg == "--init":
		if len(args) < 2 {
			exitWithError(locales.Msg("err.init_usage"))
		}
		printInit(args[1], cfg, t)

	case arg == "--list":
		listJobs(t)

	case arg == "--ids":
		printJobIDs(t)

	case arg == "--logs":
		var jobID int
		if len(args) > 1 {
			id, err := strconv.Atoi(args[1])
			if err != nil {
				exitWithError(locales.Msg("err.invalid_number", args[1]))
			}
			jobID = id
		}
		viewLogs(cfg, t, jobID)

	case arg == "--prune":
		pruneJobs(t)

	case arg == "--gc":
		garbageCollect(t)

	case arg == "--kill":
		var jobID int
		if len(args) > 1 {
			id, err := strconv.Atoi(args[1])
			if err != nil {
				exitWithError(locales.Msg("err.invalid_number", args[1]))
			}
			jobID = id
		}
		killJob(t, jobID)

	case arg == "--complete":
		// Internal command: mark job as complete
		if len(args) < 3 {
			fmt.Fprintf(os.Stderr, "%s\n", locales.Msg("err.complete_usage"))
			os.Exit(1)
		}
		jobID, err := strconv.Atoi(args[1])
		if err != nil {
			exitWithError(locales.Msg("err.invalid_job_id", args[1]))
		}
		exitCode, err := strconv.Atoi(args[2])
		if err != nil {
			exitWithError(locales.Msg("err.invalid_exit_code", args[2]))
		}
		r := runner.New(cfg, t)
		if err := r.Complete(jobID, exitCode); err != nil {
			exitWithError(locales.Msg("err.complete_failed", err))
		}

	default:
		// Check for unknown flags - don't accidentally run them as commands
		if strings.HasPrefix(arg, "-") {
			exitWithError(locales.Msg("err.unknown_flag", arg))
		}
		// Everything else is treated as a command to run
		command := strings.Join(args, " ")
		runCommand(cfg, t, command)
	}
}

// filterArgs removes global flags, sets flag values, returns remaining args
func filterArgs(args []string, jsonFlag *bool, helpFlag *bool, retryFlagOut *int, retryJobIDOut *int) []string {
	var filtered []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--json":
			*jsonFlag = true
		case arg == "--help" || arg == "-h":
			*helpFlag = true
		case arg == "--retry":
			*retryFlagOut = 0 // unlimited
		case strings.HasPrefix(arg, "--retry="):
			val := strings.TrimPrefix(arg, "--retry=")
			n, err := strconv.Atoi(val)
			if err != nil || n < 1 {
				fmt.Fprintln(os.Stderr, locales.Msg("err.retry_positive_number", val))
				os.Exit(1)
			}
			*retryFlagOut = n
		case arg == "--id":
			// --id requires a following positive number
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, locales.Msg("err.id_needs_value"))
				os.Exit(1)
			}
			i++
			id, err := strconv.Atoi(args[i])
			if err != nil {
				fmt.Fprintln(os.Stderr, locales.Msg("err.invalid_job_id", args[i]))
				os.Exit(1)
			}
			if id < 1 {
				fmt.Fprintln(os.Stderr, locales.Msg("err.job_id_minimum", id))
				os.Exit(1)
			}
			*retryJobIDOut = id
		case arg == "--running":
			listRunning = true
		case arg == "--failed":
			listFailed = true
		case arg == "--done":
			listDone = true
		case arg == "--delay":
			// --delay requires a following number
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, locales.Msg("err.delay_needs_value"))
				os.Exit(1)
			}
			i++
			d, err := strconv.Atoi(args[i])
			if err != nil || d < 0 {
				fmt.Fprintln(os.Stderr, locales.Msg("err.delay_non_negative", args[i]))
				os.Exit(1)
			}
			retryDelay = d
		case strings.HasPrefix(arg, "--delay="):
			val := strings.TrimPrefix(arg, "--delay=")
			d, err := strconv.Atoi(val)
			if err != nil || d < 0 {
				fmt.Fprintln(os.Stderr, locales.Msg("err.delay_non_negative", val))
				os.Exit(1)
			}
			retryDelay = d
		default:
			filtered = append(filtered, arg)
		}
	}
	return filtered
}

// exitWithError prints error (or JSON) and exits
func exitWithError(msg string) {
	if jsonOutput {
		outputJSON(map[string]interface{}{"error": msg})
	} else {
		fmt.Fprintln(os.Stderr, msg)
	}
	os.Exit(1)
}

// outputJSON marshals and prints JSON
func outputJSON(v interface{}) {
	data, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(data))
}

func printUsage(command string) {
	switch command {
	case "--list":
		fmt.Println(locales.Msg("help.list"))
	case "--logs":
		fmt.Println(locales.Msg("help.logs"))
	case "--prune":
		fmt.Println(locales.Msg("help.prune"))
	case "--kill":
		fmt.Println(locales.Msg("help.kill"))
	case "--gc":
		fmt.Println(locales.Msg("help.gc"))
	case "--retry":
		fmt.Println(locales.Msg("help.retry"))
	case "--completion":
		fmt.Println(locales.Msg("help.completion"))
	case "--init":
		fmt.Println(locales.Msg("help.init"))
	default:
		fmt.Println(locales.Msg("help.main"))
	}
}

func runCommand(cfg *config.Config, t *tracker.Tracker, command string) {
	r := runner.New(cfg, t)
	jobID, err := r.Run(command)
	if err != nil {
		exitWithError(locales.Msg("err.run_failed", err))
	}
	if jsonOutput {
		outputJSON(map[string]interface{}{
			"id":      jobID,
			"command": command,
			"status":  "started",
		})
	} else {
		fmt.Println(locales.Msg("job.started", jobID, command))
	}
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

// relativeTime returns a human-friendly relative time string
func relativeTime(t time.Time) string {
	d := time.Since(t)

	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		mins := int(d.Minutes())
		if mins == 1 {
			return "1 min ago"
		}
		return fmt.Sprintf("%d mins ago", mins)
	case d < 24*time.Hour:
		hours := int(d.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case d < 7*24*time.Hour:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "yesterday"
		}
		return fmt.Sprintf("%d days ago", days)
	default:
		return t.Format("Jan 02")
	}
}

func listJobs(t *tracker.Tracker) {
	jobs, err := t.List()
	if err != nil {
		exitWithError(locales.Msg("err.list_failed", err))
	}

	// Apply filters if any are set
	hasFilter := listRunning || listFailed || listDone
	if hasFilter {
		var filtered []tracker.Job
		for _, job := range jobs {
			if listRunning && job.ExitCode == nil {
				filtered = append(filtered, job)
			} else if listFailed && job.ExitCode != nil && *job.ExitCode != 0 {
				filtered = append(filtered, job)
			} else if listDone && job.ExitCode != nil && *job.ExitCode == 0 {
				filtered = append(filtered, job)
			}
		}
		jobs = filtered
	}

	if len(jobs) == 0 {
		if jsonOutput {
			outputJSON([]interface{}{})
		} else {
			if hasFilter {
				fmt.Println(locales.Msg("list.empty_filtered"))
			} else {
				fmt.Println(locales.Msg("list.empty"))
			}
		}
		return
	}

	// JSON output - return raw job data
	if jsonOutput {
		outputJSON(jobs)
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

		row.start = relativeTime(job.StartTime)

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

// printJobIDs outputs job IDs for shell completion (no jq needed)
// Respects --running, --failed, --done filters
func printJobIDs(t *tracker.Tracker) {
	jobs, err := t.List()
	if err != nil {
		os.Exit(1) // Silent fail for completions
	}

	for _, job := range jobs {
		// Apply filters if any are set
		if listRunning && job.ExitCode != nil {
			continue
		}
		if listFailed && (job.ExitCode == nil || *job.ExitCode == 0) {
			continue
		}
		if listDone && (job.ExitCode == nil || *job.ExitCode != 0) {
			continue
		}
		fmt.Println(job.ID)
	}
}

func pruneJobs(t *tracker.Tracker) {
	count, err := t.Prune()
	if err != nil {
		exitWithError(locales.Msg("err.prune_failed", err))
	}
	if jsonOutput {
		outputJSON(map[string]interface{}{"pruned": count})
	} else if count == 0 {
		fmt.Println(locales.Msg("prune.nothing"))
	} else {
		fmt.Println(locales.Msg("prune.success", count))
	}
}

func garbageCollect(t *tracker.Tracker) {
	count, err := t.GarbageCollect()
	if err != nil {
		exitWithError(locales.Msg("err.gc_failed", err))
	}
	if jsonOutput {
		outputJSON(map[string]interface{}{"collected": count})
	} else if count == 0 {
		fmt.Println(locales.Msg("gc.nothing"))
	} else {
		fmt.Println(locales.Msg("gc.success", count))
	}
}

func killJob(t *tracker.Tracker, jobID int) {
	var job *tracker.Job
	var err error

	if jobID == 0 {
		// Find the most recent running job
		job, err = t.LatestRunning()
		if err != nil {
			exitWithError(locales.Msg("err.kill_check_failed", err))
		}
		if job == nil {
			if jsonOutput {
				exitWithError("no running jobs to kill")
			}
			fmt.Println(locales.Msg("kill.no_running"))
			os.Exit(0)
		}
		jobID = job.ID
	}

	job, err = t.Kill(jobID)
	if err != nil {
		exitWithError(locales.Msg("err.kill_failed", err))
	}

	if jsonOutput {
		outputJSON(map[string]interface{}{
			"id":      job.ID,
			"command": job.Command,
			"status":  "killed",
		})
	} else {
		fmt.Println(locales.Msg("job.killed", job.ID, job.Command))
	}
}

// runCommandWithRetry runs a new command with retry logic
func runCommandWithRetry(cfg *config.Config, t *tracker.Tracker, command string, maxAttempts int, delaySecs int) {
	pwd, err := os.Getwd()
	if err != nil {
		exitWithError(locales.Msg("err.retry_pwd_failed", err))
	}

	r := runner.New(cfg, t)
	jobID, err := r.RunWithRetry(command, pwd, maxAttempts, delaySecs)
	if err != nil {
		exitWithError(locales.Msg("err.run_failed", err))
	}

	if jsonOutput {
		outputJSON(map[string]interface{}{
			"id":           jobID,
			"command":      command,
			"status":       "started",
			"max_attempts": maxAttempts,
			"delay_secs":   delaySecs,
		})
	} else {
		if maxAttempts == 0 {
			fmt.Println(locales.Msg("job.retry_unlimited", jobID, command))
		} else if maxAttempts == 1 {
			fmt.Println(locales.Msg("job.retry_one", jobID, command))
		} else {
			fmt.Println(locales.Msg("job.retry_limited", jobID, maxAttempts, command))
		}
	}
}

// retryExistingJob retries a failed job (or most recent failure if jobID is 0)
func retryExistingJob(cfg *config.Config, t *tracker.Tracker, jobID int, maxAttempts int, delaySecs int) {
	var job *tracker.Job
	var err error

	if jobID == 0 {
		// Find the most recent failed job
		jobs, err := t.List()
		if err != nil {
			exitWithError(locales.Msg("err.retry_history_failed", err))
		}
		for _, j := range jobs {
			if j.ExitCode != nil && *j.ExitCode != 0 {
				job = &j
				break
			}
		}
		if job == nil {
			if jsonOutput {
				exitWithError("no ruined jobs to retry")
			}
			fmt.Println(locales.Msg("retry.no_failed"))
			os.Exit(0)
		}
	} else {
		job, err = t.Get(jobID)
		if err != nil {
			exitWithError(locales.Msg("err.retry_find_failed", err))
		}
		if job == nil {
			exitWithError(locales.Msg("err.job_not_found", jobID))
		}
	}

	// Check if job actually failed
	if job.ExitCode == nil {
		exitWithError(locales.Msg("err.job_still_running", job.ID))
	}
	if *job.ExitCode == 0 {
		exitWithError(locales.Msg("err.job_already_succeeded", job.ID))
	}

	// Run the job with retry wrapper
	r := runner.New(cfg, t)
	newJobID, err := r.RunWithRetry(job.Command, job.PWD, maxAttempts, delaySecs)
	if err != nil {
		exitWithError(locales.Msg("err.retry_start_failed", err))
	}

	if jsonOutput {
		outputJSON(map[string]interface{}{
			"id":           newJobID,
			"command":      job.Command,
			"status":       "started",
			"max_attempts": maxAttempts,
			"delay_secs":   delaySecs,
			"original_job": job.ID,
		})
	} else {
		if maxAttempts == 0 {
			fmt.Println(locales.Msg("job.retry_unlimited", newJobID, job.Command))
		} else if maxAttempts == 1 {
			fmt.Println(locales.Msg("job.retry_one_existing", newJobID, job.Command))
		} else {
			fmt.Println(locales.Msg("job.retry_limited", newJobID, maxAttempts, job.Command))
		}
	}
}

func viewLogs(cfg *config.Config, t *tracker.Tracker, jobID int) {
	var job *tracker.Job
	var err error

	if jobID == 0 {
		job, err = t.Latest()
		if err != nil {
			exitWithError(locales.Msg("err.logs_recall_failed", err))
		}
		if job == nil {
			if jsonOutput {
				exitWithError("no jobs found")
			}
			fmt.Println(locales.Msg("logs.no_jobs"))
			os.Exit(0)
		}
	} else {
		job, err = t.Get(jobID)
		if err != nil {
			exitWithError(locales.Msg("err.logs_find_failed", err))
		}
		if job == nil {
			exitWithError(locales.Msg("err.job_not_found", jobID))
		}
	}

	// Check if log file exists
	if _, err := os.Stat(job.LogFile); os.IsNotExist(err) {
		exitWithError(locales.Msg("err.logs_not_found", job.LogFile))
	}

	// JSON mode: output log contents
	if jsonOutput {
		content, err := os.ReadFile(job.LogFile)
		if err != nil {
			exitWithError(locales.Msg("err.logs_read_failed", err))
		}
		outputJSON(map[string]interface{}{
			"job":     job,
			"content": string(content),
		})
		return
	}

	// Open with configured viewer
	cmd := exec.Command(cfg.Viewer, job.LogFile)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		exitWithError(locales.Msg("err.logs_open_failed", err))
	}
}

func printCompletion(shell string) {
	switch shell {
	case "fish":
		fmt.Print(locales.Msg("completion.fish"))
	case "zsh":
		fmt.Print(locales.Msg("completion.zsh"))
	default:
		exitWithError(locales.Msg("err.unknown_shell", shell))
	}
}

func printInit(shell string, cfg *config.Config, t *tracker.Tracker) {
	// Run housekeeping silently on shell init
	// This is a good time to clean up orphaned jobs and auto-prune
	t.GarbageCollect()
	if cfg.AutoPruneHours > 0 {
		t.PruneOlderThan(time.Duration(cfg.AutoPruneHours) * time.Hour)
	}

	switch shell {
	case "fish":
		// Write completions file if needed, then output prompt init
		writeFishCompletions()
		fmt.Print(locales.Msg("init.fish"))
	case "zsh":
		// Write completions file if needed, then output prompt init
		writeZshCompletions()
		fmt.Print(locales.Msg("init.zsh"))
	default:
		exitWithError(locales.Msg("err.unknown_shell", shell))
	}
}

// writeFishCompletions writes fish completions to ~/.config/fish/completions/bj.fish
// Only writes if file doesn't exist or content has changed
func writeFishCompletions() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return // silently fail, completions are optional
	}

	completionsDir := filepath.Join(homeDir, ".config", "fish", "completions")
	completionsFile := filepath.Join(completionsDir, "bj.fish")

	completionContent := locales.Msg("completion.fish")

	// Check if content has changed
	existing, err := os.ReadFile(completionsFile)
	if err == nil && string(existing) == completionContent {
		return // already up to date
	}

	// Ensure directory exists
	if err := os.MkdirAll(completionsDir, 0755); err != nil {
		return // silently fail
	}

	// Write completions
	os.WriteFile(completionsFile, []byte(completionContent), 0644)
}

// writeZshCompletions writes zsh completions to ~/.zsh/completions/_bj
// Only writes if file doesn't exist or content has changed
func writeZshCompletions() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return // silently fail, completions are optional
	}

	completionsDir := filepath.Join(homeDir, ".zsh", "completions")
	completionsFile := filepath.Join(completionsDir, "_bj")

	completionContent := locales.Msg("completion.zsh")

	// Check if content has changed
	existing, err := os.ReadFile(completionsFile)
	if err == nil && string(existing) == completionContent {
		return // already up to date
	}

	// Ensure directory exists
	if err := os.MkdirAll(completionsDir, 0755); err != nil {
		return // silently fail
	}

	// Write completions
	os.WriteFile(completionsFile, []byte(completionContent), 0644)
}

func printManPage() {
	fmt.Print(locales.Msg("man.page"))
}
