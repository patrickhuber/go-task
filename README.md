# go-task
a unit of work library modeled off of the dotnet tpl

## usage

go get it

```bash
go get github.com/patrickhuber/task
```

execute a simple function 

```golang
t := task.Run(func(interface{}) (interface{}, error) {
  return 1, nil
})
err := t.Wait()
```

timeout a task

```golang
ctx, _ := context.WithTimeout(context.Background(), time.Millisecond)
t := task.Run(func(interface{}) (interface{}, error) {
  ch := make(chan struct{})
  <-ch
  return nil, nil
}, task.WithContext(ctx))
t.Wait()
```

cancel a task

```golang
ctx, cancel := context.WithCancel(context.Background())
t := task.Run(func(interface{}) (interface{}, error) {
  ch := make(chan struct{})
  <-ch
  return nil, nil
}, task.WithContext(ctx))
cancel()
t.Wait()
```

when all tasks

```
t := task.WhenAll(task.Completed(), task.FromResult(1))
t.Wait()
```
