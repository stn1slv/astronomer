# Astronomer

<p align="center">
    <img width="300" src="img/logo.png"/>
</p>

> [!NOTE]
> This project is a continuation of [Ullaakut/astronomer](https://github.com/Ullaakut/astronomer), which was archived by the owner on Oct 12, 2020. This fork aims to maintain and modernize the tool for continued use.

Astronomer is a high-performance tool that analyzes GitHub repository stargazers to compute the likelihood that they are real humans. Its primary goal is to **detect illegitimate GitHub stars from bot accounts**, which are often used to artificially inflate the perceived popularity of open-source projects.

<p align="center">
    <img width="75%" src="img/astronomer.gif">
</p>

## Key Features

*   **Concurrent Analysis**: Uses modern Go concurrency primitives (`errgroup`) to fetch and analyze contribution data across multiple years and users simultaneously, significantly reducing execution time.
*   **Weighted Trust Algorithm**: Computes trust based on contribution age, private activity, and diversity of interactions (commits, issues, PRs, reviews).
*   **Comparative Reporting**: Automatically compares the "early adopters" of a repository against random samples to detect inorganic growth patterns.
*   **Local Caching**: Robust local caching of GitHub GraphQL responses to minimize API usage and respect rate limits.
*   **Signed Reports**: Generates RSA-signed reports to ensure data integrity when transmitted to Astrolab.

## Trust algorithm

Trust is computed based on several factors:

*   **Weighted Contributions**: Older contributions are weighted more heavily, as they are harder to "fake" in bulk.
*   **Activity Diversity**: Analysis of commits, issues, pull requests, and code reviews.
*   **Private Activity**: Recognition of private contributions (restricted contribution counts).
*   **Account Maturity**: Average account age; older accounts are statistically more trustworthy.
*   **Statistical Percentiles**: Evaluation of the distribution of contribution scores from the 5th to the 95th percentile.

## Getting Started

### Prerequisites

*   **Go 1.25 or later**.
*   A **GitHub Personal Access Token** with `repo` read rights. [Generate one here](https://github.com/settings/tokens).

### Installation

```bash
git clone https://github.com/ullaakut/astronomer.git
cd astronomer
make build
```

### Usage

Set your token as an environment variable:

```bash
export GITHUB_TOKEN=your_token_here
```

Run the scan:

```bash
./astronomer ullaakut/astronomer
```

## Arguments and Options

*   **`repositoryOwner/repositoryName`**: (Required) The repository to scan.
*   **`-c, --cachedir` (string)**: Directory for cached data (default: `./data`).
*   **`-s, --stars` (int)**: Maximum stars to scan in fast mode (default: `1000`).
*   **`-a, --all`**: Scan all stargazers. Overrides `--stars`. Use with caution on large repositories.
*   **`-v, --verbose`**: Enable detailed logs and comparative analysis reports.

## Development

The project includes a `Makefile` to simplify common tasks:

*   `make setup`: Bootstrap the project and download dependencies.
*   `make build`: Compile the `astronomer` binary.
*   `make test`: Run the full test suite.
*   `make lint`: Run static analysis (requires `golangci-lint`).
*   `make format`: Auto-format source code.
*   `make upgrade-deps`: Upgrade all Go dependencies to their latest versions.

## Examples

![Traefik](img/traefik.png)
![Suspicious_repo](img/suspicious_repo.png)
![envoy](img/envoy.png)

## Questions & Answers

> _Why would fake stars be an issue?_

Repositories with high star counts often appear in GitHub Trending and newsletters, attracting real users and even influencing technology choices in startups. Bot-driven stars create a false sense of security and community backing.

> _How accurate is this algorithm?_

Astronomer provides an estimate. A low score might indicate a community of casual users or low precision due to a small sample size. It is meant as a diagnostic tool rather than an absolute verdict.

> _Why do results vary slightly between scans?_

In fast mode, Astronomer scans the first 200 users and then takes random slices of the remaining stargazers. These random samples can lead to slight variations (1-3%) in the final score. Use the `--all` flag for a deterministic, comprehensive report.

## Thanks

Inspired by [spencerkimball/stargazers](https://github.com/spencerkimball/stargazers).
The original Go gopher was designed by [Renee French](http://reneefrench.blogspot.com).
