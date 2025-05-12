package bookstack

import (
	"fmt"
	"log"
	"time"

	"github.com/Flexible-Universe/bookstack-backup/internal/bookstack"
	"github.com/robfig/cron/v3"
)

// Scheduler configures and starts cron jobs for all instances in the config.
func Scheduler(cfg bookstack.Config) (*cron.Cron, error) {
	sched := cron.New(cron.WithLocation(time.Local))
	for _, inst := range cfg.Instances {
		client := bookstack.NewClient(inst)
		inst := inst
		if _, err := sched.AddFunc(inst.Schedule, func() {
			if err := client.Crawl(); err != nil {
				log.Printf("[%s] Crawl error: %v", inst.Name, err)
			}
		}); err != nil {
			return nil, fmt.Errorf("scheduling %s: %w", inst.Name, err)
		}
	}
	sched.Start()
	return sched, nil
}
