package testsource

import "scout/tools/gate/check"

func Checks() []check.Check {
	return []check.Check{
		{Name: "test.placement", Rules: []string{"TEST.placement"}, Run: runPlacement},
		{Name: "test.growth", Rules: []string{"TEST.growth.justified"}, Run: runGrowth},
		{Name: "test.duplicates", Rules: []string{"TEST.duplicate-path"}, Run: runDuplicates},
		{Name: "test.subprocess-timeouts", Rules: []string{"TEST.subprocess-timeouts"}, Run: runSubprocessTimeouts},
		{Name: "test.self-mocking", Rules: []string{"TEST.no-self-mocking"}, Run: runSelfMocking},
		{Name: "test.hooks", Rules: []string{"TEST.no-test-hooks"}, Run: runTestHooks},
		{Name: evasionCheck, Rules: []string{"TEST.no-evasion"}, Run: runEvasion},
	}
}
