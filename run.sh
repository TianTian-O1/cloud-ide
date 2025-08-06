#!/bin/bash

./bin/control-plane -zap-log-level 5 -mode dev -gateway-token YOUR_GATEWAY_TOKEN_HERE -gateway-path /internal/endpoint -gateway-service YOUR_GATEWAY_SERVICE_IP -storage-class-name nfs-csi -zap-devel -dynamic-storage-enabled

