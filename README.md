# pararun

Simple runner library for parallel operations with limits.

## Example usage

```go
var globalRunner = pararun.New(0)

func doSomething() {
    globalRunner.Run(func() {
        // this will run in background up to some point
    })
}

func scanData(data []*obj) error {
    return pararun.ForEach(globalRunner, data, func(n int, o *obj) error {
        // do something with o
        return nil
    })
}
```
