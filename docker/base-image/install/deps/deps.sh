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
    # Network and File Transfer Utilities
    curl \       # Modern HTTP client, useful for API requests and downloads
    wget \       # Another download utility, often used in scripts
    
    # Version Control
    git \        # Distributed version control system
    
    # Network Diagnostics and Monitoring
    net-tools \  # Traditional networking tools (netstat, ifconfig, etc.)
    iputils-ping \ # Tools for testing network connectivity (ping)
    
    # Time Synchronization
    ntp \        # Network Time Protocol daemon for time sync
    ntpdate \    # Client for one-time NTP sync
    
    # Web Server
    nginx \      # High-performance HTTP server and reverse proxy
    
    # File Operations
    unzip \      # Extract .zip archives
    
    # Security and Encryption
    gnupg2 \     # GNU Privacy Guard for encryption and signing
    
    # System Libraries
    libc6        # GNU C Library - essential for running C programs
