package sandbox

import (
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

type Sync struct {
	wg   sync.WaitGroup
	mu   sync.Mutex
	once sync.Once
	cond sync.Cond
	rw   sync.RWMutex

	counter         int
	stringSliceData []string
	intData         int
	isReady 		bool
	counter32       int32
}

func NewSync() *Sync {
	return &Sync{}
}

////////////////////////////////////////////////////////////////////////

func (s *Sync) RunWGExample(count int) {
	for i := 0; i < count; i++ {
		s.wg.Add(1)
		go func(i int) {
			log.Println("Горутина", i, "начала работу")
			log.Println("Горутина", i, "завершила работу")
			s.wg.Done()
		}(i)
	}
	s.wg.Wait()
	log.Println("Все горутины завершили работу")
}

////////////////////////////////////////////////////////////////////////

func (s *Sync) RunMutexExample(count int) {
	var wg sync.WaitGroup
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func() {
			s.mu.Lock()
			s.counter++
			log.Println("Значение счетчика:", s.counter)
			s.mu.Unlock()
			wg.Done()
		}()
	}
	wg.Wait()
	log.Println("Все горутины завершили работу")
}

////////////////////////////////////////////////////////////////////////

func (s *Sync) loadDataOnce() {
	log.Println("Загрузка данных")
	s.stringSliceData = []string{"данные1", "данные2", "данные3"}
	log.Println("Загружено:", s.stringSliceData)
}

func (s *Sync) RunOnceExample() {
	for i := 0; i < 3; i++ {
		s.wg.Add(1)
		go func(i int) {
			defer s.wg.Done()
			s.once.Do(s.loadDataOnce)
			log.Println("Завершен поток:", i)
		}(i)
	}
	s.wg.Wait()
}

////////////////////////////////////////////////////////////////////////

// Обертка в локи нужна, чтобы, если функции будут вызваны в множестве 
// горутин, то отстальные горутины ожидали предыдущих.

func (s *Sync) waitForCondition() {
	s.cond.L.Lock()
	defer s.cond.L.Unlock()

	for !s.isReady {
		log.Println("Ожидание кондиции...")
		s.cond.Wait()
	}
	log.Println("Конциция достигнута, продолжаем.")
}

func (s *Sync) setCondition() {
	time.Sleep(2 * time.Second)
	s.cond.L.Lock()
	s.isReady = true
	s.cond.Signal() // Signal one waiting Goroutine
	s.cond.L.Unlock()
}

func (s *Sync) RunCondExample() {
	s.isReady = false
	s.cond.L = &sync.Mutex{}
	go s.waitForCondition()
	go s.setCondition()

	// Allow Goroutines to finish before program exits
	time.Sleep(5 * time.Second)
}

////////////////////////////////////////////////////////////////////////

func (s *Sync) readData() int {
	s.rw.RLock() // Read Lock
	defer s.rw.RUnlock()
	return s.intData
}

func (s *Sync) writeData(value int) {
	s.rw.Lock() // Write Lock
	defer s.rw.Unlock()
	s.intData = value
}

func (s *Sync) RunRWMutexExample(num int) {
	for i := 1; i <= 5; i++ {
		go func() {
			log.Println("Прочитана дата:", s.readData())
		}()
	}

	s.writeData(num)

	// Allow Goroutines to finish before program exits
	time.Sleep(5 * time.Second)
}

////////////////////////////////////////////////////////////////////////

func (s *Sync) incrementCounter() {
	defer s.wg.Done()
	for i := 0; i < 100000; i++ {
		atomic.AddInt32(&s.counter32, 1)
	}
}

func (s *Sync) RunAtomicExample() {
	s.wg.Add(2)
	go s.incrementCounter()
	go s.incrementCounter()
	s.wg.Wait()

	log.Println("Счетчик:", atomic.LoadInt32(&s.counter32))
}

////////////////////////////////////////////////////////////////////////

func eventGenerator(events chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()

	for i := 1; i <= 5; i++ {
		time.Sleep(1 * time.Second)
		event := fmt.Sprintf("Event %d", i)
		events <- event
	}
	close(events)
}

func eventProcessor(events <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()

	for event := range events {
		fmt.Println("Processing:", event)
		// Simulate some processing time
		time.Sleep(500 * time.Millisecond)
	}
}

func (s *Sync) RunChannelsExample() {
	events := make(chan string)

	s.wg.Add(1)
	go eventGenerator(events, &s.wg)
	s.wg.Add(1)
	go eventProcessor(events, &s.wg)

	s.wg.Wait()
}