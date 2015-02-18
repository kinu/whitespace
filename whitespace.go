// Whitespace interpreter
// http://compsoc.dur.ac.uk/whitespace/tutorial.php
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
)

const (
	IMP_NONE = iota
	IMP_STACK
	IMP_ARITH
	IMP_HEAP
	IMP_FLOW
	IMP_IO
)

const (
	// Stack commands.
	CMD_PUSH = iota
	CMD_DUP
	CMD_COPY
	CMD_SWAP
	CMD_DISCARD
	CMD_SLIDE

	// Arithmetic commands.
	CMD_ADD
	CMD_SUB
	CMD_MUL
	CMD_DIV
	CMD_MOD

	// Heap access.
	CMD_STORE
	CMD_RETRIEVE

	// Flow control commands.
	CMD_MARK
	CMD_CALL
	CMD_JMP
	CMD_JMP_IF0
	CMD_JMP_NEG
	CMD_RET
	CMD_FINISH

	// I/O commands.
	CMD_PUTCHAR
	CMD_PUTNUM
	CMD_READCHAR
	CMD_READNUM
)

type Command struct {
	imp    int
	cmd    int
	val    int
	cmdstr string
}

type Program struct {
	commands []Command
	labels   map[int]int
}

type Parser struct {
	program  Program
	r        *bytes.Reader
	err      error
	verbose  bool
	finished bool
}

func NewParser(r *bytes.Reader, verbose bool) (p Parser) {
	p = Parser{r: r, verbose: verbose, finished: false}
	p.program.commands = make([]Command, 0, 100)
	p.program.labels = make(map[int]int)
	return
}

func (p *Parser) WriteCommand(imp int, cmd int, val int, str string, a ...interface{}) {
	s := fmt.Sprintf(str, a...)
	p.program.commands = append(p.program.commands, Command{imp, cmd, val, s})
	if cmd == CMD_MARK {
		p.program.labels[val] = len(p.program.commands) - 1
	}
	if p.verbose {
		fmt.Println(s)
	}
}

func (p *Parser) ReadSpace() (c byte) {
	if p.err != nil {
		return
	}
	for {
		c, p.err = p.r.ReadByte()
		if p.err == io.EOF || c == '\t' || c == ' ' || c == '\n' {
			return
		}
	}
}

func (p *Parser) ParseNumber() (n int) {
	n = 0
	c := p.ReadSpace()
	sign := 1
	if c == '\t' {
		sign = -1
	}
	for {
		c = p.ReadSpace()
		switch c {
		case ' ':
			n = n<<1 + 0
		case '\t':
			n = n<<1 + 1
		case '\n':
			return n * sign
		}
	}
}

func (p *Parser) ParseImp() (imp int) {
	c := p.ReadSpace()
	switch c {
	case ' ':
		return IMP_STACK
	case '\n':
		return IMP_FLOW
	case '\t':
		c = p.ReadSpace()
		switch c {
		case ' ':
			return IMP_ARITH
		case '\t':
			return IMP_HEAP
		case '\n':
			return IMP_IO
		}
	}
	return IMP_NONE
}

func (p *Parser) ParseStackCommand() {
	c := p.ReadSpace()
	switch c {
	case ' ':
		n := p.ParseNumber()
		p.WriteCommand(IMP_STACK, CMD_PUSH, n, "PUSH %d", n)
	case '\n':
		c = p.ReadSpace()
		switch c {
		case ' ':
			p.WriteCommand(IMP_STACK, CMD_DUP, -1, "DUP")
		case '\t':
			p.WriteCommand(IMP_STACK, CMD_SWAP, -1, "SWAP")
		case '\n':
			p.WriteCommand(IMP_STACK, CMD_DISCARD, -1, "DISCARD")
		}
	case '\t':
		c = p.ReadSpace()
		n := p.ParseNumber()
		switch c {
		case ' ':
			p.WriteCommand(IMP_STACK, CMD_COPY, n, "COPY %d", n)
		case '\n':
			p.WriteCommand(IMP_STACK, CMD_SLIDE, n, "SLIDE %d", n)
		}
	}
}

