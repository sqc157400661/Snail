package main
import (
	"fmt"
	"os"
	"runtime/trace"
	"time"
)

func main() {
	f, err := os.Create("trace.out")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	err = trace.Start(f)
	ch := make(chan string)
	go func(){
		time.Sleep(time.Second)
		ch<- "裸奔的蜗牛，黑乎乎"
		say := make(chan string)
		go sayHello(say)
		fmt.Println(<-say)
	}()
	fmt.Println(<-ch)
	if err != nil {
		panic(err)
	}
	defer trace.Stop()
	// Your program here
}

func sayHello(s chan string){
	time.Sleep(time.Second)
	s<- "hello"
}