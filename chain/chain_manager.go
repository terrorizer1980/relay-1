package chain

import (
	"github.com/cihub/seelog"
	"github.com/eosforce/relay/chain/base/chain-msg"
	"github.com/eosforce/relay/chain/base/handler"
	"github.com/eosforce/relay/chain/main-chain"
	"github.com/eosforce/relay/chain/side-chain/eos"
	"github.com/eosforce/relay/types"
)

// Manager manager chain to watcher
type Manager struct {
	chainMsgHandler *chainMsg.Handler

	mainHandler Handler
	mainWatcher *handler.EosWatcher

	watchers map[string]*handler.EosWatcher
	// sideHandlers
}

// NewManager create manager
func NewManager(mainOpt WatchOpt, sideOpt ...WatchOpt) *Manager {
	res := &Manager{
		chainMsgHandler: chainMsg.NewChainMsgHandler(),

		mainHandler: mainChain.NewHandler(),
		mainWatcher: handler.NewEosWatcher(mainOpt.Type, mainOpt.Name, mainOpt.ApiURL, mainOpt.P2PAddresses),

		watchers: make(map[string]*handler.EosWatcher),
	}

	for _, opt := range sideOpt {
		res.watchers[opt.Name] = handler.NewEosWatcher(opt.Type, opt.Name, opt.ApiURL, opt.P2PAddresses)
	}

	return res
}

// Start start to process msg from all chain watching
func (m *Manager) Start() error {
	// Reg
	m.mainHandler.Reg(m.chainMsgHandler)
	m.mainWatcher.RegHandler(func(action types.ActionData) {
		if action.Name != "transfer" || action.Account != "eosio" {
			return
		}
		msg, err := m.mainHandler.Builder().BuildFromEOSForce(&action)
		if err != nil {
			seelog.Errorf("build msg err By %s", err.Error())
			return
		}
		seelog.Tracef("push msg %v %v", msg.MsgTyp, msg)
		m.chainMsgHandler.PushMsg(msg)
	})

	for name, side := range m.watchers {
		func(n string, w *handler.EosWatcher) {
			// TODO support diff chain
			h := eos.NewHandler(w.Name())
			h.Reg(m.chainMsgHandler)

			side.RegHandler(func(action types.ActionData) {
				// TODO By FanYang use common handler to mainchain
				if action.Name != "transfer" || action.Account != "eosio.token" {
					return
				}

				// TODO use side
				msg, err := h.Builder().BuildFromEOS(&action)
				if err != nil {
					seelog.Errorf("build msg err By %s", err.Error())
					return
				}
				seelog.Tracef("push msg %v %v", msg.MsgTyp, msg)
				m.chainMsgHandler.PushMsg(msg)
			})
		}(name, side)
	}

	// start chain msg handler
	m.chainMsgHandler.Start()

	// start each watcher
	err := m.mainWatcher.Start()
	if err != nil {
		return err
	}

	for name, side := range m.watchers {
		err := side.Start()
		if err != nil {
			seelog.Errorf("start watcher %s %s", name, err.Error())
			return err
		}
	}

	return nil
}

// Stop stop process
func (m *Manager) Stop() {
	seelog.Warn("start stop")
	// stop all watcher first
	for name, side := range m.watchers {
		seelog.Warnf("stop watcher %s", name)
		side.Stop()
	}
	seelog.Warnf("stop main watcher")
	m.mainWatcher.Stop()

	// wait
	for name, side := range m.watchers {
		seelog.Warnf("watcher wait stop %s", name)
		side.Wait()
	}
	seelog.Warnf("watcher to main wait stop")
	m.mainWatcher.Wait()

	// stop msg handler
	seelog.Warnf("stop msg handler")
	m.chainMsgHandler.Stop()
	seelog.Warnf("wait msg handler stop")
	m.chainMsgHandler.Wait()

}
