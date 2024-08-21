package lbcommon

type Policy string

const (
	Random     Policy = "random"
	RoundRobin Policy = "round-robin"
	WRRSpeedup Policy = "wrr-speedup"
	WRRMemory  Policy = "wrr-memory"
	WRRCost    Policy = "wrr-cost"
	MAMA       Policy = "mama"
	CONST_HASH Policy = "const-hash"
)

const MAB = "[MAB]:"