func (p *Parser) ParseArithCommand() {
	c := p.ReadSpace()
	switch c {
	case ' ':
		c = p.ReadSpace()
		switch c {
		case ' ':
			p.WriteCommand(IMP_ARITH, CMD_ADD, -1, "ADD")
		case '\t':
			p.WriteCommand(IMP_ARITH, CMD_SUB, -1, "SUB")
		case '\n':
			p.WriteCommand(IMP_ARITH, CMD_MUL, -1, "MUL")
		}
	case '\t':
		c = p.ReadSpace()
		switch c {
		case ' ':
			p.WriteCommand(IMP_ARITH, CMD_DIV, -1, "DIV")
		case '\t':
			p.WriteCommand(IMP_ARITH, CMD_MOD, -1, "MOD")
		}
	}
}

func (p *Parser) ParseHeapCommand() {
	c := p.ReadSpace()
	switch c {
	case ' ':
		p.WriteCommand(IMP_HEAP, CMD_STORE, -1, "STORE")
	case '\t':
		p.WriteCommand(IMP_HEAP, CMD_RETRIEVE, -1, "RETRIEVE")
	}
}

func (p *Parser) ParseFlowCommand() {
	c := p.ReadSpace()
	switch c {
	case ' ':
		c = p.ReadSpace()
		label := p.ParseNumber()
		switch c {
		case ' ':
			p.WriteCommand(IMP_FLOW, CMD_MARK, label, "MARK %d", label)
		case '\t':
			p.WriteCommand(IMP_FLOW, CMD_CALL, label, "CALL %d", label)
		case '\n':
			p.WriteCommand(IMP_FLOW, CMD_JMP, label, "JMP %d", label)
		}
	case '\t':
		c = p.ReadSpace()
		switch c {
		case ' ':
			label := p.ParseNumber()
			p.WriteCommand(IMP_FLOW, CMD_JMP_IF0, label, "JMP_IF0 %d", label)
		case '\t':
			label := p.ParseNumber()
			p.WriteCommand(IMP_FLOW, CMD_JMP_NEG, label, "JMP_NEG %d", label)
		case '\n':
			p.WriteCommand(IMP_FLOW, CMD_RET, -1, "RET")
		}
	case '\n':
		c = p.ReadSpace()
		if c == '\n' {
			p.WriteCommand(IMP_FLOW, CMD_FINISH, -1, "FINISH")
		}
	}
}

func (p *Parser) ParseIOCommand() {
	c := p.ReadSpace()
	switch c {
	case ' ':
		c = p.ReadSpace()
		switch c {
		case ' ':
			p.WriteCommand(IMP_IO, CMD_PUTCHAR, -1, "PUTCHAR")
		case '\t':
			p.WriteCommand(IMP_IO, CMD_PUTNUM, -1, "PUTNUM")
		}
	case '\t':
		c = p.ReadSpace()
		switch c {
		case ' ':
			p.WriteCommand(IMP_IO, CMD_READCHAR, -1, "READCHAR")
		case '\t':
			p.WriteCommand(IMP_IO, CMD_READNUM, -1, "READNUM")
		}
	}
}

func (p *Parser) Parse() {
	for {
		imp := p.ParseImp()
		if p.err == io.EOF {
			return
		}
		if p.err != nil {
			panic(p.err)
		}
		switch imp {
		case IMP_NONE:
			panic("Parse error")
		case IMP_STACK:
			p.ParseStackCommand()
		case IMP_ARITH:
			p.ParseArithCommand()
		case IMP_HEAP:
			p.ParseHeapCommand()
		case IMP_FLOW:
			p.ParseFlowCommand()
		case IMP_IO:
			p.ParseIOCommand()
		}
	}
}

//--------------------------------------------------------------

type Stack []int

func NewStack(capacity int) Stack {
	return make([]int, 0, capacity)
}

func (s Stack) String() string {
	return fmt.Sprintf("%v", []int(s))
}

func (s Stack) Get(idx int) int {
	return s[len(s)-(idx+1)]
}

func (s Stack) Put(idx int, value int) {
	s[len(s)-(idx+1)] = value
}

func (s *Stack) Pop() (n int) {
	n = (*s).Get(0)
	*s = (*s)[:len(*s)-1]
	return
}

func (s *Stack) Push(n int) {
	*s = append(*s, n)
}

func (s Stack) Len() int {
	return len(s)
}

//--------------------------------------------------------------

type Heap []int

func NewHeap() Heap {
	return make([]int, 128)
}

func (h Heap) Get(idx int) int {
	if idx > len(h)-1 {
		panic("Index out of range")
	}
	return h[idx]
}

func (h *Heap) Put(idx int, value int) {
	if idx > cap(*h)-1 {
		h2 := make([]int, (value+1)*2)
		copy(h2, *h)
		*h = h2
	}
	(*h)[idx] = value
}

