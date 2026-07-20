package report

type ReliabilityRollupConfig struct {
	TimeFrame ReliabilityRollupTimeFrame `json:"timeFrame"`
	// ReliabilityScoreType selects the reliability scoring mechanism used by the
	// Reliability Rollup Report. The zero value defaults to
	// [ReliabilityScoreTypeSLOTimeWindow].
	ReliabilityScoreType ReliabilityScoreType `json:"reliabilityScoreType,omitempty"`
	CustomHierarchy      []HierarchyFolder    `json:"customHierarchy,omitempty"`
}

type ReliabilityRollupTimeFrame struct {
	Rolling  *RollingTimeFrame  `json:"rolling,omitempty"`
	Calendar *CalendarTimeFrame `json:"calendar,omitempty"`
	TimeZone string             `json:"timeZone"`
}

type HierarchyFolder struct {
	DisplayName string            `json:"displayName"`
	Children    []HierarchyFolder `json:"children,omitempty"`
	SLOs        []HierarchySLORef `json:"slos,omitempty"`
}

type HierarchySLORef struct {
	Name        string `json:"name"`
	Project     string `json:"project"`
	DisplayName string `json:"displayName,omitempty"`
}
