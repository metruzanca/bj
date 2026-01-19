//go:build ignore

package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var platforms = []struct {
	goos   string
	goarch string
	name   string
}{
	{"linux", "amd64", "bj-linux-x64"},
	{"linux", "arm64", "bj-linux-arm64"},
	{"darwin", "amd64", "bj-macos-x64"},
	{"darwin", "arm64", "bj-macos-arm64"},
}

func main() {
	tag := flag.String("tag", "", "Release tag (e.g. v0.1.0)")
	title := flag.String("title", "", "Release title")
	body := flag.String("body", "", "Release body/notes")
	updateLatest := flag.Bool("latest", true, "Also update the 'latest' release")
	flag.Parse()

	if *tag == "" {
		fatal("--tag is required")
	}
	if *title == "" {
		*title = *tag
	}
	if *body == "" {
		*body = fmt.Sprintf("Release %s", *tag)
	}

	// Create temp directory for binaries
	tmpDir, err := os.MkdirTemp("", "bj-release-*")
	if err != nil {
		fatal("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fmt.Println("Building binaries...")
	var binaries []string
	for _, p := range platforms {
		outPath := filepath.Join(tmpDir, p.name)
		binaries = append(binaries, outPath)

		fmt.Printf("  %s (%s/%s)\n", p.name, p.goos, p.goarch)
		cmd := exec.Command("go", "build", "-o", outPath, ".")
		cmd.Env = append(os.Environ(), "GOOS="+p.goos, "GOARCH="+p.goarch)
		if out, err := cmd.CombinedOutput(); err != nil {
			fatal("build failed for %s: %v\n%s", p.name, err, out)
		}
	}

	// Create and push tag
	fmt.Printf("Creating tag %s...\n", *tag)
	run("git", "tag", *tag)
	run("git", "push", "origin", *tag)

	// Create release
	fmt.Printf("Creating release %s...\n", *tag)
	args := []string{"release", "create", *tag, "--title", *title, "--notes", *body}
	args = append(args, binaries...)
	run("gh", args...)

	fmt.Printf("Release %s created successfully!\n", *tag)

	// Update latest release if requested
	if *updateLatest {
		fmt.Println("\nUpdating 'latest' release...")
		updateLatestRelease(*tag, binaries)
	}
}

func updateLatestRelease(mirrorTag string, binaries []string) {
	// Delete existing latest tag and release (ignore errors if they don't exist)
	exec.Command("gh", "release", "delete", "latest", "-y").Run()
	exec.Command("git", "push", "origin", "--delete", "latest").Run()
	exec.Command("git", "tag", "-d", "latest").Run()

	// Create new latest tag pointing to the same commit as the release tag
	run("git", "tag", "latest", mirrorTag)
	run("git", "push", "origin", "latest")

	// Create latest release
	args := []string{
		"release", "create", "latest",
		"--title", "Latest Release - Always Fresh",
		"--notes", fmt.Sprintf("This release always points to the latest version of bj. Currently mirroring %s.\n\nFor a quick and easy experience:\n```bash\nmise use -g github:metruzanca/bj\n```\n\nbj is always ready when you need it.", mirrorTag),
	}
	args = append(args, binaries...)
	run("gh", args...)

	fmt.Println("'latest' release updated!")
}

func run(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fatal("%s %s failed: %v", name, strings.Join(args, " "), err)
	}
}

func fatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(1)
}
