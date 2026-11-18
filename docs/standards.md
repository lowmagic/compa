# Coding standards

compa follows the [org standards](org-standards.md) and the standards for its topics ([go](org-standards-go.md)),
unchanged copies of [lowmagic's](https://github.com/lowmagic/.github/tree/main/code-standards). `.org-standards` lists
the topics it copies beyond the general one. This file adds compa's own rules and examples. Where it differs from an
org rule, it says so, and it wins.

## Structure

compa is a build, not a library: `main.go` passes generic parts to
[compa-engine](https://github.com/lowmagic/compa-engine) in `engine.Options`, and everything a case needs is the
engine's.

- Behavior belongs in the engine, and an expert or a tool in an expert repository. What belongs here is the choice of
  parts: the case and artifact sources, the built-in configuration, and the packages `bundle()` composes.
- `main.go`'s imports and `bundle()` are the one place that says which experts and catalogs the build ships.
- compa depends on public modules only, and `go.mod` replaces none (`TestImportsOnlyDeclaredModules`).
- `go.mod` requires the compa-engine version that the compa-experts version it requires was built on, so the
  packages ship on the engine their own checks ran against; `make pins`, in `make check`, fails on a skew
  (`go mod graph`). Move the two pins together.

## Config

The engine's configuration is typed YAML in layers, in the schema under the engine's `api/config/`: the built-in
`defaults.yaml`, the user's `~/.config/compa/config.yaml`, then `--config`. The built-in defaults allow no model
call; a user's file opts in. `main` reads one environment variable, `COMPA_CASES`.

## Privacy

`examples/config.yaml` holds the shape of a user's configuration, in place of `.env.example`. Example cases are made
up and say so.

## Testing

The tests run the command line as `main` does, over a copy of the example case, with a script answering for the
model (`testdata/`). `TestBundleLoads` is where the shipped packages and the engine they need are checked together:
each comes from its own module at its own version.

## Docs

compa's specification, architecture and roadmap are compa-engine's (`docs/SPEC.md`, `docs/ARCHITECTURE.md`,
`docs/ROADMAP.md` there). This repository documents only what the build chooses.

## Go

`go.mod`'s `go` line is compa-engine's and compa-experts', the oldest Go release line a host builds `compa` with,
so a host builds it with its own Go. This overrides the Layout rule that it names the current release.

## Known deviations

Where compa doesn't follow the org standards yet. These are debt to pay down, not overrides: each one should go.

- There's no `docs/SPEC.md` or `docs/ROADMAP.md` here: they are the engine's (see [Docs](org-standards.md#docs)).
