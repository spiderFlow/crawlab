#!/bin/bash

version="3.10"

# install pyenv
curl https://pyenv.run | bash

# add pyenv to path
echo 'export PYENV_ROOT="$HOME/.pyenv"' >> ~/.bashrc
echo 'command -v pyenv >/dev/null || export PATH="$PYENV_ROOT/bin:$PATH"' >> ~/.bashrc
echo 'eval "$(pyenv init -)"' >> ~/.bashrc
source ~/.bashrc

# install python ${version} via pyenv
pyenv install ${version}
pyenv global ${version}

# install python build dependencies
apt-get install -y \
	make build-essential libssl-dev zlib1g-dev \
	libbz2-dev libreadline-dev libsqlite3-dev wget curl llvm \
	libncursesw5-dev xz-utils tk-dev libxml2-dev libxmlsec1-dev libffi-dev liblzma-dev

# alias
rm /usr/local/bin/pip | true
rm /usr/local/bin/python | true
ln -s /usr/local/bin/pip${version} /usr/local/bin/pip
ln -s /usr/bin/python${version} /usr/local/bin/python

# verify
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

# install python dependencies
pip install -r /app/install/python/requirements.txt
