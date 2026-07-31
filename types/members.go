package types

type MemberKind uint8

const (
	MemberMethod MemberKind = iota
	MemberProperty
)

type MemberDefinition struct {
	Name            string
	Kind            MemberKind
	Parameters      []Type
	ReturnTypes     []Type
	MutatesReceiver bool
}

type MemberProvider interface {
	LookupMember(string) (MemberDefinition, bool)
}

func (s *Slice) LookupMember(name string) (MemberDefinition, bool) {
	switch name {
	case "length":
		return MemberDefinition{
			Name:        "length",
			Kind:        MemberProperty,
			ReturnTypes: []Type{Int64Type},
		}, true

	case "append":
		return MemberDefinition{
			Name:            "append",
			Kind:            MemberMethod,
			Parameters:      []Type{s.elementType},
			ReturnTypes:     nil,
			MutatesReceiver: true,
		}, true

	default:
		return MemberDefinition{}, false
	}
}

func (al *Array) LookupMember(name string) (MemberDefinition, bool) {
	switch name {
	case "length": 
		return MemberDefinition{
			Name:        "length",
			Kind:        MemberProperty,
			ReturnTypes: []Type{Int64Type},
		}, true

	case "toSlice":
		return MemberDefinition{
			Name:            "toSlice",
			Kind:            MemberMethod,
			Parameters:      nil,
			ReturnTypes:     []Type{ NewSlice(al.elementType) },
			MutatesReceiver: true,
		}, true

	default:
		return MemberDefinition{}, false
	}
}