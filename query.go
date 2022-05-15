package pl

import (
	"github.com/mndrix/golog/term"
	"github.com/mndrix/golog/read"
	"github.com/mndrix/golog"
	"fmt"
	"strings"
)

func (p *Prolog) Query(goal string, args ...interface{}) (it <-chan map[string]interface{}, ok bool, err error) {
	if len(goal) == 0 {
		err = fmt.Errorf("goal expected")
		return
	}

	defer func() {
		if r := recover(); r != nil {
			if v, o := r.(error); o {
				err = v
				return
			}
			err = fmt.Errorf("%v", r)
			return
		}
	}()
	argc, argv, vars, e := makeGoalArgs(args...)
	if e != nil {
		err = e
		return
	}

	return p.doQuery(goal, argc, argv, vars)
}

func makeGoalArgs(args ...interface{}) (argc int, argv []term.Term, vars map[int]string, err error) {
	argc = len(args)
	if argc == 0 {
		return
	}

	argv = make([]term.Term, argc)
	vars = make(map[int]string)
	for i, arg := range args {
		if plVar, ok := arg.(PlVar); ok {
			vName := string(plVar)
			if len(vName) == 0 {
				vName = fmt.Sprintf("_Var%d", i)
			}
			vars[i] = vName
			argv[i] = PlVar(vName).ToTerm()
		} else {
			argv[i], err = makePlTerm(arg)
			if err != nil {
				return
			}
		}
	}
	return
}

func (p *Prolog) doQuery(goal string, argc int, argv []term.Term, vars map[int]string) (it <-chan map[string]interface{}, ok bool, err error) {
	var qGoal term.Callable

	varCount := len(vars)
	if argc == 0 || varCount == 0 {
		qGoal = term.NewCallable(goal, argv...)
		ok = p.m.CanProve(qGoal)
		return
	}

	if varCount < argc {
		qGoal = term.NewCallable(goal, argv...)
	} else {
		as := make([]string, argc)
		for i, arg := range argv {
			as[i] = arg.String()
		}
		q := fmt.Sprintf("%s(%s).", goal, strings.Join(as, ", "))
		// fmt.Printf("q: %s\n", q)

		qTerm, e := read.Term(q)
		if e != nil {
			err = e
			return
		}
		qGoal = qTerm.(term.Callable)
	}

	qVars := term.Variables(qGoal)
	m  := p.m.PushConj(qGoal)

	var answer term.Bindings
	var e error
	m, answer, e = m.Step()
	if e == golog.MachineDone {
		// false
		return
	}

	ok = true
	// var binding
	sols := make(chan map[string]interface{})
	go func() {
		for {
			if answer != nil {
				answer = answer.WithNames(qVars)
				vals := make(map[string]interface{})
				for _, vName := range vars {
					t, err := answer.ByName(vName)
					if err != nil {
						vals[vName] = nil
					} else {
						vals[vName], _ = fromPlTerm(t)
					}
				}
				sols <- vals
			}

			m, answer, e = m.Step()
			if e == golog.MachineDone {
				break
			}
		}

		close(sols)
	}()

	it = sols
	return
}
