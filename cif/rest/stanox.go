package cif

import (
  bolt "github.com/coreos/bbolt"
  "github.com/peter-mount/golib/codec"
  "log"
  "sort"
  "strconv"
)

func (c *CIF) cleanupStanox() error {
  log.Println( "Rebuilding Stanox bucket" )

  // Clear the crs bucket
  if err := c.stanox.ForEach( func( k, v []byte) error {
    return c.stanox.Delete( k )
  }); err != nil {
    return err
  }

  // refresh stanox map
  stanox := make( map[int][]*Tiploc )

  if err := c.tiploc.ForEach( func( k, v []byte) error {
    var tiploc *Tiploc = &Tiploc{}
    codec.NewBinaryCodecFrom( v ).Read( tiploc )

    if tiploc.Stanox > 0 {
      stanox[ tiploc.Stanox ] = append( stanox[ tiploc.Stanox ], tiploc )
    }

    return nil
  }); err != nil {
    return err
  }

  // Now for each stanox, if 1 entry has a crs then use that for all entries
  for _, s := range stanox {
    var crs string
    for _, t := range s {
      // Don't use X?? or Z?? CRS codes here
      if t.CRS != "" && !( t.CRS[0:1]=="X" || t.CRS[0:1]=="Z" ) {
        crs = t.CRS
      }
    }

    // Update to the new crs field
    if crs != "" {
      for _, t := range s {
        if t.CRS != crs {
          t.CRS = crs
          codec := codec.NewBinaryCodec()
          codec.Write( t )
          if codec.Error() != nil {
            return codec.Error()
          }

          if err := c.tiploc.Put( []byte( t.Tiploc ), codec.Bytes() ); err != nil {
            return err
          }
        }
      }
    }

    // Sort the slice by NLC, hopefully making the more accurate entry first
    if len( s ) > 1 {
      sort.SliceStable( s, func( i, j int ) bool {
        return s[i].NLC < s[j].NLC
      })
    }

  }

  // Now persist
  for k, v := range stanox {
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

    if err := c.stanox.Put( []byte( strconv.FormatInt( int64( k ), 10 ) ), codec.Bytes() ); err != nil {
      return err
    }
  }

  return nil
}

func (c *CIF) GetStanox( tx *bolt.Tx, stanox int ) ( []*Tiploc, bool ) {

  b := tx.Bucket( []byte("Stanox") ).Get( []byte( strconv.FormatInt( int64( stanox ), 10 ) ) )
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