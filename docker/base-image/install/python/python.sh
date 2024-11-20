#!/bin/bash

# install pyenv
curl https://pyenv.run | bash

# add pyenv to path
echo 'export PYENV_ROOT="$HOME/.pyenv"' >> ~/.bashrc
echo 'command -v pyenv >/dev/null || export PATH="$PYENV_ROOT/bin:$PATH"' >> ~/.bashrc
echo 'eval "$(pyenv init -)"' >> ~/.bashrc
source ~/.bashrc

# install python build dependencies
apt-get install -y make build-essential libssl-dev zlib1g-dev \
libbz2-dev libreadline-dev libsqlite3-dev wget curl llvm \
libncursesw5-dev xz-utils tk-dev libxml2-dev libxmlsec1-dev libffi-dev liblzma-dev

# install python 3.10 via pyenv
pyenv install 3.10
pyenv global 3.10

# alias
rm /usr/local/bin/pip | true
rm /usr/local/bin/python | true
ln -s /usr/local/bin/pip3.10 /usr/local/bin/pip
ln -s /usr/bin/python3.10 /usr/local/bin/python

# verify
python_version=$(python -V)
if [[ $python_version =~ "Python 3.10" ]]; then
	:
else
	echo "ERROR: python version does not match. expect \"Python 3.10\", but actual is \"${python_version}\""
	exit 1
fi
pip_version=$(pip -V)
if [[ $pip_version =~ "python 3.10" ]]; then
	:
else
	echo "ERROR: pip version does not match. expected: \"python 3.10\", but actual is \"${pip_version}\""
	exit 1
fi

# install python dependencies
pip install -r /app/install/python/requirements.txt
