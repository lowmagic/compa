// Command compa is a troubleshooting companion for support cases: a chat
// with a model that reads a case's files through sandboxed, deterministic
// tools and loads domain experts on demand.
//
// It is the engine (github.com/lowmagic/compa-engine) built with generic
// parts: cases described in YAML files, their artifacts in directories
// beside them, the platform's sandbox, and the experts of
// github.com/lowmagic/compa-experts.
package main

import (
	_ "embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/lowmagic/compa-engine/engine"
	"github.com/lowmagic/compa-engine/engine/cli"
	"github.com/lowmagic/compa-engine/engine/compose"
	"github.com/lowmagic/compa-engine/engine/localdir"
	"github.com/lowmagic/compa-engine/engine/stubcases"
	"github.com/lowmagic/compa-experts/catalogs/files"
	"github.com/lowmagic/compa-experts/experts/artifacts"
	"github.com/lowmagic/compa-experts/experts/connections"
	"github.com/lowmagic/compa-experts/routing"
)

//go:embed defaults.yaml
var defaults []byte

func main() {
	code, err := run(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "compa:", err)
		os.Exit(1)
	}
	os.Exit(code)
}

// run hands the command line to the engine's and returns its exit status.
func run(args []string) (int, error) {
	cases, err := casesDir()
	if err != nil {
		return 0, err
	}
	o, err := options(cases)
	if err != nil {
		return 0, err
	}
	return cli.Main(args, o), nil
}

// options are the parts this build gives the engine, over the cases in
// one directory.
func options(cases string) (engine.Options, error) {
	experts, err := bundle()
	if err != nil {
		return engine.Options{}, err
	}
	return engine.Options{
		Cases:     stubcases.New(cases),
		Artifacts: localdir.New(cases),
		Config:    defaults,
		Bundle:    experts,
	}, nil
}

// casesDir is where cases are read from: COMPA_CASES, or ~/cases. A case
// is <dir>/<id>.yaml, which describes it, and <dir>/<id>/, which holds its
// artifacts.
func casesDir() (string, error) {
	if dir := os.Getenv("COMPA_CASES"); dir != "" {
		return dir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("find the cases directory: %w; set COMPA_CASES", err)
	}
	return filepath.Join(home, "cases"), nil
}

// bundle is the experts, catalogs and routing library this build ships.
// The import list above and this call say what they are.
func bundle() (fs.FS, error) {
	return compose.Bundle(
		routing.FS,
		files.FS,
		artifacts.FS,
		connections.FS,
	)
}
