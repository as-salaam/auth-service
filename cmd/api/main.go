package main

import (
	"flag"
	"github.com/as-salaam/auth-service/internal/database"
	"log"
)

func main() {
	DBHost := flag.String("dbhost", "localhost", "Enter the host of the DB server")
	DBName := flag.String("dbname", "auth_service", "Enter the name of the DB")
	DBUser := flag.String("dbuser", "postgres", "Enter the name of a DB user")
	DBPassword := flag.String("dbpassword", "postgres", "Enter the password of user")
	DBPort := flag.Uint("dbport", 5432, "Enter the port of DB")
	Timezone := flag.String("dbtimezone", "Asia/Dushanbe", "Enter your timezone to connect to the DB")
	DBSSLMode := flag.Bool("dbsslmode", false, "Turns on ssl mode while connecting to DB")
	HTTPPort := flag.Uint("httpport", 4000, "Which port to listen")
	RPCPort := flag.Uint("rpcport", 4001, "Which port to listen")
	flag.Parse()

	db, err := database.DBInit(*DBHost, *DBName, *DBUser, *DBPassword, *DBPort, *Timezone, *DBSSLMode)
	if err != nil {
		log.Fatal("db connection:", err)
	}

	go func() {
		err = RunGRPC(*RPCPort, db)
		if err != nil {
			log.Fatalf("cannot run grpc server: %s", err)
		}
	}()

	err = RunHTTP(*HTTPPort, db)
	if err != nil {
		log.Fatalf("cannot run http server: %s", err)
	}
}
