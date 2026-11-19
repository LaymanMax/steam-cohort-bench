# Contributing to `steam-cohort-bench`

Thank you for your interest in contributing to `steam-cohort-bench`! We welcome community contributions, bug reports, feature suggestions, and documentation improvements.

## Code of Conduct

All contributors and community members are expected to abide by our [Code of Conduct](./CODE_OF_CONDUCT.md).

---

## Development Workflow

### 1. Prerequisites

Before setting up the project locally, ensure you have the following installed:

- **Go**: 1.22+ (tested on Go 1.24+) with CGO enabled
- **C Compiler**: `clang` or `gcc` (required by DuckDB's CGO Go bindings)
- **Node.js**: v20+ (tested on Node v22)
- **pnpm**: v9+ (or via `corepack enable`)

### 2. Fork and Clone

```bash
# Fork the repository on GitHub, then clone your fork:
git clone https://github.com/<your-username>/steam-cohort-bench.git
cd steam-cohort-bench
```

### 3. Install Dependencies

```bash
make install
```

### 4. Ingest the Analytical Dataset

Before running the server, ingest the ground-truth Steam games dataset:

```bash
# Download and ingest from Hugging Face:
go run ./cmd/ingest --download

# Or if you already have the parquet file locally:
go run ./cmd/ingest --input ./data/raw/steam-games.parquet
```

### 5. Running in Development

Run the Go API server and Vite frontend concurrently with live reloading:

- **Terminal 1 (Backend API)**:
  ```bash
  go run ./cmd/server
  ```
- **Terminal 2 (Frontend Dev Server)**:
  ```bash
  cd web && pnpm run dev
  ```
  The Vite development server runs at `http://localhost:5173` and automatically proxies `/api` requests to the Go server on port 8080.

---

## Testing

Ensure all tests pass before submitting a pull request:

```bash
# Run Go unit, integration, and E2E workflow tests
make test-go

# Run frontend Vitest component tests
make test-web

# Run the complete test suite
make test
```

---

## Pull Request Guidelines

1. **Create a Topic Branch**: Use a descriptive branch name, e.g., `git checkout -b fix-price-histogram-edges` or `git checkout -b feat-tag-filter`.
2. **Adhere to Code Standards**:
   - Run `go fmt ./...` and `go vet ./...` on backend code.
   - Run `pnpm run build` in `web/` to verify TypeScript types and compilation.
3. **Write Tests**: Accompany new features, normalization rules, or bug fixes with appropriate unit or integration tests.
4. **Statistical Rigor**: Maintain the project's core philosophy: use distribution-based metrics (percentiles, ranges) rather than simple arithmetic averages. Never fabricate mock statistics or predictive success claims.
5. **Describe Your Changes**: Provide a clear pull request description outlining the problem solved, testing methodology, and any relevant screenshots for UI changes.

---

## Reporting Issues

If you encounter bugs, data discrepancies, or have ideas for enhancement, please open an issue on GitHub with:

- A clear, descriptive title.
- Steps to reproduce the issue.
- Relevant system information (OS, Go version, browser).
- Expected versus actual behavior.
