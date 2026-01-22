package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/metruzanca/bj/internal/tracker"
)

var (
	bjPath    string
	buildOnce sync.Once
	buildErr  error

	// -update flag regenerates golden files
	update = flag.Bool("update", false, "update golden files")
)

// buildBinary builds the bj binary once for all tests
func buildBinary() (string, error) {
	buildOnce.Do(func() {
		bjPath = filepath.Join(os.TempDir(), "bj-test-binary")
		cmd := exec.Command("go", "build", "-o", bjPath, ".")
		cmd.Dir = "."
		if out, err := cmd.CombinedOutput(); err != nil {
			buildErr = &exec.ExitError{Stderr: out}
		}
	})
	return bjPath, buildErr
}

// testEnv sets up an isolated test environment with temp config dir
type testEnv struct {
	t         *testing.T
	configDir string
	bjPath    string
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()

	// Create temp directory for this test's config
	configDir, err := os.MkdirTemp("", "bj-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Build bj binary (once for all tests)
	bjPath, err := buildBinary()
	if err != nil {
		t.Fatalf("failed to build bj: %v", err)
	}

	env := &testEnv{
		t:         t,
		configDir: configDir,
		bjPath:    bjPath,
	}

	t.Cleanup(func() {
		os.RemoveAll(configDir)
	})

	return env
}

// run executes bj with args and returns stdout, stderr, and exit code
func (e *testEnv) run(args ...string) (stdout, stderr string, exitCode int) {
	e.t.Helper()

	cmd := exec.Command(e.bjPath, args...)
	cmd.Env = append(os.Environ(), "BJ_CONFIG_DIR="+e.configDir)

	var stdoutBuf, stderrBuf strings.Builder
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	exitCode = 0
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode = exitErr.ExitCode()
	} else if err != nil {
		e.t.Fatalf("failed to run bj: %v", err)
	}

	return stdoutBuf.String(), stderrBuf.String(), exitCode
}

// runAndWait runs a command and waits for it to complete
func (e *testEnv) runAndWait(args ...string) (stdout, stderr string, exitCode int) {
	e.t.Helper()
	stdout, stderr, exitCode = e.run(args...)

	// Wait for job to complete (poll --list --json)
	for i := 0; i < 50; i++ { // 5 second timeout
		time.Sleep(100 * time.Millisecond)
		listOut, _, _ := e.run("--list", "--json")
		var jobs []tracker.Job
		if err := json.Unmarshal([]byte(listOut), &jobs); err != nil {
			continue
		}
		if len(jobs) > 0 && jobs[0].ExitCode != nil {
			break
		}
	}
	return
}

// writeJobsFile writes a jobs.json file directly for testing
func (e *testEnv) writeJobsFile(jobs []tracker.Job) {
	e.t.Helper()
	data, err := json.MarshalIndent(jobs, "", "  ")
	if err != nil {
		e.t.Fatalf("failed to marshal jobs: %v", err)
	}
	if err := os.WriteFile(filepath.Join(e.configDir, "jobs.json"), data, 0644); err != nil {
		e.t.Fatalf("failed to write jobs.json: %v", err)
	}
}

// assertMatch checks stdout matches a regex pattern
func assertMatch(t *testing.T, output, pattern string) {
	t.Helper()
	re := regexp.MustCompile(pattern)
	if !re.MatchString(output) {
		t.Errorf("output did not match pattern\npattern: %s\noutput: %s", pattern, output)
	}
}

// assertContains checks stdout contains a substring
func assertContains(t *testing.T, output, substr string) {
	t.Helper()
	if !strings.Contains(output, substr) {
		t.Errorf("output did not contain expected substring\nexpected: %s\noutput: %s", substr, output)
	}
}

// assertExitCode checks the exit code
func assertExitCode(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("exit code = %d, want %d", got, want)
	}
}

