# Giggles.ai Setup Guide

This guide will help you set up the giggles.ai application, including the web crawler and GUI.

## Prerequisites

- **Go 1.21 or later** - [Download Go](https://golang.org/dl/)
- **AWS Account** (for S3 storage) - [Sign up for AWS](https://aws.amazon.com/)
- **Modern web browser** (for the GUI)

## Quick Start

### 1. Clone the Repository

```bash
git clone https://github.com/your-username/giggles.ai.git
cd giggles.ai
```

### 2. Set Up AWS Credentials

#### Option A: Environment Variables (Recommended for Local Development)

Create a `.env` file in the `crawler/` directory:

```bash
cd crawler
cp .env.example .env
```

Edit `.env` with your AWS credentials:

```bash
AWS_ACCESS_KEY_ID=your_access_key_here
AWS_SECRET_ACCESS_KEY=your_secret_key_here
AWS_REGION=us-east-1
```

Then load them:

```bash
export $(cat .env | xargs)
```

#### Option B: AWS Credentials File

Create `~/.aws/credentials`:

```ini
[default]
aws_access_key_id = your_access_key_here
aws_secret_access_key = your_secret_key_here
```

Create `~/.aws/config`:

```ini
[default]
region = us-east-1
```

#### Option C: IAM Role (For EC2/ECS)

If running on AWS infrastructure, use IAM roles instead of credentials.

### 3. Create S3 Bucket

1. Go to [AWS S3 Console](https://console.aws.amazon.com/s3/)
2. Click "Create bucket"
3. Choose a unique bucket name (e.g., `giggles-ai-memes`)
4. Select your preferred region
5. Configure permissions (make sure you have `s3:PutObject` permission)
6. Create the bucket

### 4. Build the Crawler

```bash
cd crawler
go mod download
go build -o crawler .
```

### 5. Run the Crawler

```bash
./crawler \
  -start-urls "https://www.reddit.com/r/artificial,https://www.reddit.com/r/aifails,https://www.boredpanda.com/ai-fails/,https://www.cameo.com/chuffsters,https://www.unspeakable.com/,https://www.facebook.com/AliAreacts,https://cheezburger.com/38652165/28-hilarious-ai-fails-that-prove-were-safe-from-robot-overlords-for-now,https://www.quora.com/What-are-some-of-the-funniest-Artifical-Intelligence-AI-failures,https://www.facebook.com/groups/cursedaiwtf/posts/1716491672292642/" \
  -s3-bucket "your-bucket-name" \
  -s3-region "us-east-1" \
  -workers 5 \
  -delay 2s \
  -max-pages 100
```

### 6. View the GUI

Open the GUI in your browser:

```bash
# From project root
open gui/index.html  # macOS
xdg-open gui/index.html  # Linux
start gui/index.html  # Windows
```

Or simply navigate to `gui/index.html` in your browser.

## Project Structure

```
giggles.ai/
├── crawler/              # Go web crawler
│   ├── main.go          # Entry point
│   ├── internal/        # Internal packages
│   │   ├── crawler/     # Crawler logic
│   │   └── s3/          # S3 client
│   └── .env.example     # Credentials template
├── gui/                 # Web interface
│   ├── index.html       # Main GUI page
│   ├── styles.css       # Styling
│   └── app.js           # JavaScript logic
├── index.html           # Landing page
├── style.css            # Landing page styles
└── docs/               # All documentation
    └── SETUP.md        # This file
```

## Security Best Practices

⚠️ **IMPORTANT: Never commit credentials to Git!**

1. ✅ Use environment variables for local development
2. ✅ Use IAM roles for AWS infrastructure
3. ✅ Use AWS Secrets Manager for production
4. ✅ Add `.env` and credential files to `.gitignore` (already done)
5. ✅ Rotate credentials regularly
6. ✅ Use least-privilege IAM policies

### IAM Policy Example

Create an IAM user with this policy (minimum required permissions):

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:PutObject",
        "s3:PutObjectAcl"
      ],
      "Resource": "arn:aws:s3:::your-bucket-name/memes/*"
    }
  ]
}
```

## Configuration

### Crawler Options

- `-start-urls`: Comma-separated list of starting URLs to crawl (required)
- `-start-url`: Single starting URL (deprecated, use -start-urls instead)
- `-s3-bucket`: S3 bucket name (optional, but recommended)
- `-s3-region`: AWS region (default: us-east-1)
- `-workers`: Number of concurrent workers (default: 5)
- `-delay`: Delay between requests (default: 2s)
- `-max-pages`: Maximum pages to crawl (default: 100)
- `-user-agent`: User agent string (default: giggles-ai-crawler/1.0)

### Example Commands

**Crawl multiple sources for AI memes:**
```bash
./crawler -start-urls "https://www.reddit.com/r/artificial,https://www.reddit.com/r/aifails,https://www.boredpanda.com/ai-fails/,https://www.cameo.com/chuffsters,https://www.unspeakable.com/,https://www.facebook.com/AliAreacts,https://cheezburger.com/38652165/28-hilarious-ai-fails-that-prove-were-safe-from-robot-overlords-for-now,https://www.quora.com/What-are-some-of-the-funniest-Artifical-Intelligence-AI-failures,https://www.facebook.com/groups/cursedaiwtf/posts/1716491672292642/" \
  -s3-bucket "giggles-memes" \
  -workers 3 \
  -delay 3s \
  -max-pages 50
```

**Crawl without S3 (testing):**
```bash
./crawler -start-urls "https://www.boredpanda.com/ai-fails/,https://www.reddit.com/r/aifails" \
  -workers 5 \
  -max-pages 20
```

## Troubleshooting

### S3 Upload Errors

- Verify AWS credentials are set correctly
- Check S3 bucket permissions
- Ensure bucket exists in the specified region
- Check IAM policy allows `s3:PutObject`

### Crawler Issues

- **Rate limiting**: Increase `-delay` or reduce `-workers`
- **Memory issues**: Reduce `-max-pages` or process in batches
- **Robots.txt blocking**: Some sites may block crawlers - this is expected

### GUI Issues

- **Images not loading**: Check browser console for errors
- **Memes not loading / blank page**: The app fetches `memes.json`, which fails on `file://`. Serve over HTTP (`python -m http.server 8000`) instead of opening the file directly. See [GUI.md](GUI.md#data-source).
- **Pointing at S3**: Set `window.GIGGLES_MANIFEST_URL`, or regenerate `memes.json` with `crawler -gen-manifest -manifest-base-url <s3-prefix>`.

## Next Steps

1. **Hosting**: Serve the GUI and `memes.json` from S3/CloudFront (static site)
2. **Database**: Store meme metadata (title, source, tags, etc.)
3. **Search**: Implement full-text search functionality
4. **Deduplication**: Prevent downloading duplicate images
5. **Image Processing**: Resize, optimize, or add watermarks

## Getting Help

- 📖 Check the [Crawler Documentation](CRAWLER.md)
- 📖 Check the [GUI Documentation](GUI.md)
- 🐛 [Report issues](https://github.com/your-username/giggles.ai/issues)
- 💬 [Join discussions](https://github.com/your-username/giggles.ai/discussions)

## License

See [LICENSE](LICENSE) file.