func (h Heap) Len() int {
	return len(h)
}

//--------------------------------------------------------------

type Machine struct {
	stack   Stack
	frame   Stack
	heap    Heap
	err     error
	verbose bool
}

func NewMachine(verbose bool) (m Machine) {
	m = Machine{verbose: verbose}
	m.stack = NewStack(20)
	m.frame = NewStack(20)
	m.heap = NewHeap()
	return
}

func (m Machine) DebugOutput(s string, args ...interface{}) {
	if m.verbose {
		fmt.Printf(s, args...)
		fmt.Println(" [stack]", m.stack)
	}
}

func (m Machine) Run(program Program) {
	pc := 0
	for pc < len(program.commands) {
		cmd := program.commands[pc]
		m.DebugOutput(cmd.cmdstr)
		pc += 1
		switch cmd.imp {
		case IMP_ARITH:
			n2 := m.stack.Pop()
			n1 := m.stack.Pop()
			switch cmd.cmd {
			case CMD_ADD:
				m.stack.Push(n1 + n2)
			case CMD_SUB:
				m.stack.Push(n1 - n2)
			case CMD_MUL:
				m.stack.Push(n1 * n2)
			case CMD_DIV:
				m.stack.Push(n1 / n2)
			case CMD_MOD:
				m.stack.Push(n1 % n2)
			}
		default:
			switch cmd.cmd {
			case CMD_PUSH:
				m.stack.Push(cmd.val)
			case CMD_DUP:
				m.stack.Push(m.stack.Get(0))
			case CMD_SWAP:
				s := m.stack
				s[len(s)-1], s[len(s)-2] = s[len(s)-2], s[len(s)-1]
			case CMD_DISCARD:
				m.stack.Pop()
			case CMD_COPY:
				if 1+cmd.val > m.stack.Len() {
					panic("Index out of range")
				}
				m.stack.Push(cmd.val)
			case CMD_SLIDE:
				if 1+cmd.val > m.stack.Len() {
					panic("Index out of range")
				}
				idx := m.stack.Len() - (1 + cmd.val)
				m.stack = append(m.stack[:idx], m.stack[idx+1:]...)
			case CMD_STORE:
				value := m.stack.Pop()
				address := m.stack.Pop()
				m.DebugOutput("value:%d address:%d", value, address)
				m.heap.Put(address, value)
			case CMD_RETRIEVE:
				address := m.stack.Pop()
				m.stack.Push(m.heap.Get(address))
			case CMD_MARK:
				break
			case CMD_CALL:
				m.frame.Push(pc)
				pc = program.labels[cmd.val]
			case CMD_JMP:
				pc = program.labels[cmd.val]
			case CMD_JMP_IF0:
				if m.stack.Pop() == 0 {
					pc = program.labels[cmd.val]
				}
			case CMD_JMP_NEG:
				if m.stack.Pop() < 0 {
					pc = program.labels[cmd.val]
				}
			case CMD_RET:
				if len(m.frame) == 0 {
					panic("Cannot return")
				}
				pc = m.frame.Pop()
			case CMD_FINISH:
				return
			case CMD_PUTCHAR:
				fmt.Printf("%c", m.stack.Pop())
			case CMD_PUTNUM:
				fmt.Printf("%d", m.stack.Pop())
			case CMD_READCHAR:
				var c int
				fmt.Scanf("%c", &c)
				address := m.stack.Pop()
				m.heap.Put(address, int(c))
			case CMD_READNUM:
				var n int
				fmt.Scanf("%d", &n)
				address := m.stack.Pop()
				m.heap.Put(address, n)
			}
		}
	}
}

//--------------------------------------------------------------

func main() {
	verbose := flag.Bool("v", false, "verbose")
	dryRun := flag.Bool("dry_run", false, "dry run")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		panic("Usage: whitespace.go [-v][-dry_run] inputfile")
	}
	data, err := ioutil.ReadFile(args[0])
	if err != nil {
		panic(err)
	}

	if *verbose {
		fmt.Printf("\n* Parsing the program:\n\n")
	}

	r := bytes.NewReader(data)
	parser := NewParser(r, *verbose)
	parser.Parse()

	if *verbose {
		fmt.Printf("\n\n* Running the program:\n\n")
	}

	if !*dryRun {
		m := NewMachine(*verbose)
		m.Run(parser.program)
	}
}
