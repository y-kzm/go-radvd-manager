package radvd

import (
	"github.com/freeconf/yang/node"
	"github.com/freeconf/yang/nodeutil"
)

// ///////////////////////
// R A D V D    M A N A G E M E N T
//
// Manage radvd (Router Advertisement Daemon) application using FreeCONF library
// according to the radvd.yang model file.
//
// Manage is root handler from radvd.yang. i.e. module radvd { ... }
func Manage(radvd *Radvd) node.Node {

	return &nodeutil.Node{

		// Root object for the radvd configuration
		Object: radvd,

		Options: nodeutil.NodeOptions{
			ActionInputExploded:  true,
			ActionOutputExploded: true,
		},

		OnAction: func(n *nodeutil.Node, r node.ActionRequest) (node.Node, error) {
			switch r.Meta.Ident() {
			case "start":

			default:
				return n.DoAction(r)
			}
			return nil, nil
		},
	}
}
