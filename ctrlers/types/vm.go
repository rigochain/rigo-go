package types

type VMCallResult struct {
	UsedGas    uint64 `json:"usedGas,string,omitempty"`
	Err        error  `json:"vmErr,string,omitempty"`
	ReturnData []byte `json:"returnData,omitempty"`
}
