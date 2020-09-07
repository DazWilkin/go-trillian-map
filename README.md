# Exploring [Trillian](https://github.com/google/trillian)'s Maps

See: [Map Mode](https://github.com/google/trillian#map-mode)

## Exploration

```bash
go list github.com/google/trillian/... | grep integration
github.com/google/trillian/integration
github.com/google/trillian/integration/admin
github.com/google/trillian/integration/maptest
github.com/google/trillian/integration/quota
github.com/google/trillian/integration/storagetest
github.com/google/trillian/testonly/integration
github.com/google/trillian/testonly/integration/etcd
```

And

```bash
go test github.com/google/trillian/integration/...
ok  	github.com/google/trillian/integration	26.472s
ok  	github.com/google/trillian/integration/admin	4.349s
ok  	github.com/google/trillian/integration/maptest	8.232s
ok  	github.com/google/trillian/integration/quota	2.560s
?   	github.com/google/trillian/integration/storagetest	[no test files]
```

## Create Map

```bash
docker-compose up db trillian-map-server
```

```bash
GRPC="53051"
MAPID=$(\
go run github.com/google/trillian/cmd/createtree \
--admin_server=:${GRPC} \
--tree_type=MAP \
--hash_strategy=CONIKS_SHA512_256) && echo ${MAPID}
```

Results:

```SQL
SELECT * FROM `Trees`
```

And:

|TreeId|TreeState|TreeType|HashStrategy|HashAlgorithm|SignatureAlgorithm|
|------|---------|--------|------------|-------------|------------------|
|`3675790868355398218`|`ACTIVE`|`MAP`|`CONIKS_SHA512_256`|`SHA256`|`ECDSA`|
|`3962226858249318401`|`ACTIVE`|`LOG`|`RFC6962_SHA256`|`SHA256`|`ECDSA`|


> **NOTE** `LOG` was created with defaults

See: https://github.com/google/trillian/issues/1498

## Build

### Local

```bash
BUILD_TIME="$(date --rfc-3339=seconds | sed 's| |T|')" # Replaces space with "T"
GIT_COMMIT="$(git rev-parse HEAD)"

GOOS=linux \
go build -a -installsuffix cgo \
-ldflags "-X main.buildTime=${BUILD_TIME} -X main.gitCommit=${GIT_COMMIT}" \
-o ./server \
./cmd/server
```

### Docker

```bash
TOKEN="..."
BUILD_TIME="$(date --rfc-3339=seconds | sed 's| |T|')" # Replaces space with "T"
GIT_COMMIT="$(git rev-parse HEAD)"

docker build \
--build-arg=TOKEN=${TOKEN} \
--build-arg=BUILD_TIME=${BUILD_TIME} \
--build-arg=GIT_COMMIT=${GIT_COMMIT} \
--tag=gcr.io/go-trillian-map/server:${GIT_COMMIT} \
--file=./deployment/Dockerfile.server \
.
```

## Run

### Local

```bash
GRPC="53051"

go run github.com/DazWilkin/go-trillian-map/cmd/server \
--tmap_endpoint=:${GRPC} \
--tmap_id=${MAPID} \
--tmap_rev=1
```

> **NOTE** Or use `./server` if built

### Docker

```bash
MAPID="..."
GRPC="53051"
GIT_COMMIT="$(git rev-parse HEAD)"

docker run \
--interactive --tty \
--net=host \
gcr.io/go-trillian-map/server:${GIT_COMMIT} \
--tmap_endpoint=:${GRPC} \
--tmap_id=${MAPID} \
--tmap_rev=...
```

Yields:

```console
2020/09/04 11:15:20 [Client:Add] revision:1
2020/09/04 11:15:20 [Client:Get] leaves:{index:"\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00"  leaf_value:"Freddie"}
```

And:

```SQL
SELECT *, HEX(`KeyHash`) AS `KeyHash` FROM `MapLeaf`
```

And:

|TreeId|KeyHash|MapRevision|LeafValue|
|------|-------|-----------|---------|
|`3675790868355398218|0000000000000000000000000000000000000000000000000000000000000000`|`1`|`77 bytes`|


## Kubernetes


```bash
NAMESPACE="cruithne"

kubectl create namespace ${NAMESPACE}
```

Then Service Account for GCR:

```bash
ACCOUNT="microk8s"

gcloud iam service-accounts create ${ACCOUNT} \
--project=${PROJECT}

gcloud iam service-accounts keys create ./${ACCOUNT}.json  \
--iam-account=${ACCOUNT}@${PROJECT}.iam.gserviceaccount.com \
--project=${PROJECT}

gcloud projects add-iam-policy-binding ${PROJECT} \
--member=serviceAccount:${ACCOUNT}@${PROJECT}.iam.gserviceaccount.com \
--role=roles/storage.objectViewer

kubectl create secret docker-registry gcr \
--docker-server=https://gcr.io \
--docker-username=_json_key \
--docker-email=dazwilkin@gmail.com \
--docker-password="$(cat ./${ACCOUNT}.json)" \
--namespace=${NAMESPACE}
```

Perhaps:

```bash
gcloud auth print-access-token |\
docker login -u oauth2accesstoken --password-stdin https://gcr.io
```

Create Map:

```bash
```bash
HOST=$(kubectl get service/map-server --namespace=trillian --output=jsonpath="{.spec.clusterIP}")
PORT="50051"
MAPID=$(go run github.com/google/trillian/cmd/createtree --admin_server=${HOST}:${PORT} --tree_type=MAP --hash_strategy=CONIKS_SHA512_256) && echo ${MAPID}
```

Apply `personality.server.yaml`

```bash
kubectl apply --filename=./kubernetes/personality.server.yaml --namespace=${NAMESPACE}
kubectl logs deployments/server --namespace=${NAMESPACE}
```

