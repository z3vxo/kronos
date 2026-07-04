package bytemgr

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strconv"

	"github.com/z3vxo/kronos/internal/database"
	"github.com/z3vxo/kronos/internal/files"
)

func Write4(w io.Writer, val any) error {
	return binary.Write(w, binary.LittleEndian, val)
}

func WriteString(w io.Writer, str string) error {
	strLen := uint32(len(str) + 1)
	if err := Write4(w, strLen); err != nil {
		return err
	}
	if _, err := w.Write([]byte(str)); err != nil {
		return err
	}
	if _, err := w.Write([]byte{0}); err != nil {
		return err
	}
	return nil
}





// func CraftCmdBytes(tasks []database.Task, fileMgr *files.Manager) ([]byte, error) {
// 	var buffer bytes.Buffer
// 	if err := Write4(&buffer, int32(len(tasks))); err != nil {
// 		return nil, err
// 	}
// 	for _, c := range tasks {
// 		if err := Write4(&buffer, int32(c.CmdCode)); err != nil {
// 			return nil, err
// 		}
// 		if err := Write4(&buffer, c.TaskID); err != nil {
// 			return nil, err
// 		}

// 		if c.CmdCode == 13 {
// 			if err := processUploadCmd(&buffer, c, fileMgr); err != nil {
// 				return nil, err
// 			}
// 			continue
// 		}

// 		if c.Param1 != "" {
// 			if err := WriteString(&buffer, c.Param1); err != nil {
// 				return nil, err
// 			}
// 		}
// 		if c.Param2 != "" {
// 			if err := WriteString(&buffer, c.Param2); err != nil {
// 				return nil, err
// 			}
// 		}
// 	}

// 	return buffer.Bytes(), nil
// }

func processUploadCmd(cmd database.Task, fileMgr *files.Manager) ([]byte, error) {
	var tmp bytes.Buffer

	uuid := cmd.Param2
	entry, ok := fileMgr.Uploads[uuid]
	if !ok {
		return nil, fmt.Errorf("upload entry not found: %s", uuid)
	}

	data, final, err := fileMgr.ReadUploadChunk(uuid)
	if err != nil {
		return nil, err
	}

	if err := Write4(&tmp, int32(cmd.CmdCode)); err != nil {
		return nil, err
	}
	if err := Write4(&tmp, cmd.TaskID); err != nil {
		return nil, err
	}

	switch entry.Status {
	case files.UploadStatusNotStarted:
		if final {
			if err := Write4(&tmp, uint32(UPLOAD_START_NON_CHUNKED)); err != nil {
				return nil, err
			}
		} else {
			if err := Write4(&tmp, uint32(UPLOAD_START_CHUNKED)); err != nil {
				return nil, err
			}
		}
		if err := WriteString(&tmp, cmd.Param1); err != nil {
			return nil, err
		}
		if !final {
			entry.Status = files.UploadStatusOngoing
		} else {
			entry.Status = files.UploadStatusDone
		}

	case files.UploadStatusOngoing:
		if final {
			if err := Write4(&tmp, uint32(UPLOAD_DONE)); err != nil {
				return nil, err
			}
		} else {
			if err := Write4(&tmp, uint32(UPLOAD_CONTINUE)); err != nil {
				return nil, err
			}
		}
		if final {
			entry.Status = files.UploadStatusDone
		}
	}

	if err := Write4(&tmp, uint32(len(data))); err != nil {
		return nil, err
	}
	if _, err := tmp.Write(data); err != nil {
		return nil, err
	}

	return tmp.Bytes(), nil
}


func CraftAgentResponse(Tasks []database.Task, FileMgr *files.Manager) ([]byte, error) {

	var buffer bytes.Buffer
	if err := Write4(&buffer, int32(len(Tasks))); err != nil {
		return nil, err
	}

	for _, task := range Tasks {
		if task.CmdCode == 13 {
			UploadBytes, err := processUploadCmd(task, FileMgr)
			if err != nil {
				return nil, err
			}
			buffer.Write(UploadBytes)
			continue
		}
		buffer.Write(task.Params)
	}
	return buffer.Bytes(), nil
}


func CraftCmdFormat(code, TaskID uint32, param1, param2, DataType string) ([]byte, error) {
	var buffer bytes.Buffer

	if err := Write4(&buffer, code); err != nil {
		return nil, err
	}
	if err := Write4(&buffer, TaskID); err != nil {
		return nil, err
	}


	if param1 == "" && param2 == "" {
		return buffer.Bytes(), nil
	}

	if DataType == "int" {

		val, err := strconv.Atoi(param1)
		if err != nil {
			return nil, err
		}
		if err := Write4(&buffer, uint32(val)); err != nil {
			return nil, err
		}

		if param2 != "" {
			val2, err := strconv.Atoi(param2)
			if err != nil {
				return nil, err
			}
			if err := Write4(&buffer, uint32(val2)); err != nil {
				return nil, err
			}
		}
	} else {
		if param1 != "" {
			if err := WriteString(&buffer, param1); err != nil {
				return nil, err
			}
		}
		if param2 != "" {
			if err := WriteString(&buffer, param2); err != nil {
				return nil, err
			}
		}
	}

	return buffer.Bytes(), nil


}