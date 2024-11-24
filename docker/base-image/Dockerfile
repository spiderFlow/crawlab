FROM ubuntu:22.04

# Non-interactive mode
ENV DEBIAN_FRONTEND=noninteractive

# Install dependencies
COPY ./install /app/install
RUN bash /app/install/deps/deps.sh && \
    bash /app/install/python/python.sh

# Working directory
WORKDIR /app/backend

# Node path
ENV NODE_PATH=/usr/lib/node_modules

# Timezone environment
ENV TZ=Asia/Shanghai

# language environment
ENV LC_ALL=C.UTF-8
ENV LANG=C.UTF-8

# frontend port
EXPOSE 8080

# backend port
EXPOSE 8000
