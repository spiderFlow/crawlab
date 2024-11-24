#!/bin/bash

if [ "${CRAWLAB_NODE_MASTER}" == "Y" ]; then
    # Replace default api path to new one
    python /app/bin/update_docker_js_api_address.py

    # Start nginx
    service nginx start
fi

# Start backend
crawlab-server server
