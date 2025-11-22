package engine

import (
	"log"
	"os"
	"sync"

	"scaler/pkg/config"
	"scaler/pkg/deploy"
	"scaler/pkg/node"
)

type ScalerEngine struct {
	Config       config.ScalerConfig
	ActiveAgents []config.AgentConfig
	mu           sync.Mutex
	isScaling    bool
}

func NewScalerEngine(cfg config.ScalerConfig) *ScalerEngine {
	return &ScalerEngine{
		Config:       cfg,
		ActiveAgents: []config.AgentConfig{},
	}
}

func (s *ScalerEngine) CheckAndScaleUp() {
	s.mu.Lock()
	if s.isScaling {
		s.mu.Unlock()
		return
	}
	s.isScaling = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.isScaling = false
		s.mu.Unlock()
	}()

	s.ActiveAgents = []config.AgentConfig{}
	seen := make(map[string]bool)

	for _, ag := range s.Config.AvailableAgents {
		if node.IsActive(ag) {
			s.registerAgent(ag)
			seen[ag.ServerName] = true
		}
	}

	var firstInactiveAgent *config.AgentConfig
	for i := range s.Config.AvailableAgents {
		if !seen[s.Config.AvailableAgents[i].ServerName] {
			firstInactiveAgent = &s.Config.AvailableAgents[i]
			break
		}
	}

	if firstInactiveAgent != nil {
		if len(s.ActiveAgents) > 0 {
			log.Println("High load detected, scaling up...")
		}
		if err := node.ManagePower(s.Config.ServerManagerAPI, firstInactiveAgent.ServerName, "on"); err != nil {
			log.Printf("Error starting agent %s: %v", firstInactiveAgent.ServerName, err)
		}
		if !node.IsActive(*firstInactiveAgent) {
			log.Printf("Agent %s is not active", firstInactiveAgent.ServerName)
		} else {
			s.registerAgent(*firstInactiveAgent)
		}
	}
}

func (s *ScalerEngine) registerAgent(agent config.AgentConfig) {
	if err := deploy.DeployPluggableAPI(agent); err != nil {
		log.Printf("Error deploying pluggable API to %s: %v", agent.ServerName, err)
	} else {
		s.ActiveAgents = append(s.ActiveAgents, agent)
		log.Printf("Successfully deployed pluggable API to %s", agent.ServerName)
		if err := node.UpdateUpstreamConfig(s.ActiveAgents); err != nil {
			log.Printf("Error updating upstream config: %v", err)
		} else {
			log.Printf("Successfully updated upstream config")
		}
	}
}

func (s *ScalerEngine) CheckAndScaleDown() {
	s.mu.Lock()
	if s.isScaling {
		s.mu.Unlock()
		return
	}
	s.isScaling = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.isScaling = false
		s.mu.Unlock()
	}()

	currentActiveAgents := []config.AgentConfig{}
	for _, ag := range s.Config.AvailableAgents {
		if node.IsActive(ag) {
			currentActiveAgents = append(currentActiveAgents, ag)
		}
	}

	s.ActiveAgents = currentActiveAgents

	if len(currentActiveAgents) > 1 {
		log.Println("Low load detected, scaling down...")
		agentToScaleDown := currentActiveAgents[len(currentActiveAgents)-1]

		if err := node.ManagePower(s.Config.ServerManagerAPI, agentToScaleDown.ServerName, "off"); err != nil {
			log.Printf("Error stopping agent %s: %v", agentToScaleDown.ServerName, err)
		}
		if node.IsActive(agentToScaleDown) {
			log.Printf("Agent %s is still active", agentToScaleDown.ServerName)
		} else {
			updatedActiveAgents := currentActiveAgents[:len(currentActiveAgents)-1]
			s.ActiveAgents = updatedActiveAgents

			log.Printf("Successfully stopped agent %s", agentToScaleDown.ServerName)
			if err := node.UpdateUpstreamConfig(s.ActiveAgents); err != nil {
				log.Printf("Error updating upstream config: %v", err)
			}
		}
	}
}

func (s *ScalerEngine) EvaluateScaling() {
	s.mu.Lock()
	if s.isScaling {
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()

	if len(s.ActiveAgents) < 1 {
		s.CheckAndScaleUp()
		return
	}

	var totalCPU float64
	var totalMem float64
	var activeCount int

	for _, ag := range s.ActiveAgents {
		cpu, mem, err := node.GetMetrics(ag)
		if err != nil {
			if os.Getenv("DEBUG") == "true" {
				log.Printf("Error getting metrics for %s: %v", ag.ServerName, err)
			}
			continue
		}
		totalCPU += cpu
		totalMem += mem
		activeCount++
	}

	if activeCount < 1 {
		return
	}

	avgCPU := totalCPU / float64(activeCount)
	avgMem := totalMem / float64(activeCount)
	log.Printf("Average CPU Utilization: %.2f%%, Average Memory Utilization: %.2f%%", avgCPU, avgMem)

	if avgCPU > 80 || avgMem > 80 {
		s.CheckAndScaleUp()
	} else if avgCPU < 20 && avgMem < 20 {
		s.CheckAndScaleDown()
	}
}
