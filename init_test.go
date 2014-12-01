package consul_apixtra

import (
	"os"
	"os/exec"
	"time"
	"github.com/armon/consul-api"
  "log"
)

func initConsul() {
	os.RemoveAll("/tmp/consul")
	exec.Command("consul", "agent", "-server", "-bootstrap", "-data-dir=/tmp/consul").Start()
	time.Sleep(time.Second * 15)
}

func stopConsul() {
	exec.Command("killall", "-TERM", "consul").Run()
}

//Connect establishes a connection to local running consul agent.
//Currently only localhost:8500 is supported.
func Connect() *consulapi.Client {
	client, err := consulapi.NewClient(consulapi.DefaultConfig())
	if err != nil {
		log.Fatal(err)
	}

	return client
}
