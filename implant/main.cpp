#include <stdio.h>
#include "apidefs.hpp"
#include "common.hpp"


/*
	1. startup
	2. load apis and alloc needed structs
	3. collect user info
	4. go into loop
	5. dispatch tasks
	6. execute tasks

*/


int main() {


	printf("#define HASHED_GetAdaptersInfo 0x%08x\n", Hasher("GetAdaptersInfo"));
	return 1;

	if (!InitAgent()) {
		printf("Failed Loading\n");
		return 0;
	}
	

	return 1;
	
}