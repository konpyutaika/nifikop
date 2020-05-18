package templates

const (
	NodeConfigTemplate			= "%s-config"
	NodeStorageTemplate			= "%s-%d-storage"
	PrefixNodeNameTemplate    	= "%s-"
	SuffixNodeNameTemplate    	= "-node"
	RootNodeNameTemplate		= "%d"
	NodeNameTemplate 			= PrefixNodeNameTemplate+RootNodeNameTemplate+SuffixNodeNameTemplate
)
