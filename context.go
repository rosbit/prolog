package pl

import (
	"github.com/mndrix/golog"
	"os"
)

type Prolog struct {
	m golog.Machine
}

func NewProlog() *Prolog {
	return &Prolog{m: golog.NewMachine()}
}

func (p *Prolog) LoadScript(script string) {
	p.m = p.m.Consult(script)
}

func (p *Prolog) LoadFile(file string) (err error) {
	fp, e := os.Open(file)
	if e != nil {
		err = e
		return
	}
	defer fp.Close()

	p.m = p.m.Consult(fp)
	return
}
