package cli

import "sync"

type InfoCache struct {
	AgentID      int
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

type TaskMapping struct {
	Guid   string
	TaskID int
}

type Cache struct {
	mu             sync.RWMutex
	AgentInfoCache *InfoCache
	AgentsCache    []Agent
	TaskIdMap      map[int]string
	ListenersIdMap map[int]string
	AgentsIdMap    map[int]string
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

func (c *Cache) ResolveAgentID(id int) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, a := range c.AgentsCache {
		if int(a.AgentID) == id {
			return a.CodeName, true
		}
	}
	return "", false
}

func (c *Cache) InvalidateAgents() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.AgentsCache = nil
}
func (c *Cache) InvalidateOneAgent(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i, a := range c.AgentsCache {
		if a.CodeName == name {
			c.AgentsCache = append(c.AgentsCache[:i], c.AgentsCache[i+1:]...)
			for id, n := range c.AgentsIdMap {
				if n == name {
					delete(c.AgentsIdMap, id)
					break
				}
			}
			return
		}
	}
}

func (c *Cache) InvalidateInfo() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.AgentInfoCache = nil
}
