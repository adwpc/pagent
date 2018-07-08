package pagent

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"sync"
	"time"
)

type Master struct {
	pool  map[string]*Worker
	mutex sync.Mutex
}

func NewMaster() *Master {
	return &Master{
		pool: make(map[string]*Worker),
	}
}

func (m *Master) GetWorker(id string) *Worker {
	m.mutex.Lock()
	if m.pool == nil {
		m.pool = make(map[string]*Worker)
	}
	if _, ok := m.pool[id]; !ok {
		m.pool[id] = NewWorker(id)
		m.pool[id].RegMaster(m)
	}
	m.mutex.Unlock()

	return m.pool[id]
}

func (m *Master) RunWorker(id string) error {
	m.mutex.Lock()
	if m.pool == nil {
		m.mutex.Unlock()
		return errors.New("m.pool == nil")
	}
	worker, ok := m.pool[id]
	m.mutex.Unlock()
	if !ok {
		return errors.New("can't find worker id=" + id)
	}
	worker.Running = true
	go func() {
		reader := bufio.NewReader(worker.Output())
		var buf bytes.Buffer
		for {
			line, isPrefix, err := reader.ReadLine()
			if len(line) > 0 {
				buf.Write(line)
				//整行
				if !isPrefix {
					if worker.RunningCall != nil {
						worker.RunningCall(id, buf.String())
					}
					buf.Reset()
				}
			}

			if err == io.EOF {
				break
			} else if err != nil {
				break
			}

			time.Sleep(time.Millisecond * 20)
		}

		err := worker.Wait()
		if worker.FinishCall != nil {
			worker.FinishCall(id, err)
		}
		worker.Running = false
	}()

	return nil
}

func (m *Master) DelWorker(id string) error {
	m.mutex.Lock()
	if worker, ok := m.pool[id]; ok {
		m.mutex.Unlock()
		worker.Stop()
		for {
			if !worker.Running {
				delete(m.pool, id)
				break
			}
			time.Sleep(20 * time.Millisecond)
		}
	} else {
		m.mutex.Unlock()
		return errors.New("can't find worker id=" + id)
	}
	return nil
}
