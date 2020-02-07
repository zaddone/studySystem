package opencv
/*
#cgo LDFLAGS: -L. -lopencv_core
#include "copencv.h"
*/
import "C"
import (
	"unsafe"
	//"math"
)

func GetCurveFittingWeight(X []float64,Y []float64, W []float64) bool {
	_x := (*C.double)(unsafe.Pointer(&X[0]))
	_y := (*C.double)(unsafe.Pointer(&Y[0]))
	Len := C.int(len(X))
	Max := C.int(len(W))
	_w := (*C.double)(unsafe.Pointer(&W[0]))
	Out := C.GetCurveWeight(_x,_y,Len,Max,_w)
	if Out != 0 {
		//fmt.Println(W)
		return true
	}
	return false
}
