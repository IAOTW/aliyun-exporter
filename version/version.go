package version

import (
	"encoding/json"
	"fmt"
	"runtime"
)

var (
	buildDate = "unknown"
	gitCommit = "unknown"
	version   = "unknown"
)

// Version return version info
func Version() string {
	info := struct {
		BuildDate  string `json:"buildDate,omitempty"`
		Version    string `json:"version,omitempty"`
		GitVersion string `json:"gitVersion,omitempty"`
		GoVersion  string `json:"goVersion,omitempty"`
		Platform   string `json:"platform,omitempty"`
	}{buildDate, version, gitCommit, runtime.Version(), fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)}
	b, err := json.Marshal(info)
	if err != nil {
		return "unknown"
	}
	return string(b)
}
