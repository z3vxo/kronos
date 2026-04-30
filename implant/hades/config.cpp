#include "config.hpp"
#include "../utils/bytes.hpp"
#include <stdio.h>


//g_ByteMgr->InitWrite();
//g_ByteMgr->Write4(1);
//g_ByteMgr->Write4(strlen("192.168.1.24"));
//g_ByteMgr->WriteString((PBYTE)"192.168.1.24", strlen("192.168.1.24"));
//g_ByteMgr->Write4(0);
//g_ByteMgr->Write4(8080);
//g_ByteMgr->Write4(1);
//g_ByteMgr->Write4(strlen("TEST1234"));
//g_ByteMgr->WriteString((PBYTE)"TEST1234", strlen("TEST1234"));



// todo, clean this up, test code below
BOOL LoadConfig() {
	conf = (Config*)HeapAlloc(GetProcessHeap(), HEAP_ZERO_MEMORY, sizeof(struct Config));
	UINT ProfileSize = GetProfileSize();

	PBYTE ProfileBuf = AllocMemory<BYTE>(ProfileSize);
	memcpy(ProfileBuf, GetProfile(), ProfileSize);
	g_ByteMgr->InitRead(ProfileBuf, ProfileSize);
	UINT TotalDomains = g_ByteMgr->Read4();
	conf->domaincounts = TotalDomains;
	conf->domains = (DomainEntrys*)HeapAlloc(GetProcessHeap(), HEAP_ZERO_MEMORY, sizeof(struct DomainEntrys) * TotalDomains);
	for (int i = 0; i < TotalDomains; i++) {
		UINT DomainLen = g_ByteMgr->Read4();
		PCHAR Domain = (PCHAR)HeapAlloc(GetProcessHeap(), HEAP_ZERO_MEMORY, DomainLen);
		g_ByteMgr->ReadString((PBYTE)Domain, DomainLen);

		conf->domains[i].domain = Domain;
		conf->domains[i].isHttps = g_ByteMgr->Read4();
		conf->domains[i].port    = g_ByteMgr->Read4();
		conf->domains[i].UseSSL  = g_ByteMgr->Read4();
	}


	return TRUE;




}

Config* conf = NULL;