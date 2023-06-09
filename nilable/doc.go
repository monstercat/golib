// Package nilable contains utility structs, interfaces, and methods related to
// object types which may be nil. Simply use the New function to create a new
// Nilable.
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
//	    x.SomeOptionalDate = New[time.Time]()
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
//	    // Note: JSON functionality is included automatically.
//		return json.Marshal(x)
//	 }
//
// ```
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
package nilable
