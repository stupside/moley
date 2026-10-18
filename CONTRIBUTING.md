# Contributing to Moley

## Requirements

- [Go 1.27.2+](https://go.dev/dl/)
- [Task](https://taskfile.dev/) (task runner)
- [cloudflared](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/downloads/)

## Setup

```bash
git clone https://github.com/stupside/moley.git
cd moley
task go:build
```

## Making changes

1. Fork the repository and create a branch from `main`.
2. Make your changes.
3. Run `task go:test`, `task go:lint`, and `task go:build`.
4. Open a pull request against `main`.

Keep pull requests focused on a single change. If you're fixing a bug and refactoring nearby code, split them into separate PRs.

## Reporting bugs

Open an issue on GitHub with:

- What you expected to happen.
- What actually happened.
- Steps to reproduce.
- Moley version (`moley --version`) and OS.

## Code style

- Follow standard Go conventions (`gofmt`, `go vet`).
- Run `task go:modernize` to apply the modernizers included in Go, then review the diff.
- Keep changes minimal and focused.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

## Documentation

Use Node.js 24 (see `docs/.node-version`) and the Yarn version pinned in
`docs/package.json`. Enable Corepack with `corepack enable`, then run
`task docs:check` to install from the lockfile, typecheck, and build the site.
For a local preview, run `cd docs && yarn start`.

Dependabot groups weekly dependency updates for Go, docs, GitHub Actions, and
Docker. Security fixes are grouped separately from routine version updates.

`task clean` removes the local binary only. To stop and clean up resources owned
by a Moley configuration, use `moley tunnel stop` from that configuration directory.

### Dependency compatibility

Keep the Docusaurus packages on the same version. TypeScript is held at 5.9.x
because the Docusaurus tsconfig uses `baseUrl`, which TypeScript 7 removed.
The Yarn resolutions select patched transitive dependencies for structured-clone,
CSS selector parsing, JavaScript serialization, worker pools, and UUID generation.
Remove these overrides when the upstream dependency ranges include the fixes.

Run `yarn npm audit --all --recursive` from `docs/` after dependency updates.
As of October 18, 2026, it still reports the unpatched
[braces stack-exhaustion advisory](https://github.com/advisories/GHSA-vfj7-8cjw-p6xm)
through Chokidar 3 in the docs development tooling, plus a `whatwg-encoding`
deprecation. Do not force Chokidar to a new major version without checking its
watcher API and glob behavior. These packages are not part of the Moley Go binary.

## Releases

Use Conventional Commits: `fix:` for patch releases, `feat:` for minor releases,
and `!` or a `BREAKING CHANGE:` footer for major releases. Release Please maintains
one release PR against `main`, updating `CHANGELOG.md` and
`.release-please-manifest.json` from the last release (initially `v2.6.1`).

Review and merge that PR to publish. The Release workflow creates the tag and
GitHub release, then calls `_delivery.yml` in the same run. GoReleaser checks out
that exact tag, attaches binaries, publishes the multi-platform Docker image, and
updates the Homebrew tap while preserving Release Please's release notes.

The workflow uses `GITHUB_TOKEN` for release automation and the existing
`HOMEBREW_TAP_GITHUB_TOKEN` secret for the tap. GitHub Actions must be allowed to
create pull requests in the repository's Actions settings. As in Castor, bot-created
release PRs do not trigger push/PR CI with `GITHUB_TOKEN`; check the `main` CI results
and review the generated release diff before merging. To retry failed delivery,
rerun the failed jobs on the release's original workflow run.
