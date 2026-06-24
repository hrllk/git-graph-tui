package app

import "hrllk/graphkeeper/internal/graph"

type graphNode = graph.Node

type laneSide = graph.LaneSide

const (
	laneLocal  = graph.LaneLocal
	laneRemote = graph.LaneRemote
	laneOther  = graph.LaneOther
)

type laneRef = graph.LaneRef

type graphRow = graph.Row
