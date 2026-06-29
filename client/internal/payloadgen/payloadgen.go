package payloadgen


import (
	"os"
	"path/filepath"
	"gopkg.in/yaml.v3"
	"fmt"
	"unicode/utf16"
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
	Headers     map[string]string `yaml:"headers"`
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


func LongestDomain(Domains []Domain) int {
	longest = 0
	for _, entry := range Domains {
		lenght := len(entry.Host)
		if lenght > longest {
			longest = lenght
		}
	}
	return longest
}

func MultiByteLength(val string) int {
	return len(utf16.Encode([]rune(val))) + 1
}


func GeneratePayload() error {
	conf, err := LoadProfile()
	if err != nil {
		return fmt.Errorf("Failed Loading Config: %w", err)
	}

	totalDomains = len(conf.Domains)
	LongestDomain = LongestDomain(conf.Domains)
	GetLen = MultiByteLength(conf.Endpoints.Get)
	PostLen = MultiByteLength(conf.Endpoints.Post)

}