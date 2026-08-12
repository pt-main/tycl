package shared

type Config struct {
	IntV    map[string]int
	FloatV  map[string]float64
	BoolV   map[string]bool
	StringV map[string]string
	NullV   map[string]string

	IntArrV    map[string][]int
	FloatArrV  map[string][]float64
	BoolArrV   map[string][]bool
	StringArrV map[string][]string

	InnerV    map[string]*Config
	InnerArrV map[string][]*Config

	MainConf *Config
	Comments []string
	Name     string
}

func NewNilConfig() *Config {
	conf := &Config{
		IntV:    make(map[string]int),
		FloatV:  make(map[string]float64),
		BoolV:   make(map[string]bool),
		StringV: make(map[string]string),
		NullV:   make(map[string]string),

		IntArrV:    make(map[string][]int),
		FloatArrV:  make(map[string][]float64),
		BoolArrV:   make(map[string][]bool),
		StringArrV: make(map[string][]string),

		InnerV:    make(map[string]*Config),
		InnerArrV: make(map[string][]*Config),

		Comments: make([]string, 0),
		Name:     "",
	}
	conf.MainConf = conf
	return conf
}

func (c *Config) GetComments() []string {
	return c.Comments
}

func (c *Config) GetInnerV() map[string]WithDocumentation {
	res := map[string]WithDocumentation{}
	for key, val := range c.InnerV {
		res[key] = WithDocumentation(val)
	}
	return res
}

func (c *Config) GetInnerA() map[string][]WithDocumentation {
	res := map[string][]WithDocumentation{}
	for key, val := range c.InnerArrV {
		res[key] = []WithDocumentation{}
		for _, val := range val {
			res[key] = append(res[key], val)
		}
	}
	return res
}
