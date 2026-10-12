package warlock

// The warlock's spell ranks from the Forever client (1.60.1.70338, wago.tools). See
// spellRank. Index 0 is empty so the slices are indexed by rank.
//
// For the spells with a direct hit and an over time part (Immolate) the two parts are two
// tables. Drain Soul and Shadowburn keep their damage in the client's second effect, the
// first is the shard proc. Life Tap's Base is health converted, before Spirit (one point of
// health per point of Spirit) and Improved Life Tap.

var shadowBoltRanks = []spellRank{
	{},
	{SpellID: 686, Level: 1, MaxLevel: 5, Base: 13, PerLevel: 0.2, Variance: 0.2857, Coef: 0.486, Cost: 25, CastMs: 1700},
	{SpellID: 695, Level: 6, MaxLevel: 11, Base: 23, PerLevel: 0.3, Variance: 0.2308, Coef: 0.629, Cost: 40, CastMs: 2200},
	{SpellID: 705, Level: 12, MaxLevel: 17, Base: 40, PerLevel: 0.6, Variance: 0.1538, Coef: 0.8, Cost: 70, CastMs: 2800},
	{SpellID: 1088, Level: 20, MaxLevel: 25, Base: 58, PerLevel: 0.8, Variance: 0.1304, Coef: 0.857, Cost: 110, CastMs: 3000},
	{SpellID: 1106, Level: 28, MaxLevel: 33, Base: 78, PerLevel: 1.2, Variance: 0.1316, Coef: 0.857, Cost: 160, CastMs: 3000},
	{SpellID: 7641, Level: 36, MaxLevel: 41, Base: 101, PerLevel: 1.2, Variance: 0.1198, Coef: 0.857, Cost: 210, CastMs: 3000},
	{SpellID: 11659, Level: 44, MaxLevel: 49, Base: 141, PerLevel: 1.4, Variance: 0.1141, Coef: 0.857, Cost: 265, CastMs: 3000},
	{SpellID: 11660, Level: 52, MaxLevel: 57, Base: 191, PerLevel: 1.6, Variance: 0.1102, Coef: 0.857, Cost: 315, CastMs: 3000},
	{SpellID: 11661, Level: 60, MaxLevel: 65, Base: 251, PerLevel: 1.8, Variance: 0.1081, Coef: 0.857, Cost: 370, CastMs: 3000},
	{SpellID: 25307, Level: 60, MaxLevel: 65, Base: 268, PerLevel: 1.9, Variance: 0.1098, Coef: 0.857, Cost: 380, CastMs: 3000},
}

var immolateRanks = []spellRank{
	{},
	{SpellID: 348, Level: 1, MaxLevel: 5, Base: 8, PerLevel: 0.7, Variance: 0, Coef: 0.2, Cost: 25, CastMs: 2000},
	{SpellID: 707, Level: 10, MaxLevel: 15, Base: 17, PerLevel: 0.8, Variance: 0, Coef: 0.2, Cost: 45, CastMs: 2000},
	{SpellID: 1094, Level: 20, MaxLevel: 25, Base: 32, PerLevel: 1.2, Variance: 0, Coef: 0.2, Cost: 90, CastMs: 2000},
	{SpellID: 2941, Level: 30, MaxLevel: 35, Base: 56, PerLevel: 1.5, Variance: 0, Coef: 0.2, Cost: 155, CastMs: 2000},
	{SpellID: 11665, Level: 40, MaxLevel: 45, Base: 72, PerLevel: 1.6, Variance: 0, Coef: 0.2, Cost: 220, CastMs: 2000},
	{SpellID: 11667, Level: 50, MaxLevel: 55, Base: 106, PerLevel: 1.9, Variance: 0, Coef: 0.2, Cost: 295, CastMs: 2000},
	{SpellID: 11668, Level: 60, MaxLevel: 65, Base: 146, PerLevel: 2.3, Variance: 0, Coef: 0.2, Cost: 370, CastMs: 2000},
	{SpellID: 25309, Level: 60, MaxLevel: 65, Base: 158, PerLevel: 2.3, Variance: 0, Coef: 0.2, Cost: 380, CastMs: 2000},
}

