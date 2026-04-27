#include "common.hpp"
#include "apidefs.hpp"


WINAPIS* WinApis = NULL;
HTTPAPIS* HttpApis = NULL;
MODULES* kModules = NULL;

BOOL Init() {
	WinApis =  AllocMemory <WINAPIS>   (sizeof(struct WINAPIS));
	HttpApis = AllocMemory <HTTPAPIS>  (sizeof(struct HTTPAPIS));
	kModules = AllocMemory <MODULES>   (sizeof(struct MODULES));
	
	if (!LoadAPIS()) { return FALSE; }





}


