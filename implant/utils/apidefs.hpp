#pragma once
#include <Windows.h>
#include "../shared/nt.hpp"
#include "../hades/hades.h"


#define HASHED_GetUserNameA							 0xe7cb70d1
#define HASHED_LoadLibraryA							 0x8e338329
#define HASHED_GetModuleFileNameA					 0xd716f83d
#define HASHED_GetComputerNameExA					 0xcf275e49
#define HASHED_HeapAlloc						     0x1c67db00
#define HASHED_HeapFree								 0x4bdcb69a
#define HASHED_NtQueryInformationToken				 0xdeb15a53
#define HASHED_LookupPrivilegeNameA					 0xb409e14d
#define HASHED_CreateFileA							 0xc5efccc1
#define HASHED_ReadFile								 0x3fee4a9a
#define HASHED_DeleteFileA							 0xc9a828c1
#define HASHED_RemoveDirectoryA						 0x21763382
#define HASHED_CopyFileA							 0x3bbf1812
#define HASHED_FindFirstFileA						 0xf8eb0eb1
#define HASHED_FindNextFileA						 0x25efdade
#define HASHED_GetAdaptersInfo					     0xe0b877eb
#define HASHED_GetCurrentDirectoryA				 	 0x4e36830e
#define HASHED_SetCurrentDirectoryA  				 0x4e33070e
#define HASHED_MoveFileA						     0xbbb3d8d5
#define HASHED_Kernel322							 0xee3fd4f9
#define HASHED_GetProcAddress					     0xb1b6494f
#define HASHED_NTDLL								 0x2f09066c
#define HASHED_RtlGetVersion						 0x2b178a40
#define HASHED_CreateDirectoryA					     0xc1bf5ee1
#define HASHED_GenRandom							 0x813458fb
#define HASHED_WSAStartup							 0xce11f5fb
#define HASHED_GetTickCount							 0xcfec99a7
#define HASHED_OpenProcessToken						 0x467981ce
#define HASHED_OpenThreadToken						 0xb1746ccd
#define HASHED_GetTokenInformation					 0x2a3d2a14
#define HASHED_CloseHandle							 0xd411d463
#define HASHED_InternerOpenA						 0x85a10a95
#define HASHED_AddHeaders						     0xea7c7c89
#define HASHED_InternetConnect						 0xe4a86f61
#define HASHED_OpenRequest							 0x15eddc50
#define HASHED_SendRequest							 0xeda89fa7
#define HASHED_WriteFile							 0xc1b09c27
#define HASHED_InternetReadFile						 0x43d53dd0
#define HASHED_HttpQueryInfoA						 0x94887ec7
#define HASHED_InternetCloseHandle					 0x54ea3e04
#define HASHED_InternetSetOptionA					 0x2a15e3ce
#define HASHED_GetFileAttributesA					 0x6f81b238
#define HASHED_GetFullPathNameA						 0x988d675d
#define HASHED_NtQueryInformationProcess			 0x8047d43f
#define HASHED_RtlGetNtVersionNumbers				 0x2e8c8a48
#define HASHED_LookupAccountSidA					 0x932fbdf5
#define HASHED_NtOpenProcess						 0x82f3259b
#define HASHED_NtQuerySystemInformation				 0xb347a81f
#define HASHED_NtDelayExecution						 0xdafbb9c9
#define HASHED_WinHttpQueryHeaders					 0x17f2f0b8
#define HASHED_WinHttpOpen							 0x43a8f14a
#define HASHED_WinHttpConnect						 0xab17ddb5
#define HASHED_WinHttpCloseHandle					 0x3e6eddd0
#define HASHED_WinHttpOpenRequest					 0xd9c8f4e3
#define HASHED_WinHttpSendRequest					 0x859a3ee1
#define HASHED_WinHttpReceiveResponse				 0x679e5444
#define HASHED_WinHttpQueryDataAvailable			 0xb14a92df
#define HASHED_WinHttpReadData						 0xe3989deb
#define HASHED_WinHttpSetOption						 0xf399b28b
#define HASHED_WinHttpAddRequestHeaders				 0xf71b777b
#define HASHED_WinHttpGetIEProxyConfigForCurrentUser 0x782eb352
#define HASHED_WinHttpGetProxyForUrl				 0x6ab228e0
#define HASHED_WinHttpDetectAutoProxyConfigUrl		 0x9bd580a5
#define HASHED_WinHttpGetDefaultProxyConfiguration	 0x1da3264a
#define HASHED_HeapReAlloc							 0xd8e13b91
#define HASHED_ConvertSidToStringSidA				 0xe943efd1
#define HASHED_RtlNtStatusToDosError 0xE797520B
#define HASHED_NtTerminateProcess 0xF4FBDAF9
#define HASHED_RtlAdjustPrivilege 0x85E7EF8B
#define HASHED_NtDuplicateToken 0x2171600C
#define HASHED_NtSetInformationThread 0x3E370FB6





BOOL LoadAPIS();
FARPROC GetProc(HMODULE dll, DWORD hash);
HMODULE GetModule(DWORD Hash);
