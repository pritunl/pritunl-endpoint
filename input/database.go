package input

var inputs = []*Input{}

func Register(in *Input) {
	inputs = append(inputs, in)
}
