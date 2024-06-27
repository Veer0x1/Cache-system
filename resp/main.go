package resp

type RESPHandler struct {
	String     simpleString
	BulkString bulkString
	Array      array
	Integer    integer
	Error      errorString
	Nil 	  nilString
}
