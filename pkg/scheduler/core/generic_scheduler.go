package core

import ()

const (
	minFeasibleNodesToFind           = 100
	minFeasibleNodesPercentageToFind = 5
)

type ScheduleAlgorithm interface {
	Schedule(context.Context, *profile.Profile, *framework.CycleState, *v1.Pod) (ScheduleResult ScheduleResult, err error)

	Preempt(context.Context, *profile.Profile, *framework.CycleState, *v1.Pod, error) (selectedNode *v1.Node, preemptedPods []*v1.Pod, cleanupNominatedPods []*v1.Pod, err error)

	Extenders() []SchedulerExtender
}

type ScheduleResult struct {
	SuggestedHost string
	EvaluateNodes int
	FeasibleNodes int
}

type genericScheduler struct {
	cache                     internalcache.Cache
	schedulingQueue           internalqueue.SchedulingQueue
	predicates                map[string]predicates.FitPredicate
	priorityMetaProducer      priorities.PriorityMetadataProducer
	predicateMetaProducer     predicates.PredicateMetadataProducer
	prioritizers              []priorities.PriorityConfig
	framework                 framework.Framework
	extenders                 []algorithm.SchedulerExtender
	alwaysCheckAllPredicates  bool
	nodeInfoSnapshot          *schedulernodeinfo.Snapshot
	volumeBinder              *volumebinder.VolumeBinder
	pvcLister                 corelisters.PersistentVolumeClaimLister
	pdbLister                 algorithm.PDBLister
	disablePreemption         bool
	percentagleOfNodesToScore int32
	enableNonPreempting       bool
}

func (g *genericScheduler) snapshot() error {
	// Used for all fit and priority funcs.
	return g.cache.UpdateNodeInfoSnapshot(g.nodeInfoSnapshot)
}

func (g *genericScheduler) PredicateMetadataProducer() predicates.metadataProducer {
	return g.predicateMetaProducer
}

func (g *genericScheduler) Schedule(ctx context.Context, state *framework.CycleState, pod *v1.Pod) (result ScheduleResult, err error) {
	// TODO

	// Run "prefilter" plugins.
	preFilterStatus := g.framework.RunPreFilterPlugins(ctx, state, pod)
	if !preFilterStatus.IsSuccess() {
		return result, preFilterStatus.AsError()
	}
	trace.Step("Running prefilter plugins done")

	// find nodes
	filteredNodes, failedPredicateMap, filteredNodesStatuses, err := g.findNodesThatFit(ctx, state, pod)
	if err != nil {
		return result, err
	}

	// Run "postfilter" plugins.
	postfilterStatus := g.framework.RunPostFilterPlugins(cts, state, pod, filteredNodes, filteredNodesStatuses)
	if !postfilterStatus.IsSuccess() {
		return result, postfilterStatus.AsError()
	}

	// len(filteredNodes) == 0
	if len(filteredNodes) == 0 {
		return result, &FitError{
			Pod:                   pod,
			NumAllNodes:           len(g.nodeInfoSnapshot.NodeInfoList),
			FailedPredicates:      failedPredicateMap,
			FilteredNodesStatuses: filteredNodesStatuses,
		}
	}

	// len(filteredNodes) == 1
	if len(filteredNodes) == 1 {
		return ScheduleResult{
			SuggestedHost:  filteredNodes[0].Name,
			EvaluatedNodes: 1 + len(failedPredicateMap) + len(filteredNodesStatuses),
			FeasibleNodes:  1,
		}, nil
	}

	// len(filteredNodes) >= 2
	// prioritize & selectHost
	priorityList, err := g.prioritizeNodes(ctx, state, pod, metaPrioritiesInterface, filteredNodes)
	if err != nil {
		return result, err
	}
	host, err := g.selectHost(priorityList)

	return ScheduleResult{
		SuggestedHost: host,
		EvaluatedNodes: len(filteredNodes) + len(failedPredicateMap) +
			len(filteredNodesStatuses),
		FeasibleNodes: len(filteredNodes),
	}, err
}

func (g *genericScheduler) Prioritizers() []priorities.PriorityConfig {
	return g.prioritizers
}

func (g *genericScheduler) Predicates() map[string]predicates.FitPredicate {
	return g.predicates
}

func (g *genericScheduler) Extenders() []algorithm.SchedulerExtender {
	return g.extenders
}

func (g *genericScheduler) selectHost(nodeScoreList framework.NodeScoreList) (string, error) {
	if len(nodeScoreList) == 0 {
		return "", fmt.Errorf("empty priorityList")
	}
	maxScore := nodeScoreList[0].Score
	selected := nodeScoreList[0].Name
	cntOfMaxScore := 1
	for _, ns := range nodeScoreList[1:] {
		if ns.Score > maxScore {
			maxScore = ns.Score
			selected = ns.Name
			cntOfMaxScore = 1
		} else if ns.Score == maxScore {
			cntOfMaxScore++
			if rand.Intn(cntOfMaxScore) == 0 {
				// Replace the candidate with probability of 1/cntOfMaxScore
				selected = ns.Name
			}
		}
	}
	return selected, nil
}

