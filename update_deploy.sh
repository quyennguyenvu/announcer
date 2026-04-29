#!/bin/bash

# Script to update the deploy branch with a fresh binary
# Usage: ./update_deploy.sh

set -e

echo -e "\n🔄 Building fresh binary for Linux..."
GOOS=linux GOARCH=amd64 go build -o announcer

echo -e "\n📦 Backing up binary..."
cp announcer /tmp/announcer_deploy

echo -e "\n🔄 Switching to deploy branch..."
git checkout deploy

echo -e "\n📁 Updating binary..."
cp /tmp/announcer_deploy ./announcer

echo -e "\n📝 Committing updated binary..."
git add announcer
git commit -m "Update binary $(date '+%Y-%m-%d %H:%M:%S')"

echo -e "\n✅ Deploy branch updated successfully!"

echo -e "\n📤 Pushing deploy branch..."
git push origin deploy
echo -e "\n🔄 Switching back to main branch..."
git checkout main
