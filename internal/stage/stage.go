package stage

import (
	"sync"

	"github.com/EscapeBearSecond/falcon/pkg/types"
)

type Manager struct {
	m    sync.RWMutex
	core types.Stage
}

type Entry struct {
	Key   types.StageEntryName
	Value any
}

func NewEntry(key types.StageEntryName, value any) Entry {
	return Entry{
		Key:   key,
		Value: value,
	}
}

func NewManager() *Manager {
	return &Manager{}
}

func (p *Manager) Put(name types.StageName, percent float64, entries ...Entry) {
	if p == nil {
		return
	}

	p.m.Lock()
	defer p.m.Unlock()
	var core types.Stage
	core.Name = name
	core.Entries = make(map[types.StageEntryName]any)
	for _, entry := range entries {
		core.Entries[entry.Key] = entry.Value
	}
	core.Percent = percent
	p.core = core
}

func (p *Manager) Get() types.Stage {
	if p == nil {
		return types.Stage{}
	}

	p.m.RLock()
	defer p.m.RUnlock()
	return p.core
}
