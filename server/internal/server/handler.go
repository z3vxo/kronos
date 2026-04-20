package server

import (
	"bytes"
	"fmt"

	"github.com/z3vxo/kronos/internal/byte"
	"github.com/z3vxo/kronos/internal/database"
)

type AgentHandler struct {
	DB *database.DB
}

func ConvertToWindowsVer(major, minor, build int16) string {
	switch {
	case major == 10 && minor == 0 && build > 22000:
		return fmt.Sprintf("Windows 11 (Build %d)", build)
	case major == 10 && minor == 0:
		return fmt.Sprintf("Windows 10 (Build %d)", build)
	case major == 6 && minor == 3:
		return "Windows  8.1"
	case major == 6 && minor == 2:
		return "Windows 8"
	case major == 6 && minor == 1:
		return "Windows 7"
	case major == 6 && minor == 0:
		return "Windows Vista"
	case major == 5 && minor == 2:
		return "Windows XP (64-bit) / Server 2003"
	case major == 5 && minor == 1:
		return "Windows XP"
	case major == 5 && minor == 0:
		return "Windows 2000"
	default:
		return fmt.Sprintf("Unknwon Windows (%d.%d.%d)", major, minor, build)

	}
}

func (h *AgentHandler) HandleClientRegister(ip string, r *bytes.Reader) error {
	Client, err := byte.ExtractRegistrationDetails(ip, r)
	if err != nil {
		return err
	}

	ver := ConvertToWindowsVer(Client.Major, Client.Minor, Client.Build)
	CodeName := GenCodeName()
	err = h.DB.InsertAgent(Client.Guid, CodeName,
		Client.User, Client.Host,
		Client.InternaIP, Client.ExternalIP,
		Client.ProcPath, ver, Client.Pid, Client.IsElev)
	if err != nil {
		return err
	}

	return nil

}
