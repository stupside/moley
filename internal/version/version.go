package version

// These variables are set at build time using -ldflags.
// Example:
// go build -ldflags "-X 'moley/internal/version.Version=1.2.3' -X 'moley/internal/version.Commit=abcdef' -X 'moley/internal/version.BuildTime=2023-01-01T00:00:00Z'"
var (
	Version   = "dev"
	Commit    = "none"
	BuildTime = "unknown"
)
