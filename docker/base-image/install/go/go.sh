#!/bin/bash

version="1.22.9"

# Install goenv
git clone https://github.com/go-nv/goenv.git ~/.goenv

# Add goenv to path
echo 'export GOENV_ROOT="$HOME/.goenv"' >> ~/.bashrc
echo 'export PATH="$GOENV_ROOT/bin:$PATH"' >> ~/.bashrc
echo 'eval "$(goenv init -)"' >> ~/.bashrc

# Ensure changes take effect immediately
export GOENV_ROOT="$HOME/.goenv"
export PATH="$GOENV_ROOT/bin:$PATH"
eval "$(goenv init -)"

# Install go
goenv install ${version}
goenv global ${version}

# Verify
go_version=$(go version)
if [[ $go_version =~ "go${version}" ]]; then
	:
else
	echo "ERROR: go version does not match. expect \"go${version}\", but actual is \"${go_version}\""
	exit 1
fi

# Create symbolic links
ln -sf "$(goenv which go)" /usr/local/bin/go
ln -sf "$(goenv which gofmt)" /usr/local/bin/gofmt