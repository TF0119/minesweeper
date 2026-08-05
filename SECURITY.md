# Security Policy

## Supported versions

The latest release is the only supported version. Fixes land on `main` and go
out in the next tag.

## Reporting a vulnerability

Please report security issues privately through
[GitHub Security Advisories](https://github.com/TF0119/minesweeper/security/advisories/new)
rather than a public issue.

Include what you did, what happened, and the version (`minesweeper -version`).
You can expect an initial response within a week.

## Scope

This is an offline terminal game. It listens on no ports and makes no network
requests. The parts worth scrutinising are the files it writes under
`~/.config/minesweeper/` and the handling of untrusted command-line input.
