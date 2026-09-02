package cli

import (
	"testing"
)

func TestPrintNonInteractiveHelp(t *testing.T) {
	// smoke: must not panic
	printNonInteractiveHelp()
}

func TestRunShipWorkflowRequiresTTY(t *testing.T) {
	cmd := NewRoot()
	// stdin/stdout in tests are not terminals
	err := runShipWorkflow(cmd)
	if err == nil {
		t.Fatal("expected error without interactive terminal")
	}
}
