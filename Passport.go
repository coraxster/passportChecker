package passportChecker

import (
	"github.com/pkg/errors"
)

type Passport struct {
	s, n string
}

func MakePassport(s, n string) (*Passport, error) {
	p := &Passport{s, n}
	if err := p.valid(); err != nil {
		return nil, err
	}
	return p, nil
}

var Symbols []rune
var RuneDict map[rune]uint8 //symbol codes
func init() {
	Symbols = []rune{
		' ', '-', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0', 'Б', 'А', 'Е', 'Г', 'Ж', 'М', 'B', 'T', 'O', 'Ю', 'Щ', 'О', 'Д', 'С', 'V', 'Т', 'Н', 'Л', 'X', 'Ч', 'Я', 'A', 'П', 'Ф', 'К', 'C', 'M', 'Р', 'Х', 'N', 'Ш', 'В', 'И', 'I', 'З', 'K', 'У',
	}
	RuneDict = make(map[rune]uint8)
	for i, r := range Symbols {
		RuneDict[r] = uint8(i)
	}
}

func (p *Passport) valid() error {
	if len([]rune(p.String())) > 10 {
		return errors.New(">10 symbols, too big")
	}
	for _, r := range p.String() {
		if _, ok := RuneDict[r]; !ok {
			return errors.New("unsupported symbol:" + string(r))
		}
	}
	return nil
}

func (p *Passport) String() string {
	return p.s + p.n
}

func (p *Passport) Uint64() uint64 {
	result := uint64(0)
	for i, r := range p.String() {
		result = result + (uint64(RuneDict[rune(r)]) << uint(i*6))
	}
	return result
}
