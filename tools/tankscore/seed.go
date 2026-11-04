package tankscore

// SeedStride is how far apart Seed puts two consecutive user seeds' first iterations.
const SeedStride = 1_000_000_000

// Seed is the sim's RandomSeed for a tank tool's user seed (-seed of tankbench, abtest and
// rotopt's tank runs, default 1). The sim runs iteration i on RandomSeed + i (sim/core/sim.go), so
// seeds 1 and 2 would share all but one of their iterations (finding tank_rotations-20261104-13).
// Seed spaces the user seeds SeedStride apart, so two seeds' windows [Seed(s), Seed(s)+N) are
// disjoint for any N under a billion, while the default seed 1 maps to itself and reproduces the
// numbers benched before. A run stays reproducible, and the sim's replay and concurrent splits work
// on the window unchanged. Only the tank tools call it: the DPS sims keep the user's seed as is.
func Seed(userSeed int64) int64 {
	return 1 + (userSeed-1)*SeedStride
}

// SharedSeedSD is how much the standard deviation of a score's mean over several encounters
// (absolute scores, not paired differences) is understated when the error is worked out as if
// the encounters' runs were independent: every encounter of one run uses the same seed, so
// iteration i of each sees the same rolls and their scores correlate (about 0.11-0.14; the six-
// boss mean's SD is about 1.25 times the independent one, finding tank_rotations-20261104-13).
// Paired differences don't need it: their error covers the seed-to-seed spread.
const SharedSeedSD = 1.25
