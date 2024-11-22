#!/bin/bash

# Source nvm environment
export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
[ -s "$NVM_DIR/bash_completion" ] && \. "$NVM_DIR/bash_completion"

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

# Chrome path
CHROME_PATH="/chrome/linux-${version}/chrome-linux64"
CHROME_BIN="$CHROME_PATH/chrome"

# Install chrome with auto-yes
npx -y @puppeteer/browsers install chrome@${version} \
	--install-deps \
    --path="$CHROME_PATH"

# Add chrome to PATH
ln -s "$CHROME_BIN" /usr/local/bin/google-chrome

# Verify chrome is installed (with more detailed error message)
if ! command -v google-chrome &> /dev/null; then
    echo "ERROR: Chrome is not installed properly"
    echo "Chrome installation path: $(find /chrome -type f -name chrome 2>/dev/null)"
    echo "PATH environment: $PATH"
    echo "CHROME_PATH environment: $CHROME_PATH"
    exit 1
fi

# Install chromedriver with auto-yes
npx -y @puppeteer/browsers install chromedriver@${version}

# Verify chromedriver is installed (without specific version check)
if ! command -v chromedriver &> /dev/null; then
    echo "ERROR: ChromeDriver is not installed properly"
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
