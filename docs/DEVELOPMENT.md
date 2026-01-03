# Development Guide 👩‍💻👨‍💻

Welcome to the giggles.ai development guide! This document will help you get started with contributing to the project and understanding our development practices.

## Quick Start

### Prerequisites

- **Git**: For version control
- **Modern Browser**: Chrome, Firefox, Safari, or Edge (latest versions)
- **Text Editor/IDE**: VS Code, Sublime Text, Atom, or your preferred editor
- **Node.js** (optional, for future tooling): Version 16+ recommended

### Setup Instructions

1. **Fork and Clone**
   ```bash
   # Fork the repository on GitHub, then clone your fork
   git clone https://github.com/YOUR-USERNAME/giggles.ai.git
   cd giggles.ai
   
   # Add upstream remote for staying in sync
   git remote add upstream https://github.com/original-owner/giggles.ai.git
   ```

2. **Create Development Branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

3. **Start Developing**
   ```bash
   # Simply open index.html in your browser
   open index.html  # macOS
   xdg-open index.html  # Linux
   start index.html  # Windows
   
   # Or use a simple HTTP server (optional)
   python -m http.server 8000  # Python 3
   # Then visit http://localhost:8000
   ```

## Development Environment

### Recommended VS Code Extensions

```json
{
  "recommendations": [
    "bradlc.vscode-tailwindcss",
    "esbenp.prettier-vscode",
    "ms-vscode.vscode-json",
    "christian-kohler.path-intellisense",
    "formulahendry.auto-rename-tag",
    "ms-vscode.live-server"
  ]
}
```

### Editor Configuration

Create `.editorconfig` in your project root:

```ini
root = true

[*]
charset = utf-8
end_of_line = lf
insert_final_newline = true
trim_trailing_whitespace = true
indent_style = space
indent_size = 2

[*.md]
trim_trailing_whitespace = false
```

## Coding Standards

### HTML Standards

```html
<!-- ✅ Good: Semantic HTML with proper structure -->
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Page Title - giggles.ai</title>
    <meta name="description" content="Page description for SEO">
</head>
<body>
    <header role="banner">
        <nav role="navigation" aria-label="Main navigation">
            <!-- Navigation content -->
        </nav>
    </header>
    
    <main role="main">
        <section aria-labelledby="section-heading">
            <h1 id="section-heading">Section Title</h1>
            <!-- Content -->
        </section>
    </main>
    
    <footer role="contentinfo">
        <!-- Footer content -->
    </footer>
</body>
</html>
```

**HTML Guidelines:**
- Use semantic HTML5 elements
- Include proper `lang` attribute
- Add ARIA labels for accessibility
- Use meaningful `alt` text for images
- Validate HTML using W3C validator

### CSS Standards

```css
/* ✅ Good: Well-organized CSS with clear naming */

/* CSS Custom Properties (Variables) */
:root {
  --color-primary: #2563eb;
  --color-secondary: #64748b;
  --color-accent: #f59e0b;
  --color-text: #1f2937;
  --color-background: #f8fafc;
  
  --font-family-primary: 'Inter', system-ui, sans-serif;
  --font-size-base: 1rem;
  --font-size-lg: 1.125rem;
  --font-size-xl: 1.25rem;
  
  --spacing-xs: 0.25rem;
  --spacing-sm: 0.5rem;
  --spacing-md: 1rem;
  --spacing-lg: 1.5rem;
  --spacing-xl: 2rem;
}

/* Base Styles */
* {
  box-sizing: border-box;
}

body {
  font-family: var(--font-family-primary);
  font-size: var(--font-size-base);
  line-height: 1.6;
  color: var(--color-text);
  background-color: var(--color-background);
  margin: 0;
  padding: 0;
}

/* Component Styles using BEM methodology */
.button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: var(--spacing-sm) var(--spacing-md);
  border: 2px solid transparent;
  border-radius: 0.375rem;
  font-weight: 500;
  text-decoration: none;
  transition: all 0.2s ease-in-out;
  cursor: pointer;
}

.button--primary {
  background-color: var(--color-primary);
  color: white;
}

.button--primary:hover {
  background-color: #1d4ed8;
  transform: translateY(-1px);
}

.button__icon {
  margin-right: var(--spacing-xs);
}
```

**CSS Guidelines:**
- Use CSS custom properties for consistent theming
- Follow BEM naming convention
- Mobile-first responsive design
- Use `rem` and `em` units for scalability
- Group related properties together
- Add comments for complex logic

### JavaScript Standards (Future)

```javascript
// ✅ Good: Clear, documented JavaScript

/**
 * Animates an element with a fun bounce effect
 * @param {HTMLElement} element - The element to animate
 * @param {number} duration - Animation duration in milliseconds
 * @returns {Promise} Promise that resolves when animation completes
 */
async function animateElement(element, duration = 300) {
  if (!element || !(element instanceof HTMLElement)) {
    throw new Error('Invalid element provided to animateElement');
  }

  return new Promise((resolve) => {
    element.style.transition = `transform ${duration}ms cubic-bezier(0.68, -0.55, 0.265, 1.55)`;
    element.style.transform = 'scale(1.05)';
    
    setTimeout(() => {
      element.style.transform = 'scale(1)';
      setTimeout(resolve, duration);
    }, duration / 2);
  });
}

// Usage with error handling
try {
  const button = document.querySelector('.button--primary');
  await animateElement(button);
} catch (error) {
  console.error('Animation failed:', error);
}
```

