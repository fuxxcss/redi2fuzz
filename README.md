# Redi2Fuxx
* [What is it ?](#introduction)
* [Prepare DBMS](#prepare-targets)
   * [Redis](#redis)
   * [KeyDB](#keydb)
   * [Redis Stack](#redis-stack)
* [How to Install ?](#install)
* [How to Use ?](#fuzz)
* [ToDo](#todo)

## introduction
This is a fuzzing Project for redis-based Database Management System.
Target DBMS:
``` shell
1. Redis (key-value)
2. KeyDB (key-value)
3. Redis Stack (Multi-model)
4. (...)
```

## prepare targets

### redis
redis fuzz required:
- instrument redis (disable shared)
- go-redis

instrument redis-server,if you dont have afl-clang-lto,look up here [afl-clang-lto](#afl-clang-lto).
because jemalloc/tcmalloc have collision with ASAN, so 'MALLOC=libc' is needed.
``` shell
> cd /usr/local/redis
> AFL_USE_ASAN=1 CC=afl-clang-lto make MALLOC=libc -j4
```
initial testcases from [redis commands](https://redis.io/docs/latest/commands/) and DeepSeek.

### keydb
keydb fuzz required:
- instrument keydb (disable shared)
- go-redis

instrument keydb-server.
``` shell
> aptitude install libcurl4-openssl-dev
> cd /usr/local/keydb
> AFL_USE_ASAN=1 CC=afl-clang-lto CXX=afl-clang-lto++ make MALLOC=libc -j4
```
keydb is a fork of redis,so we reuse input/redis.

### redis-stack
redis stack fuzz required:
- instrument redis (disable shared)
- go-redis
- redis-stack-server

Download redis-stack-server from [redis stack server](https://redis.io/downloads/#redis-stack-downloads). Copy redis-stack-server 、etc and lib into /usr/local/redis/src/.
add
``` shell
REDIS_DATA_DIR=/usr/local/redis/src/redis-stack
echo "Starting redis-stack-server, database path ${REDIS_DATA_DIR}"
CMD=/usr/local/redis/src/redis-server
CONFFILE=/usr/local/redis/src/etc/redis-stack.conf
MODULEDIR=/usr/local/redis/src/lib
```
before
``` shell
${CMD} \
${CONFFILE} \
--dir ${REDIS_DATA_DIR} \
...
```
activate redis stack :
``` shell
> mkdir redis-stack
> chmod +x ./redis-stack-server
```

## install

### go build
please do thease after dbms init.
``` shell
> go install -buildmode=shared -linkshared std
> ./install.sh
```

### afl build
in order to use shmem for afl-fuzz, dbms server.<br>
add
``` shell
char *id_str = getenv("SHM_ID");
if (id_str) {
    shm->shm_id = atoi(id_str);
}
```
into afl-sharedmem.c, between 
``` shell
shm->shm_id =
    shmget(IPC_PRIVATE, map_size == MAP_SIZE ? map_size + 8 : map_size,
        IPC_CREAT | IPC_EXCL | DEFAULT_PERMISSION);
<here>
if (shm->cmplog_mode) {
    ...
}
```
#### afl-clang-lto
in order to use afl-clang-lto, for example, your llvm version is 16 and lld-16 was installed.
``` shell
> export LLVM_CONFIG=llvm-config-16
> make && make install
```
#### afl-gxx-fast
in order to use afl-gxx-fast, for example, your gcc version is 12 and gcc-12-plugin-dev was installed.
``` shell
> make && make install
```

## fuzz

fuxx different redis (maybe need to trash /root/dump.rdb first) : 
``` shell
...
```

## ToDo
1. fix testcase length bug.
2. learn from Redis CVEs.
``` shell
// use aflpp-havoc to mutate integer argument, identifier
[CVE-2024-51737] RediSearch – Integer Overflow with LIMIT or KNN Arguments Lead to RCE
[CVE-2024-51480] RedisTimeSeries –  Integer Overflow RCE
[CVE-2024-55656] RedisBloom –  Integer Overflow RCE
```
3. analysis bugs
