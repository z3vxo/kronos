#include "config.hpp"
#include "../utils/bytes.hpp"
#include <stdio.h>






// todo, clean this up, test code below
BOOL LoadConfig() {

	conf = AllocMemory<Config>(sizeof(struct Config));
	memcpy(conf, GetProfile(), GetProfileSize());
	
	UINT ProfileSize = GetProfileSize();

	g_ByteMgr->InitRead((PBYTE)conf, ProfileSize);
	conf->domaincounts = g_ByteMgr->Read4();

	char buf[256];
	for (int i = 0; i < conf->domaincounts; i++) {
		UINT Domainlen = g_ByteMgr->Read4();
		g_ByteMgr->ReadString((PBYTE)buf, Domainlen);
		memcpy(conf->domains[i].domain, buf, Domainlen);
		conf->domains[i].port = g_ByteMgr->Read4();
		conf->domains[i].isHttps = g_ByteMgr->Read4();	
	}

	UINT GetLen = g_ByteMgr->Read4();
	g_ByteMgr->ReadString((PBYTE)buf, GetLen);
	memcpy(conf->GetEndpoint, buf, GetLen);

	
	UINT PostLen = g_ByteMgr->Read4();
	g_ByteMgr->ReadString((PBYTE)buf, PostLen);
	memcpy(conf->PostEndpoint, buf, PostLen);
	



	
	

	
	


	return TRUE;




}

Config* conf = NULL;