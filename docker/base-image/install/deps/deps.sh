#!/bin/bash

# Ensure directory mode of /tmp is world-writable (readable, writable, executable by all users)
# This is important for temporary file operations in containerized environments
chmod 777 /tmp

# Update the package index files from the repositories
# This ensures we get the latest versions of packages
apt-get update

# Install common dependencies with detailed explanations
# -y flag means "yes" to all prompts (non-interactive installation)
apt-get install -y \
    curl \       	# Modern HTTP client
    wget \       	# Another download utility
    git \        	# Distributed version control system
    net-tools \  	# Traditional networking tools
    iputils-ping \  # Tools for testing network connectivity
    ntp \        	# Network Time Protocol daemon for time sync
    ntpdate \   	# Client for one-time NTP sync
    nginx \     	# High-performance HTTP server and reverse proxy
    unzip \     	# Extract .zip archives
    gnupg2 \    	# GNU Privacy Guard for encryption and signing
    libc6       	# GNU C Library - essential for running C programs