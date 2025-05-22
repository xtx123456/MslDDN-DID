package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
)

// Fraction struct to represent fractions
type Fraction struct {
	Numerator   *big.Int `json:"numerator"`
	Denominator *big.Int `json:"denominator"`
}

// NewFraction creates a new fraction
func NewFraction(numerator, denominator int64) Fraction {
	num := big.NewInt(numerator)
	den := big.NewInt(denominator)
	return Fraction{num, den}.Simplify()
}

// GCD calculates the greatest common divisor
func GCD(a, b *big.Int) *big.Int {
	return new(big.Int).GCD(nil, nil, a, b)
}

// Simplify simplifies the fraction
func (f Fraction) Simplify() Fraction {
	gcd := GCD(f.Numerator, f.Denominator)
	return Fraction{
		Numerator:   new(big.Int).Div(f.Numerator, gcd),
		Denominator: new(big.Int).Div(f.Denominator, gcd),
	}
}

// Float64 converts fraction to float64
func (f Fraction) Float64() float64 {
	num := new(big.Float).SetInt(f.Numerator)
	den := new(big.Float).SetInt(f.Denominator)
	res, _ := new(big.Float).Quo(num, den).Float64()
	return res
}

// Add adds two fractions
func (f Fraction) Add(other Fraction) Fraction {
	// (a/b) + (c/d) = (a*d + b*c) / (b*d)
	numerator := new(big.Int).Add(
		new(big.Int).Mul(f.Numerator, other.Denominator),
		new(big.Int).Mul(f.Denominator, other.Numerator),
	)
	denominator := new(big.Int).Mul(f.Denominator, other.Denominator)
	return Fraction{numerator, denominator}.Simplify()
}

// Multiply multiplies two fractions
func (f Fraction) Multiply(other Fraction) Fraction {
	numerator := new(big.Int).Mul(f.Numerator, other.Numerator)
	denominator := new(big.Int).Mul(f.Denominator, other.Denominator)
	return Fraction{numerator, denominator}.Simplify()
}

// Factorial calculates the factorial of a number n (n!)
func Factorial(n int64) *big.Int {
	result := big.NewInt(1)
	for i := int64(2); i <= n; i++ {
		result.Mul(result, big.NewInt(i))
	}
	return result
}

// Comb computes the combination C(n, k) = n! / (k! * (n - k)!)
func Comb(n, k int64) *big.Int {
	if k > n {
		return big.NewInt(0)
	}
	if k == 0 || k == n {
		return big.NewInt(1)
	}

	if k > n-k {
		k = n - k
	}

	num := Factorial(n)
	den := new(big.Int).Mul(Factorial(k), Factorial(n-k))
	return num.Div(num, den)
}

// pFail calculates the failure probability for passive configuration
func pFail(N, ta, ts, C int64, pMax Fraction, sa, ss Fraction) Fraction {
	p := Fraction{big.NewInt(0), big.NewInt(1)}
	denom := Comb(N, C)

	for i := int64(math.Floor(float64(C*ss.Numerator.Int64())/float64(ss.Denominator.Int64())) + 1); i <= int64(math.Ceil(float64(C*2)/3)); i++ {
		for j := int64(math.Floor(float64(C*sa.Numerator.Int64())/float64(sa.Denominator.Int64())) + 1); j <= int64(math.Ceil(float64(C*1)/3)); j++ {
			combTS := Comb(ts, i)
			combTA := Comb(ta, j)
			combNT := Comb(N-ts-ta, C-i-j)

			if C-i-j > 0 {
				p = p.Add(Fraction{combTS.Mul(combTS, combTA).Mul(combNT, combTS), denom})
			} else {
				p = p.Add(Fraction{combTS.Mul(combTS, combTA), denom})
			}
		}
	}
	return p
}

// pFail computes the probability of failure
func activeFail(n, t, s, h int64, pMax *big.Rat) *big.Rat {
	p := new(big.Rat).SetInt64(0)
	denom := Comb(n, s)

	for i := s - h + 1; i <= s; i++ {
		if i < 0 || i > t || (s-i) < 0 || (s-i) > (n-t) {
			continue // skip invalid combinations
		}

		// Calculate numerator: comb(t, i) * comb(n-t, s-i)
		numerator := new(big.Int).Mul(Comb(t, i), Comb(n-t, s-i))

		// Convert to rational and add to p
		term := new(big.Rat).SetFrac(numerator, denom)
		p.Add(p, term)
	}

	// Compare with pMax
	if p.Cmp(pMax) > 0 {
		return big.NewRat(1, 1) // return 1
	}
	return p
}

// minCsize finds the minimum committee size with corruption ratio at most cr such that pfail <= 2^(-k)
func activeMinsize(n, t int64, cr float64, k int64) int64 {
	// Calculate pMax = 1/(2^k)
	pMax := new(big.Rat).SetFrac(big.NewInt(1), new(big.Int).Exp(big.NewInt(2), big.NewInt(k), nil))

	for s := int64(1); s <= n; s++ {
		// We want at least h honest parties
		h := int64(math.Ceil((1 - cr) * float64(s)))
		if h < 0 {
			h = 0
		}

		p := activeFail(n, t, s, h, pMax)
		if p.Cmp(pMax) <= 0 {
			return s
		}
	}

	return n // return n if no smaller committee size found
}


