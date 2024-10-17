### mod example

```golang
type ExampleStruct struct {
    Name     string    `mod:"trim,lowercase"`                      // Trims spaces and converts to lowercase
    Age      int       `mod:"default=18"`                          // Sets default value to 18 if empty
    UnixTime int64     `mod:"unixtime=ms,force"`                   // Sets Unix time in milliseconds (forced)
    Counter  int64     `mod:"counter,force"`                       // Increments counter (forced)
    Email    string    `mod:"lowercase,trim"`                      // Trims spaces and converts to lowercase
    Created  time.Time `mod:"dateFormat=2006-01-02T15:04:05Z07:00"`// Formats date according to specified layout
}
```
