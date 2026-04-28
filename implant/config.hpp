#pragma once
#include <windows.h>


struct DomainEntrys {
	PCHAR domain;
	UINT  port;
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

	UINT Sleep;
	UINT Jitter;
	UINT SyscallType; // 0 = no syscalls | 1 = direct | 2 = indirect
	UINT UseHeapObf; //  0 = no | 1 = yes
	UINT SleepObf;   //   0 = no | 1 = yes
};



//PCHAR GetProfile() {
//	return <*PROFILE_DATA_REPLACE*>
//}

//
//UINT ProfileSize() {
//	return <*PROFILE_SIZE_REPLACE*>
//}



//BOOL ParseAgentConfig();