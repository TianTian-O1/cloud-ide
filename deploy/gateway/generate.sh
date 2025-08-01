#!/bin/bash

# 设置证书和密钥的文件名和路径
CERT_FILE="tls.crt"
KEY_FILE="tls.key"
CONFIGMAP_NAME="cloud-ide-gateway-secret"
NAMESPACE="cloud-ide"

# 设置证书有效期（可根据需要进行调整）
DAYS=365

# 生成自签名的证书和密钥
openssl req -new -newkey rsa:2048 -days $DAYS -nodes -x509 -keyout $KEY_FILE -out $CERT_FILE

# 输出生成的证书和密钥的信息
echo "TLS 证书和密钥已生成："
echo "证书文件：$CERT_FILE"
echo "密钥文件：$KEY_FILE"


kubectl create secret tls $CONFIGMAP_NAME \
  --cert=$CERT_FILE \
  --key=$KEY_FILE \
  --namespace=$NAMESPACE

rm -f $CERT_FILE $KEY_FILE

echo "已创建 ConfigMap: $CONFIGMAP_NAME"


