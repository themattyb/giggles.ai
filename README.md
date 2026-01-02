# 🎭 giggles.ai

> A community-driven project exploring the lighter side of AI - where learning meets laughter

## 🎯 Mission

Giggles.ai aims to create a fun and interactive space where AI can be a delightful place to learn, experiment, and share laughs together. We believe that humor and play are essential components of learning and innovation.

## 🚀 What We're Building

- **AI Meme Crawler**: A polite web crawler (written in Go) that scours the internet for funny AI memes
- **Meme Search Interface**: A beautiful, responsive web GUI for searching and viewing collected memes
- **S3 Storage**: Secure cloud storage for meme images with AWS S3 integration
- **Community Hub**: A welcoming space for AI enthusiasts, learners, and experimenters
- **Learning Platform**: Interactive experiences that make AI concepts accessible and fun

## 🎪 What This Is NOT

- ❌ An AI joke generator or comedy writing tool
- ❌ A replacement for human comedians and writers
- ❌ A commercial AI service or product

## 🌟 What This Could Become

- ✅ A repository of educational AI experiments with unexpected, amusing results
- ✅ A community where "failed" but interesting models find new life as learning tools
- ✅ A showcase of creative AI applications that prioritize fun and education
- ✅ A platform for documenting the quirky, wacky, and wonderful world of AI development

## 🏗️ Current Status

**Active Development** - The project now includes:

- ✅ **Web Crawler** (Go): Polite crawler with robots.txt support, image detection, and S3 upload
- ✅ **Web GUI**: Modern, responsive interface for searching and viewing memes
- ✅ **Landing Page**: Beautiful entry point with feature highlights
- ✅ **Security**: Secure credential management with environment variables
- ✅ **Documentation**: Comprehensive setup guides and API documentation
- ✅ **Open Source Foundation**: See [LICENSE](LICENSE)

## 🛠️ Getting Started

### Prerequisites

- **Go 1.21 or later** - [Download Go](https://golang.org/dl/)
- **AWS Account** (for S3 storage) - [Sign up for AWS](https://aws.amazon.com/)
- **Modern web browser** (Chrome, Firefox, Safari, Edge)
- Enthusiasm for AI and learning! 🎉

### Quick Start

1. **Clone the repository:**
   ```bash
   git clone https://github.com/your-username/giggles.ai.git
   cd giggles.ai
   ```

2. **Set up AWS credentials:**
   ```bash
   export AWS_ACCESS_KEY_ID=your_access_key
   export AWS_SECRET_ACCESS_KEY=your_secret_key
   export AWS_REGION=us-east-1
   ```

3. **Build the crawler:**
   ```bash
   cd crawler
   go mod download
   go build -o crawler .
   ```

4. **Run the crawler:**
   ```bash
   ./crawler -start-url "https://www.reddit.com/r/artificial" \
     -s3-bucket "your-bucket-name" \
     -workers 5 \
     -delay 2s \
     -max-pages 100
   ```

5. **View the GUI:**
   ```bash
   # From project root
   open gui/index.html  # macOS
   xdg-open gui/index.html  # Linux
   start gui/index.html  # Windows
   ```

### Detailed Setup

For complete setup instructions, including S3 bucket creation and security best practices, see [SETUP.md](SETUP.md).

### Development

- **Crawler**: Written in Go, uses standard library and AWS SDK
- **GUI**: Pure HTML/CSS/JavaScript, no build process required
- **Landing Page**: Static HTML with modern CSS

## 🤝 Contributing

We welcome contributions from developers, designers, AI researchers, educators, and anyone passionate about making AI learning fun and accessible!

### Ways to Contribute

- 🐛 Report bugs and suggest improvements
- 💡 Propose new features or experiments
- 📚 Improve documentation
- 🎨 Enhance the user interface (GUI or landing page)
- 🔧 Improve crawler functionality (better meme detection, performance)
- 🧪 Share interesting AI experiments
- 📖 Write tutorials or guides
- 🔌 Build backend API integration

### Getting Started with Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Commit your changes (`git commit -m 'Add some amazing feature'`)
5. Push to the branch (`git push origin feature/amazing-feature`)
6. Open a Pull Request

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

## 📁 Project Structure

```
giggles.ai/
├── crawler/                    # Go web crawler
│   ├── main.go                # Crawler entry point
│   ├── go.mod                 # Go module definition
│   ├── internal/
│   │   ├── crawler/           # Crawler logic (robots.txt, HTML parsing)
│   │   └── s3/                # S3 client for image uploads
│   ├── credentials.example    # Credentials template
│   └── README.md              # Crawler documentation
├── gui/                       # Web interface (isolated directory)
│   ├── index.html            # Meme search interface
│   ├── styles.css            # GUI styling
│   ├── app.js                # JavaScript application logic
│   └── README.md             # GUI documentation
├── index.html                # Landing page
├── style.css                 # Landing page styles
├── README.md                # Project documentation (this file)
├── SETUP.md                  # Detailed setup guide
├── LICENSE                   # Open source license
├── CONTRIBUTING.md           # Contribution guidelines
└── docs/                     # Additional documentation
    ├── ARCHITECTURE.md
    ├── DEVELOPMENT.md
    └── ...
```

## 🎨 Design Philosophy

- **Playful**: Everything should feel fun and engaging
- **Accessible**: Complex AI concepts made simple and approachable
- **Community-First**: Built by and for the community
- **Educational**: Learning is at the heart of everything we do

## 📜 License

This project is open source and available under the [License](LICENSE).

## 🌈 Community

- **Be Kind**: Treat everyone with respect and kindness
- **Be Curious**: Ask questions, experiment, and explore
- **Be Inclusive**: Welcome people of all backgrounds and skill levels
- **Be Playful**: Don't forget to have fun!

## 🗺️ Roadmap

### Completed ✅
- [x] Web crawler with robots.txt support
- [x] Image downloading and S3 storage
- [x] Web GUI for meme search
- [x] Secure credential management
- [x] Landing page

### In Progress 🚧
- [ ] Backend API to serve memes from S3
- [ ] Database for meme metadata
- [ ] Full-text search functionality
- [ ] Image deduplication

### Planned 📋
- [ ] Advanced meme filtering (ML-based)
- [ ] User favorites/bookmarks
- [ ] Tag system for memes
- [ ] Share functionality
- [ ] Community forums/discussions
- [ ] Educational content library

## 📚 Documentation

- 📖 [Setup Guide](SETUP.md) - Complete setup instructions
- 🤖 [Crawler README](crawler/README.md) - Crawler documentation
- 🎨 [GUI README](gui/README.md) - GUI documentation
- 🏗️ [Architecture](docs/ARCHITECTURE.md) - Technical architecture

## 📞 Contact & Support

- 🐛 **Issues**: [GitHub Issues](https://github.com/your-username/giggles.ai/issues)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/your-username/giggles.ai/discussions)
- 📧 **Email**: [Contact us](mailto:hello@giggles.ai)

## 🔒 Security

- Credentials are managed via environment variables
- Never commit `.env` files or credentials to Git
- Use IAM roles for AWS infrastructure
- See [SETUP.md](SETUP.md) for security best practices

---

<div align="center">
  <p><strong>Made with ❤️ by the giggles.ai community</strong></p>
  <p><em>Where AI meets joy, and learning never stops being fun!</em></p>
</div>