// minCsize calculates the minimum C size for passive configuration
func minCsize(N, ta, ts int64, sa, ss Fraction, k int64) int64 {
	denominator := new(big.Int).Exp(big.NewInt(2), big.NewInt(int64(k)), nil)
	pMax := Fraction{big.NewInt(1), denominator}
	for C := int64(1); C <= N; C++ {
		pf := pFail(N, ta, ts, C, pMax, sa, ss)

		left := new(big.Int).Mul(pf.Numerator, pMax.Denominator)
		right := new(big.Int).Mul(pMax.Numerator, pf.Denominator)

		if left.Cmp(right) <= 0 {
			return C
		}
	}
	return 0
}

// ComputeCSizes computes the C sizes and returns results for passive configuration
func ComputeCSizes(ns []int64, ps [][]Fraction, k int64, sa []Fraction, ss []Fraction) map[string]interface{} {
	results := make(map[string]interface{})
	details := []map[string]interface{}{}

	for _, p := range ps {
		for _, n := range ns {
			nBig := big.NewInt(n)
			ta := new(big.Int).Mul(nBig, p[0].Numerator)
			ta.Div(ta, p[0].Denominator)

			ts := new(big.Int).Mul(nBig, p[1].Numerator)
			ts.Div(ts, p[1].Denominator)

			for _, s := range sa {
				for _, s2 := range ss {
					minC := minCsize(n, ta.Int64(), ts.Int64(), s, s2, k)
					detail := map[string]interface{}{
						"n":               n,
						"psa_numerator":   p[0].Numerator.String(),
						"psa_denominator": p[0].Denominator.String(),
						"pss_numerator":   p[1].Numerator.String(),
						"pss_denominator": p[1].Denominator.String(),
						"sa_numerator":    s.Numerator.String(),
						"sa_denominator":  s.Denominator.String(),
						"ss_numerator":    s2.Numerator.String(),
						"ss_denominator":  s2.Denominator.String(),
						"minCsize":        minC,
					}
					details = append(details, detail)
				}
			}
		}
	}

	results["details"] = details
	return results
}

// MinCChaincode implements the chaincode interface
type MinCChaincode struct{}

// Init is called during chaincode instantiation
func (t *MinCChaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success(nil)
}

// Invoke is called for each transaction
func (t *MinCChaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	function, args := stub.GetFunctionAndParameters()

	switch function {
	case "computeMinC":
		return t.computeMinC(stub, args)
	case "getDefaultResults":
		return t.getDefaultResults(stub)
	case "getactiveResults":
		return t.getactiveResults(stub)
	case "activeMinC":
		return t.activeMinC(stub, args)
	default:
		return shim.Error("Invalid function name. Expecting 'computeMinC', 'getDefaultResults', 'getactiveResults' or 'activeMinC'")
	}
}

// computeMinC handles custom computation requests
func (t *MinCChaincode) computeMinC(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1 (JSON input)")
	}

	input := struct {
		Ns []int64      `json:"ns"`
		Ps [][]Fraction `json:"ps"`
		K  int64       `json:"k"`
		Sa []Fraction   `json:"sa"`
		Ss []Fraction   `json:"ss"`
	}{}

	err := json.Unmarshal([]byte(args[0]), &input)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to parse input: %s", err))
	}

	results := ComputeCSizes(input.Ns, input.Ps, input.K, input.Sa, input.Ss)
	resultsBytes, err := json.Marshal(results)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to marshal results: %s", err))
	}

	return shim.Success(resultsBytes)
}

// activeMinC handles custom computation requests for active configuration
func (t *MinCChaincode) activeMinC(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1 (JSON input)")
	}

	input := struct {
		N   int64   `json:"n"`
		T   int64   `json:"t"`
		Cr  float64 `json:"cr"`
		K   int64   `json:"k"`
	}{}

	err := json.Unmarshal([]byte(args[0]), &input)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to parse input: %s", err))
	}
	

	result := activeMinsize(input.N, input.T, input.Cr, input.K)
	resultsBytes, err := json.Marshal(map[string]int64{"minCsize": result})
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to marshal results: %s", err))
	}

	return shim.Success(resultsBytes)
}

// getDefaultResults returns the same results as the original main function for passive configuration
func (t *MinCChaincode) getDefaultResults(stub shim.ChaincodeStubInterface) peer.Response {
	ns := []int64{500, 500}
	ps := [][]Fraction{
		{NewFraction(10, 100), NewFraction(40, 100)},
		{NewFraction(10, 100), NewFraction(40, 100)},
	}
	k := int64(60)
	sa := []Fraction{NewFraction(33, 100)}
	ss := []Fraction{NewFraction(66, 100)}

	results := ComputeCSizes(ns, ps, k, sa, ss)
	resultsBytes, err := json.Marshal(results)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to marshal results: %s", err))
	}

	return shim.Success(resultsBytes)
}

// getactiveResults returns the same results as the original main function for active configuration
func (t *MinCChaincode) getactiveResults(stub shim.ChaincodeStubInterface) peer.Response {
	n := int64(2000)
	tVal := int64(600)
	k := int64(60)
	cr := float64(0.9)

	result := activeMinsize(n, tVal, cr, k)
	resultsBytes, err := json.Marshal(map[string]int64{"minCsize": result})
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to marshal results: %s", err))
	}

	return shim.Success(resultsBytes)
}

// main function starts the chaincode
func main() {
	err := shim.Start(new(MinCChaincode))
	if err != nil {
		fmt.Printf("Error starting MinCChaincode: %s", err)
	}
}