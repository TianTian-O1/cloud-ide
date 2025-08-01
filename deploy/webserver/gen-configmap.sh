#!/bin/bash

kubectl create configmap cloud-ide-mysql-init-sql --from-file=sql/init.sql --namespace=cloud-ide