var immolateDotRanks = []spellRank{
	{},
	{SpellID: 348, Level: 1, MaxLevel: 5, Base: 3, PerLevel: 0, Variance: 0, Coef: 0.13, Cost: 25, CastMs: 2000},
	{SpellID: 707, Level: 10, MaxLevel: 15, Base: 6, PerLevel: 0, Variance: 0, Coef: 0.13, Cost: 45, CastMs: 2000},
	{SpellID: 1094, Level: 20, MaxLevel: 25, Base: 12, PerLevel: 0, Variance: 0, Coef: 0.13, Cost: 90, CastMs: 2000},
	{SpellID: 2941, Level: 30, MaxLevel: 35, Base: 19, PerLevel: 0, Variance: 0, Coef: 0.13, Cost: 155, CastMs: 2000},
	{SpellID: 11665, Level: 40, MaxLevel: 45, Base: 25, PerLevel: 0, Variance: 0, Coef: 0.13, Cost: 220, CastMs: 2000},
	{SpellID: 11667, Level: 50, MaxLevel: 55, Base: 38, PerLevel: 0, Variance: 0, Coef: 0.13, Cost: 295, CastMs: 2000},
	{SpellID: 11668, Level: 60, MaxLevel: 65, Base: 52, PerLevel: 0, Variance: 0, Coef: 0.13, Cost: 370, CastMs: 2000},
	{SpellID: 25309, Level: 60, MaxLevel: 65, Base: 55, PerLevel: 0, Variance: 0, Coef: 0.13, Cost: 380, CastMs: 2000},
}

var corruptionRanks = []spellRank{
	{},
	{SpellID: 172, Level: 4, MaxLevel: 9, Base: 10, PerLevel: 0, Variance: 0, Coef: 0.2, Cost: 35, CastMs: 2000},
	{SpellID: 6222, Level: 14, MaxLevel: 19, Base: 13, PerLevel: 0, Variance: 0, Coef: 0.2, Cost: 55, CastMs: 2000},
	{SpellID: 6223, Level: 24, MaxLevel: 29, Base: 22, PerLevel: 0, Variance: 0, Coef: 0.2, Cost: 100, CastMs: 2000},
	{SpellID: 7648, Level: 34, MaxLevel: 39, Base: 28, PerLevel: 0, Variance: 0, Coef: 0.2, Cost: 160, CastMs: 2000},
	{SpellID: 11671, Level: 44, MaxLevel: 49, Base: 40, PerLevel: 0, Variance: 0, Coef: 0.2, Cost: 225, CastMs: 2000},
	{SpellID: 11672, Level: 54, MaxLevel: 59, Base: 57, PerLevel: 0, Variance: 0, Coef: 0.2, Cost: 290, CastMs: 2000},
	{SpellID: 25311, Level: 60, MaxLevel: 65, Base: 73, PerLevel: 0, Variance: 0, Coef: 0.2, Cost: 340, CastMs: 2000},
}

var baneOfAgonyRanks = []spellRank{
	{},
	{SpellID: 980, Level: 8, MaxLevel: 0, Base: 6, PerLevel: 0, Variance: 0, Coef: 0.133, Cost: 25, CastMs: 0},
	{SpellID: 1014, Level: 18, MaxLevel: 0, Base: 10, PerLevel: 0, Variance: 0, Coef: 0.133, Cost: 50, CastMs: 0},
	{SpellID: 6217, Level: 28, MaxLevel: 0, Base: 14, PerLevel: 0, Variance: 0, Coef: 0.133, Cost: 90, CastMs: 0},
	{SpellID: 11711, Level: 38, MaxLevel: 0, Base: 21, PerLevel: 0, Variance: 0, Coef: 0.133, Cost: 130, CastMs: 0},
	{SpellID: 11712, Level: 48, MaxLevel: 0, Base: 33, PerLevel: 0, Variance: 0, Coef: 0.133, Cost: 170, CastMs: 0},
	{SpellID: 11713, Level: 58, MaxLevel: 0, Base: 46, PerLevel: 0, Variance: 0, Coef: 0.133, Cost: 215, CastMs: 0},
}