// goldenFile compares output against a golden file, or updates it if -update flag is set
func goldenFile(t *testing.T, name, got string) {
	t.Helper()

	golden := filepath.Join("testdata", name+".golden")

	if *update {
		if err := os.MkdirAll("testdata", 0755); err != nil {
			t.Fatalf("failed to create testdata dir: %v", err)
		}
		if err := os.WriteFile(golden, []byte(got), 0644); err != nil {
			t.Fatalf("failed to write golden file: %v", err)
		}
		return
	}

	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("failed to read golden file %s (run with -update to create): %v", golden, err)
	}

	if got != string(want) {
		t.Errorf("output differs from golden file %s\n\ngot:\n%s\n\nwant:\n%s", golden, got, string(want))
	}
}

// =============================================================================
// Golden File Tests (Static Output Snapshots)
// =============================================================================

func TestHelp(t *testing.T) {
	env := newTestEnv(t)
	stdout, _, code := env.run("--help")
	assertExitCode(t, code, 0)
	goldenFile(t, "help", stdout)
}

func TestHelpList(t *testing.T) {
	env := newTestEnv(t)
	stdout, _, code := env.run("--list", "--help")
	assertExitCode(t, code, 0)
	goldenFile(t, "help-list", stdout)
}

func TestHelpLogs(t *testing.T) {
	env := newTestEnv(t)
	stdout, _, code := env.run("--logs", "--help")
	assertExitCode(t, code, 0)
	goldenFile(t, "help-logs", stdout)
}

func TestHelpKill(t *testing.T) {
	env := newTestEnv(t)
	stdout, _, code := env.run("--kill", "--help")
	assertExitCode(t, code, 0)
	goldenFile(t, "help-kill", stdout)
}

func TestHelpRetry(t *testing.T) {
	env := newTestEnv(t)
	stdout, _, code := env.run("--retry", "--help")
	assertExitCode(t, code, 0)
	goldenFile(t, "help-retry", stdout)
}

func TestHelpPrune(t *testing.T) {
	env := newTestEnv(t)
	stdout, _, code := env.run("--prune", "--help")
	assertExitCode(t, code, 0)
	goldenFile(t, "help-prune", stdout)
}

func TestHelpGC(t *testing.T) {
	env := newTestEnv(t)
	stdout, _, code := env.run("--gc", "--help")
	assertExitCode(t, code, 0)
	goldenFile(t, "help-gc", stdout)
}

func TestHelpCompletion(t *testing.T) {
	env := newTestEnv(t)
	stdout, _, code := env.run("--completion", "--help")
	assertExitCode(t, code, 0)
	goldenFile(t, "help-completion", stdout)
}

func TestHelpInit(t *testing.T) {
	env := newTestEnv(t)
	stdout, _, code := env.run("--init", "--help")
	assertExitCode(t, code, 0)
	goldenFile(t, "help-init", stdout)
}

func TestCompletionFish(t *testing.T) {
	env := newTestEnv(t)
	stdout, _, code := env.run("--completion", "fish")
	assertExitCode(t, code, 0)
	goldenFile(t, "completion-fish", stdout)
}

func TestCompletionZsh(t *testing.T) {
	env := newTestEnv(t)
	stdout, _, code := env.run("--completion", "zsh")
	assertExitCode(t, code, 0)
	goldenFile(t, "completion-zsh", stdout)
}

func TestInitFish(t *testing.T) {
	env := newTestEnv(t)
	stdout, _, code := env.run("--init", "fish")
	assertExitCode(t, code, 0)
	goldenFile(t, "init-fish", stdout)
}

func TestInitZsh(t *testing.T) {
	env := newTestEnv(t)
	stdout, _, code := env.run("--init", "zsh")
	assertExitCode(t, code, 0)
	goldenFile(t, "init-zsh", stdout)
}

// =============================================================================
// Error Case Tests
// =============================================================================

func TestDelayWithoutRetry(t *testing.T) {
	env := newTestEnv(t)

	_, stderr, code := env.run("--delay", "5", "echo", "test")
	assertExitCode(t, code, 1)
	assertContains(t, stderr, "--delay without --retry")
}

func TestIdWithoutRetry(t *testing.T) {
	env := newTestEnv(t)

	_, stderr, code := env.run("--id", "1", "echo", "test")
	assertExitCode(t, code, 1)
	assertContains(t, stderr, "--id only makes sense with --retry")
}

