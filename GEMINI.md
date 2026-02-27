# Development Guidelines

Specific instructions and best practices for the Astronomer project.

## Core Principles

- **Security First**: This tool deals with signing reports and handling GitHub tokens. Never log, print, or commit secrets. Use environment variables for sensitive data.
- **Concurrency with Care**: We use `errgroup` with concurrency limits (currently 10) to fetch data. Always propagate contexts and handle rate limiting carefully to avoid hitting secondary GitHub API limits.
- **Caching is Key**: Large scans consume many points of the GitHub GraphQL quota. Always check the local cache before making a network request.

## Project Structure

- `main.go`: Entry point, handles CLI arguments and orchestrates the scan.
- `pkg/gql/`: Core logic for interacting with GitHub GraphQL API, including pagination and caching.
- `pkg/trust/`: The trust algorithm logic, factors, and report rendering.
- `pkg/signature/`: Logic for signing reports and verifying signatures.
- `pkg/context/`: Shared context structure used across the application.

## Language & Framework Guidelines

### Go (Golang)

- **Version**: Go 1.25+ (pinned in `go.mod`).
- **Concurrency**: Use `golang.org/x/sync/errgroup` for parallel processing.
- **API**: GitHub GraphQL API via `pkg/gql`.
- **UI**: `github.com/Ullaakut/disgo` for CLI output and `github.com/vbauerster/mpb/v4` for progress bars.

## Automation & Tooling

We use a `Makefile` as the single entry point for development tasks.

| Target | Purpose |
| :--- | :--- |
| `setup` | Download dependencies and tidy `go.mod`. |
| `test` | Run all unit tests. |
| `lint` | Run `golangci-lint` (if installed). |
| `format` | Run `go fmt` on the entire project. |
| `build` | Compile the `astronomer` binary. |
| `docker` | Build the Docker image. |
| `run` | Build and run the application locally (requires `REPO` env). |
| `upgrade-deps` | Upgrade all dependencies to their latest versions. |

## Testing

- Place tests in the same package as the code they test (`*_test.go`).
- Use `github.com/stretchr/testify` for assertions.
- Run `make test` before submitting any PR.

## Security

- **GitHub Token**: Required via `GITHUB_TOKEN` environment variable.
- **Private Key**: Can be provided via `ASTRONOMER_PRIVATE_KEY` environment variable. Falls back to an embedded key if not provided (not recommended for production).
