#!/bin/bash

# ensure directory mode of /tmp
chmod 777 /tmp

# update
apt-get update

# common deps
apt-get install -y \
	curl \
	git \
	net-tools \
	iputils-ping \
	ntp \
	ntpdate \
	nginx \
	wget \
	dumb-init \
	cloc \
	unzip \
	build-essential \
	gnupg2 \
	libc6
