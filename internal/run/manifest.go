package run

import "time"

type Manifest struct {
	ID        string      `json:"id"`
	StartedAt time.Time   `json:"startedAt"`
	EndedAt   time.Time   `json:"endedAt,omitempty"`
	Provider  string      `json:"provider,omitempty"`
	Command   string      `json:"command"`
	Root      string      `json:"root"`
	OK        bool        `json:"ok"`
	Steps     []StepRecord `json:"steps"`
}

type StepRecord struct {
	ID        string        `json:"id"`
	Title     string        `json:"title"`
	OK        bool          `json:"ok"`
	ExitCode  int           `json:"exitCode,omitempty"`
	StartedAt time.Time     `json:"startedAt"`
	EndedAt   time.Time     `json:"endedAt"`
	Duration  string        `json:"duration"`
	StdoutLog string        `json:"stdoutLog,omitempty"`
	StderrLog string        `json:"stderrLog,omitempty"`
	Error     string        `json:"error,omitempty"`
	Hint      *Hint         `json:"hint,omitempty"`
}

type Summary struct {
	OK       bool   `json:"ok"`
	Final    bool   `json:"final"`
	Command  string `json:"command"`
	Provider string `json:"provider,omitempty"`
	URL      string `json:"url,omitempty"`
	APIURL   string `json:"apiUrl,omitempty"`
	DocsURL  string `json:"docsUrl,omitempty"`
	Message  string `json:"message,omitempty"`
	RunDir   string `json:"runDir"`
	Duration string `json:"duration"`
}

type Failure struct {
	OK          bool   `json:"ok"`
	Final       bool   `json:"final"`
	Command     string `json:"command"`
	Provider    string `json:"provider,omitempty"`
	FailedStep  string `json:"failedStep"`
	ExitCode    int    `json:"exitCode,omitempty"`
	Message     string `json:"message"`
	Hint        *Hint  `json:"hint,omitempty"`
	LogPaths    Paths  `json:"logPaths"`
	ProviderRaw string `json:"providerRawTail,omitempty"`
	RunDir      string `json:"runDir"`
}

type Paths struct {
	Combined string `json:"combined,omitempty"`
	Stdout   string `json:"stdout,omitempty"`
	Stderr   string `json:"stderr,omitempty"`
}

type Hint struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Action  string `json:"action,omitempty"`
}
