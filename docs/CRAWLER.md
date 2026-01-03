# Giggles.ai Web Crawler

A polite web crawler written in Go that searches the internet for funny AI memes, downloads images, and stores them in an S3 bucket.

## Features

- ✅ **Polite Crawling**: Respects robots.txt rules
- ✅ **Rate Limiting**: Configurable delays between requests
- ✅ **Concurrent Processing**: Multiple worker goroutines for efficient crawling
- ✅ **Image Detection**: Identifies meme images (PNG, JPG, GIF, WebP)
- ✅ **S3 Integration**: Uploads images to AWS S3
- ✅ **Secure Credentials**: Uses environment variables or AWS credentials file

## Prerequisites

- Go 1.21 or later
- AWS account with S3 bucket (optional, but recommended)
- AWS credentials configured

## Installation

```bash
cd crawler
go mod download
go build -o crawler .
```

## Configuration

### AWS Credentials

The crawler uses AWS SDK's default credential chain, which checks in this order:

1. **Environment Variables** (recommended for local development):
   ```bash
   export AWS_ACCESS_KEY_ID=your_access_key
   export AWS_SECRET_ACCESS_KEY=your_secret_key
   export AWS_REGION=us-east-1
   ```

2. **AWS Credentials File** (`~/.aws/credentials`):
   ```ini
   [default]
   aws_access_key_id = your_access_key
   aws_secret_access_key = your_secret_key
   ```

3. **IAM Role** (if running on EC2)

### Security Best Practices

⚠️ **Never commit credentials to Git!**

- Use environment variables for local development
- Use IAM roles for EC2/ECS deployments
- Use AWS Secrets Manager or Parameter Store for production
- Add `.env` and credential files to `.gitignore` (already done)

## Usage

### Basic Usage

```bash
./crawler \
  -start-urls "https://www.reddit.com/r/artificial,https://www.reddit.com/r/aifails,https://www.boredpanda.com/ai-fails/,https://www.cameo.com/chuffsters,https://www.unspeakable.com/,https://www.facebook.com/AliAreacts,https://cheezburger.com/38652165/28-hilarious-ai-fails-that-prove-were-safe-from-robot-overlords-for-now,https://www.quora.com/What-are-some-of-the-funniest-Artifical-Intelligence-AI-failures,https://www.facebook.com/groups/cursedaiwtf/posts/1716491672292642/" \
  -s3-bucket "my-memes-bucket" \
  -s3-region "us-east-1" \
  -workers 5 \
  -delay 2s \
  -max-pages 100
```

### Command Line Options

- `-start-urls`: Comma-separated list of starting URLs to crawl from (required)
- `-start-url`: Single starting URL (deprecated, use -start-urls instead)
- `-s3-bucket`: S3 bucket name for storing images (optional)
- `-s3-region`: AWS region for S3 bucket (default: us-east-1)
- `-workers`: Number of concurrent workers (default: 5)
- `-delay`: Delay between requests (default: 2s)
- `-max-pages`: Maximum number of pages to crawl (default: 100)
- `-user-agent`: User agent string (default: giggles-ai-crawler/1.0)

### Examples

**Crawl multiple sources for AI memes:**
```bash
./crawler -start-urls "https://www.reddit.com/r/artificial,https://www.reddit.com/r/aifails,https://www.boredpanda.com/ai-fails/,https://www.cameo.com/chuffsters,https://www.unspeakable.com/,https://www.facebook.com/AliAreacts,https://cheezburger.com/38652165/28-hilarious-ai-fails-that-prove-were-safe-from-robot-overlords-for-now,https://www.quora.com/What-are-some-of-the-funniest-Artifical-Intelligence-AI-failures,https://www.facebook.com/groups/cursedaiwtf/posts/1716491672292642/" \
  -s3-bucket "giggles-memes" \
  -workers 3 \
  -delay 3s \
  -max-pages 50
```

**Crawl without S3 (just download locally):**
```bash
./crawler -start-urls "https://www.boredpanda.com/ai-fails/,https://www.reddit.com/r/aifails" \
  -workers 5 \
  -max-pages 20
```

## How It Works

1. **Robots.txt Check**: Before crawling any URL, the crawler fetches and parses robots.txt to ensure it's allowed
2. **HTML Parsing**: Extracts image URLs and links from HTML pages
3. **Image Filtering**: Filters images based on:
   - File extensions (.jpg, .jpeg, .png, .gif, .webp)
   - Keywords related to AI memes
   - Known meme domains
4. **Download**: Downloads images and validates they're actually images
5. **S3 Upload**: Uploads images to S3 bucket under `memes/` prefix

## Output

The crawler prints statistics at the end:

```
=== Crawler Statistics ===
Pages crawled: 100
Images found: 342
Images downloaded: 89
Images uploaded to S3: 89
Errors: 3
Duration: 2m15s
```

## Development

### Project Structure

```
crawler/
├── main.go                 # Entry point
├── go.mod                  # Go module definition
├── internal/
│   ├── crawler/
│   │   └── crawler.go      # Main crawler logic
│   └── s3/
│       └── client.go       # S3 client wrapper
└── README.md              # This file
```

### Adding Features

- **Better Meme Detection**: Improve `isMemeImage()` function with ML or better heuristics
- **Database Storage**: Store metadata about crawled images
- **Deduplication**: Check if images already exist before downloading
- **Image Processing**: Resize, optimize, or add watermarks

## Troubleshooting

### S3 Upload Errors

- Verify AWS credentials are set correctly
- Check S3 bucket permissions (needs `s3:PutObject` permission)
- Ensure bucket exists in the specified region

### Rate Limiting

If you're getting blocked:
- Increase `-delay` value
- Reduce `-workers` count
- Check robots.txt for crawl-delay directives

### Memory Issues

For large crawls:
- Reduce `-max-pages`
- Process images in smaller batches
- Consider streaming large images instead of loading into memory

## License

See main project LICENSE file.


