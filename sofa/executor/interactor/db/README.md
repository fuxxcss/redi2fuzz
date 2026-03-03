## Join DBMS

### Definition

1. DB Struct
- interface
- proc ...
- runtime ...

2. DB Model
- DB Line Sep
- DB Token Sep

### Functions

NewDB() DB
- global sep

### Interface

DB:
1. StartUp() error
2. Restart() error
3. ShutDown()
4. CheckAlive() bool
5. CleanUp() error
6. Execute([]string) (TargetState, error)
7. Collect() (Snapshot, error)
8. Stderr() string
9. Debug()

