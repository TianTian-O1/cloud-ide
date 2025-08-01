#!/bin/bash

./bin/control-plane -zap-log-level 5 -mode dev -gateway-token XnRbVnoUZa0rT9xKAwHX0Zof3H7VpfCe -gateway-path /internal/endpoint -gateway-service 10.99.144.6 -storage-class-name nfs-csi -zap-devel -dynamic-storage-enabled