func TestInvalidId(t *testing.T) {
	env := newTestEnv(t)

	_, stderr, code := env.run("--retry", "--id", "0")
	assertExitCode(t, code, 1)
	assertContains(t, stderr, "Job IDs start at 1")
}

func TestInvalidIdNegative(t *testing.T) {
	env := newTestEnv(t)

	_, stderr, code := env.run("--retry", "--id", "-5")
	assertExitCode(t, code, 1)
	assertContains(t, stderr, "Job IDs start at 1")
}

func TestUnknownShell(t *testing.T) {
	env := newTestEnv(t)

	_, stderr, code := env.run("--completion", "bash")
	assertExitCode(t, code, 1)
	assertContains(t, stderr, "Unknown shell")
}

func TestNoArgs(t *testing.T) {
	env := newTestEnv(t)

	stdout, _, code := env.run()
	assertExitCode(t, code, 1)
	// Should print help on no args
	assertContains(t, stdout, "bj - Background Jobs")
}

func TestUnknownFlag(t *testing.T) {
	env := newTestEnv(t)

	_, stderr, code := env.run("--unknown-flag")
	assertExitCode(t, code, 1)
	assertContains(t, stderr, "Unknown flag")
}

func TestUnknownFlagShort(t *testing.T) {
	env := newTestEnv(t)

	_, stderr, code := env.run("-x")
	assertExitCode(t, code, 1)
	assertContains(t, stderr, "Unknown flag")
}

// =============================================================================
// Job Lifecycle Tests
// =============================================================================

func TestRunAndList(t *testing.T) {
	env := newTestEnv(t)

	// Run a quick command
	stdout, _, code := env.runAndWait("echo", "hello")
	assertExitCode(t, code, 0)
	assertMatch(t, stdout, `\[\d+\] bj is on it: echo hello`)

	// Check it appears in list
	listOut, _, _ := env.run("--list")
	assertContains(t, listOut, "echo hello")
	assertContains(t, listOut, "done")
}

func TestRunAndListJSON(t *testing.T) {
	env := newTestEnv(t)

	// Run a command
	env.runAndWait("echo", "json test")

	// Get JSON output
	stdout, _, code := env.run("--list", "--json")
	assertExitCode(t, code, 0)

	var jobs []tracker.Job
	if err := json.Unmarshal([]byte(stdout), &jobs); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}
	if jobs[0].Command != "echo json test" {
		t.Errorf("command = %q, want %q", jobs[0].Command, "echo json test")
	}
	if jobs[0].ExitCode == nil || *jobs[0].ExitCode != 0 {
		t.Errorf("expected exit code 0, got %v", jobs[0].ExitCode)
	}
}

func TestIds(t *testing.T) {
	env := newTestEnv(t)

	// Run a few commands
	env.runAndWait("echo", "one")
	env.runAndWait("false") // will fail
	env.runAndWait("echo", "two")

	// All IDs
	stdout, _, _ := env.run("--ids")
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 IDs, got %d: %v", len(lines), lines)
	}

	// Failed only
	stdout, _, _ = env.run("--ids", "--failed")
	lines = strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) != 1 {
		t.Errorf("expected 1 failed ID, got %d: %v", len(lines), lines)
	}

	// Done only
	stdout, _, _ = env.run("--ids", "--done")
	lines = strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 done IDs, got %d: %v", len(lines), lines)
	}
}

