package version

// Version of the program. Is injected via ldflags during compilation with latest tag value.
var Version = "N/A"

// Get returns program version
func Get() string {
	return Version
}
