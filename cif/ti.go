package cif

import (
  "github.com/peter-mount/golib/codec"
)

func (c *CIF) parseTI( l string ) error {
  var t Tiploc = Tiploc{}
  i := 2
  i = parseStringTrim( l, i, 7, &t.Tiploc )
  i += 2
  i = parseInt( l, i, 6, &t.NLC )
  i = parseString( l, i, 1, &t.NLCCheck )
  i = parseStringTitle( l, i, 26, &t.Desc )
  i = parseInt( l, i, 5, &t.Stanox )
  i += 4
  i = parseStringTrim( l, i, 3, &t.CRS )
  i = parseStringTitle( l, i, 16, &t.NLCDesc )

  return c.putTiploc( &t )
}

// Store/replace a tiploc only if the entry is newer than an existing one
func (c *CIF) putTiploc( t *Tiploc ) error {

  // Link it to this CIF file
  t.DateOfExtract = c.importhd.DateOfExtract

  // Retrieve the existing entry (if any)
  b := c.tiploc.Get( []byte( t.Tiploc ) )

  var e Tiploc
  if( b != nil ) {
    codec.NewBinaryCodecFrom( b ).Read( &e )
  }

  // If we don't have an entry or this one is newer then persist
  if t.Tiploc != e.Tiploc || t.DateOfExtract.After( e.DateOfExtract ) {
    //return c.put( c.tiploc, t.Tiploc, &t )
    codec := codec.NewBinaryCodec()
    codec.Write( t )
    if codec.Error() != nil {
      return codec.Error()
    }

    return c.tiploc.Put( []byte( t.Tiploc ), codec.Bytes() )
  }

  return nil
}
