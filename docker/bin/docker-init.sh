#!/bin/bash

if [ "${CRAWLAB_NODE_MASTER}" == "Y" ]; then
    # Start nginx
    service nginx start
fi

# Start backend
crawlab-server server
