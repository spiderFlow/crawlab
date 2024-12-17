#!/bin/bash

# Fail on error
set -e

# Ensure directory mode of /tmp is world-writable (readable, writable, executable by all users)
# This is important for temporary file operations in containerized environments
chmod 777 /tmp

# Update the package index files from the repositories
# This ensures we get the latest versions of packages
apt-get update

# Install common dependencies with detailed explanations
# -y flag means "yes" to all prompts (non-interactive installation)
apt-get install -y \
    curl \
    wget \
    git \
    net-tools \
    iputils-ping \
    ntp \
    ntpdate \
    nginx \
    unzip \
    gnupg2 \
    libc6

# Add source /etc/profile to ~/.bashrc
echo "source /etc/profile" >> ~/.bashrc