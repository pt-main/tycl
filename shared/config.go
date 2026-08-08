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
	Name      string
}

func NewNilConfig() *Config {
	return &Config{
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
		Name:      "",
	}
}
