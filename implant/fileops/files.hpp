#pragma once
#include <windows.h>

#define FILE_CHUNK_SIZE (256 * 1024)

#define UPLOAD_CHUNKED 1
#define UPLOAD_NO_CHUNKED 2
#define UPLOAD_DONE 3
#define TASK_TYPE_UPLOAD 6

enum TaskType {
	FileUpload,
	FileDownload,
};

struct FileTasks {
	HANDLE hProc;
	UINT Status;
	UINT TaskID;
	TaskType type;


	FileTasks* next;
};



class FileMgr {
	enum FileStatus {
		FileStart,
		FileOnGoing,
		FileDone,
		FileFail,
	};

	FileTasks* head;
	UINT ChunkSize = FILE_CHUNK_SIZE;

	BOOL ProcessEntry(FileTasks* task);
public:
	BOOL CheckTasks();
	BOOL InsertTask(UINT32 TaskID, HANDLE handle, TaskType type);
	HANDLE GetHandle(UINT32 TaskID);
	BOOL RemoveTask(UINT32 TaskID);
};

extern FileMgr* g_FileMgr;