# Working on compa

A troubleshooting companion for support cases: a chat with a model that reads a case's artifacts through sandboxed,
deterministic tools and loads domain experts on demand. This repository is the build; the product is
[compa-engine](https://github.com/lowmagic/compa-engine)'s.

Read before changing behavior:

- [README](README.md): what compa is and how to run it
- [Coding standards](docs/standards.md): what belongs in this repository and how it is checked, on top of the org's
  [principles](https://github.com/lowmagic/.github/blob/main/PRINCIPLES.md) and standards (`docs/org-standards*.md`)
- compa-engine's [specification](https://github.com/lowmagic/compa-engine/blob/main/docs/SPEC.md) and
  [architecture](https://github.com/lowmagic/compa-engine/blob/main/docs/ARCHITECTURE.md), for anything about what
  compa does

Docs live in `docs/` and change with behavior. Never commit hostnames, tokens, or deployment config (see
Privacy in the [org standards](docs/org-standards.md)).

Run `make check` before committing. It runs everything CI runs, except the org standards check:
`make standards` (needs the network).
