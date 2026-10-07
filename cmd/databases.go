package main


import (
  "fmt"
  "os"
  "os/exec"
  "sync"
  "time"
  "log"
  "path/filepath"
  "bytes"
  "strings"
)


type DumpResults struct {
  DBName            string            `json:db_name`
  DumpDBStatus      string            `json:dump_db_status`
  Duration          float64            `json:duration`
  Error             string             `json: error`
}

type DumpSummary struct {
  DBNames         []string              `json:db_names`
  DBCount         int                   `json:db_count`
  DumpSucceed     int                   `json:dump_succeed`
  DumpFailed      int                   `json:dump_failed`
}


func RunCommand(host, port, username, password string) (string, error) {

  cmd := exec.Command(
      "mysql",
      "-h", host,
      "-P", port,
      "-u", username,
      "-p"+password,
      "-N",
      "-e", "SHOW DATABASES;",
  )

  var stderr bytes.Buffer
  cmd.Stderr = &stderr

  outBytes, err := cmd.Output()

  if err != nil {
    return "", fmt.Errorf(
      "mysql command failed: %v: %s",
      err,
      stderr.String(),
    )
  }
  //fmt.Println(outBytes)
  //fmt.Println(string(outBytes))
  return string(outBytes), nil
}


func ConcurrentDumpDB(databases []string, host, port, username, password string, maxConcurrent int) error {
    if len(databases) == 0 {
      return fmt.Errorf("there's no databse in your rds instance")
    }


    if maxConcurrent <= 0 {
  		return fmt.Errorf("maxConcurrent must be greater than 0")
  	}


    n := len(databases)
    var wg sync.WaitGroup
    limiter := make(chan struct {}, maxConcurrent)
    dumpResult := make(chan DumpResults, n)

    directory := strings.Split(host, ".")

    err := os.MkdirAll(directory[0]+"-dump", 0755)
    if err != nil {
        return fmt.Errorf("error creating dump directory: %w", err)
    }
    //progress := NewProgressManager(databases)

    // Initial display
    //progress.render()

    for _, db := range databases {

      wg.Add(1)
      go func(db string) {
        defer wg.Done()

        limiter <- struct{}{}

        // We acquired a slot.
        //progress.Update(db, "RUNNING")

        defer func() { <- limiter }()

        start := time.Now()

        cmd := exec.Command(
          "caffeinate",
          "-i",
          "mysqldump",
          "-h", host,
          "-P", port,
          "-u", username,
          "-p"+password,
          "--single-transaction",
          "--quick",
          "--skip-lock-tables",
          db,
        )

        var stderr bytes.Buffer
        cmd.Stderr = &stderr

        outputFile, err := os.Create(filepath.Join(directory[0]+"-dump", db+"_dump.sql"))
        if err != nil {
          //progress.Update(db, "FAILED")

          dumpResult <- DumpResults{
              DBName:       db,
              Error:        stderr.String(),
              Duration:     time.Since(start).Seconds(),
              DumpDBStatus: "failed",
          }

          return

        }
        defer outputFile.Close()

        cmd.Stdout = outputFile

        err = cmd.Run()
        if err != nil {
          //progress.Update(db, "FAILED")

          dumpResult <- DumpResults{
            DBName:           db,
            Error:            stderr.String(),
            Duration:         time.Since(start).Seconds(),
            DumpDBStatus:     "failed",
          }

          //WriteReport("db_report", dumpResult)

          return
        }

        //progress.Update(db, "DONE")

        //timeSince := time.Since(start)

        dumpResult <- DumpResults {
          DBName: db,
          DumpDBStatus: "succeeded",
          Duration: time.Since(start).Seconds(),
          Error: "",
        }

        //WriteReport("db_report", dumpResult)

      }(db)
    }

    go func(){
      wg.Wait()
      close(dumpResult)
    }()

    // Write db report and summary to json file

    summary := &DumpSummary{
      DBCount: len(databases),
    }

    for result := range dumpResult{
      err := WriteReport(result.DBName, result)

      if err != nil {
        log.Printf(
          "failed to write report for %s",
          result.DBName,
          err,
        )
      }


      if result.DumpDBStatus == "succeeded" {
        summary.DumpSucceed ++
        summary.DBNames = append(summary.DBNames, result.DBName)
      }else{
        summary.DumpFailed++
      }
    }

    err = WriteReport("summary", *summary)
    if err != nil {
      return fmt.Errorf("Failed to write summary report: %w", err)
    }

    fmt.Printf(
      "\n%d succeeded, %d failed\n",
      summary.DumpSucceed,
      summary.DumpFailed,
    )
    return nil
}