**JavaScript Guidelines:**
- Use modern ES6+ features
- Write descriptive function and variable names
- Add JSDoc comments for functions
- Use `const` and `let` instead of `var`
- Handle errors gracefully
- Use async/await for asynchronous operations

## Testing Strategy

### Manual Testing Checklist

#### Browser Testing
- [ ] Chrome (latest)
- [ ] Firefox (latest)
- [ ] Safari (latest)
- [ ] Edge (latest)
- [ ] Mobile Safari (iOS)
- [ ] Chrome Mobile (Android)

#### Accessibility Testing
- [ ] Keyboard navigation works
- [ ] Screen reader compatibility (test with NVDA/JAWS)
- [ ] Color contrast meets WCAG 2.1 AA standards
- [ ] Focus indicators are visible
- [ ] Alt text for images is descriptive

#### Performance Testing
- [ ] Page loads in under 3 seconds
- [ ] Images are optimized
- [ ] CSS and JS are minified (when applicable)
- [ ] No unnecessary network requests

### Automated Testing (Future)

```bash
# Install testing dependencies (future)
npm install --save-dev jest @testing-library/dom

# Run tests
npm test

# Run tests with coverage
npm run test:coverage
```

## Git Workflow

### Branch Naming

```bash
# Feature branches
feature/add-homepage-animation
feature/improve-accessibility

# Bug fixes
fix/navigation-mobile-issue
fix/css-overflow-problem

# Documentation
docs/update-contributing-guide
docs/add-api-documentation

# Refactoring
refactor/reorganize-css-structure
refactor/simplify-html-markup
```

### Commit Messages

Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```bash
# Format
type(scope): description

[optional body]

[optional footer(s)]

# Examples
feat(ui): add animated welcome message to homepage
fix(css): resolve mobile navigation overflow issue
docs(readme): update installation instructions
style(css): improve button hover animations
refactor(html): simplify navigation structure
test(ui): add accessibility tests for main navigation
```

### Pull Request Process

1. **Keep PRs Small**: Focus on one feature or fix per PR
2. **Write Clear Descriptions**: Explain what changed and why
3. **Include Screenshots**: For UI changes, show before/after
4. **Test Thoroughly**: Check your changes across browsers
5. **Update Documentation**: Keep docs in sync with code changes

## Code Review Guidelines

### For Authors
- [ ] Self-review your code before submitting
- [ ] Write clear commit messages
- [ ] Add comments for complex logic
- [ ] Test on multiple browsers
- [ ] Update relevant documentation

### For Reviewers
- Be kind and constructive
- Focus on code quality and maintainability
- Ask questions if something is unclear
- Suggest improvements, don't just point out problems
- Approve when code meets our standards

## Performance Best Practices

### HTML Optimization
- Use semantic elements for better parsing
- Minimize DOM depth and nesting
- Optimize images with proper formats and sizes
- Use lazy loading for below-the-fold content

### CSS Optimization
- Minimize unused CSS
- Use efficient selectors
- Leverage CSS custom properties
- Consider critical CSS inlining

### Image Optimization
```bash
# Recommended image formats and sizes
- WebP for modern browsers (fallback to JPEG/PNG)
- SVG for icons and simple graphics
- Responsive images with srcset
- Compress images without quality loss
```

## Debugging Tips

### Browser DevTools
- Use Chrome DevTools for performance profiling
- Test responsive design with device emulation
- Debug accessibility with Lighthouse
- Monitor network requests and loading times

### Common Issues
- **CSS not applying**: Check selector specificity
- **Layout breaking**: Inspect box model and overflow
- **Performance issues**: Use Lighthouse audit
- **Accessibility problems**: Use axe DevTools extension

## Resources

### Learning Resources
- [MDN Web Docs](https://developer.mozilla.org/) - Comprehensive web development reference
- [Web.dev](https://web.dev/) - Modern web development best practices
- [CSS-Tricks](https://css-tricks.com/) - CSS techniques and tips
- [A11y Project](https://www.a11yproject.com/) - Accessibility guidelines

### Tools
- [W3C HTML Validator](https://validator.w3.org/)
- [W3C CSS Validator](https://jigsaw.w3.org/css-validator/)
- [WAVE Web Accessibility Evaluation](https://wave.webaim.org/)
- [Lighthouse](https://lighthouse-ci.com/) - Performance and accessibility auditing

## Getting Help

### Community Support
- **GitHub Discussions**: Ask questions and share ideas
- **Discord**: Real-time community chat
- **Email**: [dev@giggles.ai](mailto:dev@giggles.ai) for development questions

### Documentation
- [Architecture Guide](ARCHITECTURE.md) - Technical architecture overview
- [Contributing Guide](CONTRIBUTING.md) - Contribution guidelines
- [README](../README.md) - Project overview and setup

---

<div align="center">
  <p><strong>Happy coding! 🎉💻</strong></p>
  <p><em>Building the future of AI education, one commit at a time!</em></p>
</div>