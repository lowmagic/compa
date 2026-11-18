<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="docs/assets/compa-lockup-dark.svg">
    <img src="docs/assets/compa-lockup-light.svg" alt="compa" width="240">
  </picture>
</p>

# compa

A troubleshooting companion for support cases. You talk to a model about a case in a terminal chat; it reads the
case's files through deterministic tools that run in a sandbox, keeps what they return as evidence it must cite, and
loads domain experts when the problem needs them. It writes the update for the customer or the note for a colleague,
and you copy it: compa posts nothing.

This repository is the build people install: the engine,
[compa-engine](https://github.com/lowmagic/compa-engine), with generic parts and the experts of
[compa-experts](https://github.com/lowmagic/compa-experts). What the product does is in the engine's
[specification](https://github.com/lowmagic/compa-engine/blob/main/docs/SPEC.md).

## Running it

```sh
go build -o compa . && ./compa version
```

**Cases.** compa reads cases from one directory, `COMPA_CASES`, by default `~/cases`: `<id>.yaml` describes a case,
its details and comments, and `<id>/` holds its files, such as an unpacked diagnostic archive.
[examples/cases](examples/cases) has one to try:

```sh
COMPA_CASES=examples/cases ./compa case open demo   # read the case and count its files; no model
COMPA_CASES=examples/cases ./compa demo             # the chat
./compa chat --dir /path/to/files                   # a chat over a directory, with no case
```

**Models.** Out of the box compa calls no model: its built-in configuration ([defaults.yaml](defaults.yaml)) allows
nothing to leave the machine, and everything that needs no model works. To use one, copy
[examples/config.yaml](examples/config.yaml) to `~/.config/compa/config.yaml`, which allows calls to Vertex AI and
names your Google Cloud project, and sign in with `gcloud auth application-default login`. Secrets are redacted from
everything a call sends. `--config <file>` layers one more file on top.

**Experts.** [main.go](main.go) lists the expert and catalog packages this build ships. `./compa expert check <dir>`
checks packages of your own, and `experts.local` in the configuration loads them.

`./compa` with no arguments prints every command. The case's store is `.compa/` inside its folder.

- `make check` runs the checks CI runs, offline: `make pins` (compa-experts on the engine this build requires), the
  linter, the tests and `go mod tidy -diff`.
- `make standards` checks the org standards copies and needs the network.
