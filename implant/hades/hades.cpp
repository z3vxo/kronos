#include "../shared/common.hpp"
#include "hades.h"
#include <stdio.h>


BOOL RunHades() {
	printf("#define HASHED_NtDelayExecution 0x%08x\n", Hasher((PCHAR)"NtDelayExecution"));
	return 1;
	if (!InitAgent()) {
		printf("Failed Loading\n");
		return FALSE;
	}

}


Hades* hades = NULL;