package jsonutil

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"time"

	"cloud.google.com/go/civil"
)

// DateMarshalIntUnix is a custom marshaler for civil.Date, marshaling them as integers representing unix time.
func DateMarshalIntUnix(enc *jsontext.Encoder, d civil.Date) error {
	if d.IsZero() {
		return enc.WriteToken(jsontext.Int(0))
	}

	return enc.WriteToken(jsontext.Int(int64(
		time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, time.UTC).Unix(),
	)))
}

// DateUnmarshalIntUnix is a custom unmarshaler for civil.Date, unmarshaling them from integers and assuming they represent unix time.
func DateUnmarshalIntUnix(dec *jsontext.Decoder, d *civil.Date) error {
	var seconds int64
	if err := json.UnmarshalDecode(dec, &seconds); err != nil {
		return err
	}

	if seconds == 0 {
		*d = civil.Date{}
	} else {
		*d = civil.DateOf(time.Unix(seconds, 0).UTC())
	}

	return nil
}

// DateIntUnix is a civil.Date written and read as an integer of unix time, as [DateMarshalIntUnix] and
// [DateUnmarshalIntUnix] do. It is for one field among others written as strings: a pointer to the field converts to a
// *DateIntUnix, where a marshaler for all of civil.Date would be too many.
type DateIntUnix civil.Date

// IsZero reports whether d is the zero date, for omitzero.
func (d DateIntUnix) IsZero() bool { return civil.Date(d).IsZero() }

// MarshalJSONTo implements [json.MarshalerTo].
func (d DateIntUnix) MarshalJSONTo(enc *jsontext.Encoder) error {
	return DateMarshalIntUnix(enc, civil.Date(d))
}

// UnmarshalJSONFrom implements [json.UnmarshalerFrom].
func (d *DateIntUnix) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	return DateUnmarshalIntUnix(dec, (*civil.Date)(d))
}
