# go-task

a unit of work library modeled off of the dotnet tpl

## usage

go get it

```bash 
go get github.com/patrickhuber/go-task
```

execute a simple function 

```golang
t := task.RunFunc(func() interface{} {
  return 1
})
t.Wait()
```

passing in data

```golang
t := task.RunWith(func(state interface{}){
  data := state.(string)
  fmt.Println(data)
}, task.WithState("this is data"))
t.Wait()
```

timeout a task

```golang
ctx, _ := context.WithTimeout(context.Background(), time.Millisecond)
t := task.RunAction(func(){
  ch := make(chan struct{})
  <-ch
}, task.WithContext(ctx))
t.Wait()
```

cancel a task

```golang
ctx, cancel := context.WithCancel(context.Background())
t := task.RunAction(func() {
  ch := make(chan struct{})
  <-ch
}, task.WithContext(ctx))
cancel()
err := t.Wait() // error contains context cancellation error
```

when all tasks

```
t := task.WhenAll(task.Completed(), task.FromResult(1))
t.Wait()
```

when any tasks

```
t := task.WhenAll(task.Completed(), task.FromResult(1))
t.Wait()
```

aggregate errors

```
```