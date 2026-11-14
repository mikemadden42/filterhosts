# TODO

## Bugs (`main.go`)

- [x] **Wi-Fi "5 bars" devices disappear from the connection report** (`processLanHostsFile`). Rows whose 4th column is `"Wi-Fi   Wi-Fi 5 bars"` go into "Connection Summary" as raw lines instead of being grouped, so devices with the best signal never appear under "Devices Grouped by Connection".
- [x] **Output order changes every run.** All reports loop over Go maps, which have no fixed order. Sort the keys before printing.
- [x] **`fields[1] == "Status"` check is dead code.** The header is already skipped, so "Status Summary" is almost always empty. Remove it or define what the summary should be.
- [x] **Errors look like success.** Errors go to stdout and the program exits 0. Write to stderr, return errors from the functions, and `os.Exit(1)` in `main`.
- [x] **Short rows are dropped silently.** Warn on rows with too few fields. Make header skipping consistent between the two parsers (one matches a string, the other drops the first line).

## Design

- [ ] Take file names as flags/arguments instead of hard-coding `hosts.tsv` and `lan-hosts.tsv`.
- [x] Replace the three copied "group and print" loops with one helper (e.g. `printGroups(title string, groups map[string][]string)` with sorted keys).
- [ ] `LanDevice` ignores `fields[3]`; `Device`/`LanDevice` are exported without doc comments (unnecessary in `package main`).
- [x] Add tests: split parsing (`io.Reader` → groups) from printing and cover it with table tests using inline TSV.

## Tooling and config

- [x] Migrate `.golangci.yml` to v2 schema (`golangci-lint config verify` fails):
  - `version: "2"` (string)
  - `disable-all` → `linters.default: none`
  - `linters-settings` → `linters.settings`
  - `issues.exclude-dirs` → `linters.exclusions.paths`
  - `goimports.local-prefixes` → under `formatters.settings`, as a list
- [x] Fix current lint findings: unchecked `file.Close` ×2, gosec G304 ×2, missing package comment.
- [ ] Rename module to `github.com/mikemadden42/filterhosts` so the goimports local-prefix applies and `go install` works.
- [x] Fix `.goreleaser.yaml` (`goreleaser check` fails): add `version: 2`, rename `snapshot.name_template` → `version_template`, drop the no-op `go generate` hook.
- [ ] Update CI: test Go `oldstable` and `stable` (done), bump `go.mod` from `go 1.21`, update `actions/checkout` and `actions/setup-go` to current majors, add `golangci/golangci-lint-action`.
- [ ] Expand README: expected TSV columns for each file and sample output.

## Suggested order

1. ~~Bugs 1–2 (wrong or unstable output)~~ done
2. ~~Lint and release configs~~ done
3. ~~Parser tests~~ done
