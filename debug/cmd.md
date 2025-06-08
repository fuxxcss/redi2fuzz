**queue**
``` shell
HSET myhash field2 "value2" field3 "value3"
HGETALL myhash  
HSET myhash field1 "new_value"
HGET myhash field1  
HSETNX myhash field4 "value4"  
HDEL myhash field4  
HEXISTS myhash field2
HMGET myhash field1 field2 field3
HKEYS myhash  
HVALS myhash  
HLEN myhash  
HSET myhash num 5
HINCRBY myhash num 3     
HINCRBY myhash new_num 10 
HSET myhash float 10.5
HINCRBYFLOAT myhash float -5.5  
HEXPIRE myhash 60 FIELDS 1 field1  
```

**snapshot**
``` shell
2025/06/07 17:33:27 execute HEXPIRE myhash 60 FIELDS 1 field1  
2025/06/07 17:33:27 old snapshot
2025/06/07 17:33:27 key myhash size 6  ->  field  size 0
2025/06/07 17:33:27 key myhash size 6  ->  field field2 size 6
2025/06/07 17:33:27 key myhash size 6  ->  field field3 size 6
2025/06/07 17:33:27 key myhash size 6  ->  field field1 size 6
2025/06/07 17:33:27 key myhash size 6  ->  field num size 3
2025/06/07 17:33:27 key myhash size 6  ->  field new_num size 7
2025/06/07 17:33:27 key myhash size 6  ->  field float size 5
2025/06/07 17:33:27 new snapshot
2025/06/07 17:33:27 key myhash size 6  ->  field  size 0
2025/06/07 17:33:27 key myhash size 6  ->  field field1 size 6
2025/06/07 17:33:27 key myhash size 6  ->  field field2 size 6
2025/06/07 17:33:27 key myhash size 6  ->  field field3 size 6
2025/06/07 17:33:27 key myhash size 6  ->  field num size 3
2025/06/07 17:33:27 key myhash size 6  ->  field new_num size 7
2025/06/07 17:33:27 key myhash size 6  ->  field float size 5
2025/06/07 17:33:27 keep snapshot
2025/06/07 17:33:27 key myhash size 6  ->  field  size 0
2025/06/07 17:33:27 key myhash size 6  ->  field field1 size 6
2025/06/07 17:33:27 key myhash size 6  ->  field field2 size 6
2025/06/07 17:33:27 key myhash size 6  ->  field field3 size 6
2025/06/07 17:33:27 key myhash size 6  ->  field num size 3
2025/06/07 17:33:27 key myhash size 6  ->  field new_num size 7
2025/06/07 17:33:27 key myhash size 6  ->  field float size 5
```

**graph**
``` shell
2025/06/07 17:33:29 cmdV type 0
2025/06/07 17:33:29 cmdV data HEXPIRE myhash 60 FIELDS 1 field1  
2025/06/07 17:33:29 cmdV prev
2025/06/07 17:33:29 type 1 myhash size 6 -> cmdV
2025/06/07 17:33:29 cmdV next
2025/06/07 17:33:29 all vertexs
2025/06/07 17:33:29 vertex type 1 myhash size 6
2025/06/07 17:33:29 myhash  -> HEXPIRE myhash 60 FIELDS 1 field1   size 35
2025/06/07 17:33:29 myhash  -> field1 size 6
2025/06/07 17:33:29 vertex type 2 field1 size 6
```

