// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0
// Copyright (c) 2023 HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package command

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zclconf/go-cty/cty"

	"github.com/opentofu/opentofu/internal/addrs"
	"github.com/opentofu/opentofu/internal/states"
)

func TestOutput(t *testing.T) {
	originalState := states.BuildState(func(s *states.SyncState) {
		s.SetOutputValue(
			addrs.OutputValue{Name: "foo"}.Absolute(addrs.RootModuleInstance),
			cty.StringVal("bar"),
			false,
			"",
		)
	})

	statePath := testStateFile(t, originalState)

	view, done := testView(t)
	c := &OutputCommand{
		Meta: Meta{
			testingOverrides: metaOverridesForProvider(testProvider()),
			View:             view,
		},
	}

	args := []string{
		"-state", statePath,
		"foo",
	}
	code := c.Run(args)
	output := done(t)
	if code != 0 {
		t.Fatalf("bad: \n%s", output.Stderr())
	}

	actual := strings.TrimSpace(output.Stdout())
	if actual != `"bar"` {
		t.Fatalf("bad: %#v", actual)
	}
}

func TestOutput_json(t *testing.T) {
	originalState := states.BuildState(func(s *states.SyncState) {
		s.SetOutputValue(
			addrs.OutputValue{Name: "foo"}.Absolute(addrs.RootModuleInstance),
			cty.StringVal("bar"),
			false,
			"",
		)
	})

	statePath := testStateFile(t, originalState)

	view, done := testView(t)
	c := &OutputCommand{
		Meta: Meta{
			testingOverrides: metaOverridesForProvider(testProvider()),
			View:             view,
		},
	}

	args := []string{
		"-state", statePath,
		"-json",
	}
	code := c.Run(args)
	output := done(t)
	if code != 0 {
		t.Fatalf("bad: \n%s", output.Stderr())
	}

	actual := strings.TrimSpace(output.Stdout())
	expected := "{\n  \"foo\": {\n    \"sensitive\": false,\n    \"type\": \"string\",\n    \"value\": \"bar\"\n  }\n}"
	if actual != expected {
		t.Fatalf("wrong output\ngot:  %#v\nwant: %#v", actual, expected)
	}
}

func TestOutput_emptyOutputs(t *testing.T) {
	originalState := states.NewState()
	statePath := testStateFile(t, originalState)

	p := testProvider()
	view, done := testView(t)
	c := &OutputCommand{
		Meta: Meta{
			testingOverrides: metaOverridesForProvider(p),
			View:             view,
		},
	}

	args := []string{
		"-no-color",
		"-state", statePath,
	}
	code := c.Run(args)
	output := done(t)
	if code != 0 {
		t.Fatalf("bad: \n%s", output.Stderr())
	}
	// Warning diagnostics should go to stdout
	if got, want := output.Stdout(), "Warning: No outputs found"; !strings.Contains(got, want) {
		t.Fatalf("bad output: expected to contain %q, got:\n%s", want, got)
	}
}

func TestOutput_badVar(t *testing.T) {
	originalState := states.BuildState(func(s *states.SyncState) {
		s.SetOutputValue(
			addrs.OutputValue{Name: "foo"}.Absolute(addrs.RootModuleInstance),
			cty.StringVal("bar"),
			false,
			"",
		)
	})
	statePath := testStateFile(t, originalState)

	view, done := testView(t)
	c := &OutputCommand{
		Meta: Meta{
			testingOverrides: metaOverridesForProvider(testProvider()),
			View:             view,
		},
	}

	args := []string{
		"-state", statePath,
		"bar",
	}
	code := c.Run(args)
	output := done(t)
	if code != 1 {
		t.Fatalf("bad: \n%s", output.Stderr())
	}
}

func TestOutput_blank(t *testing.T) {
	originalState := states.BuildState(func(s *states.SyncState) {
		s.SetOutputValue(
			addrs.OutputValue{Name: "foo"}.Absolute(addrs.RootModuleInstance),
			cty.StringVal("bar"),
			false,
			"",
		)
		s.SetOutputValue(
			addrs.OutputValue{Name: "name"}.Absolute(addrs.RootModuleInstance),
			cty.StringVal("john-doe"),
			false,
			"",
		)
	})
	statePath := testStateFile(t, originalState)

	view, done := testView(t)
	c := &OutputCommand{
		Meta: Meta{
			testingOverrides: metaOverridesForProvider(testProvider()),
			View:             view,
		},
	}

	args := []string{
		"-state", statePath,
		"",
	}

	code := c.Run(args)
	output := done(t)
	if code != 0 {
		t.Fatalf("bad: \n%s", output.Stderr())
	}

	expectedOutput := "foo = \"bar\"\nname = \"john-doe\"\n"
	if got := output.Stdout(); got != expectedOutput {
		t.Fatalf("wrong output\ngot:  %#v\nwant: %#v", got, expectedOutput)
	}
}

