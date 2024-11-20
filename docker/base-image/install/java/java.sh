#!/bin/bash

# Exit on error
set -e

# Update package list and install OpenJDK 11
DEBIAN_FRONTEND=noninteractive apt-get update && \
    apt-get install -y --no-install-recommends openjdk-11-jdk && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*
