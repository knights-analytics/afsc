package s3

import "sync"

//ByteWriterAt  represents a bytes writer at
type Writer struct {
	mutex    *sync.Mutex
	Buffer   []byte
	position int
}

//WriteAt returns number of written bytes or error
func (w *Writer) WriteAt(p []byte, offset int64) (n int, err error) {
	w.mutex.Lock()
	if int(offset) == w.position {
		w.Buffer = append(w.Buffer, p...)
		w.position += len(p)
		w.mutex.Unlock()
		return len(p), nil
	} else if w.position < int(offset) {
		var diff = (int(offset) - w.position)
		var fillingBytes = make([]byte, diff)
		w.position += len(fillingBytes)
		w.Buffer = append(w.Buffer, fillingBytes...)
		w.mutex.Unlock()
		return w.WriteAt(p, offset)
	} else {
		for i := 0; i < len(p); i++ {
			var index = int(offset) + i
			if index < len(w.Buffer) {
				w.Buffer[int(offset)+i] = p[i]
			} else {
				w.Buffer = append(w.Buffer, p[i:]...)
				break
			}
		}
		w.mutex.Unlock()
		return len(p), nil
	}
}

//Writer returns a writer
func NewWriter() *Writer {
	return &Writer{
		mutex:  &sync.Mutex{},
		Buffer: make([]byte, 0),
	}
}