func TestOutput_manyArgs(t *testing.T) {
	view, done := testView(t)
	c := &OutputCommand{
		Meta: Meta{
			testingOverrides: metaOverridesForProvider(testProvider()),
			View:             view,
		},
	}

	args := []string{
		"bad",
		"bad",
	}
	code := c.Run(args)
	output := done(t)
	if code != 1 {
		t.Fatalf("bad: \n%s", output.Stdout())
	}
}

func TestOutput_noArgs(t *testing.T) {
	view, done := testView(t)
	c := &OutputCommand{
		Meta: Meta{
			testingOverrides: metaOverridesForProvider(testProvider()),
			View:             view,
		},
	}

	args := []string{}
	code := c.Run(args)
	output := done(t)
	if code != 0 {
		t.Fatalf("bad: \n%s", output.Stdout())
	}
}

func TestOutput_noState(t *testing.T) {
	originalState := states.NewState()
	statePath := testStateFile(t, originalState)

	view, done := testView(t)
	c := &OutputCommand{
		Meta: Meta{
			testingOverrides: metaOverridesForProvider(testProvider()),
			View:             view,
		},
	}

	args := []string{
		"-state", statePath,
		"foo",
	}
	code := c.Run(args)
	output := done(t)
	if code != 0 {
		t.Fatalf("bad: \n%s", output.Stderr())
	}
}

func TestOutput_noVars(t *testing.T) {
	originalState := states.NewState()

	statePath := testStateFile(t, originalState)

	view, done := testView(t)
	c := &OutputCommand{
		Meta: Meta{
			testingOverrides: metaOverridesForProvider(testProvider()),
			View:             view,
		},
	}

	args := []string{
		"-state", statePath,
		"bar",
	}
	code := c.Run(args)
	output := done(t)
	if code != 0 {
		t.Fatalf("bad: \n%s", output.Stderr())
	}
}

func TestOutput_stateDefault(t *testing.T) {
	originalState := states.BuildState(func(s *states.SyncState) {
		s.SetOutputValue(
			addrs.OutputValue{Name: "foo"}.Absolute(addrs.RootModuleInstance),
			cty.StringVal("bar"),
			false,
			"",
		)
	})

	// Write the state file in a temporary directory with the
	// default filename.
	td := testTempDirRealpath(t)
	statePath := filepath.Join(td, DefaultStateFilename)

	f, err := os.Create(statePath)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	err = writeStateForTesting(originalState, f)
	f.Close()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Change to that directory
	t.Chdir(filepath.Dir(statePath))

	view, done := testView(t)
	c := &OutputCommand{
		Meta: Meta{
			testingOverrides: metaOverridesForProvider(testProvider()),
			View:             view,
		},
	}

	args := []string{
		"foo",
	}
	code := c.Run(args)
	output := done(t)
	if code != 0 {
		t.Fatalf("bad: \n%s", output.Stderr())
	}

	actual := strings.TrimSpace(output.Stdout())
	if actual != `"bar"` {
		t.Fatalf("bad: %#v", actual)
	}
}

func TestOutput_showSensitiveArg(t *testing.T) {
	originalState := stateWithSensitiveValueForOutput()

	statePath := testStateFile(t, originalState)

	view, done := testView(t)
	c := &OutputCommand{
		Meta: Meta{
			testingOverrides: metaOverridesForProvider(testProvider()),
			View:             view,
		},
	}

	args := []string{
		"-state", statePath,
		"-show-sensitive",
	}
	code := c.Run(args)
	output := done(t)
	if code != 0 {
		t.Fatalf("bad: \n%s", output.Stderr())
	}

	actual := strings.TrimSpace(output.Stdout())
	if actual != "foo = \"bar\"" {
		t.Fatalf("bad: %#v", actual)
	}
}

func TestOutput_withoutShowSensitiveArg(t *testing.T) {
	originalState := stateWithSensitiveValueForOutput()

	statePath := testStateFile(t, originalState)

	view, done := testView(t)
	c := &OutputCommand{
		Meta: Meta{
			testingOverrides: metaOverridesForProvider(testProvider()),
			View:             view,
		},
	}

	args := []string{
		"-state", statePath,
	}
	code := c.Run(args)
	output := done(t)
	if code != 0 {
		t.Fatalf("bad: \n%s", output.Stderr())
	}

	actual := strings.TrimSpace(output.Stdout())
	if actual != "foo = <sensitive>" {
		t.Fatalf("bad: %#v", actual)
	}
}

// stateWithSensitiveValueForOutput return a state with an output value
// marked as sensitive.
func stateWithSensitiveValueForOutput() *states.State {
	state := states.BuildState(func(s *states.SyncState) {
		s.SetOutputValue(
			addrs.OutputValue{Name: "foo"}.Absolute(addrs.RootModuleInstance),
			cty.StringVal("bar"),
			true,
			"",
		)
	})
	return state
}
