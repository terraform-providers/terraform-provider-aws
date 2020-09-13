package analysisutil

import (
	"golang.org/x/tools/go/ssa"
)

// IfInstr returns *ssa.If which is contained in the block b.
// If the block b has not any if instruction, IfInstr returns nil.
func IfInstr(b *ssa.BasicBlock) *ssa.If {
	if len(b.Instrs) == 0 {
		return nil
	}

	ifinstr, ok := b.Instrs[len(b.Instrs)-1].(*ssa.If)
	if !ok {
		return nil
	}

	return ifinstr
}

// Phi returns phi values which are contained in the block b.
func Phi(b *ssa.BasicBlock) (phis []*ssa.Phi) {
	for _, instr := range b.Instrs {
		if phi, ok := instr.(*ssa.Phi); ok {
			phis = append(phis, phi)
		} else {
			// no more phi
			break
		}
	}
	return
}

// Returns returns a slice of *ssa.Return in the function.
func Returns(v ssa.Value) []*ssa.Return {
	var fn *ssa.Function
	switch v := v.(type) {
	case *ssa.Function:
		fn = v
	case *ssa.MakeClosure:
		return Returns(v.Fn)
	default:
		return nil
	}

	var rets []*ssa.Return
	done := map[*ssa.BasicBlock]bool{}
	for _, b := range fn.Blocks {
		rets = append(rets, returnsInBlock(b, done)...)
	}
	return rets
}

func returnsInBlock(b *ssa.BasicBlock, done map[*ssa.BasicBlock]bool) (rets []*ssa.Return) {
	if done[b] {
		return
	}
	done[b] = true

	if len(b.Instrs) != 0 {
		switch instr := b.Instrs[len(b.Instrs)-1].(type) {
		case *ssa.Return:
			rets = append(rets, instr)
		}
	}

	for _, s := range b.Succs {
		rets = append(rets, returnsInBlock(s, done)...)
	}
	return
}
