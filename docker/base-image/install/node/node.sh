#!/bin/bash

version="22"

# Install nvm (Node Version Manager)
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.0/install.sh | bash

# Create a file in /etc/profile.d/
cat > /etc/profile.d/node-env.sh << 'EOF'
export NVM_DIR="/root/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
[ -s "$NVM_DIR/bash_completion" ] && \. "$NVM_DIR/bash_completion"
export NODE_PATH=/usr/lib/node_modules
EOF

# Make the file executable
chmod +x /etc/profile.d/node-env.sh

# Source the file to apply the environment variables
source /etc/profile.d/node-env.sh

# Download and install Node.js (you may need to restart the terminal)
nvm install ${version}

# Set node version and make it the default
nvm use ${version}
nvm alias default ${version}

# Verify the right Node.js version is in the environment
if [[ ! "$(node -v)" =~ ^v${version} ]]; then
	echo "Node.js version is not v${version}.x"
	exit 1
fi

# Install node dependencies
npm install -g \
	npm@latest \
	yarn \
	pnpm \
	crawlab-sdk@latest \
	puppeteer \
	playwright \
	playwright-chromium \
	crawlee

# Create symbolic links
ln -sf "$(nvm which node)" /usr/local/bin/node
ln -sf "$(nvm which npm)" /usr/local/bin/npm
ln -sf "$(nvm which yarn)" /usr/local/bin/yarn
ln -sf "$(nvm which pnpm)" /usr/local/bin/pnpm

# Clean up
npm cache clean --force && \
rm -rf ~/.npm