var siphonLifeRanks = []spellRank{
	{},
	{SpellID: 18265, Level: 30, MaxLevel: 38, Base: 11, PerLevel: 0, Variance: 0, Coef: 0.05, Cost: 150, CastMs: 0},
	{SpellID: 18879, Level: 38, MaxLevel: 48, Base: 19, PerLevel: 0, Variance: 0, Coef: 0.05, Cost: 205, CastMs: 0},
	{SpellID: 18880, Level: 48, MaxLevel: 58, Base: 29, PerLevel: 0, Variance: 0, Coef: 0.05, Cost: 285, CastMs: 0},
	{SpellID: 18881, Level: 58, MaxLevel: 68, Base: 41, PerLevel: 0, Variance: 0, Coef: 0.05, Cost: 365, CastMs: 0},
}

var drainLifeRanks = []spellRank{
	{},
	{SpellID: 689, Level: 14, MaxLevel: 19, Base: 10, PerLevel: 0, Variance: 0, Coef: 0.1, Cost: 55, CastMs: 0},
	{SpellID: 699, Level: 22, MaxLevel: 27, Base: 14, PerLevel: 0, Variance: 0, Coef: 0.1, Cost: 85, CastMs: 0},
	{SpellID: 709, Level: 30, MaxLevel: 35, Base: 22, PerLevel: 0, Variance: 0, Coef: 0.1, Cost: 135, CastMs: 0},
	{SpellID: 7651, Level: 38, MaxLevel: 43, Base: 28, PerLevel: 0, Variance: 0, Coef: 0.1, Cost: 185, CastMs: 0},
	{SpellID: 11699, Level: 46, MaxLevel: 51, Base: 39, PerLevel: 0, Variance: 0, Coef: 0.1, Cost: 240, CastMs: 0},
	{SpellID: 11700, Level: 54, MaxLevel: 59, Base: 51, PerLevel: 0, Variance: 0, Coef: 0.1, Cost: 300, CastMs: 0},
}

var drainSoulRanks = []spellRank{
	{},
	{SpellID: 1120, Level: 10, MaxLevel: 0, Base: 17, PerLevel: 0, Variance: 0, Coef: 0.1, Cost: 55, CastMs: 0},
	{SpellID: 8288, Level: 24, MaxLevel: 0, Base: 34, PerLevel: 0, Variance: 0, Coef: 0.1, Cost: 125, CastMs: 0},
	{SpellID: 8289, Level: 38, MaxLevel: 0, Base: 54, PerLevel: 0, Variance: 0, Coef: 0.1, Cost: 210, CastMs: 0},
	{SpellID: 11675, Level: 52, MaxLevel: 0, Base: 84, PerLevel: 0, Variance: 0, Coef: 0.1, Cost: 290, CastMs: 0},
}

var searingPainRanks = []spellRank{
	{},
	{SpellID: 5676, Level: 18, MaxLevel: 24, Base: 23, PerLevel: 0.6, Variance: 0.2105, Coef: 0.429, Cost: 45, CastMs: 1500},
	{SpellID: 17919, Level: 26, MaxLevel: 32, Base: 33, PerLevel: 0.7, Variance: 0.1846, Coef: 0.429, Cost: 68, CastMs: 1500},
	{SpellID: 17920, Level: 34, MaxLevel: 40, Base: 44, PerLevel: 0.8, Variance: 0.1895, Coef: 0.429, Cost: 91, CastMs: 1500},
	{SpellID: 17921, Level: 42, MaxLevel: 48, Base: 62, PerLevel: 0.9, Variance: 0.1791, Coef: 0.429, Cost: 118, CastMs: 1500},
	{SpellID: 17922, Level: 50, MaxLevel: 56, Base: 85, PerLevel: 1, Variance: 0.1734, Coef: 0.429, Cost: 141, CastMs: 1500},
	{SpellID: 17923, Level: 58, MaxLevel: 64, Base: 114, PerLevel: 1.2, Variance: 0.1622, Coef: 0.429, Cost: 168, CastMs: 1500},
}

