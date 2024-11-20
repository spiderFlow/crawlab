#!/bin/bash

version="22"

# installs nvm (Node Version Manager)
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.0/install.sh | bash

# ensure changes take effect immediately
export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"  # This loads nvm
[ -s "$NVM_DIR/bash_completion" ] && \. "$NVM_DIR/bash_completion"  # This loads nvm bash_completion

# download and install Node.js (you may need to restart the terminal)
nvm install ${version}

# set node version and make it the default
nvm use ${version}
nvm alias default ${version}

# verifies the right Node.js version is in the environment
if [[ ! "$(node -v)" =~ ^v${version} ]]; then
	echo "Node.js version is not v${version}.x"
	exit 1
fi

# install node dependencies
npm install -g \
	yarn \
	pnpm \
	crawlab-sdk@latest \
	puppeteer \
	playwright \
	playwright-chromium \
	crawlee
