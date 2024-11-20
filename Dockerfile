ARG CRAWLAB_TAG=latest

FROM crawlabteam/crawlab-backend:${CRAWLAB_TAG} AS backend-build

FROM crawlabteam/crawlab-frontend:${CRAWLAB_TAG} AS frontend-build

FROM crawlabteam/crawlab-base:${CRAWLAB_TAG}

# copy backend files
RUN mkdir -p /opt/bin
COPY --from=backend-build /go/bin/crawlab /opt/bin
RUN cp /opt/bin/crawlab /usr/local/bin/crawlab-server

# copy backend config files
COPY ./backend/conf /app/backend/conf

# copy frontend files
COPY --from=frontend-build /app/dist /app/dist

# copy nginx config files
COPY ./docker/nginx/crawlab.conf /etc/nginx/conf.d

# copy docker bin files
RUN mkdir -p /app/bin
COPY ./docker/bin/* /app/bin

# start backend
CMD ["/bin/bash", "/app/bin/docker-init.sh"]
