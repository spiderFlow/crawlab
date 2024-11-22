#!/bin/bash

version="11.0.12-open"

# Install SDKMAN!
curl -s "https://get.sdkman.io" | bash

# Source SDKMAN!
source "$HOME/.sdkman/bin/sdkman-init.sh"

# Install Java 11 (you can specify vendor, e.g., 11.0.12-open for OpenJDK)
sdk install java ${version}

# Set Java 11 as default
sdk default java ${version}

# Create symbolic links
ln -sf "$(sdkman which java)" /usr/local/bin/java
ln -sf "$(sdkman which javac)" /usr/local/bin/javac

# Verify
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