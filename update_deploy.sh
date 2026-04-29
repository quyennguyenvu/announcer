#!/bin/bash

# Script to update the deploy branch with a fresh binary
# Usage: ./update_deploy.sh

set -e

echo "\n🔄 Building fresh binary for Linux..."
GOOS=linux GOARCH=amd64 go build -o announcer

echo "\n📦 Backing up binary..."
cp announcer /tmp/announcer_deploy

echo "\n🔄 Switching to deploy branch..."
git checkout deploy

echo "\n📁 Updating binary..."
cp /tmp/announcer_deploy ./announcer

echo "\n📝 Committing updated binary..."
git add announcer
git commit -m "Update binary $(date '+%Y-%m-%d %H:%M:%S')"

echo "\n✅ Deploy branch updated successfully!"

echo "\n📤 Pushing deploy branch..."
git push origin deploy
echo "\n🔄 Switching back to main branch..."
git checkout main
