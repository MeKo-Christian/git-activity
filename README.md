# Git Activity CLI

`git-activity` is a command-line tool to analyze Git repository activity. It generates charts to visualize commit activity, enabling insights into developer workflows and repository trends.

## Features

- Analyze multiple Git repositories.
- Support for stacked bar charts.
- Filter by date range.
- Output charts in `png` or `svg` format.
- Analyze commits or lines of code.
- Customizable developer aliases for stacking activities by devs.

## Installation

1. Clone the repository:

    ```bash
    git clone https://github.com/meko-christian/git-activity.git
    cd git-activity
    ```

2. Build the binary:

    ```bash
    go build -o git-activity
    ```

3. Run the CLI:

    ```bash
    ./git-activity
    ```

## Usage

### Commands

#### Analyze

Analyze Git repositories to generate activity charts:

```bash
./git-activity analyze [options] <repo1> <repo2> ...
```

Example:

```bash
./git-activity analyze --start=2023-01-01 --end=2023-12-31 --format=png --grouped --mode=commits ./repo1 ./repo2
```

#### Options:

| Flag         | Default     | Description                                                              |
|--------------|-------------|--------------------------------------------------------------------------|
| `--start, -s` | `""`        | Start date for analysis (YYYY-MM-DD).                                    |
| `--end, -e`   | `""`        | End date for analysis (YYYY-MM-DD).                                      |
| `--format, -f`| `png`       | Output format for charts (`png` or `svg`).                               |
| `--grouped, -g`| `false`    | Generate grouped bar charts.                                             |
| `--mode, -m`  | `commits`   | Analysis mode (`commits` or `lines`).                                    |
| `--people, -p`| `""`        | Path to a file defining developer aliases.                               |
| `--bars, -b`  | `""`        | Stacking mode for charts (`repository`, `developer`, or flat by default).|

### Debugging

The CLI exposes profiling data for debugging and performance analysis:

- Profiling data available at `http://localhost:6060/debug/fgprof`
- Start the CLI with profiling enabled:

    ```bash
    ./git-activity
    ```

## Developer Aliases

To map multiple Git identities to a single developer, provide a file with the following format:

```
name|alias1|alias2|...
```

Use the `--people` flag to specify the path to this file.

## Development

### Requirements

- Go 1.20+

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.
