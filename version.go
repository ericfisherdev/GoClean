package main

import (
	"fmt"
	"runtime"
)

// Version information - this will be set during build
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
	GoVersion = runtime.Version()
)

// VersionInfo contains all version-related information
type VersionInfo struct {
	Version   string `json:"version"`
	GitCommit string `json:"git_commit"`
	BuildDate string `json:"build_date"`
	GoVersion string `json:"go_version"`
	Platform  string `json:"platform"`
}

// GetVersionInfo returns structured version information
func GetVersionInfo() *VersionInfo {
	return &VersionInfo{
		Version:   Version,
		GitCommit: GitCommit,
		BuildDate: BuildDate,
		GoVersion: GoVersion,
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// GetVersionString returns a formatted version string
func GetVersionString() string {
	if GitCommit != "unknown" && len(GitCommit) > 7 {
		GitCommit = GitCommit[:7]
	}
	
	return fmt.Sprintf("GoClean %s (%s) built with %s on %s", 
		Version, GitCommit, GoVersion, BuildDate)
}