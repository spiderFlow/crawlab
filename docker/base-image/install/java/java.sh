#!/bin/bash

version="11.0.12-open"

# Install SDKMAN!
curl -s "https://get.sdkman.io" | bash

# Create persistent environment config for SDKMAN
cat > /etc/profile.d/sdkman.sh << 'EOF'
# SDKMAN configuration
export SDKMAN_DIR="$HOME/.sdkman"
[[ -s "$SDKMAN_DIR/bin/sdkman-init.sh" ]] && source "$SDKMAN_DIR/bin/sdkman-init.sh"

# Java environment variables
export JAVA_HOME="$SDKMAN_DIR/candidates/java/current"
export PATH="$JAVA_HOME/bin:$PATH"
EOF

# Make the file executable
chmod +x /etc/profile.d/sdkman.sh

# Source it immediately for the rest of the installation
source /etc/profile.d/sdkman.sh

# Install Java 11 (OpenJDK)
sdk install java ${version}

# Set Java 11 as default
sdk default java ${version}

# Create symbolic links
ln -sf "$(sdkman which java)" /usr/local/bin/java
ln -sf "$(sdkman which javac)" /usr/local/bin/javac

# Verify installations
java_version=$(java -version)
if [[ $java_version =~ "${version}" ]]; then
    :
else
    echo "ERROR: java version does not match. expect \"${version}\", but actual is \"${java_version}\""
    exit 1
fi
javac_version=$(javac -version)
if [[ $javac_version =~ "${version}" ]]; then
    :
else
    echo "ERROR: javac version does not match. expect \"${version}\", but actual is \"${javac_version}\""
    exit 1
fi