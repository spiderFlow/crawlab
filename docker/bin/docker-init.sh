#!/bin/bash

if [ "${CRAWLAB_NODE_MASTER}" == "Y" ]; then
    # Start nginx
    service nginx start >> /dev/null 2>&1
fi

# Start backend
crawlab-server server
