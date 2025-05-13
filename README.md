# Redi2Fuxx
* [What is it ?](#introduction)
* [Prepare DBMS](#prepare-targets)
   * [Redis](#redis)
   * [KeyDB](#keydb)
   * [Valkey](#valkey)
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

### valkey
valkey fuzz required:
- instrument valkey (disable shared)
- go-redis

instrument valkey-server.
``` shell
> cd /usr/local/valkey
> AFL_USE_ASAN=1 CC=afl-clang-lto CXX=afl-clang-lto++ make MALLOC=libc -j4
```

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

## prepare testcases

The key point : ensuring that the initial testcases are grammatically and semantically correct.

initial testcases from [redis commands](https://redis.io/docs/latest/commands/) and DeepSeek/DouBao.

keydb and valkey is fork of redis,so we reuse input/redis.

## install

### go build
please do thease after dbms init.
``` shell
> go install -buildmode=shared -linkshared std
> ./install.sh
```

### afl build
in order to use shmem for afl-fuzz, dbms server.<br>
rewrite afl-sharedmem.c afl_shm_init()
``` shell
char *id_str = getenv("COVERAGE_MAP");
if (id_str) {
  shm->shm_id = atoi(id_str);
}else {
  shm->shm_id =
    shmget(IPC_PRIVATE, map_size == MAP_SIZE ? map_size + 8 : map_size,
           IPC_CREAT | IPC_EXCL | DEFAULT_PERMISSION);
}
```
add 
``` shell
setvbuf(stdout,NULL,_IOLBF,0);
setvbuf(stderr,NULL,_IOLBF,0);
```
into afl-fuzz.c main()

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

redi2fuxx useage :
``` shell
root@debian: r2f -h
A fuxxing tool for redis-based dbms with three mutation modes.

Usage:
  redi2fuxx [command]

Available Commands:
  analyze     Analyze Bugs.
  completion  Generate the autocompletion script for the specified shell
  fuxx        Ready to Fuxx.
  help        Help about any command

Flags:
  -h, --help            help for redi2fuxx
  -t, --target string   Fuxx Target (redis, keydb, redis-stack) (default "redis")
  -T, --tool string     Fuxx Base (afl, honggfuzz) (default "afl")

Use "redi2fuxx [command] --help" for more information about a command.
```

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
