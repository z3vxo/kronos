#include "files.hpp"


BOOL FileMgr::InsertTask(UINT32 TaskID, HANDLE handle) {
	FileTasks* task = (FileTasks*)HeapAlloc(GetProcessHeap(), HEAP_ZERO_MEMORY, sizeof(FileTasks));
	if (!task) {
		return FALSE;
	}

	task->TaskID = TaskID;
	task->hProc = handle;
	task->Status = FileOnGoing;

	if (!this->head) {
		this->head = task;
		return TRUE;
	}

	FileTasks* cur = this->head;
	while (cur->next) {
		if (cur->TaskID == TaskID) {
			HeapFree(GetProcessHeap(), 0, task);
			return FALSE;
		}
		cur = cur->next;
	}

	if (cur->TaskID == TaskID) {
		HeapFree(GetProcessHeap(), 0, task);
		return FALSE;
	}

	cur->next = task;

	return TRUE;
}

BOOL FileMgr::ProcessEntry(FileTasks* task) {
	return TRUE;
}

BOOL FileMgr::CheckTasks() {
	return TRUE;
}