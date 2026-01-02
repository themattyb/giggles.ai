# Giggles.ai - Quick Reference

## Overview

Giggles.ai consists of two main components:

1. **Web Crawler** (Go) - Scours the internet for AI memes and stores them in S3
2. **Web GUI** (HTML/JavaScript) - Beautiful interface to search and view memes

## Quick Start

### 1. Set Up AWS Credentials

```bash
export AWS_ACCESS_KEY_ID=your_key
export AWS_SECRET_ACCESS_KEY=your_secret
export AWS_REGION=us-east-1
```

### 2. Build Crawler

```bash
cd crawler
go mod download
go build -o crawler .
```

### 3. Run Crawler

```bash
./crawler -start-url "https://www.reddit.com/r/artificial" \
  -s3-bucket "your-bucket" \
  -workers 5 \
  -delay 2s \
  -max-pages 100
```

### 4. View GUI

Open `gui/index.html` in your browser.

## Project Structure

- `crawler/` - Go web crawler with robots.txt support
- `gui/` - Isolated HTML/JavaScript interface
- `index.html` - Landing page
- `SETUP.md` - Detailed setup instructions

## Security

- ✅ Credentials via environment variables
- ✅ `.gitignore` configured to exclude credential files
- ✅ IAM roles supported for AWS infrastructure

See [SETUP.md](SETUP.md) for complete setup instructions.

