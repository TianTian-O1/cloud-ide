## build & deploy

**notice: first you should make sure you have a kubernetes cluster and version >= 1.23**

### 1、build

#### 1.1 build git-cloner
git-cloner is an image used to clone a specified repository from github or other git repository in an initialization container.
```sh
cd build/git-cloner
./build.sh
```

#### 1.2 build control-plane
control-plane is a kubernetes controller for reconcile Workspace resources.

```sh
# make sure you are in root path of the project
make docker WHAT=control-plane VERSION=v1.0   
```

#### 1.3 build gateway
gateway is based on openresty, which is used for service discovery for workspace.

```sh
# make sure you are in root path of the project
make docker WHAT=gateway VERSION=v1.0   
```

#### 1.4 build webserver
webserver is an web application based on gin framework, which is used for manage cloud ide

```sh
# make sure you are in root path of the project
make docker WHAT=webserver VERSION=v1.0   
```

`finally make sure you have the following images`
```sh
# docker images
REPOSITORY               TAG    IMAGE ID       SIZE
cloud-ide-webserver      v1.0  138b14aa3016   22.1MB
cloud-ide-gateway        v1.0  702e3b5dcc0d   96.7MB
cloud-ide-control-plane  v1.0  6beb47cb6dfc   41MB
git-cloner               v1.0  34782bbb7407   116MB
```

### 2、deploy

#### step1: deploy nfs and nfs-csi-driver
nfs is used to create a storage volume mounted to the workspace, which is used to store data such as the user's Code in the workspace and the VS Code plugin

nfs-csi-driver is a controller for dynamically preparing nfs storage volumes, you can also not deploy it, but you need to manually create PVS

`deploy nfs`
```sh
cd deploy/nfs-csi
kubectl create -f .
```
`deploy nfs-csi-driver`
```sh
# make sure you in deploy/nfs-csi
./install-driver.sh
```

#### step2: deploy control-plane
```sh
kubectl create ns cloud-ide
kubectl create ns cloud-ide-ws
# make sure you are in deploy/control-plane 
kubectl create -f .
```

#### step3: deploy webserver
```sh
# make sure you are in deploy/webserver
./gen-configmap.sh  # this is used to create configmap from sql/init.sql
kubectl create -f .
```

#### step4: deploy gateway
```sh
# make sure you are in deploy/gateway
# generate the nginx https certificate and key
./generate.sh
# deploy gateway
kubectl create -f .
```

in the end you will see the following result:
```sh
# kubectl get all -n cloud-ide
NAME                                          READY   STATUS    RESTARTS 
pod/cloud-ide-control-plane-5f8589cb5-bn29l   1/1     Running   0        
pod/cloud-ide-gateway-575ccf45d6-txz2b        1/1     Running   0        
pod/cloud-ide-web-6d778cf4c9-lrh4d            1/1     Running   0        
pod/mysql-7d8b965bcf-tsr7l                    1/1     Running   0        

NAME                                  TYPE        CLUSTER-IP       PORT(S)          
service/cloud-ide-control-plane-svc   ClusterIP   10.104.220.72    6387/TCP   
service/cloud-ide-gateway-svc         NodePort    10.99.212.202    443:30443/TCP    
service/cloud-ide-mysql-svc           ClusterIP   10.96.168.101    3306/TCP         
service/cloud-ide-web-svc             ClusterIP   10.100.194.142   8088/TCP         

NAME                                      READY   UP-TO-DATE   AVAILABLE  
deployment.apps/cloud-ide-control-plane   1/1     1            1          
deployment.apps/cloud-ide-gateway         1/1     1            1          
deployment.apps/cloud-ide-web             1/1     1            1          
deployment.apps/mysql                     1/1     1            1          

NAME                                                DESIRED   CURRENT   READY   
replicaset.apps/cloud-ide-control-plane-5f8589cb5   1         1         1       
replicaset.apps/cloud-ide-gateway-575ccf45d6        1         1         1       
replicaset.apps/cloud-ide-web-6d778cf4c9            1         1         1       
replicaset.apps/mysql-7d8b965bcf                    1         1         1       

```

