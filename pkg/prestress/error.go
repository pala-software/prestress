package prestress

import (
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type HttpError interface {
	error
	Message() string
	Status() int
}

func HandleDatabaseError(writer http.ResponseWriter, err error) {
	switch err {
	case pgx.ErrNoRows:
		writer.WriteHeader(404)
		return
	case ErrForbiddenSchema:
		writer.WriteHeader(403)
		return
	}

	if err, ok := err.(*pgconn.PgError); ok {
		switch {
		case err.Code == "42501":
			writer.WriteHeader(403)
			writer.Write([]byte(err.Message))
			return
		case err.Code == "42P01":
			writer.WriteHeader(404)
			writer.Write([]byte(err.Message))
			return
		case err.Code == "42501":
			writer.WriteHeader(403)
			writer.Write([]byte(err.Message))
			return
		case err.Code[0:2] == "22" ||
			err.Code[0:2] == "23" ||
			err.Code[0:2] == "3F" ||
			err.Code[0:2] == "42" ||
			err.Code[0:2] == "44" ||
			err.Code[0:2] == "54":
			writer.WriteHeader(400)
			writer.Write([]byte(err.Message))
			return
		case err.Code[0:2] == "53":
			fmt.Println(err)
			writer.WriteHeader(403)
			return
		case err.Code[0:2] == "55":
			writer.WriteHeader(409)
			writer.Write([]byte(err.Message))
			return
		case err.Code == "P0001":
			writer.WriteHeader(400)
			writer.Write([]byte(err.Message))
			return
		}
	}

	if err, ok := err.(HttpError); ok {
		writer.WriteHeader(err.Status())
		writer.Write([]byte(err.Message()))
	}

	fmt.Println(err)
	writer.WriteHeader(500)
}
