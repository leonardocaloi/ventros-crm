package version

// Version is the current semantic version of Ventros CRM
const Version = "0.1.0"

// BuildDate is the date when the binary was built
// This can be overridden at build time using ldflags:
// go build -ldflags "-X github.com/ventros/crm/internal/version.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
var BuildDate = "2025-10-06"

// GitCommit is the git commit hash
// This can be overridden at build time using ldflags:
// go build -ldflags "-X github.com/ventros/crm/internal/version.GitCommit=$(git rev-parse HEAD)"
var GitCommit = "unknown"

// Info contains all version information
type Info struct {
	Version   string `json:"version"`
	BuildDate string `json:"build_date"`
	GitCommit string `json:"git_commit"`
	GoVersion string `json:"go_version"`
}

// GetInfo returns the version information
func GetInfo(goVersion string) Info {
	return Info{
		Version:   Version,
		BuildDate: BuildDate,
		GitCommit: GitCommit,
		GoVersion: goVersion,
	}
}

// GetVersion returns just the version string
func GetVersion() string {
	return Version
}

// GetFullVersion returns a formatted version string with all info
func GetFullVersion() string {
	return "Ventros CRM v" + Version + " (built " + BuildDate + ", commit " + GitCommit + ")"
}
