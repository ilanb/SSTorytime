module forensicinvestigator

go 1.24.2

require (
	SSTorytime v0.0.0
	github.com/google/uuid v1.6.0
)

require github.com/lib/pq v1.10.9 // indirect

replace SSTorytime => ../../pkg/SSTorytime
