## Modifier Example

Modifiers are functions that can be applied to the fields of your struct during serialization and deserialization. They are particularly useful for tasks like trimming spaces, converting to lowercase, setting default values, etc. The modifiers are defined using struct tags.

### Example Usage

```golang

type ExampleStruct struct {
	Name     string    `mod:"trim,lowercase"`                      // Trims spaces and converts to lowercase
	Age      int       `mod:"default=18"`                          // Sets default value to 18 if empty
	UnixTime int64     `mod:"unixtime=ms,force"`                   // Sets Unix time in milliseconds (forced)
	Counter  int64     `mod:"counter,force"`                       // Increments counter (forced)
	Email    string    `mod:"lowercase,trim"`                      // Trims spaces and converts to lowercase
}

// Example of how to use modifiers in your code
func ApplyModifiersToStruct() {
	example := ExampleStruct{
		Name:     "  John Doe  ",
		Age:      0,
		UnixTime: 0,
		Counter:  0,
		Email:    "  john.doe@domain.com  ",
		Created:  time.Now(),
	}

	// Apply modifiers
	if err := redisdb.ApplyModifiers(&example); err == nil {
		// Log or print the modified struct
		fmt.Println(example)
	}
}
```