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

	Comments []string
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
		Comments:  make([]string, 0),
	}
}

func (c *Contract) GetComments() []string {
	return c.Comments
}

func (c *Contract) GetInnerV() map[string]WithDocumentation {
	res := map[string]WithDocumentation{}
	for key, val := range c.Inner {
		res[key] = WithDocumentation(val)
	}
	return res
}

func (c *Contract) GetInnerA() map[string][]WithDocumentation {
	res := map[string][]WithDocumentation{}
	for key, val := range c.InnerArrV {
		res[key] = []WithDocumentation{WithDocumentation(val)}
	}
	return res
}
