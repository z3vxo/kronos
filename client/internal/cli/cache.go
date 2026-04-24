package cli

import "sync"

type InfoCache struct {
	User         string
	Host         string
	ProcPath     string
	Pid          int32
	PPid         int32
	WinVer       string
	InternalIP   string
	ExternalIP   string
	IsElevated   bool
	Arch         byte
	LastCheckin  int64
	RegisterTime int64
}

type Cache struct {
	mu             sync.RWMutex
	AgentInfoCache *InfoCache
	AgentsCache    []Agent
}

func (c *Cache) PopulateInfoCache(a AgentInfoResp) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.AgentInfoCache = &InfoCache{
		User:         a.User,
		Host:         a.Host,
		ProcPath:     a.ProcPath,
		Pid:          a.Pid,
		PPid:         a.PPid,
		WinVer:       a.WinVer,
		InternalIP:   a.InternalIP,
		ExternalIP:   a.ExternalIP,
		IsElevated:   a.IsElevated,
		Arch:         a.Arch,
		LastCheckin:  a.LastCheckin,
		RegisterTime: a.RegisterTime,
	}
}

func (c *Cache) PopulateAgentsCache(a Agents) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.AgentsCache = make([]Agent, 0, len(a.Agent))
	for _, i := range a.Agent {
		c.AgentsCache = append(c.AgentsCache, i)
	}
}

func (c *Cache) GetAgentsCache() []Agent {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.AgentsCache
}

func (c *Cache) GetInfoCache() *InfoCache {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.AgentInfoCache
}

func (c *Cache) InvalidateAgents() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.AgentsCache = nil
}

func (c *Cache) InvalidateInfo() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.AgentInfoCache = nil
}
