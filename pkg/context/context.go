// Package context provides a shared context for the staraudit application.
package context

// Context represents the context of an StarAudit scan.
type Context struct {
	RepoOwner          string
	RepoName           string
	GithubToken        string
	CacheDirectoryPath string

	// ScanAll makes staraudit scan every stargazer
	// when set to true.
	ScanAll bool

	// Amount of stars to scan in fastMode.
	Stars uint

	// Verbose enables the verbose mode.
	Verbose bool
}
