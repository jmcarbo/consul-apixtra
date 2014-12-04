package consul_apixtra

import (
	"errors"
	//"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/armon/consul-api"
	//"github.com/nu7hatch/gouuid"
	//"time"
)

type Lock struct {
	key              string
	session          *Session
	client           *consulapi.Client
	stopCh           chan struct{}
	doneCh           chan struct{}
	bInternalSession bool
}

func NewLock(client *consulapi.Client, key string) *Lock {
	if key == "" {
		log.Error("must supply key name")
		return nil
	}

	lock := &Lock{key: key, client: client}

	kvp, _, err := lock.client.KV().Get(lock.key, nil)
	if err != nil {
		return nil
	}
	if kvp == nil {
		_, err := lock.client.KV().Put(&consulapi.KVPair{Key: lock.key}, nil)
		if err != nil {
			return nil
		}
	}

	lock.doneCh = make(chan struct{})
	lock.stopCh = make(chan struct{})
	return lock
}

func (lock *Lock) Lock(session *Session) error {
	if lock.IsLeader() {
		//log.Info("Lock is leader ------")
		return nil
	}
	if session == nil && lock.session == nil {
		sess := NewSession(lock.client, "")
		if sess == nil {
			errors.New("Unable to get new session for lock")
		}
		lock.session = sess
		lock.bInternalSession = true
	}
	if session != nil && lock.session == nil {
		lock.session = session
	}

	if lock.session.IsHealthy() == false {
		log.Error("Session not healthy")
		lock.session.Destroy()
		sess2 := NewSession(lock.client, "")
		if sess2 == nil {
			errors.New("Unable to get new session for lock")
		}
		lock.session = sess2
		lock.bInternalSession = true
	}

	kvp, _, err := lock.client.KV().Get(lock.key, nil)
	if err != nil {
		return err
	}

	if kvp == nil {
		_, err := lock.client.KV().Put(&consulapi.KVPair{Key: lock.key}, nil)
		if err != nil {
			return errors.New("Non existant key")
		}
		kvp, _, err = lock.client.KV().Get(lock.key, nil)
		if err != nil {
			return err
		}
	}

	kvp.Session = lock.session.sessionID
	done, _, err := lock.client.KV().Acquire(kvp, nil)
	if err != nil {
		return err
	}
	if done {
		return nil
	} else {
		return errors.New("Unable to get lock")
	}

}

func (lock *Lock) IsLeader() bool {
	opts := &consulapi.QueryOptions{RequireConsistent: true}
	kvp, _, err := lock.client.KV().Get(lock.key, opts)
	if err != nil {
		//log.Error("@@@@@@@@ error getting log\n")
		return false
	}

	if kvp == nil {
		//log.Error("@@@@@@@@ key is nil\n")
		return false
	}

	if lock.session == nil {
		//log.Error("@@@@@@@@ session is nil\n")
		return false
	}

	if kvp.Session != lock.session.sessionID {
		//log.Error("@@@@@@@@ session not owned \n")
		return false
	}
	if kvp.Session == "" {
		//log.Error("@@@@@@@@ no sesson in key \n")
		return false
	}
	return true
}

func (lock *Lock) IsLocked() error {
	if lock.session == nil {
		return errors.New("No session in lock")
	}

	opts := &consulapi.QueryOptions{RequireConsistent: true}
	kvp, _, err := lock.client.KV().Get(lock.key, opts)
	if err != nil {
		return err
	}
	if kvp == nil {
		return errors.New("Key does not exist")
	}

	if kvp.Session == "" {
		return errors.New("Key not locked")
	}

	if kvp.Session != lock.session.sessionID {
		return errors.New("key locked by another session")
	}

	return nil
}

func (lock *Lock) Unlock() error {
	kvp, _, err := lock.client.KV().Get(lock.key, nil)
	if err != nil {
		return err
	}
	if kvp == nil {
		return errors.New("Key does not exist")
	}
	done, _, err := lock.client.KV().Release(kvp, nil)
	if err != nil {
		return err
	}
	if !done {
		return errors.New("Unable to unlock key")
	}
	return nil
}

func (lock *Lock) IsUnlocked() bool {
	kvp, _, err := lock.client.KV().Get(lock.key, nil)
	if err != nil {
		return false
	}
	if kvp == nil {
		return false
	}
	if kvp.Session != "" {
		return false
	}
	return true
}

func (lock *Lock) Destroy() error {
	if lock.IsLeader() {
		err := lock.Unlock()
		if err != nil {
			return err
		}
	}

	if lock.bInternalSession && lock.session != nil {
		err := lock.session.Destroy()
		if err != nil {
			return err
		}
	}
	return nil
}
