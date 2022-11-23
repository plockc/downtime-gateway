package resource

type Named string

func (n Named) String() string {
	return string(n)
}

func (n Named) Id() string {
	return string(n)
}
