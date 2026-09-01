package run

// Session carries cross-provider state during a deploy run.
type Session struct {
	RunDir string
	APIURL string
}
