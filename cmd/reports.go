package main

import (
  "os"
  "encoding/json"
  "fmt"
  "path/filepath"
)



func WriteReport(reportType string, report any) error {
  // reporType can be either db_report, or summary

  err := os.MkdirAll("reports", 0755)
  if err != nil {
    return fmt.Errorf("error creating your 'reports' folder: %w", err)
  }

  jsonData, err := json.MarshalIndent(report, "", "  ")
  if err != nil {
    return fmt.Errorf("error converting report to json: %w", err)
  }

  outputFile, err := os.Create(filepath.Join("reports", reportType+".json"))
  if err != nil{
    return fmt.Errorf("error writing data in your file: %w", err)
  }
  defer outputFile.Close()

  _, err = outputFile.Write(jsonData)
  if err != nil{
    return fmt.Errorf("error writing report: %w", err)
  }

  return nil
}
