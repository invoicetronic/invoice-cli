# Changelog

## In Development

- chore: release binaries are built with `CGO_ENABLED=0`, so the Linux builds are statically linked and no longer depend on the glibc of the build machine. The release workflow no longer needs to pin an old Ubuntu runner (`ubuntu-22.04`, soon retired by GitHub) and runs on `ubuntu-latest`.

## [1.1.0] - 2026-04-15

- new: send a `User-Agent: Invoicetronic-CLI/<version>` header on every request. The API logs it per event (filterable via `user_agent` on `GET /log/`), aligned with the SDK convention `Invoicetronic-{Lang}-SDK/{version}`.
- fix: `send` now correctly preserves the original file name. `SendItem` marshals `payload` and `file_name` in snake_case (matching the API naming policy) and excludes the local-only `FilePath`. Previously the API, receiving an unrecognized `FileName` field, silently auto-generated a file name and rewrote the XML `DatiTrasmissione` block.

## [1.0.0] - 2026-02-11

- fix: `get --json` was also writing the file to disk instead of only outputting JSON.
- new: Comprehensive test suite (28 tests) with testability refactoring of critical functions.
- new: Contributing section in README with build, test and cross-compile instructions.
- new: CI with GitHub Actions to run vet and test suite on every push and pull request.
- new: Release workflow runs the test suite before proceeding.

## [1.0.0-beta.1] - 2025-07-08

_First pre-release._
