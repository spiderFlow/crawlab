#!/bin/bash

version="3.12"

# install pyenv
curl https://pyenv.run | bash

# add pyenv to path
echo 'export PYENV_ROOT="$HOME/.pyenv"' >> ~/.bashrc
echo '[[ -d $PYENV_ROOT/bin ]] && export PATH="$PYENV_ROOT/bin:$PATH"' >> ~/.bashrc
echo 'eval "$(pyenv init -)"' >> ~/.bashrc
echo 'eval "$(pyenv virtualenv-init -)"' >> ~/.bashrc

# ensure changes take effect immediately
export PYENV_ROOT="$HOME/.pyenv"
[[ -d $PYENV_ROOT/bin ]] && export PATH="$PYENV_ROOT/bin:$PATH"
eval "$(pyenv init -)"
eval "$(pyenv virtualenv-init -)"

# install python ${version} via pyenv
pyenv install ${version}
pyenv global ${version}

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
