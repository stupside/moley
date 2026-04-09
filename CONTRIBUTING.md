# Contributing to Moley

## Requirements

- [Go 1.26+](https://go.dev/dl/)
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
3. Make sure the project builds: `task go:build`.
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
- Keep changes minimal and focused.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
