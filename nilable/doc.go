// Package nilable contains utility structs, interfaces, and methods related to
// object types which may be nil. Simply use the New function to create a new
// Nilable and instantiate its type with its With function.
//
// Typed nilables can be created through functions prefixed by New. Each "New"
// function creates a new Nilable of a different type, with IsNull set to false
// and Data being the zero value of its respective type.
//
// # Implementing Interfaces
//
// Nilables are only guaranteed to contain two methods: IsNil and Value. To
// implement other interfaces, use an Add-prefixed function. The provided
// Add-prefixed functions such as AddJSON or AddScanner. They
// mutate the internal object to include the desired functionality
// (re-assignment is unnecessary).
//
// Extending functionality beyond the provided JSON and Scanner functions
// is simple.
//  1. Create your own struct which implements Nilable and Wraps
//  2. Create a Has and Add function for your struct. Utility methods for Has
//     and Add are provided.
//
// # Options
//
// Options modify the way that nilables work. They are implemented currently on
// JSON during unmarshaling and marshaling.
//
// Usage:
// ```
//
//	 type MyStruct struct {
//	    ID 				  string
//	    SomeOptionalDate  Nilable[time.Time]
//	 }
//
//	 func GetFromDB(db *sqlx.DB, id string) (*MyStruct, error) {
//	    // ... some code
//	    // At this point, all references to nullable.Nilable are nil. Therefore,
//	    // we cannot directly parse the DB values into here.
//	    var x MyStruct
//	    x.SomeOptionalDate = New().With(NewTime())
//	    AddScanner(x.SomeOptionalDate)
//	    if err := qry.Scan() ; err != nil {
//	       return nil, err
//	    }
//	    return &x, nil
//	 }
//
//	 // This is a handler for an application. We assume it returns a JSON
//	 // object
//	 func someHandler() ([]byte, error) {
//	    // ... some code
//	    x, err := GetFromDB(db, "id")
//	    // ... some code
//
//	    // Note: the line below can be performed in a function which
//	    // prepares it for JSON marshalling and unmarshalling.
//		AddJSON(x.SomeOptionalDate)
//		return json.Marshal(x)
//	 }
//
// ```
package nilable
