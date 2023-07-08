# go-task

a unit of work library modeled off of the dotnet tpl

## install

go get it

```bash 
go get github.com/patrickhuber/go-task
```

## usage

```golang
package main

import (
	"net/http"

	"github.com/patrickhuber/go-task"
)

func main() {
	urls := []string{
		"http://www.golang.org/",
		"http://www.google.com/",
		"http://www.yahoo.com",
	}

	tasks := []task.Task{}
	for _, url := range urls {
		t := task.RunActionWith(func(state interface{}) {
			url := state.(string)
			http.Get(url)
		}, task.WithState(url))
		tasks = append(tasks, t)
	}

	task.WhenAll(tasks...).Wait()
}
```

Try It [here](https://go.dev/play/p/Ur4z-KBabvV)

## feature usage


### return data

```golang
t := task.RunFunc(func() interface{} {
  return 1
})
t.Wait()
t.Result()
```

### passing in data

```golang
t := task.RunWith(func(state interface{}){
  data := state.(string)
  fmt.Println(data)
}, task.WithState("this is data"))
t.Wait()
```

### timeout a task

```golang
t := task.RunAction(func(){
  	ch := make(chan struct{})
  	defer close(ch)
	select {
		case <-ch:
		case <-time.After(time.Second):
	}
}, task.WithTimeout(time.Millisecond))
t.Wait()
```

### cancel a task

```golang
ctx, cancel := context.WithCancel(context.Background())
t := task.RunAction(func() {
  	ch := make(chan struct{})
  	defer close(ch)
	select {
		case <-ch:
		case <-time.After(time.Second):
	}
}, task.WithContext(ctx))
cancel()
err := t.Wait() // error contains context cancellation error
```

### when all tasks

```golang
t := task.WhenAll(task.Completed(), task.FromResult(1))
t.Wait()
```

### when any tasks

```golang
t := task.WhenAny(task.Completed(), task.FromResult(1))
t.Wait()
```

### aggregate errors

```golang
tasks := []task.Task{}
for i := 0; i < 3; i++{
  t:= task.FromError(fmt.Errorf("%d",d))
  tasks = append(tasks, t)
}
err := task.WhenAll(tasks).Wait()

// prints 3
fmt.Println(len(err.(task.AggregateError).Errors()))
```

### continuation

```golang
t := task.RunFunc(func() interface{} {
  return 1
})
cont := t.ContinueFunc(func(t task.Task) interface{} {
  value := t.Result()
  i, ok := value.(int)
  if !ok {
    return nil
  }
  return i + 1
})
cont.Wait()
count := cont.Result()
fmt.Println(count) // prints 2
```
