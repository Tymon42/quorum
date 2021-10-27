// +build js,wasm

package wasm

import (
	"context"
	"crypto/rand"
	"fmt"
	"sync/atomic"
	"time"

	ethKeystore "github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/rumsystem/quorum/internal/pkg/options"
	quorumP2P "github.com/rumsystem/quorum/internal/pkg/p2p"
)

func StartQuorum(qchan chan struct{}, bootAddrsStr string) {
	ctx, cancel := context.WithCancel(context.Background())
	config := NewBrowserConfig([]string{bootAddrsStr})

	nodeOpt := options.NodeOptions{}
	nodeOpt.EnableNat = false
	nodeOpt.NetworkName = config.NetworkName
	nodeOpt.EnableDevNetwork = config.UseTestNet

	/* Randomly genrate a key
	TODO: should load from somewhere(IndexedDB or user localfile etc.) */
	key := ethKeystore.NewKeyForDirectICAP(rand.Reader)

	node, err := quorumP2P.NewBrowserNode(ctx, &nodeOpt, key)
	if err != nil {
		panic(nil)
	}

	wasmCtx := NewQuorumWasmContext(qchan, config, node, ctx, cancel)

	/* Bootstrap will connect to all bootstrap nodes in config.
	since we can not listen in browser, there is no need to anounce */
	Bootstrap(wasmCtx)

	/* TODO: should also try to connect known peers in peerstore which is
	   not implemented yet */

	/* keep finding peers, and try to connect to them */
	go startBackgroundWork(wasmCtx)
}

func Bootstrap(wasmCtx *QuorumWasmContext) {
	bootstraps := []peer.AddrInfo{}
	for _, peerAddr := range wasmCtx.Config.BootstrapPeers {
		peerinfo, _ := peer.AddrInfoFromP2pAddr(peerAddr)
		bootstraps = append(bootstraps, *peerinfo)
	}

	connectedPeers := wasmCtx.QNode.AddPeers(wasmCtx.Ctx, bootstraps)
	println(fmt.Sprintf("Connected to %d peers", connectedPeers))
}

func startBackgroundWork(wasmCtx *QuorumWasmContext) {
	/* first job will start after 1 second */
	go func() {
		time.Sleep(1 * time.Second)
		backgroundWork(wasmCtx)
	}()

	ticker := time.NewTicker(30 * time.Second)
	for {
		select {
		case <-ticker.C:
			backgroundWork(wasmCtx)
		case <-wasmCtx.Qchan:
			ticker.Stop()
			wasmCtx.Cancel()
		}
	}
}

func backgroundWork(wasmCtx *QuorumWasmContext) {
	println("Searching for other peers...")
	peerChan, err := wasmCtx.QNode.RoutingDiscovery.FindPeers(wasmCtx.Ctx, DefaultRendezvousString)
	if err != nil {
		panic(err)
	}

	var connectCount uint32 = 0

	for peer := range peerChan {
		curPeer := peer
		println("Found peer:", curPeer.String())
		go func() {
			pctx, cancel := context.WithTimeout(wasmCtx.Ctx, time.Second*10)
			defer cancel()
			err := wasmCtx.QNode.Host.Connect(pctx, curPeer)
			if err != nil {
				println("Failed to connect peer: ", curPeer.String())
				cancel()
			} else {
				curConnectedCount := atomic.AddUint32(&connectCount, 1)
				println(fmt.Sprintf("Connected to peer(%d): %s", curConnectedCount, curPeer.String()))
			}
		}()
	}
}
