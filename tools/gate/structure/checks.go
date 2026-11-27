package structure

import "scout/tools/gate/check"

func Checks() []check.Check {
	return []check.Check{
		{Name: "structure", Rules: []string{ruleDirection, rulePlacement, ruleReadership, ruleSingleOwner, ruleSibling}, Run: runDirection},
		{Name: "prefix", Rules: []string{rulePrefix}, Run: runPrefix},
		{Name: "rename", Rules: []string{ruleRename}, Run: runRename},
	}
}