func TestLogs(t *testing.T) {
	env := newTestEnv(t)

	// Run a command that outputs something
	env.runAndWait("echo", "log output test")

	// Get logs as JSON (avoids opening less)
	stdout, _, code := env.run("--logs", "--json")
	assertExitCode(t, code, 0)

	var result struct {
		Job     tracker.Job `json:"job"`
		Content string      `json:"content"`
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	assertContains(t, result.Content, "log output test")
}

func TestPrune(t *testing.T) {
	env := newTestEnv(t)

	// Run successful and failed commands
	env.runAndWait("echo", "success")
	env.runAndWait("false") // fails

	// Verify both exist
	stdout, _, _ := env.run("--list", "--json")
	var jobs []tracker.Job
	json.Unmarshal([]byte(stdout), &jobs)
	if len(jobs) != 2 {
		t.Fatalf("expected 2 jobs before prune, got %d", len(jobs))
	}

	// Prune - should remove ALL completed jobs (both success and failed)
	stdout, _, code := env.run("--prune")
	assertExitCode(t, code, 0)
	assertMatch(t, stdout, `Wiped away 2 finished job`)

	// Verify all jobs are gone
	stdout, _, _ = env.run("--list", "--json")
	json.Unmarshal([]byte(stdout), &jobs)
	if len(jobs) != 0 {
		t.Fatalf("expected 0 jobs after prune, got %d", len(jobs))
	}
}

func TestPruneKeepsRunning(t *testing.T) {
	env := newTestEnv(t)

	// Start a long-running job
	env.run("sleep", "30")
	time.Sleep(200 * time.Millisecond)

	// Run a completed job
	env.runAndWait("echo", "done")

	// Prune - should only remove completed job
	stdout, _, code := env.run("--prune")
	assertExitCode(t, code, 0)
	assertMatch(t, stdout, `Wiped away 1 finished job`)

	// Verify running job remains
	stdout, _, _ = env.run("--list", "--json")
	var jobs []tracker.Job
	json.Unmarshal([]byte(stdout), &jobs)
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job after prune, got %d", len(jobs))
	}
	if jobs[0].ExitCode != nil {
		t.Error("expected remaining job to still be running")
	}

	// Clean up
	env.run("--kill")
}

func TestPruneResetsIDCounter(t *testing.T) {
	env := newTestEnv(t)

	// Run a job
	env.runAndWait("echo", "first")

	// Prune it
	env.run("--prune")

	// Run another job - ID should be 1 again
	stdout, _, _ := env.run("echo", "second")
	assertMatch(t, stdout, `\[1\] bj is on it`)
}

