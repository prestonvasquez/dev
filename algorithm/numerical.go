package algorithm

// GCD algorithm E1 from Knuth's TAOCP Volume 1, section 1.1.
func gcdAlgorithm1(high, low, remainder, iter int) (newH int, newL int, newR int, newI int) {
	if high == low {
		return high, 0, 0, 0
	}

IterSwitch:
	switch iter {
	case 1:
		newH, newL, newR, newI = gcdAlgorithm1(high, low, high%low, 2)
	case 2:
		if remainder == 0 {
			iter = 0

			goto IterSwitch
		}

		newH, newL, newR, newI = gcdAlgorithm1(high, low, remainder, 3)
	case 3:
		newH, newL, newR, newI = gcdAlgorithm1(low, remainder, remainder, 1)
	default:
		newH, newL, newR, newI = low, 0, 0, 0
	}

	return
}

// GCD algorithm proposed by ChatGPT:
func gcdAlgorithm2(high, low int) int {
	// Continue until the remainder becomes zero
	for low != 0 {
		high, low = low, high%low
	}

	return high
}

// Stein's GCD algorithm, AKA Binary GCD algorithm.
func gcdAlgorithm3(a, b int) int {
	// Base cases
	if a == 0 {
		return b
	}

	if b == 0 {
		return a
	}

	// Determine the greatest power of 2 that divides both a and b
	shift := 0
	for ((a | b) & 1) == 0 {
		a >>= 1
		b >>= 1
		shift++
	}

	// Divide a by 2 until it becomes odd
	for (a & 1) == 0 {
		a >>= 1
	}

	// Main loop
	for b != 0 {
		// Remove all factors of 2 in b, since they don't affect the GCD
		for (b & 1) == 0 {
			b >>= 1
		}

		// Ensure that a <= b
		if a > b {
			a, b = b, a
		}

		// Subtract a from b (which is now guaranteed to be even)
		b -= a
	}

	// Restore common factors of 2
	return a << shift
}

// GCD returns the greatest common divisor of two numbers using the Knuth
// algorithm.
func GCD(a, b int) int {
	res, _, _, _ := gcdAlgorithm1(a, b, 0, 1)

	return res
}
