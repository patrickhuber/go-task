# go-task

a unit of work library modeled off of the dotnet tpl

## install

go get it

```bash 
go get github.com/patrickhuber/go-task
```

## usage

### execute a simple function 

```golang
t := task.RunFunc(func() interface{} {
  return 1
})
t.Wait()
```

using goroutines

```golang
intChan := make(chan int)
go func(){
  intChan <- 1
}()
<- intChan
```

### passing in data

```golang
t := task.RunWith(func(state interface{}){
  data := state.(string)
  fmt.Println(data)
}, task.WithState("this is data"))
t.Wait()
```

using goroutines

```golang
stringChan := make(chan string)
go func(data string){
  stringChan <- data
}("this is data")
fmt.Println(<-stringChan)
```

### timeout a task

```golang
ctx, _ := context.WithTimeout(context.Background(), time.Millisecond)
t := task.RunAction(func(){
  ch := make(chan struct{})
  <-ch
}, task.WithContext(ctx))
t.Wait()
```

### cancel a task

```golang
ctx, cancel := context.WithCancel(context.Background())
t := task.RunAction(func() {
  ch := make(chan struct{})
  <-ch
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
t := task.WhenAll(task.Completed(), task.FromResult(1))
t.Wait()
```

### aggregate errors

```golang
tasks := []task.ObservableTask{}
for i := 0; i < 3; i++{
  task.FromError(fmt.Errorf("%d",d))
}
err := task.WhenAll(tasks).Wait()

// prints 3
fmt.Println(len(err.(task.AggregateError).Errors()))
```

```golang
var errorChans := [3]chan error
for i, errChan := range errorChans{
  errChan = make(chan error)
  go func(i int){
    errChan <- fmt.Errorf("%d", i)
  }(i)
}

errors := []error{}
for _, errChan := range errorChans{
  err := <- errChan
  errors = append(errors, err)
}
// prints 3
fmt.Println(len(errors))
```