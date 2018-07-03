package pagent

import (
	"fmt"
	"testing"
	"time"
)

type MyBiz struct {
	Master
}

func NewBiz() *MyBiz {
	return &MyBiz{}
}

func (a *MyBiz) BizRunning(id, str string) error {
	fmt.Println("[MyBiz BizRunning] str=" + str)
	return nil
}

func (a *MyBiz) BizFinish(id string, err error) error {
	fmt.Println("[MyBiz BizFinish] id=" + id)
	return err
}

func TestDecode(t *testing.T) {

	a := NewBiz()

	fmt.Println("worker1-------------------------")
	a.GetWorker("worker1").Start("bash", a.BizRunning, a.BizFinish)
	a.GetWorker("worker1").Input("ls")
	time.Sleep(1 * time.Second)
	a.DelWorker("worker1")

	fmt.Println("worker2-------------------------")
	a.GetWorker("worker2").Start("ifconfig", nil, a.BizFinish)
	time.Sleep(1 * time.Second)
	a.DelWorker("worker2")

	fmt.Printf("end!----------------------------")

}
