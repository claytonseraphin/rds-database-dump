package main

import (
  "fmt"
  "flag"
  "log"
  "strings"
  "os"

  "github.com/joho/godotenv"
)


// func DumpDB(db_name, host, port, username, password string) error {
//     cmd := exec.Command(
//       "mysqldump",
//       "-h", host,
//       "-P", port,
//       "-u", username,
//       "-p"+password,
//       db_name,
//     )
//
//     outputFile, err := os.Create(db_name+"_dump.sql")
//     if err != nil {
//       return fmt.Errorf("error creating your dump file: %v", err)
//
//     }
//     defer outputFile.Close()
//
//     cmd.Stdout = outputFile
//
//     err = cmd.Run()
//     if err != nil {
//       log.Fatalf("error running mysqldump command: %v", err)
//
//     }
//     return nil
// }



func main () {

  if err:= godotenv.Load(); err != nil {
    log.Fatalf("Error loading .env file")
  }


  host := flag.String("h", os.Getenv("HOST"), "The host of the RDS")
  port := flag.String("P", os.Getenv("PORT"), "Mysql Port")
  username := flag.String("u", os.Getenv("USERNAME"), "Mysql user")
  password := flag.String("p", os.Getenv("PASSWORD"), "User password")
  //database := flag.String("d", "", "The target database")

  flag.Parse()

  getCommandOutput, err := RunCommand(*host, *port, *username, *password)
  if err != nil {
    log.Fatalf("an error occured %v", err)
  }

  databases := strings.Split(strings.TrimSpace(getCommandOutput), "\n")

  fmt.Println(databases)

  systemsDB := map[string]bool{
    "mysql":                  true,
    "information_schema":     true,
    "performance_schema":     true,
    "sys":                    true,
  }

  var userDatabases []string

  for _, db := range databases{
    if !systemsDB[db] {
      userDatabases = append(userDatabases, db)
    }
  }
  // nonValidDB := []string{"mysql", "information_schema", "performance_schema", "sys"}
  // for _, db := range nonValidDB {
  //   if i := slices.Index(databases, db); i != -1{
  //     databases = slices.Delete(databases, i, i+1)
  //   }
  // }

  err = ConcurrentDumpDB(userDatabases, *host, *port, *username, *password, 2)
  if err !=nil {
    log.Fatalf("an error occured: %v", err)
  }


}
