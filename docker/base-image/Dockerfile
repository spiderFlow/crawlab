# images
FROM ubuntu:22.04

# set as non-interactive
ENV DEBIAN_FRONTEND=noninteractive

# copy install scripts
COPY ./install /app/install

# install all dependencies in parallel
RUN bash /app/install/deps/deps.sh & \
    bash /app/install/python/python.sh & \
    bash /app/install/golang/golang.sh & \
    bash /app/install/node/node.sh & \
    wait

# install chromedriver
RUN bash /app/install/chromedriver/chromedriver.sh

# working directory
WORKDIR /app/backend

# node path
ENV NODE_PATH=/usr/lib/node_modules

# timezone environment
ENV TZ=Asia/Shanghai

# language environment
ENV LC_ALL=C.UTF-8
ENV LANG=C.UTF-8

# frontend port
EXPOSE 8080

# backend port
EXPOSE 8000
