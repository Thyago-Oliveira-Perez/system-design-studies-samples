// Package shared holds the types exchanged between the RPC server and client.
//
// In net/rpc both sides must agree on the method signature and the argument
// and reply types. Putting them in one package keeps server and client in
// sync — change it once and both recompile against the same contract.
package shared

// CalcArgs is the argument for the Calculator methods.
type CalcArgs struct {
	A, B float64
}

// The RPC endpoint name the client calls: "Calculator.Add", "Calculator.Multiply".
const ServiceName = "Calculator"

// Addr is where the server listens and the client dials.
const Addr = "localhost:1234"
