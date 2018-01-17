package cif

import (
  bolt "github.com/coreos/bbolt"
  "bytes"
  "encoding/json"
//  "encoding/xml"
  "github.com/peter-mount/golib/rest"
  "sort"
  "strings"
)

type TiplocMap struct {
  m map[string]*Tiploc
  s bool
}

func (r *Response) tiplocMap() *TiplocMap {
  if r.Tiploc == nil {
    r.Tiploc = &TiplocMap{ m: make( map[string]*Tiploc ) }
  }
  return r.Tiploc
}

func (r *Response) sortTiplocs() {
  r.tiplocMap().s = true
}

// AddTiploc adds a Tiploc to the response
func (r *Response) AddTiploc( t *Tiploc ) {
  tm := r.tiplocMap()
  if _, ok := tm.m[ t.Tiploc ]; !ok {
    tm.m[ t.Tiploc ] = t
  }
}

// AddTiplocs adds an array of Tiploc's to the response
func (r *Response) AddTiplocs( t []*Tiploc ) {
  for _, e := range t {
    r.AddTiploc( e )
  }
}

func (r *Response) GetTiploc( n string ) ( *Tiploc, bool ) {
  t, e := r.tiplocMap().m[ n ]
  return t, e
}

// ResolveScheduleTiplocs resolves the Tiploc's in a Schedule
func (c *CIF) ResolveScheduleTiplocs( tx *bolt.Tx, s *Schedule, r *Response ) {
  for _, l := range s.Locations {
    if _, ok := r.GetTiploc( l.Tiploc ); !ok {
      if t, exists := c.GetTiploc( tx, l.Tiploc ); exists {
        r.AddTiploc( t )
      }
    }
  }
}

// SetSelf sets the Self field to match this request
func (r *Response) TiplocsSetSelf( rs *rest.Rest ) {
  t := r.tiplocMap()
  for _, v := range t.m {
    v.SetSelf( rs )
  }
}

func (t *TiplocMap) MarshalJSON() ( []byte, error ) {
  // Tiploc sorted by NLC
  var vals []*Tiploc
  for _, v := range t.m {
    vals = append( vals, v )
  }

  if t.s {
    // Tiploc sorted by NLC
    sort.SliceStable( vals, func( i, j int ) bool {
      return vals[i].NLC < vals[j].NLC
    })
  } else {
    // Default sort by name
    sort.SliceStable( vals, func( i, j int ) bool {
      return strings.Compare( vals[i].Tiploc, vals[j].Tiploc ) < 0
    })
  }

  b := &bytes.Buffer{}
  b.WriteByte( '{' )

  for i, v := range vals {
    if i > 0 {
      b.WriteByte( ',' )
    }
    b.WriteByte( '"' )
    b.WriteString( v.Tiploc )
    b.WriteByte( '"' )
    b.WriteByte( ':' )

    if eb, err := json.Marshal( v ); err != nil {
      return nil, err
    } else {
      b.Write( eb )
    }
  }

  b.WriteByte( '}' )
  return b.Bytes(), nil
}
