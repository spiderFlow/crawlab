#!/bin/bash

# version
version="stable"

# deps
apt-get install -y xvfb libxi6 libgconf-2-4

# install chrome
npx @puppeteer/browsers install chrome@${version}

# verify chrome version
if [[ ! "$(google-chrome --version)" =~ ^Google\ Chrome\ ${version} ]]; then
  echo "ERROR: chrome version does not match. expected: \"Google Chrome ${version}\", but actual is \"$(google-chrome --version)\""
  exit 1
fi

# install chromedriver
npx @puppeteer/browsers install chromedriver@${version}

# verify chromedriver version
if [[ ! "$(chromedriver --version)" =~ ^ChromeDriver\ ${version} ]]; then
  echo "ERROR: chromedriver version does not match. expected: \"ChromeDriver ${version}\", but actual is \"$(chromedriver --version)\""
  exit 1
fi
