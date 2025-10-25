# Stats Application

A Go application that queries SabNZBD for download queue size, calculates media directory sizes, and generates an HTML stats page.

## Features

- Queries SabNZBD API for download queue information
- Calculates media directory sizes using `du` command
- Generates human-readable size formats
- Creates HTML statistics page
- Handles maintenance notices
- Configurable via environment variables

## Project Structure

This project follows Go best practices with a clean, idiomatic directory structure:

```
stats/
├── cmd/
│   └── stats/           # Main application entry point
│       ├── main.go      # Application entry point
│       └── logic.go     # Main application logic
├── internal/            # Private application code
│   ├── api/            # API clients (SabNZBD)
│   ├── config/          # Configuration management
│   ├── logger/          # Logging functionality
│   ├── models/          # Data structures
│   └── utils/           # Internal utilities
├── pkg/                 # Public library code
│   ├── html/           # HTML generation
│   └── size/           # Size calculation utilities
├── test/               # Test files
├── .github/workflows/  # CI/CD pipeline
├── build.sh           # Build script
├── Makefile           # Development tasks
└── README.md          # This file
```

## Environment Variables

- `SABAPIKEY`: SabNZBD API key (default: "YOURKEY")
- `SABHOST`: SabNZBD host URL (default: "https://sab.zaldre.com")
- `SABPORT`: SabNZBD port (default: 443)
- `UPTIME`: Status page URL (default: "https://app.statuscake.com/button/index.php?Track=lmmBTReo4c&Days=30&Design=2")
- `LOGLEVEL`: Log level - None, Normal, Debug (default: "Normal")
- `STATSFILE`: Output HTML file path (default: "/container/data/stats/index.html")
- `WEBTIMEOUT`: HTTP timeout in seconds (default: 15)

## Testing

### Running Tests

```bash
# Run all tests
make test

# Run tests with verbose output
make test-verbose

# Run tests with coverage
make test-coverage

# Run the test script
./run-tests.sh
```

### Test Coverage

The test suite includes comprehensive tests for:

- **env.go**: Environment variable handling
- **calcsize.go**: Size calculation and formatting
- **log.go**: Logging functionality
- **models.go**: Data structures and JSON handling
- **html.go**: HTML generation
- **maintenance.go**: Maintenance notice handling
- **sab.go**: SabNZBD API integration
- **size.go**: Media directory size calculation
- **logic.go**: Main application logic
- **main.go**: Application initialization

### Test Files

- `env_test.go` - Tests for environment variable handling
- `calcsize_test.go` - Tests for size calculation functions
- `log_test.go` - Tests for logging functionality
- `models_test.go` - Tests for data structures
- `html_test.go` - Tests for HTML generation
- `maintenance_test.go` - Tests for maintenance notice handling
- `sab_test.go` - Tests for SabNZBD API integration
- `size_test.go` - Tests for media directory size calculation
- `logic_test.go` - Tests for main application logic
- `main_test.go` - Tests for application initialization

## Building

### Local Build

```bash
# Build for current platform
make build

# Build for all platforms
make build-all
```

### Build Script

The application includes a build script (`build.sh`) that creates optimized binaries:

```bash
./build.sh
```

## CI/CD Pipeline

The project includes a GitHub Actions CI/CD pipeline (`.github/workflows/ci.yml`) that:

1. **Test**: Runs unit tests with coverage
2. **Build**: Builds for multiple platforms (Linux, Windows, macOS)
3. **Lint**: Runs code quality checks
4. **Security**: Performs security scans
5. **Release**: Creates releases for main branch pushes

### Pipeline Features

- Multi-platform builds (Linux AMD64/ARM64, Windows AMD64, macOS AMD64/ARM64)
- Test coverage reporting
- Code quality checks with golangci-lint
- Security scanning with Gosec
- Automatic releases for main branch

## Development

### Prerequisites

- Go 1.25.3 or later
- Make (optional, for using Makefile)

### Setup

```bash
# Install dependencies
make dev-deps

# Install development tools
make install-tools

# Run quality checks
make check
```

### Code Quality

The project includes several code quality tools:

- **golangci-lint**: Comprehensive Go linter
- **gosec**: Security scanner
- **go vet**: Built-in Go analysis tool
- **go fmt**: Code formatting

## Usage

### Basic Usage

```bash
# Run the application
make run

# Or build and run manually
make build
./stats
```

### Configuration

The application can be configured using environment variables. See the Environment Variables section above.

## Package Structure

### `cmd/stats/`
- **main.go**: Application entry point and initialization
- **logic.go**: Main application logic and orchestration

### `internal/` (Private packages)
- **api/**: SabNZBD API client
- **config/**: Configuration management and environment variables
- **logger/**: Logging functionality with consistent timestamp formatting
- **models/**: Data structures and JSON handling
- **utils/**: Internal utilities (size calculation, maintenance notices)

### `pkg/` (Public packages)
- **html/**: HTML generation and templating
- **size/**: Media directory size calculation utilities

### `test/`
- Comprehensive test suite for all components

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run the test suite: `make test`
6. Run quality checks: `make check`
7. Submit a pull request

## License

This project is open source. Please check the license file for details.
