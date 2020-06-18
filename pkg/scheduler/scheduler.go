package scheduler

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"time"

	"k8s.io/kubernetes/pkg/scheduler/core"
	framework "k8s.io/kuernetes/pkg/scheduler/framework/v1alpha1"
	internalcache "k8s.io/kuernetes/pkg/scheduler/internal/cache"
	internalqueue "k8s.io/kuernetes/pkg/scheduler/internal/queue"
)


// Scheduler watches for new unscheduled pods. It attempts to fine
// nodes that they fit on and writes bindings back to the api server.
type Scheduler struct {
	// It is expected that changes made via SchedulerCache will be oberved
	// by NodeLister and Algorithm.
	SchedulerCache internalcache.Cache

	Algorithm core.ScheduleAlgorithm
	// PodConditionUpdater is used only in case of scheduling errors. If we succeed
	// with scheduling, PodScheduled condition will be updated in apiserver in /bind
	// handler so that binding and setting PodCondition it is atomic.
	podConditionUpdater podContitionUpdater
	// PodPreemptor is used to evict pods and update 'NominateNode' field of 
	// the preemptor pod.
	podPreemptor podPreemptor

	// NextPod should be a function that blocks until the next pod
	// is available. We don't use a channel for this, because scheduling
	// a pod may take some amount of time and we don't want pods to get
	// stale while they sit in a channel.
	NextPod func() *framework.PodInfo

	// Error is called if there is an error. It is passed the pod in
	// question, and the error
	Error func(*framework.PodInfo, error)

	// Close this to shut down the scheduler.
	StopEverything <-chan struct{}

	// VolumeBinder handles PVC/PV binding for the pod.
	VolumeBinder *volumebinder.VolumeBinder

	// Disable pod preemption or not.
	DisablePreemption bool

	// SchedulingQueue holds pods to be scheduled
	SchedulingQueue internalqueue.SchedulingQueue

	// Profiles are the scheduling profiles.
	Profiles profile.Map

	scheduledPodsHasSynced func() bool
}
