package shared

type ContractType int

const (
	ContractDynamic ContractType = iota
	ContractFlexible
	ContractStrict
)

type Contract struct {
	Type    ContractType
	BoolV   []string
	IntV    []string
	FloatV  []string
	StringV []string

	BoolArrV   []string
	IntArrV    []string
	FloatArrV  []string
	StringArrV []string
	InnerArrV  []string

	Inner map[string]*Contract
}

func NewNillContract() *Contract {
	return &Contract{
		Type:    0,
		BoolV:   make([]string, 0),
		IntV:    make([]string, 0),
		FloatV:  make([]string, 0),
		StringV: make([]string, 0),
		Inner:   make(map[string]*Contract),
	}
}