var soulFireRanks = []spellRank{
	{},
	{SpellID: 6353, Level: 48, MaxLevel: 54, Base: 377, PerLevel: 1.7, Variance: 0.2276, Coef: 1, Cost: 305, CastMs: 6000},
	{SpellID: 17924, Level: 56, MaxLevel: 62, Base: 431, PerLevel: 1.9, Variance: 0.2247, Coef: 1, Cost: 335, CastMs: 6000},
}

var shadowburnRanks = []spellRank{
	{},
	{SpellID: 17877, Level: 20, MaxLevel: 24, Base: 66, PerLevel: 0.9, Variance: 0.129, Coef: 0.429, Cost: 105, CastMs: 0},
	{SpellID: 18867, Level: 24, MaxLevel: 30, Base: 80, PerLevel: 1, Variance: 0.1301, Coef: 0.429, Cost: 130, CastMs: 0},
	{SpellID: 18868, Level: 32, MaxLevel: 38, Base: 118, PerLevel: 1.3, Variance: 0.1212, Coef: 0.429, Cost: 190, CastMs: 0},
	{SpellID: 18869, Level: 40, MaxLevel: 46, Base: 148, PerLevel: 1.3, Variance: 0.1155, Coef: 0.429, Cost: 245, CastMs: 0},
	{SpellID: 18870, Level: 48, MaxLevel: 54, Base: 203, PerLevel: 1.6, Variance: 0.1132, Coef: 0.429, Cost: 305, CastMs: 0},
	{SpellID: 18871, Level: 56, MaxLevel: 62, Base: 266, PerLevel: 1.8, Variance: 0.1092, Coef: 0.429, Cost: 365, CastMs: 0},
}

var conflagrateRanks = []spellRank{
	{},
	{SpellID: 1293817, Level: 25, MaxLevel: 30, Base: 95, PerLevel: 0.9, Variance: 0.2418, Coef: 0.429, Cost: 100, CastMs: 0},
	{SpellID: 1293818, Level: 32, MaxLevel: 38, Base: 122, PerLevel: 0.9, Variance: 0.2418, Coef: 0.429, Cost: 130, CastMs: 0},
	{SpellID: 17962, Level: 40, MaxLevel: 46, Base: 146, PerLevel: 1, Variance: 0.2418, Coef: 0.429, Cost: 165, CastMs: 0},
	{SpellID: 18930, Level: 48, MaxLevel: 54, Base: 194, PerLevel: 1.1, Variance: 0.2247, Coef: 0.429, Cost: 200, CastMs: 0},
	{SpellID: 18931, Level: 54, MaxLevel: 60, Base: 239, PerLevel: 1.2, Variance: 0.2227, Coef: 0.429, Cost: 230, CastMs: 0},
	{SpellID: 18932, Level: 60, MaxLevel: 66, Base: 282, PerLevel: 1.3, Variance: 0.2191, Coef: 0.429, Cost: 255, CastMs: 0},
}

var incinerateRanks = []spellRank{
	{},
	{SpellID: 412758, Level: 40, MaxLevel: 49, Base: 97, PerLevel: 1.1, Variance: 0.15, Coef: 0.714, Cost: 205, CastMs: 2500},
	{SpellID: 1293812, Level: 50, MaxLevel: 59, Base: 145, PerLevel: 1.3, Variance: 0.15, Coef: 0.714, Cost: 265, CastMs: 2500},
	{SpellID: 1293813, Level: 60, MaxLevel: 69, Base: 217, PerLevel: 1.4, Variance: 0.15, Coef: 0.714, Cost: 325, CastMs: 2500},
}

var deathCoilRanks = []spellRank{
	{},
	{SpellID: 6789, Level: 42, MaxLevel: 48, Base: 272, PerLevel: 2.2, Variance: 0, Coef: 0.214, Cost: 435, CastMs: 0},
	{SpellID: 17925, Level: 50, MaxLevel: 56, Base: 359, PerLevel: 2.6, Variance: 0, Coef: 0.214, Cost: 525, CastMs: 0},
	{SpellID: 17926, Level: 58, MaxLevel: 64, Base: 454, PerLevel: 3, Variance: 0, Coef: 0.214, Cost: 600, CastMs: 0},
}

