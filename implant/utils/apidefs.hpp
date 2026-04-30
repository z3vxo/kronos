#pragma once
#include <Windows.h>
#include <wininet.h>
#include <iphlpapi.h>
#include "../shared/nt.hpp"

#define DECL(x) decltype(x) * x;

#define HASHED_GetUserNameA				 0xe7cb70d1
#define HASHED_LoadLibraryA				 0x8e338329
#define HASHED_GetModuleFileNameA		 0xd716f83d
#define HASHED_GetComputerNameExA		 0xcf275e49
#define HASHED_HeapAlloc				 0x1c67db00
#define HASHED_HeapFree					 0x4bdcb69a
#define HASHED_GetAdaptersInfo			 0xe0b877eb
#define HASHED_Kernel322				 0xee3fd4f9
#define HASHED_NTDLL					 0x2f09066c
#define HASHED_RtlGetVersion			 0x2b178a40
#define HASHED_GenRandom				 0x813458fb
#define HASHED_WSAStartup				 0xce11f5fb
#define HASHED_GetTickCount				 0xcfec99a7
#define HASHED_OpenProcessToken			 0x467981ce
#define HASHED_OpenThreadToken			 0xb1746ccd
#define HASHED_GetTokenInformation		 0x2a3d2a14
#define HASHED_CloseHandle				 0xd411d463
#define HASHED_InternerOpenA			 0x85a10a95
#define HASHED_AddHeaders				 0xea7c7c89
#define HASHED_InternetConnect			 0xe4a86f61
#define HASHED_OpenRequest				 0x15eddc50
#define HASHED_SendRequest				 0xeda89fa7
#define HASHED_InternetReadFile			 0x43d53dd0
#define HASHED_InternetCloseHandle		 0x54ea3e04
#define HASHED_InternetSetOptionA		 0x2a15e3ce
#define HASHED_NtQueryInformationProcess 0x8047d43f
#define HASHED_RtlGetNtVersionNumbers    0x2e8c8a48




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
	DECL(GetTickCount);
	DECL(CloseHandle);


		





	DECL(RtlGetVersion);
	DECL(RtlRandomEx);
	DECL(RtlIpv4StringToAddressA);
	DECL(NtOpenProcessToken);
	DECL(NtOpenThreadToken);
	DECL(NtQueryInformationProcess);
	DECL(RtlGetNtVersionNumbers);



	DECL(GetUserNameA);
	DECL(GetTokenInformation)

	
	DECL(GetAdaptersInfo);

	DECL(WSAStartup);



};






extern WINAPIS * WinApis;
extern MODULES* kModules;

BOOL LoadAPIS();
FARPROC GetProc(HMODULE dll, DWORD hash);
HMODULE GetModule(DWORD Hash);