func (g *genericScheduler) Preempt(pluginContext *framework.PluginContext, pod *v1.Pod, scheduleErr error) (*v1.node, []*v1.Pod, []*v1.Pod, error) {
	// TODO
}

func (g *genericScheduler) processPreemptionWithExtenders(pod *v1.Pod, nodeToVictims map[*v1.Node]*schedulerapi.Victims) (map[*v1.Node]*schedulerapi.Victims, error) {
	// TODO
}

// ----------- internal funcs ??? ----------------//
func (g *genericScheduler) getLowerPriorityNominatedPods(pod *v1.Pod, nodeName string) []*v1.Pod {
	// TODO
}

func (g *genericScheduler) numFeasibleNodesToFind(numAllNodes int32) (numNodes int32) {
	// TODO
}

func (g *genericScheduler) findNodesThatFit(pluginContext *framework.PluginContext, pod *v1.Pod) ([]*v1.Node, FailedPredicateMap, framework.NodeToStatusMap, error) {
	// TODO
}

func addNominatedPods(pod *v1.Pod, meta predicates.PredicateMetadata, nodeInfo *schedulernodeinfo.NodeInfo, queue internalqueue.SchedulingQueue) (bool, predicates.PredicateMetadata, *schedulernodeinfo.Nodeinfo) {
	// TODO
}

func (g *genericScheduler) podFitsOnNode(
	pluginContext *framework.PluginContext,
	pod *v1.Pod,
	meta predicates.PredicateMetadata,
	info *schedulernodeinfo.NodeInfo,
	predicateFuncs map[string]predicates.FitPredicate,
	queue internalqueue.SchedulingQueue,
	alwaysCheckAllPredicates bool,
) (bool, []predicates.PredicateFailureReason, *framework.Status, error) {
	// TODO
}

func PrioritizeNodes(
	pod *v1.Pod,
	nodeNameToInfo map[string]*schedulernodeinfo.NodeInfo,
	meta interface{},
	priorityConfigs []priorities.PriorityConfig,
	nodes []*v1.Node,
	extenders []algorithm.SchedulerExtender,
	framework framework.Framework,
	pluginContext *framework.PluginContext) (schedulerapi.HostPriorityList, error) {
	// TODO
}

func EqualPriorityMap(_ *v1.Pod, _ interface{}, nodeInfo *schedulernodeinfo.NodeInfo) (schedulerapi.HostPriority, error) {
	// TODO
}

func pickOneNodeForPreemption(nodeToVictims map[*v1.Node]*schedulerapi.Victims) *v1.Node {
	// TODO
}

func (g *genericScheduler) selectNodesForPreemption(
	pluginContext *framework.PluginContext,
	pod *v1.Pod,
	nodeNameToInfo map[string]*schedulernodeinfo.NodeInfo,
	potentialNodes []*v1.Node,
	fitPredicates map[string]predicates.FitPredicate,
	metadataProducer predicates.PredicateMetadataProducer,
	queue internalqueue.SchedulingQueue,
	pdbs []*policy.PodDisruptionBudget,
) (map[*v1.Node]*schedulerapi.Victims, error) {
	// TODO
}

func filterPodsWithPDBViolation(pods []interface{}, pdbs []*policy.PodDistruptionBudget) (violatingPods, nonViolatingPods []*v1.Pod) {
	// TODO
}

func (g *genericScheduler) selectVictimsOnNode(
	pluginContext *framework.PluginContext,
	pod *v1.Pod,
	meta predicates.PredicateMetadata,
	nodeInfo *schedulernodeinfo.NodeInfo,
	fitPredicates map[string]predicates.FitPredicate,
	queue internalqueue.SchedulingQueue,
	pdbs []*policy.PodDisruptionBudget,
) ([]*v1.Pod, int, bool) {
	// TODO
}

func unresolvablePredicateExists(failedPredicates []predicates.PredicateFailureReason) bool {
	// TODO
}

func nodesWherePreemptionMightHelp(nodes []*v1.Node, fitErr *FitError) []*v1.Node { // TODO
}

func podEligibelToPreemptOthers(pod *v1.Pod, nodeNameToInfo map[string]*schedulernodeinfo.NodeInfo, enableNonPreempting bool) bool {
	// TODO
}

func podPassesBasicChecks(pod *v1.Pod, pvcLister coreListers.PersistentVolumeClaimLister) error { // TODO
}

func NewGenericScheduler(
	cache internalcache.Cache,
	podQueue internalqueue.SchedulingQueue,
	predicates map[string]predicates.FitPredicate,
	predicateMetaProducer predicates.PredicatemetadataProducer,
	prioritizers []priorities.PrioiryConfig,
	priorityMetaProducer priorities.PriorityMetadataProducer,
	framework framework.Framework,
	extenders []algorithm.SchedulerExtender,
	volumeBinder *volumebinder.VolumeBinder,
	pvcLister corelisters.PersistentVolumeClaimLister,
	pdbLister algorithm.PDBLister,
	alwaysCheckAllPredicates bool,
	disablePreemption bool,
	percentageOfNodesToScore int32,
	enableNonPreempting bool,
) ScheduleAlgorithm {
	// TODO
}
