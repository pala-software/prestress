package prestress

import (
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// TODO: Test
func handleOperationError(writer http.ResponseWriter, err error) {
	switch err {
	case pgx.ErrNoRows:
		writer.WriteHeader(200)
		return
	case ErrForbiddenSchema:
		writer.WriteHeader(404)
		return
	}

	if err, ok := err.(*pgconn.PgError); ok {
		switch err.Code[0:2] {
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
