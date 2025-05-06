# 🕵️‍♂️ HTML Analyzer

A concurrent web analyzer CLI and API service built in Go — capable of scanning HTML documents for structural metadata like heading tags, HTML version, internal/external links, and login form detection.
Listen on port 8080

## 📦 Features

- **Analyze and extract from any public web page:**
    - HTML version
    - Page title
    - Headings count (h1–h6)
    - Internal, external, and inaccessible links
    - Login form detection
- **CLI mode** for batch analysis from a CSV file
- **Web API mode** for use with frontend applications
- **Dockerized** CLI and Web versions

## 🚀 Usage

### CLI Usage
We can provide list of urls in a csv file as input and get the required output for a csv file.

Navigate to the folder path that input and output files lives in then run below

```bash
cd data && docker run --rm -v "$(pwd)":/data eranga567/html-analyzer:latest-cli /data/input.csv /data/output.csv
```

### 🌐 Web API Usage

This will start the backend web server

```bash
docker run -p 8080:8080 eranga567/html-analyzer:latest-web
```

Then send a POST request:

```bash
curl -X POST http://localhost:8080/analyze \
     -H "Content-Type: application/json" \
     -d '{"url": "https://example.com"}'
```

## 🧰 Development

### Build CLI & Web binaries

```bash
make cli-build
make web-build
```

### Run tests

```bash
make test
```

### Lint code

```bash
make lint
```

## 🐳 Docker

### Build Docker images

```bash
make docker-web-build
make docker-cli-build
```
```

## 📁 Project Structure

.
├── cmd/
│   ├── cli/          # CLI entrypoint
│   └── server/       # Web API entrypoint
├── internal/
│   ├── app/          # Core services
│   ├── handlers/     # CLI and HTTP handlers
│   ├── core/         # Adapters
│   └── config/       # Configuration management
├── build/            # Compiled binaries
├── mocks/            # Auto-generated mocks
|__ pkg               # Constants and Entities
├── Makefile
├── web.Dockerfile
├── cli.Dockerfile
