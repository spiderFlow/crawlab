FROM golang:latest AS backend-build

WORKDIR /go/src/app
COPY ./backend .

ENV GO111MODULE on
ENV GOPROXY https://goproxy.io

RUN go install -v ./...

FROM node:latest AS frontend-build

ADD ./frontend /app
WORKDIR /app

# install frontend
#RUN npm config set unsafe-perm true
#RUN npm install -g yarn && yarn install

#RUN yarn install && yarn run build:prod 

RUN npm install && npm run build:prod
# images
FROM ubuntu:latest

# set as non-interactive
ENV DEBIAN_FRONTEND noninteractive

# set CRAWLAB_IS_DOCKER
ENV CRAWLAB_IS_DOCKER Y

# install packages
RUN chmod 777 /tmp \
	&& apt-get update \
	&& apt-get install -y curl git net-tools iputils-ping ntp ntpdate python3 python3-pip pipx nginx wget dumb-init \
	&& ln -s /usr/bin/pip3 /usr/local/bin/pip \
	&& ln -s /usr/bin/python3 /usr/local/bin/python


# install backend
# RUN pipx ensurepath && pipx install scrapy pymongo bs4 requests crawlab-sdk scrapy-splash

# add files
COPY ./backend/conf /app/backend/conf
COPY ./backend/data /app/backend/data
COPY ./backend/scripts /app/backend/scripts
COPY ./backend/template /app/backend/template
COPY ./nginx /app/nginx
COPY ./docker_init.sh /app/docker_init.sh

# copy backend files
RUN mkdir -p /opt/bin
COPY --from=tikazyq/crawlab:0.5.1 /opt/bin/crawlab /opt/bin
RUN cp /opt/bin/crawlab /usr/local/bin/crawlab-server

# copy frontend files
COPY  --from=tikazyq/crawlab:0.5.1 /app/dist /app/dist

# copy nginx config files
COPY ./nginx/crawlab.conf /etc/nginx/conf.d

COPY --from=ghcr.io/astral-sh/uv:0.7.5 /uv /uvx /bin/
# working directory
WORKDIR /app/backend

# timezone environment
ENV TZ Asia/Shanghai

# language environment
ENV LC_ALL C.UTF-8
ENV LANG C.UTF-8

# frontend port
EXPOSE 8080

# backend port
EXPOSE 8000

# start backend
CMD ["/bin/bash", "/app/docker_init.sh"]
