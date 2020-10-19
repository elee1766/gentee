// Copyright 2019 Alexey Krivonogov. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package vm

import (
	"fmt"
	"os"
	"sync"

	"github.com/gentee/gentee/core"
)

//go:generate go run generate/generate.go

const (
	STACKSIZE = 256
	// CYCLE is the limit of loops
	CYCLE = uint64(16000000)
	// DEPTH is the maximum size of blocks stack
	DEPTH = uint32(1000)
)

const (
	sysClose = iota
	SysSuspend
	SysResume
	SysTerminate
)

type Settings struct {
	CmdLine        []string
	Stdin          *os.File
	Stdout         *os.File
	Stderr         *os.File
	Input          []byte   // stdin
	Cycle          uint64   // limit of loops
	Depth          uint32   // limit of blocks stack
	SysChan        chan int // system chan
	IsPlayground   bool
	Playground     Playground
	ProgressHandle ProgressFunc
}

type Const struct {
	Type  uint16
	Value interface{}
}

// VM is the main structure of the virtual machine
type VM struct {
	Settings    Settings
	Exec        *core.Exec
	Consts      map[int32]Const
	Runtimes    []*Runtime
	CtxMutex    sync.RWMutex
	ThreadMutex sync.RWMutex
	LockMutex   sync.Mutex
	WaitGroup   sync.WaitGroup
	Context     map[string]string
	Count       int64 // count of active threads
	WaitCount   int64
	Stopped     bool
	ChCount     chan int64
	ChError     chan error
	ChWait      chan int64
	Playground  PlaygroundFS
}

type OptValue struct {
	Var   int32       // id of variable
	Type  int         // type of variable
	Value interface{} // value
}

// Runtime is the one thread structure
type Runtime struct {
	Owner    *VM
	ParCount int32
	Calls    []Call
	Thread   Thread
	ThreadID int64
	Optional *[]OptValue
	// These are stacks for different types
	SInt   [STACKSIZE]int64       // int, char, bool
	SFloat [STACKSIZE]float64     // float
	SStr   [STACKSIZE]string      // str
	SAny   [STACKSIZE]interface{} // all other types
}

// Call stores stack of blocks
type Call struct {
	IsFunc   bool
	IsLocal  bool
	Cycle    uint64
	Offset   int32
	Int      int32
	Float    int32
	Str      int32
	Any      int32
	Optional *[]OptValue
	// for loop blocks
	Flags    int16
	Start    int32
	Continue int32 // shift for continue
	Break    int32 // shift for break
	Try      int32 // shift for try
	Recover  int32 // shift for recover
	Retry    int32 // shift for retry
}

func (vm *VM) runConsts(offset int64) (interface{}, error) {
	rt := &Runtime{
		Owner: vm,
	}
	vm.Runtimes = append(vm.Runtimes, rt)
	return rt.Run(offset)
}

