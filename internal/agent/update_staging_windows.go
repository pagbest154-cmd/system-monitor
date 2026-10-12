//go:build windows

package agent

func cleanupUpdateStagingOnStart() {
	if err := CleanupUpdateStaging(); err != nil {
		AgentLogf("update: staging cleanup on start: %v", err)
	}
}
