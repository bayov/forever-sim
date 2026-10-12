package core

// Global configuration settings

// Whether to include spell ranks only accessible after the release of Ahn'Qiraj
// TODO: add phase switch instead of hardcoded const later?
const IncludeAQ = true

// IncludeAQRanks reports whether spell ranks taught only by Ahn'Qiraj books are available. Forever
// launches before Molten Core, and its client has the AQ books with no source.
func (env *Environment) IncludeAQRanks() bool { return IncludeAQ && !env.IsForever() }