var lifeTapRanks = []spellRank{
	{},
	{SpellID: 1454, Level: 6, MaxLevel: 16, Base: 20, PerLevel: 1, Variance: 0, Coef: 0, Cost: 0, CastMs: 0},
	{SpellID: 1455, Level: 16, MaxLevel: 26, Base: 65, PerLevel: 1, Variance: 0, Coef: 0, Cost: 0, CastMs: 0},
	{SpellID: 1456, Level: 26, MaxLevel: 36, Base: 130, PerLevel: 1, Variance: 0, Coef: 0, Cost: 0, CastMs: 0},
	{SpellID: 11687, Level: 36, MaxLevel: 46, Base: 210, PerLevel: 1, Variance: 0, Coef: 0, Cost: 0, CastMs: 0},
	{SpellID: 11688, Level: 46, MaxLevel: 56, Base: 300, PerLevel: 1, Variance: 0, Coef: 0, Cost: 0, CastMs: 0},
	{SpellID: 11689, Level: 56, MaxLevel: 66, Base: 420, PerLevel: 1, Variance: 0, Coef: 0, Cost: 0, CastMs: 0},
}

var impFireboltRanks = []spellRank{
	{},
	{SpellID: 3110, Level: 1, MaxLevel: 5, Base: 4, PerLevel: 0.1, Variance: 0.2857, Coef: 0.164, Cost: 10, CastMs: 2000},
	{SpellID: 7799, Level: 8, MaxLevel: 13, Base: 7, PerLevel: 0.2, Variance: 0.1538, Coef: 0.314, Cost: 20, CastMs: 2000},
	{SpellID: 7800, Level: 18, MaxLevel: 23, Base: 12, PerLevel: 0.3, Variance: 0.1667, Coef: 0.529, Cost: 35, CastMs: 2000},
	{SpellID: 7801, Level: 28, MaxLevel: 33, Base: 17, PerLevel: 0.3, Variance: 0.1143, Coef: 0.571, Cost: 50, CastMs: 2000},
	{SpellID: 7802, Level: 38, MaxLevel: 43, Base: 25, PerLevel: 0.4, Variance: 0.1176, Coef: 0.571, Cost: 70, CastMs: 2000},
	{SpellID: 11762, Level: 48, MaxLevel: 53, Base: 35, PerLevel: 0.5, Variance: 0.1127, Coef: 0.571, Cost: 95, CastMs: 2000},
	{SpellID: 11763, Level: 58, MaxLevel: 63, Base: 44, PerLevel: 0.6, Variance: 0.1136, Coef: 0.571, Cost: 115, CastMs: 2000},
}

var lashOfPainRanks = []spellRank{
	{},
	{SpellID: 7814, Level: 20, MaxLevel: 0, Base: 16, PerLevel: 0, Variance: 0, Coef: 0.429, Cost: 65, CastMs: 0},
	{SpellID: 7815, Level: 28, MaxLevel: 0, Base: 22, PerLevel: 0, Variance: 0, Coef: 0.429, Cost: 80, CastMs: 0},
	{SpellID: 7816, Level: 36, MaxLevel: 0, Base: 30, PerLevel: 0, Variance: 0, Coef: 0.429, Cost: 105, CastMs: 0},
	{SpellID: 11778, Level: 44, MaxLevel: 0, Base: 36, PerLevel: 0, Variance: 0, Coef: 0.429, Cost: 125, CastMs: 0},
	{SpellID: 11779, Level: 52, MaxLevel: 0, Base: 43, PerLevel: 0, Variance: 0, Coef: 0.429, Cost: 145, CastMs: 0},
	{SpellID: 11780, Level: 60, MaxLevel: 0, Base: 50, PerLevel: 0, Variance: 0, Coef: 0.429, Cost: 160, CastMs: 0},
}
