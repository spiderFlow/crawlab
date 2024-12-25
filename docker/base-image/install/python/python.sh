#!/bin/bash

# Fail on error
set -e

# Get version from first argument
version="${1:-3.12}"

# Check if version is provided
if [ -z "$version" ]; then
	echo "Please provide a version number"
	exit 1
fi

# Install build dependencies
apt-get install -y \
	make \
	build-essential \
	libssl-dev \
	zlib1g-dev \
	libbz2-dev \
	libreadline-dev \
	libsqlite3-dev \
	wget \
	curl \
	llvm \
	libncursesw5-dev \
	xz-utils \
	tk-dev \
	libxml2-dev \
	libxmlsec1-dev \
	libffi-dev \
	liblzma-dev

# Install pyenv
curl https://pyenv.run | bash

# Create a file in $HOME/.pyenv-env.sh
cat > $HOME/.pyenv-env.sh << 'EOF'
export PYENV_ROOT="$HOME/.pyenv"
[[ -d $PYENV_ROOT/bin ]] && export PATH="$PYENV_ROOT/bin:$PATH"
eval "$(pyenv init -)"
eval "$(pyenv virtualenv-init -)"
EOF

# Make the file executable
chmod +x $HOME/.pyenv-env.sh

# Source it immediately for the rest of the installation
source $HOME/.pyenv-env.sh

# Install python ${version} via pyenv
pyenv install ${version}
pyenv global ${version}

# Verify
python_version=$(python -V)
if [[ $python_version =~ "Python ${version}" ]]; then
	:
else
	echo "ERROR: python version does not match. expect \"Python ${version}\", but actual is \"${python_version}\""
	exit 1
fi
pip_version=$(pip -V)
if [[ $pip_version =~ "python ${version}" ]]; then
	:
else
	echo "ERROR: pip version does not match. expected: \"python ${version}\", but actual is \"${pip_version}\""
	exit 1
fi

# Install python dependencies
pip install -r /app/install/python/requirements.txt

# Create symbolic links
ln -sf $(pyenv which python) /usr/local/bin/python
ln -sf $(pyenv which python3) /usr/local/bin/python3
ln -sf $(pyenv which pip) /usr/local/bin/pip

# After pip install
pip cache purge && \
rm -rf ~/.cache/pip/* && \
apt-get remove -y make build-essential && \
apt-get autoremove -y