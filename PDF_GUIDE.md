# 📚 Dolphin Framework - PDF Guide Generation

This guide explains how to generate a comprehensive PDF guide for Dolphin Framework from the documentation.

## 🚀 Quick Start

### Method 1: Using Puppeteer (Recommended)

**Prerequisites:**
- Node.js installed
- Docusaurus website source code

**Steps:**

1. **Navigate to the documentation website directory**:
```bash
cd dolphin-website
```

2. **Install dependencies** (if not already installed):
```bash
npm install
```

3. **Start Docusaurus server** (in one terminal):
```bash
npm start
```

4. **Generate PDF** (in another terminal):
```bash
npm run pdf
```

**Output:** `build/dolphin-framework-guide.pdf`

### Method 2: Using Pandoc

**Prerequisites:**
- Pandoc installed
- LaTeX or wkhtmltopdf

**Steps:**

1. **Install Pandoc**:

**macOS:**
```bash
brew install pandoc wkhtmltopdf
```

**Ubuntu/Debian:**
```bash
sudo apt install pandoc wkhtmltopdf
```

**Windows:**
```powershell
choco install pandoc wkhtmltopdf
```

2. **Run the script**:
```bash
cd dolphin-website
./scripts/generate-pdf.sh
```

### Method 3: Browser Print to PDF

1. **Build the documentation**:
```bash
cd dolphin-website
npm run build
npm run serve
```

2. **Open in browser**: http://localhost:3000/docs

3. **Print to PDF**:
   - Press `Ctrl+P` (Windows/Linux) or `Cmd+P` (Mac)
   - Select "Save as PDF"
   - Choose "Save"

## 📋 What's Included in the PDF

The comprehensive PDF guide includes:

### Getting Started
- Introduction to Dolphin Framework
- Installation Guide (Windows, macOS, Linux)
- Recent Bug Fixes

### Core Features
- Routing
- Controllers
- Middleware
- Request & Response
- API Resources

### ORM & Database
- Models
- Relationships
- Query Builder
- Migrations
- Seeders
- Factories

### Templates
- Templates Overview
- Template Directives
- Components
- Layouts
- HTMX Integration

### Authentication
- Authentication Overview
- Guards
- JWT Authentication
- Auth Middleware

### Additional Sections
- Validation
- Caching
- Storage
- Events & Queues
- Advanced Features
- Enterprise Features
- Deployment Guides
- CLI Commands

## 🔧 Customization

### Add/Remove Sections

Edit `scripts/generate-pdf.sh` to customize included sections:

```bash
# Add a section
add_section "$DOCS_DIR/path/to/file.md" "Section Title"

# Remove a section
# Comment out or delete the add_section line
```

### Customize PDF Format

Edit `scripts/generate-pdf-online.js` for Puppeteer:

```javascript
await page.pdf({
    format: 'A4',  // or 'Letter', 'Legal', etc.
    margin: {
        top: '2cm',
        right: '2cm',
        bottom: '2cm',
        left: '2cm',
    },
});
```

Edit `scripts/generate-pdf.sh` for Pandoc:

```bash
pandoc "$COMBINED_MD" \
    -o "$PDF_OUTPUT" \
    --variable=geometry:margin=2cm \
    --variable=fontsize:11pt \
    # ... other options
```

## 📦 Package Scripts

The following npm scripts are available:

```bash
# Generate PDF using Puppeteer (requires server running)
npm run pdf

# Generate PDF using Pandoc (requires Pandoc installed)
npm run pdf:pandoc

# Build site and generate PDF
npm run pdf:build
```

## 🆘 Troubleshooting

### Puppeteer Issues

**Error: "Failed to launch browser"**

Install Chromium dependencies (Linux):
```bash
sudo apt-get install -y \
    libasound2 libatk1.0-0 libc6 libcairo2 libcups2 \
    libdbus-1-3 libexpat1 libfontconfig1 libgcc1 \
    libgconf-2-4 libgdk-pixbuf2.0-0 libglib2.0-0 \
    libgtk-3-0 libnspr4 libpango-1.0-0 libpangocairo-1.0-0 \
    libstdc++6 libx11-6 libx11-xcb1 libxcb1 libxcomposite1 \
    libxcursor1 libxdamage1 libxext6 libxfixes3 libxi6 \
    libxrandr2 libxrender1 libxss1 libxtst6 ca-certificates \
    fonts-liberation libappindicator1 libnss3 lsb-release \
    xdg-utils wget
```

### Pandoc Issues

**Error: "pdflatex not found"**

Install LaTeX:
```bash
# macOS
brew install --cask mactex

# Ubuntu/Debian
sudo apt install texlive-full

# Or use wkhtmltopdf instead
brew install wkhtmltopdf  # macOS
sudo apt install wkhtmltopdf  # Linux
```

### Docusaurus Server Not Running

Make sure the server is running before generating PDF:
```bash
npm start
# Keep this running, then in another terminal:
npm run pdf
```

## 📖 Output Details

- **Location**: `dolphin-website/build/dolphin-framework-guide.pdf`
- **Format**: A4 (customizable)
- **Features**:
  - Table of Contents
  - Page numbers
  - Clickable hyperlinks
  - Code syntax highlighting
  - Professional formatting
  - Headers and footers

## 🔄 Automated Generation

### GitHub Actions

Add to `.github/workflows/pdf.yml`:

```yaml
name: Generate PDF Guide

on:
  push:
    branches: [ main ]
  workflow_dispatch:

jobs:
  generate-pdf:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
        with:
          node-version: '20'
      - name: Install dependencies
        run: |
          cd dolphin-website
          npm install
      - name: Build Docusaurus
        run: |
          cd dolphin-website
          npm run build
      - name: Install Puppeteer dependencies
        run: |
          sudo apt-get update
          sudo apt-get install -y libgbm-dev
      - name: Generate PDF
        run: |
          cd dolphin-website
          npm run serve &
          sleep 10
          npm run pdf
      - name: Upload PDF
        uses: actions/upload-artifact@v3
        with:
          name: dolphin-framework-guide
          path: dolphin-website/build/dolphin-framework-guide.pdf
```

## 📚 Alternative: Manual PDF Creation

If automated generation doesn't work, you can:

1. **Visit the documentation site**: https://dolphin-docs.netlify.app/
2. **Use browser extensions**:
   - Chrome: "Print Friendly & PDF"
   - Firefox: "Save as PDF"
3. **Use online tools**:
   - https://www.markdowntopdf.com/
   - https://www.markdown-pdf.com/

## 🎯 Next Steps

- Generate your PDF guide
- Share it with your team
- Use it for offline reference
- Include it in releases

---

**Happy PDF generation! 🐬📄**

For more details, see the [PDF Guide in Documentation](https://dolphin-docs.netlify.app/docs/getting-started/pdf-guide).

