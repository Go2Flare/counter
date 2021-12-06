package main

import (
	"fmt"
	"sync"
	"time"
)

/*
作业功能：实现一个计数器模块，不依赖外部三方模块和存储
要求进程内协程安全、异步、高性能按指标key-value统计
使用示例：counter.Init()
counter.Flush2broker(5000, FuncCbFlush)//每 5 秒调用一次 FlushCb 并重置计数器
counter.Incr("get.called", 123)
counter.Incr("get.called", 456)
*/

type Counter struct {
	//线程安全map
	M *sync.Map
}

func (c *Counter) Init()  {
	*c = Counter{&sync.Map{}}
}

func (c *Counter) Incr(s interface{}, n int) {
	//不存在保存,存在则取出
	if v, ok := c.M.LoadOrStore(s, n); ok {
		c.M.Store(s, v.(int)+n)
	}
}

func (c *Counter) Get(s interface{}) int{
	v, _ := c.M.Load(s)
	return v.(int)
}

func (c *Counter) Flush2Broker(ms int, FuncCbFlush func()) {
	go func(){
		//循环调用FuncCbFlush
		ticker := time.NewTicker(time.Millisecond * time.Duration(ms))
		go func() {
			for range ticker.C {
				fmt.Printf("超时%.2fs, 重新刷新计数器\n", float64(ms/1000))
				FuncCbFlush()
			}
		}()

		time.Sleep(time.Millisecond * 60000)//只循环60s内
		ticker.Stop()
	}()
}

// 重置计数器
func (c *Counter) FuncCbFlush() {
	c.Init()
}


func TestUse() {
	counter := &Counter{}
	counter.Init()
	FuncCbFlush := counter.FuncCbFlush
	counter.Incr("get.called", 123)
	counter.Incr("get.called", 456)
	counter.Flush2Broker(5000, FuncCbFlush)
	//每2s打印一次结果
	for i := 0; i < 5; i++ {
		fmt.Println(counter.Get("get.called"))
		time.Sleep(2 * time.Second)
	}
}

//协程安全
func TestConcurrent() {
	counter := &Counter{}
	counter.Init()
	defer func(){
		fmt.Println(counter.Get("0001"))
	}()
	go func(){
		for i:= 0; i<1000; i++{
			counter.Incr("0001", 1)
		}
		for i:=0; i<1000; i++{
			counter.Incr("0001", 1)
		}
		for i:=0; i<1000; i++{
			counter.Incr("0001", 1)
		}
	}()
	time.Sleep(time.Second*1)
}

//异步
func testConsistency(){
	var wg sync.WaitGroup
	wg.Add(4)
	ss := []string{"1","2","3","4"}

	var cs []*Counter
	for i:=0; i<4; i++{
		c := &Counter{}
		c.Init()
		cs= append(cs, c)
	}
	fmt.Println(cs)
	for i:=0; i<len(ss); i++{
		cur := i
		go func(){
			defer wg.Done()
			for j:=0; j<10000; j++{
				cs[cur].Incr(ss[cur], 1)
			}
			fmt.Printf("%v计数器操作完成\n", cur)
		}()
	}
	wg.Wait()
	for i:=0; i<len(cs); i++{
		fmt.Println(cs[i].Get(ss[i]))
	}
}

func main(){
	//TestUse()
	//TestConcurrent()
	//testConsistency()
}

