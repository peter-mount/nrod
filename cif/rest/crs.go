package cif

import (
  bolt "github.com/coreos/bbolt"
  "github.com/peter-mount/golib/codec"
  "log"
  "sort"
)

func (c *CIF) cleanupCRS() error {
  log.Println( "Rebuilding CRS bucket" )

  // Clear the crs bucket
  if err := c.crs.ForEach( func( k, v []byte) error {
    return c.crs.Delete( k )
  }); err != nil {
    return err
  }

  // Refresh CRS map
  crs := make( map[string][]*Tiploc )

  if err := c.tiploc.ForEach( func( k, v []byte) error {
    var tiploc *Tiploc = &Tiploc{}
    codec.NewBinaryCodecFrom( v ).Read( tiploc )

    if tiploc.CRS != "" {
      crs[ tiploc.CRS ] = append( crs[ tiploc.CRS ], tiploc )
    }

    return nil
  }); err != nil {
    return err
  }

  // Sort each crs slice by NLC, hopefully making the more accurate entry first
  // e.g. Look at VIC as an example
  for _, t := range crs {
    if len( t ) > 1 {
      sort.SliceStable( t, func( i, j int ) bool {
        return t[i].NLC < t[j].NLC
      })
    }
  }

  log.Println( "crs", len( crs ) )

  // Now persist
  for k, v := range crs {
    // Array of just Tiploc codes to save space
    var ar []string
    for _, t := range v {
      ar = append( ar, t.Tiploc )
    }

    codec := codec.NewBinaryCodec()
    codec.WriteStringArray( ar )
    if codec.Error() != nil {
      return codec.Error()
    }

    if err := c.crs.Put( []byte( k ), codec.Bytes() ); err != nil {
      return err
    }
  }

  return nil
}

// GetCRS retrieves an array of Tiploc records for the CRS/3Alpha code of a station.
func (c *CIF) GetCRS( tx *bolt.Tx, crs string ) ( []*Tiploc, bool ) {

  b := tx.Bucket( []byte("Crs") ).Get( []byte( crs ) )
  if b == nil {
    return nil, false
  }

  var ar []string
  codec.NewBinaryCodecFrom( b ).ReadStringArray( &ar )

  if len( ar ) == 0 {
    return nil, false
  }

  var t []*Tiploc
  for _, k := range ar {
    if tiploc, exists := c.GetTiploc( tx, k ); exists {
      t = append( t, tiploc )
    }
  }

  return t, len( t ) > 0
}