package schedule

import (
	"fmt"
	"time"
	"strings"
	"math/rand"
	
	"github.com/golang/glog"
	
	"github.com/asobti/kube-monkey/chaos"
	"github.com/asobti/kube-monkey/config"
	"github.com/asobti/kube-monkey/calendar"
	"github.com/asobti/kube-monkey/deployments"
)

type Schedule struct {
	entries []*chaos.Chaos
}

func (s *Schedule) Entries() []*chaos.Chaos {
	return s.entries
}

func (s *Schedule) Add(entry *chaos.Chaos) {
	s.entries = append(s.entries, entry)
}

func (s *Schedule) String() string {
	schedString := []string{}
	schedString = append(schedString, fmt.Sprint("\t********** Today's schedule **********"))
	if len(s.entries) == 0 {
		schedString = append(schedString, fmt.Sprint("No terminations scheduled"))
	} else {
		schedString = append(schedString, fmt.Sprint("\tDeployment\t\t\tTermination time"))
		schedString = append(schedString, fmt.Sprint("\t----------\t\t\t----------------"))
		for _, chaos := range s.entries {
			schedString = append(schedString, fmt.Sprintf("\t%s\t\t\t%s", chaos.Deployment().Name(), chaos.KillAt().Format("01/01/2017 18:42:05 Z0700 UTC")))
		}
	}
	schedString = append(schedString, fmt.Sprint("\t********** End of schedule **********"))

	return strings.Join(schedString, "\n")
}

func (s Schedule) Print() {
	glog.V(4).Infof("Status Update: %v terminations scheduled today", len(s.entries))
	for _, chaos := range s.entries {
		glog.V(4).Infof("Deployment %s scheduled for termination at %s", chaos.Deployment().Name(), chaos.KillAt().Format("01/01/2017 18:42:05 Z0700 UTC"))
	}
}

func New() (*Schedule, error) {
	glog.V(3).Info("Status Update: Generating schedule for terminations")
	deployments, err := deployments.EligibleDeployments()
	if err != nil {
		return nil, err
	}

	schedule := &Schedule{
		entries: []*chaos.Chaos{},
	}

	for _, dep := range deployments {
		killtime := CalculateKillTime()

		if ShouldScheduleChaos(dep.Mtbf()) {
			schedule.Add(chaos.New(killtime, dep))
		}
	}

	return schedule, nil
}

func CalculateKillTime() time.Time {
	if config.DebugEnabled() && config.DebugScheduleImmediateKill() {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		// calculate a second-offset in the next minute
		secOffset := r.Intn(60)
		return time.Now().Add(time.Duration(secOffset) * time.Second)
	} else {
		return calendar.RandomTimeInRange(config.StartHour(), config.EndHour(), config.Timezone())
	}
}

func ShouldScheduleChaos(mtbf int) bool {
	if config.DebugEnabled() && config.DebugForceShouldKill() {
		return true
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var probability float64 = 1 / float64(mtbf)
	return probability > r.Float64()
}
