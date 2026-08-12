package shared

type WithDocumentation interface {
	GetComments() []string
	GetInnerV() map[string]WithDocumentation
	GetInnerA() map[string][]WithDocumentation
}
