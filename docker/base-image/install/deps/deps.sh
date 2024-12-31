#!/bin/bash

# Fail on error
set -e

# Ensure directory mode of /tmp is world-writable
chmod 777 /tmp

# Update package index
apt-get update

# Install essential dependencies
apt-get install -y \
    curl \
    wget \
    zip \
    unzip \
    git \
    iputils-ping \
    nginx \
    jq \
    net-tools

# Add source /etc/profile to ~/.bashrc
echo "source /etc/profile" >> ~/.bashrc