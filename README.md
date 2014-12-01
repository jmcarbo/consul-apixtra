```go
PACKAGE DOCUMENTATION

package consul_apixtra
    import "."


TYPES

type Lock struct {
    // contains filtered or unexported fields
}

func NewLock(client *consulapi.Client, key string) *Lock

func (lock *Lock) Destroy() error

func (lock *Lock) IsLeader() bool

func (lock *Lock) IsLocked() error

func (lock *Lock) IsUnlocked() bool

func (lock *Lock) Lock(session *Session) error

func (lock *Lock) Unlock() error

type Session struct {
    // contains filtered or unexported fields
}

func NewSession(client *consulapi.Client, name string) *Session

func (sess *Session) Destroy() error

func (sess *Session) IsDestroyed() bool

func (sess *Session) IsHealthy() bool
```

