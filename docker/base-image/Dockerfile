# Build stage
FROM ubuntu:22.04 AS builder
ENV DEBIAN_FRONTEND=noninteractive
COPY ./install /app/install
RUN bash /app/install/deps/deps.sh && \
    bash /app/install/python/python.sh && \
    bash /app/install/go/go.sh && \
    bash /app/install/node/node.sh && \
    bash /app/install/browser/browser.sh

# Final stage
FROM ubuntu:22.04
ENV DEBIAN_FRONTEND=noninteractive
COPY --from=builder /usr/local/bin /usr/local/bin
COPY --from=builder /usr/local/lib /usr/local/lib
COPY --from=builder /usr/lib /usr/lib
COPY --from=builder /app/install /app/install

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
