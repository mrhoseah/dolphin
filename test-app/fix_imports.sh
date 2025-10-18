#!/bin/bash

# Fix all import statements in Go files
echo "🔧 Fixing import statements..."

# Find all .go files recursively and replace the import paths
find . -name "*.go" -type f | while read -r file; do
    echo "Processing: $file"
    # Replace github.com/mrhoseah/dolphin with dolphin
    sed -i 's|github\.com/mrhoseah/dolphin|dolphin|g' "$file"
done

echo "✅ Import statements fixed!"
