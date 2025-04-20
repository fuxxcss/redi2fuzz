
#include <assert.h>
#include <errno.h>
#include <fcntl.h>
#include <signal.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/mman.h>
#include <sys/shm.h>
#include <sys/types.h>
#include <sys/wait.h>
#include <unistd.h>
#include <afl/types.h>
#include <afl/config.h>
#include "afl/afl-fuzz.h"
#ifndef _AFL_H_
#define _AFL_H_

// Fork server logic. 
void afl_forkserver_start(void) {

	uint8_t tmp[4] = {0, 0, 0, 0};
  
   /* Phone home and tell the parent that we're OK. */
	write(FORKSRV_FD + 1, tmp, 4);
}
  
size_t afl_next_testcase(uint8_t *buf, size_t max_len) {

	size_t status, res = 0xffffff;

  	/* Wait for parent by reading from the pipe. Abort if read fails. */
	if (read(FORKSRV_FD, &status, 4) != 4) return 0;

  	/* we have a testcase - read it */
	status = read(0, buf, max_len);

   	/* report that we are starting the target */
	if (write(FORKSRV_FD + 1, &res, 4) != 4) return 0;
  
	return status;
}

void afl_end_testcase(void) {

	size_t waitpid_status = 0xffffff;

	if (write(FORKSRV_FD + 1, &waitpid_status, 4) != 4) exit(1);
}

#endif