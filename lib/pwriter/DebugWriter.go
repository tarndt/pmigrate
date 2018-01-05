package pwriter

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"

	"lib"
	"lib/errs"
)

//Ensure DebugConsumer implements StateConsumer
var _ lib.StateConsumer = new(DebugConsumer)

type DebugConsumer struct {
	debugInfo string
}

func NewDebugConsumer() *DebugConsumer {
	return new(DebugConsumer)
}

func (this *DebugConsumer) Consume(provider lib.StateProvider) error {
	regs, err := provider.GetRegisters()
	if err != nil {
		return errs.Append(err, "Could not get registers")
	}
	spans, err := provider.GetMemoryMeta()
	if err != nil {
		return errs.Append(err, "Could not get memory metadata")
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "Process Name: %s\nProcess Identifer (PID): %d\n", provider.GetName(), provider.GetPID())

	buf.WriteString("\nRegisters:\n")
	for i, reg := range structToMap(regs) {
		fmt.Fprintf(&buf, "%d %s: 0x%X", i, reg.Name, reg.Value)
		if (i+1)%4 == 0 {
			buf.WriteByte('\n')
		} else {
			buf.WriteByte('\t')
		}
	}
	buf.Truncate(buf.Len() - 1)

	buf.WriteString("\n\nMemory Map Metadata & Content MD5 hashes:\n")
	hash := md5.New()
	for i, spanMeta := range spans {
		fmt.Fprintf(&buf, " %d Meta: %s\n", i, spanMeta)
		span, err := provider.GetMemorySpan(spanMeta)
		if err != nil {
			return errs.Append(err, "Could not get memory span")
		}
		hash.Reset()
		if _, err = io.Copy(hash, span.ReadCloser); err != nil {
			return errs.Append(err, "Could not read memory span data")
		}
		span.Close()
		fmt.Fprintf(&buf, " %d Hash: %s\n", i, strings.ToUpper(hex.EncodeToString(hash.Sum(nil))))
	}

	buf.WriteString("\nOpen File Handles:\n")
	for _, fd := range provider.GetFiles() {
		fmt.Fprintf(&buf, "%s\n", fd.String())
	}

	this.debugInfo = buf.String()
	return nil
}

func (this *DebugConsumer) Close() error {
	return nil
}

func (this *DebugConsumer) DebugInfo() string {
	return this.debugInfo
}

type regDesc struct {
	Name  string
	Value interface{}
}

type regDescs []regDesc

func (this regDescs) Len() int           { return len(this) }
func (this regDescs) Swap(i, j int)      { this[i], this[j] = this[j], this[i] }
func (this regDescs) Less(i, j int) bool { return this[i].Name < this[j].Name }

func structToMap(structure interface{}) []regDesc {
	structVal := reflect.ValueOf(structure)
	if structVal.Kind() == reflect.Ptr {
		structVal = structVal.Elem()
	}
	if structVal.Kind() != reflect.Struct { //we only handle structures
		return nil
	}

	fieldCount := structVal.NumField()
	result := make(regDescs, 0, fieldCount)
	for i := 0; i < fieldCount; i++ {
		result = append(result, regDesc{structVal.Type().Field(i).Name, structVal.Field(i).Interface()})
	}
	sort.Sort(result)

	return result
}
