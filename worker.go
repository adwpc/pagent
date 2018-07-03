package pagent

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"syscall"
)

type _RUNFUNC func(id, info string) error
type _FINFUNC func(id string, err error) error

type Worker struct {
	//私有成员
	_in     io.WriteCloser
	_out    io.Reader
	_pid    int
	_cmd    *exec.Cmd
	_master *Master

	Running bool
	Cond    *sync.Cond

	//业务。。。
	Id          string
	RunningCall _RUNFUNC
	FinishCall  _FINFUNC
}

func NewWorker(id string) *Worker {
	return &Worker{
		Id:   id,
		Cond: sync.NewCond(&sync.Mutex{}),
	}
}

func (r *Worker) Start(cmd string, RunningCall _RUNFUNC, FinishCall _FINFUNC, arg ...string) error {
	r._cmd = exec.Command(cmd, arg...)
	var err error
	if r._in, err = r._cmd.StdinPipe(); err != nil {
		return err
	}

	if r._out, err = r._cmd.StdoutPipe(); err != nil {
		return err
	}

	if err := r._cmd.Start(); err != nil {
		return err
	}

	r.RunningCall = RunningCall
	r.FinishCall = FinishCall
	r._pid = r._cmd.Process.Pid
	if r._master != nil {
		r._master.RunWorker(r.Id)
	}

	return nil

}

func (r *Worker) RegMaster(m *Master) error {
	if m == nil {
		return errors.New("master == nil")
	}
	r._master = m
	return nil
}

func (r *Worker) Wait() error {
	if r._pid == 0 {
		return errors.New("r._pid == 0")
	}

	defer r._in.Close()
	err := r._cmd.Wait()
	if err != nil {
		return err
	}

	return nil
}

func (r *Worker) Stop() error {
	if r._pid == 0 {
		return errors.New("r._pid == 0")
	}

	pro, err := os.FindProcess(r._pid)
	if err != nil {
		return err
	}

	err = pro.Signal(syscall.SIGINT)
	if err != nil {
		return err
	}
	return nil
}

func (r *Worker) Input(in string) error {
	if r._in == nil {
		return errors.New("_in == nil")
	}

	if !strings.Contains(in, "\n") {
		in += "\n"
	}

	i := 0
	total := 0
	for {
		if n, err := r._in.Write([]byte(in)); err != nil {
			return err
		} else {
			total += n
		}

		if total >= len(in) {
			return nil
		}
		i++
		if i > 3 {
			return errors.New("write data not enough!")
		}
		time.Sleep(time.Millisecond * 100)
	}
	return nil

}

func (r *Worker) Output() io.Reader {
	return r._out
}
