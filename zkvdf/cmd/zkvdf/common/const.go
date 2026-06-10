package common

type CircuitName string

const (
	HashToFormCircuitName       CircuitName = "hash_to_form"
	IntermediatePowCircuitName  CircuitName = "intermediate_pow"
	RCVerifierCircuitName       CircuitName = "rc_verifier"
	RCVerifierPhase1CircuitName CircuitName = "rc_verifier_phase_1"
	RCVerifierPhase2CircuitName CircuitName = "rc_verifier_phase_2"
	VerifierCircuitName         CircuitName = "verifier"
)
