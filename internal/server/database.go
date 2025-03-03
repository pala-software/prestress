package server

import "database/sql"

func (server *Server) connectToDatabase() error {
	var err error

	server.DB, err = sql.Open("postgres", server.dbConnStr)
	if err != nil {
		return err
	}

	// Test database connection.
	err = server.DB.Ping()
	if err != nil {
		return err
	}

	return nil
}
