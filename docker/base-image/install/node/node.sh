#!/bin/bash

# installs nvm (Node Version Manager)
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.0/install.sh | bash

# download and install Node.js (you may need to restart the terminal)
nvm install 22

# set node version
nvm use 22

# verifies the right Node.js version is in the environment
if [[ ! "$(node -v)" =~ ^v22 ]]; then
  echo "Node.js version is not v22.x"
  exit 1
fi

# verifies the right npm version is in the environment
if [[ ! "$(npm -v)" =~ ^10 ]]; then
  echo "npm version is not 10.x"
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
