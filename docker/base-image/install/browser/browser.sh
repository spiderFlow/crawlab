#!/bin/bash

# Source nvm environment from profile.d instead
source /etc/profile.d/node-env.sh

# Version - using "stable" for installation but not for verification
version="stable"

# Install dependencies
apt-get install -y \
	xvfb \
	libxi6 \
	libgconf-2-4 \
	libglib2.0-0 \
	libnss3 \
	libx11-6

# Install puppeteer browsers package globally first
npm install -g @puppeteer/browsers

# Install chrome with auto-yes and capture the output
INSTALL_OUTPUT=$(npx -y @puppeteer/browsers install chrome@${version} --install-deps)
echo "Installation output: $INSTALL_OUTPUT"

# Extract the actual version and path from the output
ACTUAL_VERSION=$(echo "$INSTALL_OUTPUT" | grep -o 'chrome@[^ ]*' | cut -d'@' -f2)
CHROME_BIN=$(echo "$INSTALL_OUTPUT" | awk '{print $2}')

echo "Detected Chrome version: $ACTUAL_VERSION"
echo "Chrome binary path: $CHROME_BIN"

# Update version variable for ChromeDriver
version="$ACTUAL_VERSION"

# Create symbolic links
ln -sf "$CHROME_BIN" /usr/local/bin/google-chrome

# Verify chrome is installed (with more detailed error message)
if ! command -v google-chrome &> /dev/null; then
    echo "ERROR: Chrome is not installed properly"
    echo "Chrome installation path: $(find /chrome -type f -name chrome 2>/dev/null)"
    echo "PATH environment: $PATH"
    echo "CHROME_PATH environment: $CHROME_BIN"
    exit 1
fi

# Install chromedriver with auto-yes and capture the output
CHROMEDRIVER_OUTPUT=$(npx -y @puppeteer/browsers install chromedriver@${version})
echo "ChromeDriver installation output: $CHROMEDRIVER_OUTPUT"

# Extract ChromeDriver path from the output
CHROMEDRIVER_BIN=$(echo "$CHROMEDRIVER_OUTPUT" | awk '{print $2}')
echo "ChromeDriver binary path: $CHROMEDRIVER_BIN"

# Create symbolic links
ln -sf "$CHROMEDRIVER_BIN" /usr/local/bin/chromedriver

# Verify chromedriver is installed
if ! command -v chromedriver &> /dev/null; then
    echo "ERROR: ChromeDriver is not installed properly"
    echo "ChromeDriver installation path: $CHROMEDRIVER_BIN"
    echo "PATH environment: $PATH"
    exit 1
fi

# Print installed versions for reference
echo "Installed Chrome version: $(google-chrome --version)"
echo "Installed ChromeDriver version: $(chromedriver --version)"

# Create a temporary directory for the test
TEST_DIR=$(mktemp -d)
cd "$TEST_DIR"

# Create a simple test script
cat > test.py << 'EOL'
from selenium import webdriver
from selenium.webdriver.chrome.options import Options

try:
    chrome_options = Options()
    chrome_options.add_argument('--headless')
    chrome_options.add_argument('--no-sandbox')
    chrome_options.add_argument('--disable-dev-shm-usage')
    
    driver = webdriver.Chrome(options=chrome_options)
    driver.get('https://www.google.com')
    print("ChromeDriver test successful!")
    driver.quit()
except Exception as e:
    print(f"ChromeDriver test failed: {str(e)}")
    exit(1)
EOL

# Run the test
python3 test.py

# Clean up
cd -
rm -rf "$TEST_DIR"

# After successful Chrome and ChromeDriver installation, create persistent env config
cat > /etc/profile.d/browser-env.sh << 'EOF'
# Chrome and ChromeDriver paths
export CHROME_BIN="$(which google-chrome)"
export CHROMEDRIVER_BIN="$(which chromedriver)"
# Common Chrome flags for headless environments
export CHROME_FLAGS="--headless --no-sandbox --disable-dev-shm-usage"
EOF

# Make the file executable
chmod +x /etc/profile.d/browser-env.sh

# Source it immediately
source /etc/profile.d/browser-env.sh
