#include "apidefs.hpp"
#include "common.hpp"




// shitty  algrorthim but just made it quick for testing
// ill fix it later
DWORD Hasher(const char* str) {
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

	kModules->K32 = GetModuleHandleA("kernel32");

	WinApis->GetUserNameA = (decltype(WinApis->GetUserNameA))GetProc(kModules->K32, 0x12222);
	WinApis->GetComputerNameExA = (decltype(WinApis->GetComputerNameExA))GetProc(kModules->K32, 0x12222);
	WinApis->GetModuleFileNameA = (decltype(WinApis->GetModuleFileNameA))GetProc(kModules->K32, 0x12222);








}