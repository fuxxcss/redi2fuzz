package main

/*
#include "afl/afl-fuzz.h"
*/
import "C"

import (
    "unsafe"
    "io"
    "os"
    "log"
)

// Fuxxer Server File
const (
	FDRIVER_R uintptr = iota +  3
	FDRIVER_W
	FMUTATOR_R
	FMUTATOR_W
)

// Fuxxer Server phone string
const (
	FSERVER_OK string = "ok"
	FSERVER_BAD string = "bad"
	FSERVER_ERR string = "err"
)

/**
 * Initialize this custom mutator
 *
 * @param[in] afl a pointer to the internal state object. Can be ignored for
 * now.
 * @param[in] seed A seed for this mutator - the same seed should always mutate
 * in the same way.
 * @return Pointer to the data object this custom mutator instance should use.
 *         There may be multiple instances of this mutator in one afl-fuzz run!
 *         Return NULL on error.
 */


//export afl_custom_init
func afl_custom_init(afl *C.afl_state_t,seed uint32) *int {
    
    return nil
}

/**
 * Perform custom mutations on a given input
 *
 * (Optional for now. Required in the future)
 *
 * @param[in] data pointer returned in afl_custom_init for this fuzz case
 * @param[in] buf Pointer to input data to be mutated
 * @param[in] buf_size Size of input data
 * @param[out] out_buf the buffer we will work on. we can reuse *buf. NULL on
 * error.
 * @param[in] add_buf Buffer containing the additional test case
 * @param[in] add_buf_size Size of the additional test case
 * @param[in] max_size Maximum size of the mutated output. The mutation must not
 *     produce data larger than max_size.
 * @return Size of the mutated output.
 */


//export afl_custom_fuzz
func afl_custom_fuzz(unused *int,buf *C.uint8_t,buf_size int,out_buf **C.uint8_t,
add_buf *uint8,add_buf_size int,max_size int) int {
    
    pipeR := os.NewFile(FMUTATOR_R, "Read")

    // read testcase
    testcase,err := io.ReadAll(pipeR)

    if err != nil {
        log.Fatalln("fuxxer io failed")
    }

    *out_buf = (*C.uint8_t)(unsafe.Pointer(C.CString(string(testcase))))

    return len(testcase)
}

/**
   * Deinitialize the custom mutator.
   *
   * @param data pointer returned in afl_custom_init by this custom mutator
   */

   
//export afl_custom_deinit
func afl_custom_deinit(not_use *int) {}

// needed by c-shared
func main() { /* empty */ }