package palakit

import "database/sql"

// TODO: Test
func (server *Server) ConnectToDatabase() error {
	var err error

	server.DB, err = sql.Open("postgres", server.DbConnStr)
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
