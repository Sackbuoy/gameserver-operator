package crds

import "sync"

type CRDInstanceMap struct {
  instances map[string]*GameServer
  accessMut sync.Mutex
}

func NewInstanceMap() (*CRDInstanceMap, error) {
  instances := make(map[string]*GameServer)

  return &CRDInstanceMap{
    instances: instances,
  }, nil
}

func (m *CRDInstanceMap) Create(instance *GameServer) error {
  m.accessMut.Lock()
  m.instances[instance.Name] = instance
  m.accessMut.Unlock()
  return nil
}

func (m *CRDInstanceMap) Update(instance *GameServer) error {
  m.accessMut.Lock()
  m.instances[instance.Name] = instance
  m.accessMut.Unlock()
  return nil
}

func (m *CRDInstanceMap) Delete(name string) error {
  m.accessMut.Lock()
  delete(m.instances, name)
  m.accessMut.Unlock()
  return nil
}

func (m *CRDInstanceMap) Read(name string) *GameServer {
  val, ok := m.instances[name]; if !ok {
    return nil
  }
  return val
}

func (m *CRDInstanceMap) List() []*GameServer {
  result := make([]*GameServer, len(m.instances))
  for _, instance := range m.instances {
    result = append(result, instance)
  }

  return result
}
