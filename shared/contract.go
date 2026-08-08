package shared

type ContractType int

func (c ContractType) String() string {
	switch c {
	case ContractStrict:
		return "strict"
	case ContractFlexible:
		return "flexible"
	}
	return "dynamic"
}

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

	InnerArrV map[string]*Contract
	Inner     map[string]*Contract
}

func NewNillContract() *Contract {
	return &Contract{
		Type:      0,
		BoolV:     make([]string, 0),
		IntV:      make([]string, 0),
		FloatV:    make([]string, 0),
		StringV:   make([]string, 0),
		Inner:     make(map[string]*Contract),
		InnerArrV: make(map[string]*Contract),
	}
}
