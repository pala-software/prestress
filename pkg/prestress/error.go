package prestress

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/lib/pq"
)

// TODO: Test
func handleOperationError(writer http.ResponseWriter, err error) {
	switch err {
	case sql.ErrNoRows:
		writer.WriteHeader(200)
		return
	case ErrForbiddenSchema:
		writer.WriteHeader(404)
		return
	}

	if err, ok := err.(*pq.Error); ok {
		switch err.Code.Class() {
		case "23":
			writer.WriteHeader(400)
			writer.Write([]byte(err.Message))
			return
		case "42":
			writer.WriteHeader(404)
			writer.Write([]byte(err.Message))
			return
		}
		fmt.Printf("CODE: %s\n", err.Code)
	}

	fmt.Println(err)
	writer.WriteHeader(500)
}
