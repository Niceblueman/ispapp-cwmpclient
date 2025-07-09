#!/bin/bash

# GitHub Actions compatible dependency installer
# This script is specifically designed for Ubuntu runners in GitHub Actions

set -e

echo "Installing OpenWrt build dependencies for GitHub Actions..."

# Update package list
sudo apt-get update

# Install dependencies with python3-setuptools (Ubuntu 20.04+ compatible)
sudo apt-get install -y \
  build-essential \
  clang \
  flex \
  bison \
  g++ \
  gawk \
  gcc-multilib \
  g++-multilib \
  gettext \
  git \
  libncurses5-dev \
  libssl-dev \
  python3-setuptools \
  python3-dev \
  rsync \
  unzip \
  zlib1g-dev \
  file \
  wget \
  curl \
  time

echo "Dependencies installed successfully!"

# Verify critical tools
echo "Verifying installation..."
for tool in wget curl tar make gcc g++ git unzip python3; do
    if command -v "$tool" >/dev/null 2>&1; then
        echo "✓ $tool"
    else
        echo "✗ $tool - MISSING"
        exit 1
    fi
done

echo "All dependencies verified successfully!"
