#!/bin/bash

version="1.22.9"

# Install goenv
git clone https://github.com/go-nv/goenv.git ~/.goenv

# Create persistent environment config
cat > /etc/profile.d/goenv.sh << 'EOF'
export GOENV_ROOT="$HOME/.goenv"
export PATH="$GOENV_ROOT/bin:$PATH"
eval "$(goenv init -)"
EOF

# Make the file executable
chmod +x /etc/profile.d/goenv.sh

# Source it immediately for the rest of the installation
source /etc/profile.d/goenv.sh

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