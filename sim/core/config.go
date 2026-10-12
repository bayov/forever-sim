package core

// Global configuration settings

// Whether to include spell ranks only accessible after the release of Ahn'Qiraj
// TODO: add phase switch instead of hardcoded const later?
const IncludeAQ = true

// IncludeAQRanks is IncludeAQ for this sim's ruleset. Forever is set before Molten Core, and
// its client (1.60.1.70338) has the Ahn'Qiraj books that teach these ranks with no source and no
// trainer, so a Forever sim uses the ranks before them.
func (env *Environment) IncludeAQRanks() bool {
	return IncludeAQ && !env.IsForever()
}
