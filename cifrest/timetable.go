package cifrest

import (
  "github.com/peter-mount/golib/rest"
  "log"
  "strings"
)

// TimetableHandler implements a REST endpoint which returns the scheduled
// services at a specific CRS endpoint for a given date and hour of the day.
func (c *CIFRest) TimetableHandler( r *rest.Rest ) error {
  crs := r.Var( "crs" )

  date := r.Var( "date" )

  row := c.dbService.QueryRow( "SELECT timetable.timetable( $1, $2 ) AS json", crs, date )

  var tt []uint8
  err := row.Scan( &tt )
  if err != nil {
    r.Status( 500 )
    log.Printf( "500: crs %s date %s: %s", crs, date, err )
    return err
  }

  // As the JSON is generated by postgreSQL then just write the raw string
  // to the result
  r.Status( 200 ).Reader( strings.NewReader( string(tt) ) )

  return nil
}