func Run(exec *core.Exec, settings Settings) (interface{}, error) {
	if exec == nil {
		return nil, fmt.Errorf(ErrorText(ErrNotRun))
	}
	if exec.CRCStdlib != CRCStdlib || (exec.CRCCustom != 0 && exec.CRCCustom != CRCCustom) {
		return nil, fmt.Errorf(ErrorText(ErrCRC))
	}
	if settings.IsPlayground {
		if err := InitPlayground(&settings); err != nil {
			return nil, err
		}
	}
	vm := &VM{
		Settings: settings,
		Exec:     exec,
		Consts:   make(map[int32]Const),
		Context:  make(map[string]string),
		Runtimes: make([]*Runtime, 0, 32),
		ChCount:  make(chan int64, 16),
		ChError:  make(chan error, 16),
		ChWait:   make(chan int64, 16),
	}
	if settings.IsPlayground {
		vm.Playground.Files = make(map[string]int64)
	}
	if vm.Settings.Cycle == 0 {
		vm.Settings.Cycle = CYCLE
	}
	if vm.Settings.Depth == 0 {
		vm.Settings.Depth = DEPTH
	}
	var (
		pStdin, pStdout, pStderr *os.File
	)
	if vm.Settings.Stdin != nil {
		pStdin = os.Stdin
		os.Stdin = vm.Settings.Stdin
	}
	if vm.Settings.Stdout != nil {
		pStdout = os.Stdout
		os.Stdout = vm.Settings.Stdout
	}
	if vm.Settings.Stderr != nil {
		pStderr = os.Stderr
		os.Stderr = vm.Settings.Stderr
	}
	//	fmt.Println(`CODE`, vm.Exec.Code)
	//fmt.Println(`POS`, vm.Exec.Pos)
	//fmt.Println(`STRING`, vm.Exec.Strings)
	var iotaShift int32
	for i, id := range vm.Exec.Init {
		if i == 0 {
			iotaShift = id
			vm.Consts[id] = Const{Type: core.TYPEINT, Value: int64(0)}
			continue
		}
		switch id - iotaShift {
		case core.ConstDepthID:
			vm.Consts[id] = Const{Type: core.TYPEINT, Value: int64(vm.Settings.Depth)}
			continue
		case core.ConstCycleID:
			vm.Consts[id] = Const{Type: core.TYPEINT, Value: int64(vm.Settings.Cycle)}
			continue
		case core.ConstScriptID:
			vm.Consts[id] = Const{Type: core.TYPESTR, Value: exec.Path}
			continue
		}
		val, err := vm.runConsts(int64(vm.Exec.Funcs[id]))
		if err != nil {
			return nil, err
		}
		var constType uint16
		switch v := val.(type) {
		case int64:
			constType = core.TYPEINT
		case float64:
			constType = core.TYPEFLOAT
		case bool:
			constType = core.TYPEBOOL
			if v {
				val = int64(1)
			} else {
				val = int64(0)
			}
			//				case reflect.TypeOf(float64(0.0)):
			//					retType = core.STACKFLOAT
		case rune:
			constType = core.TYPECHAR
			val = int64(v)
		case string:
			constType = core.TYPESTR
		}
		vm.Consts[id] = Const{Type: constType, Value: val}
	}
	vm.Runtimes = vm.Runtimes[:0]
	rt := vm.newThread(ThWork)
	go func() {
		x := int64(1)
		for x != 0 {
			select {
			case x = <-vm.ChCount:
				if x != 0 {
					vm.ThreadMutex.Lock()
					vm.Count--
					vm.ThreadMutex.Unlock()
				}
			}
		}
	}()
	if settings.SysChan != nil {
		go func() {
		sysChan:
			for {
				x := <-settings.SysChan
				switch x {
				case sysClose:
					break sysChan
				case SysSuspend:
					vm.Stopped = true
				case SysResume:
					vm.Stopped = false
				case SysTerminate:
					rt.Owner.ChError <- fmt.Errorf(ErrorText(ErrTerminated))
					vm.Stopped = false // if it has been suspended
				}
			}
		}()
	}
	result, errResult := rt.Run(0)
	if settings.SysChan != nil {
		settings.SysChan <- sysClose
	}
	if errResult != nil {
		vm.closeAll()
	}
	for vm.Count > 0 {
		select {
		case err := <-vm.ChError:
			vm.closeAll()
			if errResult == nil {
				errResult = err
			}
		default:
		}
	}
	if err, ok := errResult.(*RuntimeError); ok && err.Message == ErrorText(ErrExit) {
		result = err.ID
		errResult = nil
	}
	vm.ChCount <- 0
	close(vm.Runtimes[0].Thread.Chan)
	close(vm.ChCount)
	close(vm.ChError)
	if pStdin != nil {
		os.Stdin = pStdin
	}
	if pStdout != nil {
		os.Stdout = pStdout
	}
	if pStderr != nil {
		os.Stderr = pStderr
	}
	if vm.Settings.IsPlayground {
		DeinitPlayground(vm)
	}
	return result, errResult

}