func TestPruneJSON(t *testing.T) {
	env := newTestEnv(t)

	env.runAndWait("echo", "to prune")

	stdout, _, code := env.run("--prune", "--json")
	assertExitCode(t, code, 0)

	var result struct {
		Pruned int `json:"pruned"`
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if result.Pruned != 1 {
		t.Errorf("pruned = %d, want 1", result.Pruned)
	}
}

// =============================================================================
// Kill Tests
// =============================================================================

func TestKill(t *testing.T) {
	env := newTestEnv(t)

	// Start a long-running job
	env.run("sleep", "30")

	// Give it a moment to start
	time.Sleep(200 * time.Millisecond)

	// Verify it's running
	listOut, _, _ := env.run("--ids", "--running")
	if strings.TrimSpace(listOut) == "" {
		t.Fatal("expected job to be running before kill")
	}

	// Kill it
	stdout, _, code := env.run("--kill")
	assertExitCode(t, code, 0)
	assertMatch(t, stdout, `\[\d+\] bj stopped abruptly: sleep 30`)

	// The kill should immediately mark the job as done
	listOut, _, _ = env.run("--ids", "--running")
	if strings.TrimSpace(listOut) != "" {
		t.Errorf("expected no running jobs after kill, got: %s", listOut)
	}
}

func TestKillNoRunning(t *testing.T) {
	env := newTestEnv(t)

	stdout, _, code := env.run("--kill")
	assertExitCode(t, code, 0)
	assertContains(t, stdout, "bj isn't doing anything right now")
}

func TestKillByID(t *testing.T) {
	env := newTestEnv(t)

	// Start a long-running job
	stdout, _, _ := env.run("sleep", "30")
	// Extract job ID from output like "[1] bj is on it: sleep 30"
	var jobID string
	if _, err := fmt.Sscanf(stdout, "[%s]", &jobID); err == nil {
		jobID = strings.TrimSuffix(jobID, "]")
	}

	time.Sleep(200 * time.Millisecond)

	// Kill by specific ID
	stdout, _, code := env.run("--kill", "1")
	assertExitCode(t, code, 0)
	assertMatch(t, stdout, `\[1\] bj stopped abruptly`)
}

func TestKillJSON(t *testing.T) {
	env := newTestEnv(t)

	// Start a long-running job
	env.run("sleep", "30")
	time.Sleep(200 * time.Millisecond)

	// Kill with JSON output
	stdout, _, code := env.run("--kill", "--json")
	assertExitCode(t, code, 0)

	var result struct {
		ID      int    `json:"id"`
		Command string `json:"command"`
		Status  string `json:"status"`
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if result.Status != "killed" {
		t.Errorf("status = %q, want %q", result.Status, "killed")
	}
	if result.Command != "sleep 30" {
		t.Errorf("command = %q, want %q", result.Command, "sleep 30")
	}
}

// =============================================================================
// GC Tests
// =============================================================================

func TestGCOrphanedJob(t *testing.T) {
	env := newTestEnv(t)

	// Manually create an orphaned job (running state, but PID doesn't exist)
	// Use PID 1 which we can't kill and will show as "exists" on most systems,
	// so use a very high PID that almost certainly doesn't exist
	startTime := time.Now().Add(-10 * time.Second) // older than grace period
	jobs := []tracker.Job{
		{
			ID:        1,
			Command:   "orphaned command",
			PWD:       "/tmp",
			StartTime: startTime,
			LogFile:   "/tmp/fake.log",
			PID:       999999, // almost certainly doesn't exist
		},
	}
	env.writeJobsFile(jobs)

	// Run GC
	stdout, _, code := env.run("--gc")
	assertExitCode(t, code, 0)
	assertMatch(t, stdout, `Found 1 ruined job`)

	// Verify job is now marked as failed
	listOut, _, _ := env.run("--list", "--json")
	var resultJobs []tracker.Job
	json.Unmarshal([]byte(listOut), &resultJobs)

	if len(resultJobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(resultJobs))
	}
	if resultJobs[0].ExitCode == nil {
		t.Error("expected job to have exit code after GC")
	} else if *resultJobs[0].ExitCode != -1 {
		t.Errorf("expected exit code -1, got %d", *resultJobs[0].ExitCode)
	}
}

func TestGCRespectsGracePeriod(t *testing.T) {
	env := newTestEnv(t)

	// Create a job that looks orphaned but is very recent (within grace period)
	jobs := []tracker.Job{
		{
			ID:        1,
			Command:   "recent job",
			PWD:       "/tmp",
			StartTime: time.Now(), // just now
			LogFile:   "/tmp/fake.log",
			PID:       999999,
		},
	}
	env.writeJobsFile(jobs)

	// Run GC
	stdout, _, code := env.run("--gc")
	assertExitCode(t, code, 0)
	assertContains(t, stdout, "No orphaned jobs found")

	// Verify job is still running
	listOut, _, _ := env.run("--list", "--json")
	var resultJobs []tracker.Job
	json.Unmarshal([]byte(listOut), &resultJobs)

	if resultJobs[0].ExitCode != nil {
		t.Error("expected job to still be running (within grace period)")
	}
}

func TestGCJSON(t *testing.T) {
	env := newTestEnv(t)

	startTime := time.Now().Add(-10 * time.Second)
	jobs := []tracker.Job{
		{
			ID:        1,
			Command:   "orphaned",
			PWD:       "/tmp",
			StartTime: startTime,
			LogFile:   "/tmp/fake.log",
			PID:       999999,
		},
	}
	env.writeJobsFile(jobs)

	stdout, _, code := env.run("--gc", "--json")
	assertExitCode(t, code, 0)

	var result struct {
		Collected int `json:"collected"`
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if result.Collected != 1 {
		t.Errorf("collected = %d, want 1", result.Collected)
	}
}

// =============================================================================
// Retry Tests
// =============================================================================

func TestRetryWithCommand(t *testing.T) {
	env := newTestEnv(t)

	// Run with retry (command that succeeds immediately)
	stdout, _, code := env.run("--retry", "echo", "retry test")
	assertExitCode(t, code, 0)
	assertMatch(t, stdout, `\[\d+\] bj will keep edging until it succeeds: echo retry test`)
}

func TestRetryWithLimit(t *testing.T) {
	env := newTestEnv(t)

	stdout, _, code := env.run("--retry=3", "echo", "limited")
	assertExitCode(t, code, 0)
	assertMatch(t, stdout, `\[\d+\] bj will tease up to 3 times before giving up: echo limited`)
}

func TestRetryExistingJob(t *testing.T) {
	env := newTestEnv(t)

	// Create a failed job
	env.runAndWait("false")

	// Retry it
	stdout, _, code := env.run("--retry")
	assertExitCode(t, code, 0)
	assertMatch(t, stdout, `\[\d+\] bj will keep edging until it succeeds: false`)
}

func TestRetryNoFailedJobs(t *testing.T) {
	env := newTestEnv(t)

	// Run a successful job
	env.runAndWait("echo", "success")

	// Try to retry (no failures)
	stdout, _, code := env.run("--retry")
	assertExitCode(t, code, 0)
	assertContains(t, stdout, "bj hasn't ruined anything yet")
}

func TestRetryWithID(t *testing.T) {
	env := newTestEnv(t)

	// Create two failed jobs
	env.runAndWait("false")
	env.runAndWait("sh", "-c", "exit 2")

	// Retry specific job by ID (job 1)
	stdout, _, code := env.run("--retry", "--id", "1")
	assertExitCode(t, code, 0)
	assertMatch(t, stdout, `\[\d+\] bj will keep edging until it succeeds: false`)
}

func TestRetryWithDelay(t *testing.T) {
	env := newTestEnv(t)

	// Run with retry and custom delay
	stdout, _, code := env.run("--retry=2", "--delay", "0", "echo", "quick")
	assertExitCode(t, code, 0)
	assertMatch(t, stdout, `\[\d+\] bj will tease up to 2 times before giving up: echo quick`)
}

func TestRetryJSON(t *testing.T) {
	env := newTestEnv(t)

	stdout, _, code := env.run("--retry=3", "--json", "echo", "test")
	assertExitCode(t, code, 0)

	var result struct {
		ID          int    `json:"id"`
		Command     string `json:"command"`
		Status      string `json:"status"`
		MaxAttempts int    `json:"max_attempts"`
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if result.MaxAttempts != 3 {
		t.Errorf("max_attempts = %d, want 3", result.MaxAttempts)
	}
	if result.Command != "echo test" {
		t.Errorf("command = %q, want %q", result.Command, "echo test")
	}
}

// =============================================================================
// List Filter Tests
// =============================================================================

func TestListFilters(t *testing.T) {
	env := newTestEnv(t)

	// Create jobs in different states
	env.runAndWait("echo", "success1")
	env.runAndWait("false") // fails
	env.runAndWait("echo", "success2")

	// Test --done filter
	stdout, _, _ := env.run("--list", "--done")
	assertContains(t, stdout, "success1")
	assertContains(t, stdout, "success2")
	if strings.Contains(stdout, "false") {
		t.Error("--done should not show failed jobs")
	}

	// Test --failed filter
	stdout, _, _ = env.run("--list", "--failed")
	assertContains(t, stdout, "false")
	if strings.Contains(stdout, "success") {
		t.Error("--failed should not show successful jobs")
	}
}

func TestListRunningFilter(t *testing.T) {
	env := newTestEnv(t)

	// Start a long-running job
	env.run("sleep", "30")
	time.Sleep(200 * time.Millisecond)

	// Also run a completed job
	env.runAndWait("echo", "done")

	// Test --running filter
	stdout, _, _ := env.run("--list", "--running")
	assertContains(t, stdout, "sleep 30")
	if strings.Contains(stdout, "echo done") {
		t.Error("--running should not show completed jobs")
	}

	// Clean up
	env.run("--kill")
}

func TestListEmpty(t *testing.T) {
	env := newTestEnv(t)

	stdout, _, code := env.run("--list")
	assertExitCode(t, code, 0)
	assertContains(t, stdout, "bj has nothing going on")
}

func TestListFilterEmpty(t *testing.T) {
	env := newTestEnv(t)

	env.runAndWait("echo", "success")

	stdout, _, _ := env.run("--list", "--failed")
	assertContains(t, stdout, "No jobs match your criteria")
}
