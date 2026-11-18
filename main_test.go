package main

import (
	"bytes"
	"context"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/lowmagic/compa-engine/engine/cli"
	"github.com/lowmagic/compa-engine/engine/expertcheck"
)

// TestBundleLoads: the experts and catalogs this build ships compose, and
// load through the engine's own check with no error and no warning. The
// packages come from another module than the engine, so this is where a
// package and the engine it needs are checked together.
func TestBundleLoads(t *testing.T) {
	b, err := bundle()
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	if err := os.CopyFS(root, b); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := expertcheck.Command([]string{"--fail-on-warning", filepath.Join(root, "experts")}, &out); err != nil {
		t.Fatalf("%v\n%s", err, out.String())
	}
	for _, want := range []string{"catalog generic.files ", "expert generic.artifacts "} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("the bundle lacks %q:\n%s", want, out.String())
		}
	}
}

// compa runs the command line as main does, over a copy of the example
// cases and an empty home, and returns what it printed.
type compa struct {
	t     *testing.T
	cases string
	home  string
}

func newCompa(t *testing.T) compa {
	t.Helper()
	c := compa{t: t, cases: t.TempDir(), home: t.TempDir()}
	if err := os.CopyFS(c.cases, os.DirFS(filepath.Join("examples", "cases"))); err != nil {
		t.Fatal(err)
	}
	return c
}

func (c compa) run(config []byte, args ...string) (stdout, stderr string, code int) {
	c.t.Helper()
	o, err := options(c.cases)
	if err != nil {
		c.t.Fatal(err)
	}
	o.Home = c.home
	if config != nil {
		o.Config = config
	}
	var out, errs bytes.Buffer
	code = cli.Run(context.Background(), args, o, strings.NewReader(""), &out, &errs)
	return out.String(), errs.String(), code
}

// TestExampleCaseOpens: with the built-in defaults, which allow no model
// call, the example case opens, its files are counted and the shipped
// expert is installed.
func TestExampleCaseOpens(t *testing.T) {
	c := newCompa(t)
	out, errs, code := c.run(nil, "case", "open", "demo")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errs)
	}
	for _, want := range []string{"case demo: Worker stops taking jobs", "attachments: 1, holding 1 file", "generic.artifacts 0.1.0 installed"} {
		if !strings.Contains(out, want) {
			t.Errorf("case open lacks %q:\n%s", want, out)
		}
	}
	if _, errs, code := c.run(nil, "case", "open", "absent"); code == 0 || !strings.Contains(errs, "absent") {
		t.Errorf("a case with no file: exit %d, %q", code, errs)
	}
}

// TestAskRunsAShippedTool: one turn on the example case, a script
// answering for the model. The model's call of a tool from the shipped
// catalog runs, its result is kept as evidence under a name, and the
// answer citing it is printed.
func TestAskRunsAShippedTool(t *testing.T) {
	script, err := filepath.Abs(filepath.Join("testdata", "script.json"))
	if err != nil {
		t.Fatal(err)
	}
	config := []byte("apiVersion: compa.lowmag.cc/config/v1alpha1\nkind: Config\nreasoning:\n" +
		"  models: {script: {provider: fixture, model: " + strconv.Quote(script) + "}}\n" +
		"  profiles: {companion: {model: script}}\n")
	c := newCompa(t)
	out, errs, code := c.run(config, "ask", "demo", "what failed")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errs)
	}
	if !strings.Contains(out, "refused its database connection five times [e4:1]") {
		t.Errorf("the answer:\n%s\n%s", out, errs)
	}
	if strings.Contains(out+errs, "⚠") || strings.Contains(out+errs, "flagged") {
		t.Errorf("the answer's citations of the tool's result are flagged:\n%s\n%s", out, errs)
	}
	list, errs, code := c.run(config, "evidence", "demo")
	if code != 0 || !strings.Contains(list, "e4") || !strings.Contains(list, `generic.files.count@0.1.0 {"path":"logs/worker.log"`) {
		t.Errorf("evidence, exit %d:\n%s\n%s", code, list, errs)
	}
	item, _, _ := c.run(config, "evidence", "demo", "e4")
	if !strings.Contains(item, `"count":5`) || !strings.Contains(item, "connection to db.example.com refused") {
		t.Errorf("e4 does not hold the count of the refused connections:\n%s", item)
	}
}

// TestImportsOnlyDeclaredModules: every import is the standard library,
// this module or a module go.mod requires, and go.mod replaces none, so the
// build is what anyone fetches. A workspace spanning other checkouts would
// let an undeclared import build.
func TestImportsOnlyDeclaredModules(t *testing.T) {
	raw, err := os.ReadFile("go.mod")
	if err != nil {
		t.Fatal(err)
	}
	allowed := []string{}
	block := false
	for _, line := range strings.Split(string(raw), "\n") {
		fields := strings.Fields(line)
		switch {
		case len(fields) == 0:
		case fields[0] == "replace":
			t.Errorf("go.mod replaces a module: %s", line)
		case fields[0] == "module":
			allowed = append(allowed, fields[1])
		case block && fields[0] == ")":
			block = false
		case block:
			allowed = append(allowed, fields[0])
		case fields[0] == "require" && len(fields) == 2 && fields[1] == "(":
			block = true
		case fields[0] == "require" && len(fields) >= 3:
			allowed = append(allowed, fields[1])
		}
	}
	fset := token.NewFileSet()
	err = filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if name := d.Name(); path != "." && (strings.HasPrefix(name, ".") || name == "testdata") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
	imports:
		for _, spec := range file.Imports {
			imported, _ := strconv.Unquote(spec.Path.Value)
			if first, _, _ := strings.Cut(imported, "/"); !strings.Contains(first, ".") {
				continue
			}
			for _, mod := range allowed {
				if imported == mod || strings.HasPrefix(imported, mod+"/") {
					continue imports
				}
			}
			t.Errorf("%s imports %s, which go.mod does not require", path, imported)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
