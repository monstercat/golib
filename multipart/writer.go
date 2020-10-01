package multipart

import (
	"fmt"
	"io"
	"mime/multipart"
	"reflect"

	structTag "github.com/monstercat/golib/struct-tag"
)

const StructTag = "multipart"

var (
	reader = reflect.TypeOf((*io.Reader)(nil)).Elem()
)

// Writer wraps multipart writer and adds some reflection logic
type Writer struct {
	*multipart.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		Writer: multipart.NewWriter(w),
	}
}

func (w *Writer) Marshal(values interface{}) (err error) {
	var errs Errors
	structTag.IterateStructFields(values, func(f reflect.StructField, v reflect.Value) {
		t := &tags{}
		t.Parse(f.Tag.Get(StructTag))

		if t.Ignore {
			return
		}

		// For now, we aren't dealing with arrays or structs.
		if f.Type.Kind() == reflect.Struct {
			return
		}
		if f.Type.Kind() == reflect.Array {
			return
		}

		if f.Type.Implements(reader) {
			r := v.Interface().(io.Reader)
			fw, err := w.CreateFormFile(t.FieldName, "")
			if err != nil {
				errs.AddError(err)
				return
			}
			if _, err := io.Copy(fw, r); err != nil {
				errs.AddError(err)
				return
			}
			return
		}

		fw, err := w.CreateFormField(t.FieldName)
		if err != nil {
			errs.AddError(err)
			return
		}

		if _, err := fw.Write([]byte(fmt.Sprintf("%v", v.Interface()))); err != nil {
			errs.AddError(err)
		}
	})
	if len(errs) == 0 {
		return nil
	}
	return errs
}
