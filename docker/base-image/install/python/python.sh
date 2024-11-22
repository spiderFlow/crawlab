#!/bin/bash

version="3.12"

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

# Add pyenv to path
echo 'export PYENV_ROOT="$HOME/.pyenv"' >> ~/.bashrc
echo '[[ -d $PYENV_ROOT/bin ]] && export PATH="$PYENV_ROOT/bin:$PATH"' >> ~/.bashrc
echo 'eval "$(pyenv init -)"' >> ~/.bashrc
echo 'eval "$(pyenv virtualenv-init -)"' >> ~/.bashrc

# Ensure changes take effect immediately
export PYENV_ROOT="$HOME/.pyenv"
[[ -d $PYENV_ROOT/bin ]] && export PATH="$PYENV_ROOT/bin:$PATH"
eval "$(pyenv init -)"
eval "$(pyenv virtualenv-init -)"

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
ln -sf $(pyenv which pip) /usr/local/bin/pip

# After pip install
pip cache purge && \
rm -rf ~/.cache/pip/* && \
apt-get remove -y make build-essential && \
apt-get autoremove -y