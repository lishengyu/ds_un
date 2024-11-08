package fileproc

const (
	DpiLogtarName = "ds_data_bak/logtar"
	IdentifyName  = "ds_data_identify"
	MonitorName   = "ds_data_monitor"
	EvidenceName  = "ds_evidence_file"
	KeywordName   = "ds_keyword_file"
	KeywordNameB  = "ds_data_keyword"
	RulesName     = "ds_data_identify_rules"
	AuditName     = "ds_audit_log"
)

const (
	IdentifyIndex = iota
	MonitorIndex
	EvidenceIndex
	RulesIndex
	KeywordIndex
	AuditIndex
	LogIndexMax
)
