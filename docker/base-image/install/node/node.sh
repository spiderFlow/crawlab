#!/bin/bash

version="22"

# installs nvm (Node Version Manager)
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.0/install.sh | bash

# add nvm to path
echo 'export NVM_DIR="$HOME/.nvm"' >> ~/.bashrc
echo '[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"  # This loads nvm' >> ~/.bashrc
echo '[ -s "$NVM_DIR/bash_completion" ] && \. "$NVM_DIR/bash_completion"  # This loads nvm bash_completion' >> ~/.bashrc

# ensure changes take effect immediately
export NVM_DIR="$HOME/.nvm"
[[ -s "$NVM_DIR/nvm.sh" ]] && \. "$NVM_DIR/nvm.sh"  # This loads nvm
[[ -s "$NVM_DIR/bash_completion" ]] && \. "$NVM_DIR/bash_completion"  # This loads nvm bash_completion

# download and install Node.js (you may need to restart the terminal)
nvm install ${version}

# set node version and make it the default
nvm use ${version}
nvm alias default ${version}

# Create symbolic links
ln -sf "$(nvm which node)" /usr/local/bin/node
ln -sf "$(nvm which npm)" /usr/local/bin/npm
ln -sf "$(nvm which yarn)" /usr/local/bin/yarn
ln -sf "$(nvm which pnpm)" /usr/local/bin/pnpm

# verifies the right Node.js version is in the environment
if [[ ! "$(node -v)" =~ ^v${version} ]]; then
	echo "Node.js version is not v${version}.x"
	exit 1
fi

# install node dependencies
npm install -g \
	npm@latest \
	yarn \
	pnpm \
	crawlab-sdk@latest \
	puppeteer \
	playwright \
	playwright-chromium \
	crawlee

# Clean up
npm cache clean --force && \
rm -rf ~/.npm