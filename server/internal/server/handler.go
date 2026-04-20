package server

import (
	"bytes"
	"fmt"

	"github.com/z3vxo/kronos/internal/byte"
)

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

func HandleClientRegister(ip string, r *bytes.Reader) error {
	Client, err := byte.ExtractRegistrationDetails(ip, r)
	if err != nil {
		return err
	}

	fmt.Printf("[*] NEW AGENT\n")
	fmt.Printf("\t[+] Guid: %s\n", Client.Guid)
	fmt.Printf("\t[+] User: %s\n", Client.User)
	fmt.Printf("\t[+] Host: %s\n", Client.Host)
	fmt.Printf("\t[+] Internal IP: %s\n", Client.InternaIP)
	fmt.Printf("\t[+] External IP: %s\n", Client.ExternalIP)
	fmt.Printf("\t[+] Process Path: %s\n", Client.ProcPath)
	fmt.Printf("\t[+] Proc Identifier: %d\n", Client.Pid)
	if Client.IsElev == 0 {
		fmt.Printf("\t[+] Is Elevated: NO\n")
	} else {
		fmt.Printf("\t[+] Is Elevated: YES\n")
	}
	fmt.Printf("\t[+] Minor: %d\n", Client.Minor)
	fmt.Printf("\t[+] Major: %d\n", Client.Major)
	fmt.Printf("\t[+] Build: %d\n", Client.Build)
	fmt.Printf("\t[+] Human Readable: %s\n", ConvertToWindowsVer(Client.Major, Client.Minor, Client.Build))

	return nil

}
