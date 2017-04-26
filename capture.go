package capture

import (
	"bytes"
	"errors"
	"io"
	"os"
	"regexp"
)

/*
Temporarily redirects writes from *os.File to a string.

Swaps value of the variable holding *os.File to a os.Pipe then reads
any thing written to the pipe produced by fn.  Once fn completes,
another swap reverts the variable's value to its original value.

Note

Code encapsulated within fn must directly reference or execute
an assignment involving the variable holding the *os.File, otherwise,
the encapsulate code won't notice (be affected by) the redirection.

Inspiration

http://stackoverflow.com/questions/10473800/in-go-how-do-i-capture-stdout-of-a-function-into-a-string

Danger!

This routine implements a race condition when the
scope of the variable holding the pointer to the
os.File structure is declared outside the function that
calls this function or the calling function spawns goroutines
that reference the variable.

Unfortunately, I'm unable to devise a means, other than this one,
to capture writes to the standard (STDOUT/STDERR) devices without
changing the code that's targeted by the test.
*/
func It(osf **os.File, fn func()) string {
	std := *osf
	defer func() {
		*osf = std
	}()
	rdr, wrt, _ := os.Pipe()
	defer rdr.Close()
	*osf = wrt
	capture := make(chan string)
	defer close(capture)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, rdr)
		capture <- buf.String()
	}()
	fn()
	wrt.Close()
	return <-capture
}
func Match(osf **os.File, fn func(), regex string) (err error) {
	msg := It(osf, fn)
	var ok bool
	ok, err = regexp.MatchString(regex, msg)
	if err != nil {
		return
	}
	if !ok {
		err = errors.New("String: '" + msg + "' fails to match regexp: '" + regex + "'.")
	}
	return
}
