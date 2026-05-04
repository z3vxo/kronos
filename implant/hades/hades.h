#pragma once
#include <Windows.h>
#include "../shared/nt.hpp"
#include "../shared/common.hpp"
#include "../networkd/network.hpp"
#include "../utils/bytes.hpp"
#include <wininet.h>
#include <iphlpapi.h>


#define DECL(x) decltype(x) * x;

#ifdef _DEBUG
#define DEBUG_LOG(msg, ...) printf(msg "\n", ##__VA_ARGS__)
#else
#define DEBUG_LOG(msg, ...)
#endif


typedef struct {
	struct _MODULES {
		HMODULE K32;
		HMODULE IPHLPAPI;
		HMODULE WININET;
		HMODULE NTDLL;
		HMODULE ADVAPI32;
	} Modules;


	struct _NTAPIS {

		DECL(RtlGetVersion);
		DECL(RtlRandomEx);
		DECL(RtlIpv4StringToAddressA);
		DECL(NtOpenProcessToken);
		DECL(NtOpenThreadToken);
		DECL(NtQueryInformationProcess);
		DECL(RtlGetNtVersionNumbers);
		DECL(NtDelayExecution);

	} NtApis;

	struct _WINAPIS {

		// kernel32
		DECL(ReadFile);
		DECL(GetModuleFileNameA);
		DECL(GetComputerNameExA);
		DECL(LoadLibraryA);
		DECL(HeapAlloc);
		DECL(GetTickCount);
		DECL(CloseHandle);

		DECL(GetUserNameA);
		DECL(GetTokenInformation)


		DECL(GetAdaptersInfo);
		DECL(WSAStartup);

	} WinApis;

} Hades;





extern Hades* hades;

BOOL RunHades();