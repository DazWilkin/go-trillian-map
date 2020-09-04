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
--hash_strategy=CONIKS_SHA512_256 \
--display_name="freddiemap" \
--description="First ever Trillian Map") && echo ${MAPID}
```

Results:

```SQL
SELECT * FROM `Trees`
```

And:

|TreeId|TreeState|TreeType|HashStrategy|HashAlgorithm|SignatureAlgorithm|DisplayName|Description|
|------|---------|--------|------------|-------------|------------------|-----------|-----------|
|3675790868355398218|ACTIVE|MAP|CONIKS_SHA512_256|SHA256|ECDSA|freddiemap|Trillian Map|
|3962226858249318401|ACTIVE|LOG|RFC6962_SHA256|SHA256|ECDSA|


> **NOTE** `LOG` was created with defaults

See: https://github.com/google/trillian/issues/1498


## Run

```bash
GRPC="53051"
go run github.com/DazWilkin/go-trillian-map/cmd/server \
--tmap_endpoint=:${GRPC} \
--tmap_id=${MAPID}
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
|3675790868355398218|0000000000000000000000000000000000000000000000000000000000000000|1|77 bytes|


