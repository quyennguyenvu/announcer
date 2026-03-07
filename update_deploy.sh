#!/bin/bash

# Script to update the deploy branch with a fresh binary
# Usage: ./update_deploy.sh

set -e

echo "🔄 Building fresh binary for Linux..."
GOOS=linux GOARCH=amd64 go build -o announcer

echo "📦 Backing up binary..."
cp announcer /tmp/announcer_deploy

echo "🔄 Switching to deploy branch..."
git checkout deploy

echo "📁 Updating binary..."
cp /tmp/announcer_deploy ./announcer

echo "📝 Committing updated binary..."
git add announcer
git commit -m "Update binary $(date '+%Y-%m-%d %H:%M:%S')"

echo "✅ Deploy branch updated successfully!"

echo "📤 Pushing deploy branch..."
git push origin deploy
echo "🔄 Switching back to main branch..."
git checkout main
