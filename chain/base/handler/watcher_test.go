package handler

import (
	"testing"

	"github.com/eosforce/relay/const"
	"github.com/eosforce/relay/types"

	"time"

	"github.com/cihub/seelog"
)

func TestEosWatcher_Start(t *testing.T) {
	defer seelog.Flush()

	// start a chain http api :8889, p2p address 9001
	apiURL := "http://127.0.0.1:8890"
	p2pAddrs := []string{
		"127.0.0.1:9002",
		"127.0.0.1:9003",
		"127.0.0.1:9004",
		"127.0.0.1:9005",
		"127.0.0.1:9006",
		"127.0.0.1:9007",
		"127.0.0.1:9008",
		"127.0.0.1:9009",
		"127.0.0.1:9010",
	}

	watcher := NewEosWatcher(consts.TypeBaseEosforce, "eosforce1", apiURL, p2pAddrs)
	watcher.RegHandler(func(action types.ActionData) {
		seelog.Infof("action %d, %s, %s,",
			action.BlockNum, action.BlockID, action.TrxID)
	})
	err := watcher.Start()
	if err != nil {
		t.Errorf("start err by %s", err.Error())
		t.FailNow()
		return
	}

	time.Sleep(20 * time.Second)

	watcher.Stop()
	watcher.Wait()

}
