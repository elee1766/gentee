// Copyright 2018 Alexey Krivonogov. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package gentee

// Run executes run block
func (vm *VirtualMachine) Run() (interface{}, error) {
	if vm.RunID == Undefined {
		return nil, runtimeError(vm, ErrNoRun, nil)
	}
	rt := newRunTime(vm)
	rt.Run()
	return nil, nil
}
