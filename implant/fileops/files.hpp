#pragma once
#include <windows.h>
#include "../shared/list.hpp"

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
};



class FileMgr {
	enum FileStatus {
		FileStart,
		FileOnGoing,
		FileDone,
		FileFail,
	};



	BOOL ProcessEntry(FileTasks* task);
	List<FileTasks> Tasks;
public:
	BOOL CheckTasks();
	BOOL InsertTask(UINT32 TaskID, HANDLE handle, TaskType type);
	HANDLE GetHandle(UINT32 TaskID);
	BOOL RemoveTask(UINT32 TaskID);


};

extern FileMgr* g_FileMgr;
