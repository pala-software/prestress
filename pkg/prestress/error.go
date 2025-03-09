package prestress

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/lib/pq"
)

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
		case "42":
			writer.WriteHeader(404)
			return
		}
	}

	fmt.Println(err)
	writer.WriteHeader(500)
}
