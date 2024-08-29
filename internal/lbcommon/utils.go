package lbcommon

// GetAllPolicies returns a list of load balancing policies that are currently enabled.
func GetAllPolicies() []Policy {
	return []Policy{
		Random,
		RoundRobin,
		//MAMA,
		//CONST_HASH,
		//WRRSpeedup,
		//WRRMemory,
		//WRRCost,
	}
}
